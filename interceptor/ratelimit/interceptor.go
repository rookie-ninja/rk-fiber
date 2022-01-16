// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberlimit is a middleware of fiber framework for adding rate limit in RPC response
package rkfiberlimit

import (
	"context"
	"github.com/gofiber/fiber/v2"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidlimit "github.com/rookie-ninja/rk-entry/middleware/ratelimit"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
)

// Interceptor Add rate limit interceptors.
func Interceptor(opts ...rkmidlimit.Option) fiber.Handler {
	set := rkmidlimit.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		beforeCtx := set.BeforeCtx(req)
		set.Before(beforeCtx)

		if beforeCtx.Output.ErrResp != nil {
			ctx.Response().SetStatusCode(beforeCtx.Output.ErrResp.Err.Code)
			return ctx.JSON(beforeCtx.Output.ErrResp)
		}

		return ctx.Next()
	}
}
