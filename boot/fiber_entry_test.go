// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiber

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/interceptor/log/zap"
	"github.com/rookie-ninja/rk-fiber/interceptor/metrics/prom"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	defaultBootConfigStr = `
---
fiber:
 - name: greeter
   port: 8080
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
`
)

func TestWithZapLoggerEntryFiber_HappyCase(t *testing.T) {
	loggerEntry := rkentry.NoopZapLoggerEntry()
	entry := RegisterFiberEntry()

	option := WithZapLoggerEntryFiber(loggerEntry)
	option(entry)

	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestWithEventLoggerEntryFiber_HappyCase(t *testing.T) {
	entry := RegisterFiberEntry()

	eventLoggerEntry := rkentry.NoopEventLoggerEntry()

	option := WithEventLoggerEntryFiber(eventLoggerEntry)
	option(entry)

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
}

func TestWithInterceptorsFiber_WithNilInterceptorList(t *testing.T) {
	entry := RegisterFiberEntry()

	option := WithInterceptorsFiber(nil)
	option(entry)

	assert.NotNil(t, entry.Interceptors)
}

func TestWithInterceptorsFiber_HappyCase(t *testing.T) {
	entry := RegisterFiberEntry()

	loggingInterceptor := rkfiberlog.Interceptor()
	metricsInterceptor := rkfibermetrics.Interceptor()

	interceptors := []fiber.Handler{
		loggingInterceptor,
		metricsInterceptor,
	}

	option := WithInterceptorsFiber(interceptors...)
	option(entry)

	assert.NotNil(t, entry.Interceptors)
	// should contains logging, metrics and panic interceptor
	// where panic interceptor is inject by default
	assert.Len(t, entry.Interceptors, 3)
}

func TestWithCommonServiceEntryFiber_WithEntry(t *testing.T) {
	entry := RegisterFiberEntry()

	option := WithCommonServiceEntryFiber(NewCommonServiceEntry())
	option(entry)

	assert.NotNil(t, entry.CommonServiceEntry)
}

func TestWithCommonServiceEntryFiber_WithoutEntry(t *testing.T) {
	entry := RegisterFiberEntry()

	assert.Nil(t, entry.CommonServiceEntry)
}

func TestWithTVEntryFiber_WithEntry(t *testing.T) {
	entry := RegisterFiberEntry()

	option := WithTVEntryFiber(NewTvEntry())
	option(entry)

	assert.NotNil(t, entry.TvEntry)
}

func TestWithTVEntry_WithoutEntry(t *testing.T) {
	entry := RegisterFiberEntry()

	assert.Nil(t, entry.TvEntry)
}

func TestWithCertEntryFiber_HappyCase(t *testing.T) {
	entry := RegisterFiberEntry()
	certEntry := &rkentry.CertEntry{}

	option := WithCertEntryFiber(certEntry)
	option(entry)

	assert.Equal(t, entry.CertEntry, certEntry)
}

func TestWithSWEntryFiber_HappyCase(t *testing.T) {
	entry := RegisterFiberEntry()
	sw := NewSwEntry()

	option := WithSwEntryFiber(sw)
	option(entry)

	assert.Equal(t, entry.SwEntry, sw)
}

func TestWithPortFiber_HappyCase(t *testing.T) {
	entry := RegisterFiberEntry()
	port := uint64(1111)

	option := WithPortFiber(port)
	option(entry)

	assert.Equal(t, entry.Port, port)
}

func TestWithNameFiber_HappyCase(t *testing.T) {
	entry := RegisterFiberEntry()
	name := "unit-test-entry"

	option := WithNameFiber(name)
	option(entry)

	assert.Equal(t, entry.EntryName, name)
}

func TestRegisterFiberEntriesWithConfig_WithInvalidConfigFilePath(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, true)
		} else {
			// this should never be called in case of a bug
			assert.True(t, false)
		}
	}()

	RegisterFiberEntriesWithConfig("/invalid-path")
}

func TestRegisterFiberEntriesWithConfig_WithNilFactory(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterFiberEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)
}

