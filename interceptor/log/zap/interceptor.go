// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberlog is a middleware for fiber framework for logging RPC.
package rkfiberlog

import (
	"context"
	"github.com/gofiber/fiber/v2"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidlog "github.com/rookie-ninja/rk-entry/middleware/log"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
	"strconv"
)

// Interceptor returns a fiber.Handler (middleware) that logs requests using uber-go/zap.
func Interceptor(opts ...rkmidlog.Option) fiber.Handler {
	set := rkmidlog.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		// call before
		beforeCtx := set.BeforeCtx(req)
		set.Before(beforeCtx)

		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EventKey, beforeCtx.Output.Event))
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.LoggerKey, beforeCtx.Output.Logger))

		err := ctx.Next()

		afterCtx := set.AfterCtx(
			rkfiberctx.GetRequestId(ctx),
			rkfiberctx.GetTraceId(ctx),
			strconv.Itoa(ctx.Response().StatusCode()))
		set.After(beforeCtx, afterCtx)

		return err
	}
}
