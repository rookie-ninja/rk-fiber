// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberpanic is a middleware of fiber framework for recovering from panic
package rkfiberpanic

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-common/error"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidpanic "github.com/rookie-ninja/rk-entry/middleware/panic"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
)

// Interceptor returns a fiber.Handler(middleware)
func Interceptor(opts ...rkmidpanic.Option) fiber.Handler {
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