func TestRegisterFiberEntriesWithConfig_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	// write config file in unit test temp directory
	tempDir := path.Join(t.TempDir(), "boot.yaml")
	assert.Nil(t, ioutil.WriteFile(tempDir, []byte(defaultBootConfigStr), os.ModePerm))
	entries := RegisterFiberEntriesWithConfig(tempDir)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 2)

	// validate entry element based on boot.yaml config defined in defaultBootConfigStr
	greeter := entries["greeter"].(*FiberEntry)
	assert.NotNil(t, greeter)
	assert.Equal(t, uint64(8080), greeter.Port)
	assert.NotNil(t, greeter.SwEntry)
	assert.NotNil(t, greeter.CommonServiceEntry)
	assert.NotNil(t, greeter.TvEntry)
	// logging, metrics, auth and panic interceptor should be included
	assert.True(t, len(greeter.Interceptors) > 0)

	greeter2 := entries["greeter2"].(*FiberEntry)
	assert.NotNil(t, greeter2)
	assert.Equal(t, uint64(2008), greeter2.Port)
	assert.NotNil(t, greeter2.SwEntry)
	assert.NotNil(t, greeter2.CommonServiceEntry)
	assert.NotNil(t, greeter2.TvEntry)
	// logging, metrics, auth and panic interceptor should be included
	assert.Len(t, greeter2.Interceptors, 4)
}

func TestRegisterFiberEntry_WithZapLoggerEntry(t *testing.T) {
	loggerEntry := rkentry.NoopZapLoggerEntry()
	entry := RegisterFiberEntry(WithZapLoggerEntryFiber(loggerEntry))
	assert.Equal(t, loggerEntry, entry.ZapLoggerEntry)
}

func TestRegisterFiberEntry_WithEventLoggerEntry(t *testing.T) {
	loggerEntry := rkentry.NoopEventLoggerEntry()

	entry := RegisterFiberEntry(WithEventLoggerEntryFiber(loggerEntry))
	assert.Equal(t, loggerEntry, entry.EventLoggerEntry)
}

func TestNewFiberEntry_WithInterceptors(t *testing.T) {
	loggingInterceptor := rkfiberlog.Interceptor()
	entry := RegisterFiberEntry(WithInterceptorsFiber(loggingInterceptor))
	assert.Len(t, entry.Interceptors, 2)
}

func TestNewFiberEntry_WithCommonServiceEntry(t *testing.T) {
	entry := RegisterFiberEntry(WithCommonServiceEntryFiber(NewCommonServiceEntry()))
	assert.NotNil(t, entry.CommonServiceEntry)
}

func TestNewFiberEntry_WithTVEntry(t *testing.T) {
	entry := RegisterFiberEntry(WithTVEntryFiber(NewTvEntry()))
	assert.NotNil(t, entry.TvEntry)
}

func TestNewFiberEntry_WithCertStore(t *testing.T) {
	certEntry := &rkentry.CertEntry{}

	entry := RegisterFiberEntry(WithCertEntryFiber(certEntry))
	assert.Equal(t, certEntry, entry.CertEntry)
}

func TestNewFiberEntry_WithSWEntry(t *testing.T) {
	sw := NewSwEntry()
	entry := RegisterFiberEntry(WithSwEntryFiber(sw))
	assert.Equal(t, sw, entry.SwEntry)
}

func TestNewFiberEntry_WithPort(t *testing.T) {
	entry := RegisterFiberEntry(WithPortFiber(8080))
	assert.Equal(t, uint64(8080), entry.Port)
}

func TestNewFiberEntry_WithName(t *testing.T) {
	entry := RegisterFiberEntry(WithNameFiber("unit-test-greeter"))
	assert.Equal(t, "unit-test-greeter", entry.GetName())
}

func TestNewFiberEntry_WithDefaultValue(t *testing.T) {
	entry := RegisterFiberEntry()
	assert.True(t, strings.HasPrefix(entry.GetName(), "FiberServer-"))
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.Len(t, entry.Interceptors, 1)
	assert.Nil(t, entry.App)
	assert.Nil(t, entry.SwEntry)
	assert.Nil(t, entry.CertEntry)
	assert.False(t, entry.IsSwEnabled())
	assert.False(t, entry.IsTlsEnabled())
	assert.Nil(t, entry.CommonServiceEntry)
	assert.Nil(t, entry.TvEntry)
	assert.Equal(t, "FiberEntry", entry.GetType())
}

