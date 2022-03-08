// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibertimeout is a middleware of fiber framework for timing out request in RPC response
package rkfibertimeout

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/timeout"
	"github.com/rookie-ninja/rk-fiber/middleware/context"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
)

// Middleware Add timeout interceptors.
func Middleware(opts ...rkmidtimeout.Option) fiber.Handler {
	set := rkmidtimeout.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		// case 1: return to user if error occur
		beforeCtx := set.BeforeCtx(req, rkfiberctx.GetEvent(ctx))
		toCtx := &timeoutCtx{
			fiberCtx: ctx,
			before:   beforeCtx,
		}
		// assign handlers
		beforeCtx.Input.InitHandler = initHandler(toCtx)
		beforeCtx.Input.NextHandler = nextHandler(toCtx)
		beforeCtx.Input.PanicHandler = panicHandler(toCtx)
		beforeCtx.Input.FinishHandler = finishHandler(toCtx)
		beforeCtx.Input.TimeoutHandler = timeoutHandler(toCtx)

		// call before
		set.Before(beforeCtx)

		beforeCtx.Output.WaitFunc()

		return toCtx.nextError
	}
}

type timeoutCtx struct {
	fiberCtx  *fiber.Ctx
	before    *rkmidtimeout.BeforeCtx
	nextError error
}

func timeoutHandler(ctx *timeoutCtx) func() {
	return func() {
		ctx.fiberCtx.Response().SetStatusCode(ctx.before.Output.TimeoutErrResp.Err.Code)
		ctx.fiberCtx.JSON(ctx.before.Output.TimeoutErrResp)
	}
}

func finishHandler(ctx *timeoutCtx) func() {
	return func() {}
}

func panicHandler(ctx *timeoutCtx) func() {
	return func() {}
}

func nextHandler(ctx *timeoutCtx) func() {
	return func() {
		ctx.nextError = ctx.fiberCtx.Next()
	}
}

func initHandler(ctx *timeoutCtx) func() {
	return func() {}
}
