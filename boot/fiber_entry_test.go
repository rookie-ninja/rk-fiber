// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiber

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	rkentry "github.com/rookie-ninja/rk-entry/entry"
	rkmidmetrics "github.com/rookie-ninja/rk-entry/middleware/metrics"
	rkfibermeta "github.com/rookie-ninja/rk-fiber/interceptor/meta"
	rkfibermetrics "github.com/rookie-ninja/rk-fiber/interceptor/metrics/prom"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

const (
	defaultBootConfigStr = `
---
fiber:
 - name: greeter
   port: 1949
   enabled: true
   sw:
     enabled: true
     path: "sw"
   commonService:
     enabled: true
   tv:
     enabled: true
   prom:
     enabled: true
     pusher:
       enabled: false
   interceptors:
     loggingZap:
       enabled: true
     metricsProm:
       enabled: true
     auth:
       enabled: true
       basic:
         - "user:pass"
     meta:
       enabled: true
     tracingTelemetry:
       enabled: true
     ratelimit:
       enabled: true
     timeout:
       enabled: true
     cors:
       enabled: true
     jwt:
       enabled: true
     secure:
       enabled: true
     csrf:
       enabled: true
     gzip:
       enabled: true
 - name: greeter2
   port: 2008
   enabled: true
   sw:
     enabled: true
     path: "sw"
   commonService:
     enabled: true
   tv:
     enabled: true
   interceptors:
     loggingZap:
       enabled: true
     metricsProm:
       enabled: true
     auth:
       enabled: true
       basic:
         - "user:pass"
 - name: greeter3
   port: 2022
   enabled: false
`
)

func TestGetFiberEntry(t *testing.T) {
	// expect nil
	assert.Nil(t, GetFiberEntry("entry-name"))

	// happy case
	echoEntry := RegisterFiberEntry(WithName("ut"))
	assert.Equal(t, echoEntry, GetFiberEntry("ut"))

	rkentry.GlobalAppCtx.RemoveEntry("ut")
}

func TestRegisterFiberEntry(t *testing.T) {
	// without options
	entry := RegisterFiberEntry()
	assert.NotNil(t, entry)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	rkentry.GlobalAppCtx.RemoveEntry(entry.GetName())

	// with options
	entry = RegisterFiberEntry(
		WithZapLoggerEntry(nil),
		WithEventLoggerEntry(nil),
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithTvEntry(rkentry.RegisterTvEntry()),
		WithStaticFileHandlerEntry(rkentry.RegisterStaticFileHandlerEntry()),
		WithCertEntry(rkentry.RegisterCertEntry()),
		WithSwEntry(rkentry.RegisterSwEntry()),
		WithPort(8080),
		WithName("ut-entry"),
		WithDescription("ut-desc"),
		WithPromEntry(rkentry.RegisterPromEntry()))

	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.True(t, entry.IsSwEnabled())
	assert.True(t, entry.IsStaticFileHandlerEnabled())
	assert.True(t, entry.IsPromEnabled())
	assert.True(t, entry.IsCommonServiceEnabled())
	assert.True(t, entry.IsTvEnabled())
	assert.True(t, entry.IsTlsEnabled())

	bytes, err := entry.MarshalJSON()
	assert.NotEmpty(t, bytes)
	assert.Nil(t, err)
	assert.Nil(t, entry.UnmarshalJSON([]byte{}))
}

func TestFiberEntry_AddInterceptor(t *testing.T) {
	defer assertNotPanic(t)
	entry := RegisterFiberEntry()
	inter := rkfibermeta.Interceptor()
	entry.AddInterceptor(inter)
}

func TestFiberEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	// without enable sw, static, prom, common, tv, tls
	entry := RegisterFiberEntry(WithPort(8080))
	entry.Bootstrap(context.TODO())
	validateServerIsUp(t, 8080, entry.IsTlsEnabled())
	assert.Empty(t, entry.ListRoutes())

	entry.Interrupt(context.TODO())
	time.Sleep(time.Second)

	// with enable sw, static, prom, common, tv, tls
	certEntry := rkentry.RegisterCertEntry()
	certEntry.Store.ServerCert, certEntry.Store.ServerKey = generateCerts()

	entry = RegisterFiberEntry(
		WithPort(8081),
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithTvEntry(rkentry.RegisterTvEntry()),
		WithStaticFileHandlerEntry(rkentry.RegisterStaticFileHandlerEntry()),
		WithCertEntry(certEntry),
		WithSwEntry(rkentry.RegisterSwEntry()),
		WithPromEntry(rkentry.RegisterPromEntry()))
	entry.Bootstrap(context.TODO())
	validateServerIsUp(t, 8081, entry.IsTlsEnabled())
	assert.NotEmpty(t, entry.ListRoutes())

	entry.Interrupt(context.TODO())
	time.Sleep(time.Second)
}

