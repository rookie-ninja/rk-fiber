// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiberctx

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"testing"
)

func newCtx() (*fiber.Ctx, *fasthttp.RequestCtx) {
	app := fiber.New()
	reqCtx := &fasthttp.RequestCtx{
		Request:  fasthttp.Request{},
		Response: fasthttp.Response{},
	}
	reqCtx.Request.SetRequestURI("/ut-path")
	reqCtx.Request.Header.SetMethod(http.MethodGet)
	ctx := app.AcquireCtx(reqCtx)
	return ctx, reqCtx
}

func TestGetIncomingHeaders(t *testing.T) {
	ctx, reqCtx := newCtx()
	reqCtx.Request.Header.Set("ut-key", "ut-value")

	assert.NotNil(t, GetIncomingHeaders(ctx))
	assert.Equal(t, "ut-value", string(GetIncomingHeaders(ctx).Peek("ut-key")))
}

func TestAddHeaderToClient(t *testing.T) {
	defer assertNotPanic(t)

	ctx, _ := newCtx()

	// With nil context
	AddHeaderToClient(nil, "", "")

	// With nil writer
	AddHeaderToClient(ctx, "", "")

	// Happy case
	AddHeaderToClient(ctx, "ut-key", "ut-value")
	assert.Equal(t, "ut-value", string(ctx.Response().Header.Peek("ut-key")))
}

func TestSetHeaderToClient(t *testing.T) {
	defer assertNotPanic(t)

	ctx, _ := newCtx()

	// With nil context
	SetHeaderToClient(nil, "", "")

	// With nil writer
	SetHeaderToClient(ctx, "", "")

	// Happy case
	SetHeaderToClient(ctx, "ut-key", "ut-value")
	assert.Equal(t, "ut-value", string(ctx.Response().Header.Peek("ut-key")))
}

func TestGetEvent(t *testing.T) {
	// With nil context
	assert.Equal(t, noopEvent, GetEvent(nil))

	// With no event in context
	ctx, _ := newCtx()
	assert.Equal(t, noopEvent, GetEvent(ctx))

	// Happy case
	event := rkquery.NewEventFactory().CreateEventNoop()
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEventKey, event))
	assert.Equal(t, event, GetEvent(ctx))
}

func TestGetLogger(t *testing.T) {
	// With nil context
	assert.Equal(t, rklogger.NoopLogger, GetLogger(nil))

	ctx, _ := newCtx()

	// With no logger in context
	assert.Equal(t, rklogger.NoopLogger, GetLogger(ctx))

	// Happy case
	// Add request id and trace id
	ctx.Response().Header.Set(RequestIdKey, "ut-request-id")
	ctx.Response().Header.Set(TraceIdKey, "ut-trace-id")

	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcLoggerKey, rklogger.NoopLogger))
	assert.Equal(t, rklogger.NoopLogger, GetLogger(ctx))
}

func TestGetRequestId(t *testing.T) {
	// With nil context
	assert.Empty(t, GetRequestId(nil))

	ctx, _ := newCtx()

	// With no requestId in context
	assert.Empty(t, GetRequestId(ctx))

	// Happy case
	ctx.Response().Header.Set(RequestIdKey, "ut-request-id")
	assert.Equal(t, "ut-request-id", GetRequestId(ctx))
}

func TestGetTraceId(t *testing.T) {
	// With nil context
	assert.Empty(t, GetTraceId(nil))

	ctx, _ := newCtx()

	// With no traceId in context
	assert.Empty(t, GetTraceId(ctx))

	// Happy case
	ctx.Response().Header.Set(TraceIdKey, "ut-trace-id")
	assert.Equal(t, "ut-trace-id", GetTraceId(ctx))
}

func TestGetEntryName(t *testing.T) {
	// With nil context
	assert.Empty(t, GetEntryName(nil))

	ctx, _ := newCtx()

	// With no entry name in context
	assert.Empty(t, GetEntryName(ctx))

	// Happy case
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, "ut-entry-name"))
	assert.Equal(t, "ut-entry-name", GetEntryName(ctx))
}

func TestGetTraceSpan(t *testing.T) {
	ctx, _ := newCtx()

	// With nil context
	assert.NotNil(t, GetTraceSpan(nil))

	// With no span in context
	assert.NotNil(t, GetTraceSpan(ctx))

	// Happy case
	_, span := noopTracerProvider.Tracer("ut-trace").Start(ctx.Context(), "noop-span")
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcSpanKey, span))
	assert.Equal(t, span, GetTraceSpan(ctx))
}

func TestGetTracer(t *testing.T) {
	ctx, _ := newCtx()

	// With nil context
	assert.NotNil(t, GetTracer(nil))

	// With no tracer in context
	assert.NotNil(t, GetTracer(ctx))

	// Happy case
	tracer := noopTracerProvider.Tracer("ut-trace")
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcTracerKey, tracer))
	assert.Equal(t, tracer, GetTracer(ctx))
}

func TestGetTracerProvider(t *testing.T) {
	ctx, _ := newCtx()

	// With nil context
	assert.NotNil(t, GetTracerProvider(nil))

	// With no tracer provider in context
	assert.NotNil(t, GetTracerProvider(ctx))

	// Happy case
	provider := trace.NewNoopTracerProvider()
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcTracerProviderKey, provider))
	assert.Equal(t, provider, GetTracerProvider(ctx))
}

func TestGetTracerPropagator(t *testing.T) {
	ctx, _ := newCtx()

	// With nil context
	assert.Nil(t, GetTracerPropagator(nil))

	// With no tracer propagator in context
	assert.Nil(t, GetTracerPropagator(ctx))

	// Happy case
	prop := propagation.NewCompositeTextMapPropagator()
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcPropagatorKey, prop))
	assert.Equal(t, prop, GetTracerPropagator(ctx))
}

func TestInjectSpanToHttpRequest(t *testing.T) {
	defer assertNotPanic(t)

	// With nil context and request
	InjectSpanToHttpRequest(nil, nil)

	// Happy case
	ctx, _ := newCtx()

	prop := propagation.NewCompositeTextMapPropagator()
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcPropagatorKey, prop))

	InjectSpanToHttpRequest(ctx, &http.Request{
		Header: http.Header{},
	})
}

func TestNewTraceSpan(t *testing.T) {
	ctx, _ := newCtx()

	assert.NotNil(t, NewTraceSpan(ctx, "ut-span"))
}

func TestEndTraceSpan(t *testing.T) {
	defer assertNotPanic(t)

	ctx, _ := newCtx()

	// With success
	span := GetTraceSpan(ctx)
	EndTraceSpan(ctx, span, true)

	// With failure
	span = GetTraceSpan(ctx)
	EndTraceSpan(ctx, span, false)
}

func TestGetJwtToken(t *testing.T) {
	ctx, _ := newCtx()

	// with nil context
	assert.Nil(t, GetJwtToken(nil))

	// without jwt token
	assert.Nil(t, GetJwtToken(ctx))

	// happy case
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcJwtTokenKey, &jwt.Token{}))
	assert.NotNil(t, GetJwtToken(ctx))
}

func TestGetCsrfToken(t *testing.T) {
	ctx, _ := newCtx()

	// with nil context
	assert.Empty(t, GetCsrfToken(nil))

	// without csrf token
	assert.Empty(t, GetCsrfToken(ctx))

	// happy case
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcCsrfTokenKey, "csrf-token"))
	assert.NotEmpty(t, GetCsrfToken(ctx))
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
