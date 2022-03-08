// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberpanic is a middleware of fiber framework for recovering from panic
package rkfiberpanic

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/v2/error"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/panic"
	"github.com/rookie-ninja/rk-fiber/middleware/context"
)

// Middleware returns a fiber.Handler(middleware)
func Middleware(opts ...rkmidpanic.Option) fiber.Handler {
	set := rkmidpanic.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		handlerFunc := func(resp *rkerror.ErrorResp) {
			ctx.Response().SetStatusCode(resp.Err.Code)
			ctx.JSON(resp)
		}
		beforeCtx := set.BeforeCtx(rkfiberctx.GetEvent(ctx), rkfiberctx.GetLogger(ctx), handlerFunc)
		set.Before(beforeCtx)

		defer beforeCtx.Output.DeferFunc()

		return ctx.Next()
	}
}
