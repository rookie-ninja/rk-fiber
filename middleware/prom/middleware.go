// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberprom is a middleware for fiber framework which record prometheus metrics for RPC
package rkfiberprom

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/prom"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
	"strconv"
)

// Middleware create a new prometheus metrics interceptor with options.
func Middleware(opts ...rkmidprom.Option) fiber.Handler {
	set := rkmidprom.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		beforeCtx := set.BeforeCtx(req)
		set.Before(beforeCtx)

		err := ctx.Next()

		afterCtx := set.AfterCtx(strconv.Itoa(ctx.Response().StatusCode()))
		set.After(beforeCtx, afterCtx)

		return err
	}
}