func TestFiberEntry_GetName_HappyCase(t *testing.T) {
	entry := RegisterFiberEntry(WithNameFiber("unit-test-entry"))
	assert.Equal(t, "unit-test-entry", entry.GetName())
}

func TestFiberEntry_GetType_HappyCase(t *testing.T) {
	assert.Equal(t, "FiberEntry", RegisterFiberEntry().GetType())
}

func TestFiberEntry_String_HappyCase(t *testing.T) {
	assert.NotEmpty(t, RegisterFiberEntry().String())
}

func TestFiberEntry_IsSwEnabled_ExpectTrue(t *testing.T) {
	sw := NewSwEntry()
	entry := RegisterFiberEntry(WithSwEntryFiber(sw))
	assert.True(t, entry.IsSwEnabled())
}

func TestFiberEntry_IsSwEnabled_ExpectFalse(t *testing.T) {
	entry := RegisterFiberEntry()
	assert.False(t, entry.IsSwEnabled())
}

func TestFiberEntry_IsTlsEnabled_ExpectTrue(t *testing.T) {
	certEntry := &rkentry.CertEntry{
		Store: &rkentry.CertStore{},
	}

	entry := RegisterFiberEntry(WithCertEntryFiber(certEntry))
	assert.True(t, entry.IsTlsEnabled())
}

func TestFiberEntry_IsTlsEnabled_ExpectFalse(t *testing.T) {
	entry := RegisterFiberEntry()
	assert.False(t, entry.IsTlsEnabled())
}

func TestFiberEntry_GetFiber_HappyCase(t *testing.T) {
	entry := RegisterFiberEntry()
	assert.Nil(t, entry.App)
}

func TestFiberEntry_Bootstrap_WithSwagger(t *testing.T) {
	sw := NewSwEntry(
		WithPathSw("sw"),
		WithZapLoggerEntrySw(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntrySw(rkentry.NoopEventLoggerEntry()))
	entry := RegisterFiberEntry(
		WithNameFiber("unit-test-entry"),
		WithPortFiber(8080),
		WithZapLoggerEntryFiber(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryFiber(rkentry.NoopEventLoggerEntry()),
		WithSwEntryFiber(sw))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)
	assert.True(t, len(entry.ListRoutes()) >= 3)

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestFiberEntry_Bootstrap_WithoutSwagger(t *testing.T) {
	entry := RegisterFiberEntry(
		WithNameFiber("unit-test-entry"),
		WithPortFiber(8080),
		WithZapLoggerEntryFiber(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryFiber(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)
	assert.Empty(t, entry.ListRoutes())

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestFiberEntry_Bootstrap_WithoutTLS(t *testing.T) {
	entry := RegisterFiberEntry(
		WithNameFiber("unit-test-entry"),
		WithPortFiber(8080),
		WithZapLoggerEntryFiber(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryFiber(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestFiberEntry_Shutdown_WithBootstrap(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterFiberEntry(
		WithNameFiber("unit-test-entry"),
		WithPortFiber(8080),
		WithZapLoggerEntryFiber(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryFiber(rkentry.NoopEventLoggerEntry()))

	go entry.Bootstrap(context.Background())
	time.Sleep(time.Second)
	// endpoint should be accessible with 8080 port
	validateServerIsUp(t, entry.Port)

	entry.Interrupt(context.Background())
	time.Sleep(time.Second)
}

func TestFiberEntry_Shutdown_WithoutBootstrap(t *testing.T) {
	defer assertNotPanic(t)

	entry := RegisterFiberEntry(
		WithNameFiber("unit-test-entry"),
		WithPortFiber(8080),
		WithZapLoggerEntryFiber(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryFiber(rkentry.NoopEventLoggerEntry()))

	entry.Interrupt(context.Background())
}

func validateServerIsUp(t *testing.T, port uint64) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("0.0.0.0", strconv.FormatUint(port, 10)), time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, conn)
	if conn != nil {
		assert.Nil(t, conn.Close())
	}
}