func TestFiberEntry_startServer_ServerFail(t *testing.T) {
	// let's give an invalid port
	entry := RegisterFiberEntry(
		WithPort(808080))

	event := rkentry.NoopEventLoggerEntry().GetEventFactory().CreateEventNoop()
	logger := rkentry.NoopZapLoggerEntry().GetLogger()

	entry.startServer(event, logger)
}

func TestRegisterFiberEntriesWithConfig(t *testing.T) {
	assertNotPanic(t)

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterFiberEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)

	// validate entry element based on boot.yaml config defined in defaultBootConfigStr
	greeter := entries["greeter"].(*FiberEntry)
	assert.NotNil(t, greeter)

	greeter2 := entries["greeter2"].(*FiberEntry)
	assert.NotNil(t, greeter2)

	greeter3 := entries["greeter3"]
	assert.Nil(t, greeter3)
}

func TestFiberEntry_Apis(t *testing.T) {
	entry := RegisterFiberEntry()

	app := fiber.New()
	app.Get("/apis", entry.Apis)
	entry.App = app

	req := httptest.NewRequest(http.MethodGet, "/apis", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestFiberEntry_Req_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterFiberEntry(
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithPort(8082),
		WithName("ut"))

	entry.AddInterceptor(rkfibermetrics.Interceptor(
		rkmidmetrics.WithEntryNameAndType("ut", "Fiber"),
		rkmidmetrics.WithRegisterer(prometheus.NewRegistry())))

	app := fiber.New()
	app.Get("/req", entry.Req)
	entry.App = app

	entry.Bootstrap(context.TODO())

	req := httptest.NewRequest(http.MethodGet, "/req", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	entry.Interrupt(context.TODO())
	time.Sleep(time.Second)
}

func TestFiberEntry_TV(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterFiberEntry(
		WithCommonServiceEntry(rkentry.RegisterCommonServiceEntry()),
		WithTvEntry(rkentry.RegisterTvEntry()),
		WithPort(8083),
		WithName("ut"))

	entry.AddInterceptor(rkfibermetrics.Interceptor(
		rkmidmetrics.WithEntryNameAndType("ut", "Echo")))

	app := fiber.New()
	app.Get("/ut/*", entry.TV)
	entry.App = app

	entry.Bootstrap(context.TODO())

	req := httptest.NewRequest(http.MethodGet, "/ut/apis", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// for default
	req = httptest.NewRequest(http.MethodGet, "/ut/other", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	entry.Interrupt(context.TODO())
	time.Sleep(time.Second)
}

func generateCerts() ([]byte, []byte) {
	// Create certs and return as []byte
	ca := &x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{"Fake cert."},
		},
		SerialNumber:          big.NewInt(42),
		NotAfter:              time.Now().Add(2 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Create a Private Key
	key, _ := rsa.GenerateKey(rand.Reader, 4096)

	// Use CA Cert to sign a CSR and create a Public Cert
	csr := &key.PublicKey
	cert, _ := x509.CreateCertificate(rand.Reader, ca, ca, csr, key)

	// Convert keys into pem.Block
	c := &pem.Block{Type: "CERTIFICATE", Bytes: cert}
	k := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}

	return pem.EncodeToMemory(c), pem.EncodeToMemory(k)
}

func validateServerIsUp(t *testing.T, port uint64, isTls bool) {
	// sleep for 2 seconds waiting server startup
	time.Sleep(2 * time.Second)

	if !isTls {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), time.Second)
		assert.Nil(t, err)
		assert.NotNil(t, conn)
		if conn != nil {
			assert.Nil(t, conn.Close())
		}
		return
	}

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}

	tlsConn, err := tls.Dial("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), tlsConf)
	assert.Nil(t, err)
	assert.NotNil(t, tlsConn)
	if tlsConn != nil {
		assert.Nil(t, tlsConn.Close())
	}
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
