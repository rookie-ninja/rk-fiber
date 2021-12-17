// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberlimit is a middleware of fiber framework for adding rate limit in RPC response
package rkfiberlimit

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"net/http"
)

// Interceptor Add rate limit interceptors.
func Interceptor(opts ...Option) fiber.Handler {
	set := newOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, set.EntryName))

		event := rkfiberctx.GetEvent(ctx)

		if duration, err := set.Wait(ctx); err != nil {
			event.SetCounter("rateLimitWaitMs", duration.Milliseconds())
			event.AddErr(err)

			ctx.JSON(rkerror.New(
				rkerror.WithHttpCode(http.StatusTooManyRequests),
				rkerror.WithDetails(err)))
			return fiber.ErrTooManyRequests
		}

		return ctx.Next()
	}
}
