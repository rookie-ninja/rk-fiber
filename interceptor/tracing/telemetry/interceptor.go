// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibertrace is a middleware of fiber framework for recording trace info of RPC
package rkfibertrace

import (
	"context"
	"github.com/gofiber/fiber/v2"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidtrace "github.com/rookie-ninja/rk-entry/middleware/tracing"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
)

// Interceptor create a interceptor with opentelemetry.
func Interceptor(opts ...rkmidtrace.Option) fiber.Handler {
	set := rkmidtrace.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.TracerKey, set.GetTracer()))
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.TracerProviderKey, set.GetProvider()))
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.PropagatorKey, set.GetPropagator()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		beforeCtx := set.BeforeCtx(req, false)
		set.Before(beforeCtx)

		ctx.SetUserContext(beforeCtx.Output.NewCtx)

		// add to context
		if beforeCtx.Output.Span != nil {
			traceId := beforeCtx.Output.Span.SpanContext().TraceID().String()
			rkfiberctx.GetEvent(ctx).SetTraceId(traceId)
			ctx.Response().Header.Set(rkmid.HeaderTraceId, traceId)
			ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.SpanKey, beforeCtx.Output.Span))
		}

		err := ctx.Next()

		afterCtx := set.AfterCtx(ctx.Response().StatusCode(), "")
		set.After(beforeCtx, afterCtx)

		return err
	}
}
