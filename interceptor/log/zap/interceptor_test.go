// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiberlog

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidlog "github.com/rookie-ninja/rk-entry/middleware/log"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var defaultMiddlewareFunc = func(context *fiber.Ctx) error {
	return nil
}

func TestInterceptor_WithShouldNotLog(t *testing.T) {
	defer assertNotPanic(t)

	app := fiber.New()

	handler := Interceptor(
		rkmidlog.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidlog.WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		rkmidlog.WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))

	app.Use(handler)
	app.Get("/rk/v1/assets", func(ctx *fiber.Ctx) error {
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/rk/v1/assets", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInterceptor_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	app := fiber.New()

	handler := Interceptor(
		rkmidlog.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidlog.WithZapLoggerEntry(rkentry.NoopZapLoggerEntry()),
		rkmidlog.WithEventLoggerEntry(rkentry.NoopEventLoggerEntry()))

	var eventForValidation rkquery.Event

	app.Use(handler)
	app.Use(func(ctx *fiber.Ctx) error {
		ctx.Next()
		eventForValidation = rkfiberctx.GetEvent(ctx)
		return nil
	})
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		ctx.Response().Header.Set(rkmid.HeaderRequestId, "ut-request-id")
		ctx.Response().Header.Set(rkmid.HeaderTraceId, "ut-trace-id")
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.NotEmpty(t, eventForValidation.GetRemoteAddr())
	assert.NotEmpty(t, eventForValidation.ListPayloads())
	assert.NotEmpty(t, eventForValidation.GetOperation())
	assert.NotEmpty(t, eventForValidation.GetRequestId())
	assert.NotEmpty(t, eventForValidation.GetTraceId())
	assert.NotEmpty(t, eventForValidation.GetResCode())
	assert.Equal(t, rkquery.Ended, eventForValidation.GetEventStatus())
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
