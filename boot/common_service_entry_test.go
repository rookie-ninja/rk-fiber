// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiber

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func newCtx() (*fiber.App, *fiber.Ctx) {
	app := fiber.New()
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{
		Request:  *fasthttp.AcquireRequest(),
		Response: *fasthttp.AcquireResponse(),
	})

	return app, ctx
}

func TestWithNameCommonService_WithEmptyString(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithNameCommonService(""))

	assert.NotEmpty(t, entry.GetName())
}

func TestWithNameCommonService_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithNameCommonService("unit-test"))

	assert.Equal(t, "unit-test", entry.GetName())
}

func TestWithEventLoggerEntryCommonService_WithNilParam(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithEventLoggerEntryCommonService(nil))

	assert.NotNil(t, entry.EventLoggerEntry)
}

func TestWithEventLoggerEntryCommonService_HappyCase(t *testing.T) {
	eventLoggerEntry := rkentry.NoopEventLoggerEntry()
	entry := NewCommonServiceEntry(
		WithEventLoggerEntryCommonService(eventLoggerEntry))

	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
}

func TestWithZapLoggerEntryCommonService_WithNilParam(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(nil))

	assert.NotNil(t, entry.ZapLoggerEntry)
}

func TestWithZapLoggerEntryCommonService_HappyCase(t *testing.T) {
	zapLoggerEntry := rkentry.NoopZapLoggerEntry()
	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(zapLoggerEntry))

	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
}

func TestNewCommonServiceEntry_WithoutOptions(t *testing.T) {
	entry := NewCommonServiceEntry()

	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
}

func TestNewCommonServiceEntry_HappyCase(t *testing.T) {
	zapLoggerEntry := rkentry.NoopZapLoggerEntry()
	eventLoggerEntry := rkentry.NoopEventLoggerEntry()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(zapLoggerEntry),
		WithEventLoggerEntryCommonService(eventLoggerEntry),
		WithNameCommonService("ut"))

	assert.Equal(t, zapLoggerEntry, entry.ZapLoggerEntry)
	assert.Equal(t, eventLoggerEntry, entry.EventLoggerEntry)
	assert.Equal(t, "ut", entry.GetName())
	assert.NotEmpty(t, entry.GetType())
}

func TestCommonServiceEntry_Bootstrap_WithoutRouter(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(rkentry.NoopEventLoggerEntry()))
	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestCommonServiceEntry_Bootstrap_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(rkentry.NoopEventLoggerEntry()))
	entry.Bootstrap(context.TODO())
	entry.Interrupt(context.TODO())
}

func TestCommonServiceEntry_Interrupt_HappyCase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry(
		WithZapLoggerEntryCommonService(rkentry.NoopZapLoggerEntry()),
		WithEventLoggerEntryCommonService(rkentry.NoopEventLoggerEntry()))
	entry.Interrupt(context.Background())
}

func TestCommonServiceEntry_GetName_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry(
		WithNameCommonService("ut"))

	assert.Equal(t, "ut", entry.GetName())
}

func TestCommonServiceEntry_GetType_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	assert.Equal(t, "CommonServiceEntry", entry.GetType())
}

func TestCommonServiceEntry_String_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	assert.NotEmpty(t, entry.String())
}

func TestCommonServiceEntry_Healthy_WithNilContext(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry := NewCommonServiceEntry()
	entry.Healthy(nil)
}

func TestCommonServiceEntry_Healthy_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	entry.Healthy(ctx)

	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.Equal(t, `{"healthy":true}`, strings.TrimSuffix(string(ctx.Response().Body()), "\n"))
}

func TestCommonServiceEntry_GC_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Gc(nil)
}

func TestCommonServiceEntry_GC_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Gc(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_Info_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Info(nil)
}

func TestCommonServiceEntry_Info_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Info(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_Config_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Configs(nil)
}

func TestCommonServiceEntry_Config_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	vp := viper.New()
	vp.Set("unit-test-key", "unit-test-value")

	viperEntry := rkentry.RegisterConfigEntry(
		rkentry.WithNameConfig("unit-test"),
		rkentry.WithViperInstanceConfig(vp))

	rkentry.GlobalAppCtx.AddConfigEntry(viperEntry)

	entry.Configs(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	resStr := string(ctx.Response().Body())

	assert.NotEmpty(t, resStr)
	assert.Contains(t, resStr, "unit-test-key")
	assert.Contains(t, resStr, "unit-test-value")
}

func TestCommonServiceEntry_APIs_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Apis(nil)
}

func TestCommonServiceEntry_APIs_WithEmptyEntries(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Apis(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_APIs_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, "unit-test-fiber"))

	entry.Apis(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_Sys_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Sys(nil)
}

func TestCommonServiceEntry_Sys_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	entry.Sys(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_Req_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, "unit-test-fiber"))

	entry.Req(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_Req_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Req(nil)
}

func TestCommonServiceEntry_Entries_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Entries(nil)
}

func TestCommonServiceEntry_Entries_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	entry.Entries(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_Certs_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Certs(nil)
}

func TestCommonServiceEntry_Certs_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	rkentry.RegisterCertEntry(rkentry.WithNameCert("ut-cert"))
	certEntry := rkentry.GlobalAppCtx.GetCertEntry("ut-cert")
	certEntry.Retriever = &rkentry.CredRetrieverLocalFs{}
	certEntry.Store = &rkentry.CertStore{}

	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, "unit-test-fiber"))

	entry.Certs(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_Logs_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Logs(nil)
}

