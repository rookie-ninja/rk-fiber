// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberctx defines utility functions and variables used by fiber middleware
package rkfiberctx

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	rkcursor "github.com/rookie-ninja/rk-entry/v2/cursor"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-logger"
	"github.com/rookie-ninja/rk-query"
	"github.com/valyala/fasthttp"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"
)

var (
	noopTracerProvider = trace.NewNoopTracerProvider()
	noopEvent          = rkquery.NewEventFactory().CreateEventNoop()
	pointerCreator     rkcursor.PointerCreator
)

// GetIncomingHeaders extract call-scoped incoming headers
func GetIncomingHeaders(ctx *fiber.Ctx) *fasthttp.RequestHeader {
	return &ctx.Request().Header
}

// AddHeaderToClient headers that would be sent to client.
// Values would be merged.
func AddHeaderToClient(ctx *fiber.Ctx, key, value string) {
	if ctx == nil || ctx.Response() == nil {
		return
	}

	ctx.Response().Header.Add(key, value)
}

// SetHeaderToClient headers that would be sent to client.
// Values would be overridden.
func SetHeaderToClient(ctx *fiber.Ctx, key, value string) {
	if ctx == nil || ctx.Response() == nil {
		return
	}

	ctx.Response().Header.Set(key, value)
}

// SetPointerCreator override  rkcursor.PointerCreator
func SetPointerCreator(creator rkcursor.PointerCreator) {
	pointerCreator = creator
}

// GetCursor create rkcursor.Cursor instance
func GetCursor(ctx *fiber.Ctx) *rkcursor.Cursor {
	res := rkcursor.NewCursor(
		rkcursor.WithLogger(GetLogger(ctx)),
		rkcursor.WithEvent(GetEvent(ctx)),
		rkcursor.WithEntryNameAndType(GetEntryName(ctx), "FiberEntry"))

	if pointerCreator != nil {
		res.Creator = pointerCreator
	}

	return res
}

// GetEvent extract takes the call-scoped EventData from middleware.
func GetEvent(ctx *fiber.Ctx) rkquery.Event {
	if ctx == nil {
		return noopEvent
	}

	if raw := ctx.UserContext().Value(rkmid.EventKey); raw != nil {
		return raw.(rkquery.Event)
	}

	return noopEvent
}

// GetLogger extract takes the call-scoped zap logger from middleware.
func GetLogger(ctx *fiber.Ctx) *zap.Logger {
	if ctx == nil {
		return rklogger.NoopLogger
	}

	if raw := ctx.UserContext().Value(rkmid.LoggerKey); raw != nil {
		requestId := GetRequestId(ctx)
		traceId := GetTraceId(ctx)
		fields := make([]zap.Field, 0)
		if len(requestId) > 0 {
			fields = append(fields, zap.String("requestId", requestId))
		}
		if len(traceId) > 0 {
			fields = append(fields, zap.String("traceId", traceId))
		}

		return raw.(*zap.Logger).With(fields...)
	}

	return rklogger.NoopLogger
}

func GormCtx(ctx *fiber.Ctx) context.Context {
	res := context.Background()
	res = context.WithValue(res, rkmid.LoggerKey.String(), GetLogger(ctx))
	res = context.WithValue(res, rkmid.EventKey.String(), GetEvent(ctx))
	return res
}

// GetRequestId extract request id from context.
// If user enabled meta interceptor, then a random request Id would e assigned and set to context as value.
// If user called AddHeaderToClient() with key of RequestIdKey, then a new request id would be updated.
func GetRequestId(ctx *fiber.Ctx) string {
	if ctx == nil || ctx.Response() == nil {
		return ""
	}

	return string(ctx.Response().Header.Peek(rkmid.HeaderRequestId))
}

// GetTraceId extract trace id from context.
func GetTraceId(ctx *fiber.Ctx) string {
	if ctx == nil || ctx.Response() == nil {
		return ""
	}

	return string(ctx.Response().Header.Peek(rkmid.HeaderTraceId))
}

