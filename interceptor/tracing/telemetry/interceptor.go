// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibertrace is aa middleware of fiber framework for recording trace info of RPC
package rkfibertrace

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"net/http"
)

// Interceptor create a interceptor with opentelemetry.
func Interceptor(opts ...Option) fiber.Handler {
	set := newOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, set.EntryName))
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcTracerKey, set.Tracer))
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcTracerProviderKey, set.Provider))
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcPropagatorKey, set.Propagator))

		span := before(ctx, set)
		defer span.End()

		err := ctx.Next()

		after(ctx, span)

		return err
	}
}

func before(ctx *fiber.Ctx, set *optionSet) oteltrace.Span {
	// let's do the trick here, create a http request based on fiber request
	fakeHttpRequest := &http.Request{}
	fasthttpadaptor.ConvertRequest(ctx.Context(), fakeHttpRequest, true)

	opts := []oteltrace.SpanStartOption{
		oteltrace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", fakeHttpRequest)...),
		oteltrace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(fakeHttpRequest)...),
		oteltrace.WithAttributes(semconv.HTTPServerAttributesFromHTTPRequest(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName, ctx.Path(), fakeHttpRequest)...),
		oteltrace.WithAttributes(localeToAttributes()...),
		oteltrace.WithSpanKind(oteltrace.SpanKindServer),
	}

	// 1: extract tracing info from request header
	spanCtx := oteltrace.SpanContextFromContext(
		set.Propagator.Extract(ctx.Context(), &FastHttpHeaderCarrier{&ctx.Request().Header}))

	spanName := ctx.Path()
	if len(spanName) < 1 {
		spanName = "rk-span-default"
	}

	// 2: start new span
	newRequestCtx, span := set.Tracer.Start(
		oteltrace.ContextWithRemoteSpanContext(ctx.Context(), spanCtx),
		spanName, opts...)
	ctx.SetUserContext(newRequestCtx)

	// 3: read trace id, tracer, traceProvider, propagator and logger into event data and echo context
	rkfiberctx.GetEvent(ctx).SetTraceId(span.SpanContext().TraceID().String())
	ctx.Response().Header.Set(rkfiberctx.TraceIdKey, span.SpanContext().TraceID().String())
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcSpanKey, span))

	return span
}

func after(ctx *fiber.Ctx, span oteltrace.Span) {
	attrs := semconv.HTTPAttributesFromHTTPStatusCode(ctx.Response().StatusCode())
	spanStatus, spanMessage := semconv.SpanStatusFromHTTPStatusCode(ctx.Response().StatusCode())
	span.SetAttributes(attrs...)
	span.SetStatus(spanStatus, spanMessage)
}

// Convert locale information into attributes.
func localeToAttributes() []attribute.KeyValue {
	res := []attribute.KeyValue{
		attribute.String(rkfiberinter.Realm.Key, rkfiberinter.Realm.String),
		attribute.String(rkfiberinter.Region.Key, rkfiberinter.Region.String),
		attribute.String(rkfiberinter.AZ.Key, rkfiberinter.AZ.String),
		attribute.String(rkfiberinter.Domain.Key, rkfiberinter.Domain.String),
	}

	return res
}