func TestCommonServiceEntry_Logs_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	fiberEntry := RegisterFiberEntry(
		WithCommonServiceEntryFiber(entry),
		WithNameFiber("unit-test-fiber"))
	rkentry.GlobalAppCtx.AddEntry(fiberEntry)

	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, "unit-test-fiber"))

	entry.Logs(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_Deps_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Deps(nil)
}

func TestCommonServiceEntry_Deps_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	fiberEntry := RegisterFiberEntry(
		WithCommonServiceEntryFiber(entry),
		WithNameFiber("unit-test-fiber"))
	rkentry.GlobalAppCtx.AddEntry(fiberEntry)

	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, "unit-test-fiber"))

	entry.Deps(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_License_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.License(nil)
}

func TestCommonServiceEntry_License_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	fiberEntry := RegisterFiberEntry(
		WithCommonServiceEntryFiber(entry),
		WithNameFiber("unit-test-fiber"))
	rkentry.GlobalAppCtx.AddEntry(fiberEntry)

	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, "unit-test-fiber"))

	entry.License(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_Readme_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Readme(nil)
}

func TestCommonServiceEntry_Readme_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	fiberEntry := RegisterFiberEntry(
		WithCommonServiceEntryFiber(entry),
		WithNameFiber("unit-test-fiber"))
	rkentry.GlobalAppCtx.AddEntry(fiberEntry)

	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, "unit-test-fiber"))

	entry.Readme(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestCommonServiceEntry_Git_WithNilContext(t *testing.T) {
	entry := NewCommonServiceEntry()

	defer func() {
		if r := recover(); r != nil {
			// expect panic to be called with non nil error
			assert.True(t, false)
		} else {
			// this should never be called in case of a bug
			assert.True(t, true)
		}
	}()

	entry.Git(nil)
}

func TestCommonServiceEntry_Git_HappyCase(t *testing.T) {
	entry := NewCommonServiceEntry()

	app, ctx := newCtx()
	defer app.Shutdown()

	fiberEntry := RegisterFiberEntry(
		WithCommonServiceEntryFiber(entry),
		WithNameFiber("unit-test-fiber"))
	rkentry.GlobalAppCtx.AddEntry(fiberEntry)
	rkentry.GlobalAppCtx.SetRkMetaEntry(&rkentry.RkMetaEntry{
		RkMeta: &rkcommon.RkMeta{
			Git: &rkcommon.Git{
				Commit: &rkcommon.Commit{
					Committer: &rkcommon.Committer{},
				},
			},
		},
	})

	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, "unit-test-fiber"))

	entry.Git(ctx)
	assert.Equal(t, http.StatusOK, ctx.Response().StatusCode())
	assert.NotEmpty(t, string(ctx.Response().Body()))
}

func TestGetEntry_WithNilContext(t *testing.T) {
	assert.Nil(t, getEntry(nil))
}

func TestConstructSwUrl_WithNilEntry(t *testing.T) {
	app, ctx := newCtx()
	defer app.Shutdown()
	assert.Equal(t, "N/A", constructSwUrl(nil, ctx))
}

func TestConstructSwUrl_WithNilContext(t *testing.T) {
	path := "ut-sw"
	port := 1111
	sw := NewSwEntry(WithPathSw(path))
	entry := RegisterFiberEntry(WithSwEntryFiber(sw), WithPortFiber(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://localhost:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, nil))
}

func TestConstructSwUrl_WithNilRequest(t *testing.T) {
	path := "ut-sw"
	port := 1111

	app, ctx := newCtx()
	defer app.Shutdown()
	ctx.Request().SetHost(fmt.Sprintf("localhost:%d", port))

	sw := NewSwEntry(WithPathSw(path))
	entry := RegisterFiberEntry(WithSwEntryFiber(sw), WithPortFiber(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://localhost:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, ctx))
}

func TestConstructSwUrl_WithEmptyHost(t *testing.T) {
	path := "ut-sw"
	port := 1111

	app, ctx := newCtx()
	defer app.Shutdown()
	ctx.Request().SetHost(fmt.Sprintf("localhost:%d", port))

	sw := NewSwEntry(WithPathSw(path))
	entry := RegisterFiberEntry(WithSwEntryFiber(sw), WithPortFiber(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://localhost:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, ctx))
}

func TestConstructSwUrl_HappyCase(t *testing.T) {
	app, ctx := newCtx()
	defer app.Shutdown()
	ctx.Request().SetHost("8.8.8.8:1111")

	path := "ut-sw"
	port := 1111

	sw := NewSwEntry(WithPathSw(path), WithPortSw(uint64(port)))
	entry := RegisterFiberEntry(WithSwEntryFiber(sw), WithPortFiber(uint64(port)))

	assert.Equal(t, fmt.Sprintf("http://8.8.8.8:%s/%s/",
		strconv.Itoa(port), path), constructSwUrl(entry, ctx))
}

func TestContainsMetrics_ExpectFalse(t *testing.T) {
	api := "/rk/v1/non-exist"
	metrics := make([]*rkentry.ReqMetricsRK, 0)
	metrics = append(metrics, &rkentry.ReqMetricsRK{
		RestPath: "/rk/v1/exist",
	})

	assert.False(t, containsMetrics(api, metrics))
}

func TestContainsMetrics_ExpectTrue(t *testing.T) {
	api := "/rk/v1/exist"
	metrics := make([]*rkentry.ReqMetricsRK, 0)
	metrics = append(metrics, &rkentry.ReqMetricsRK{
		RestPath: api,
	})

	assert.True(t, containsMetrics(api, metrics))
}