// GetEntryName extract entry name from context.
func GetEntryName(ctx *fiber.Ctx) string {
	if ctx == nil {
		return ""
	}

	if raw := ctx.UserContext().Value(rkmid.EntryNameKey); raw != nil {
		return raw.(string)
	}

	return ""
}

// GetTraceSpan extract the call-scoped span from context.
func GetTraceSpan(ctx *fiber.Ctx) trace.Span {
	_, span := noopTracerProvider.Tracer("rk-trace-noop").Start(context.TODO(), "noop-span")

	if ctx == nil || ctx.Request() == nil {
		return span
	}

	_, span = noopTracerProvider.Tracer("rk-trace-noop").Start(ctx.Context(), "noop-span")

	if raw := ctx.UserContext().Value(rkmid.SpanKey); raw != nil {
		return raw.(trace.Span)
	}

	return span
}

// GetTracer extract the call-scoped tracer from context.
func GetTracer(ctx *fiber.Ctx) trace.Tracer {
	if ctx == nil {
		return noopTracerProvider.Tracer("rk-trace-noop")
	}

	if raw := ctx.UserContext().Value(rkmid.TracerKey); raw != nil {
		return raw.(trace.Tracer)
	}

	return noopTracerProvider.Tracer("rk-trace-noop")
}

// GetTracerProvider extract the call-scoped tracer provider from context.
func GetTracerProvider(ctx *fiber.Ctx) trace.TracerProvider {
	if ctx == nil {
		return noopTracerProvider
	}

	if raw := ctx.UserContext().Value(rkmid.TracerProviderKey); raw != nil {
		return raw.(trace.TracerProvider)
	}

	return noopTracerProvider
}

// GetTracerPropagator extract takes the call-scoped propagator from middleware.
func GetTracerPropagator(ctx *fiber.Ctx) propagation.TextMapPropagator {
	if ctx == nil {
		return nil
	}

	if raw := ctx.UserContext().Value(rkmid.PropagatorKey); raw != nil {
		return raw.(propagation.TextMapPropagator)
	}

	return nil
}

// InjectSpanToHttpRequest inject span to http request
func InjectSpanToHttpRequest(ctx *fiber.Ctx, req *http.Request) {
	if req == nil {
		return
	}

	newCtx := trace.ContextWithRemoteSpanContext(req.Context(), GetTraceSpan(ctx).SpanContext())

	if propagator := GetTracerPropagator(ctx); propagator != nil {
		propagator.Inject(newCtx, propagation.HeaderCarrier(req.Header))
	}
}

// NewTraceSpan start a new span
func NewTraceSpan(ctx *fiber.Ctx, name string) trace.Span {
	tracer := GetTracer(ctx)

	newCtx, span := tracer.Start(ctx.UserContext(), name)

	ctx.SetUserContext(newCtx)

	GetEvent(ctx).StartTimer(name)

	return span
}

// EndTraceSpan end span
func EndTraceSpan(ctx *fiber.Ctx, span trace.Span, success bool) {
	if success {
		span.SetStatus(otelcodes.Ok, otelcodes.Ok.String())
	}

	span.End()
}

// GetJwtToken return jwt.Token if exists
func GetJwtToken(ctx *fiber.Ctx) *jwt.Token {
	if ctx == nil {
		return nil
	}

	if raw := ctx.UserContext().Value(rkmid.JwtTokenKey); raw != nil {
		if res, ok := raw.(*jwt.Token); ok {
			return res
		}
	}

	return nil
}

// GetCsrfToken return csrf token if exists
func GetCsrfToken(ctx *fiber.Ctx) string {
	if ctx == nil {
		return ""
	}

	if raw := ctx.UserContext().Value(rkmid.CsrfTokenKey); raw != nil {
		if res, ok := raw.(string); ok {
			return res
		}
	}

	return ""
}
