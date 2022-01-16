// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibermeta is a middleware of fiber framework for adding metadata in RPC response
package rkfibermeta

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/meta"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
)

// Interceptor will add common headers as extension style in http response.
func Interceptor(opts ...rkmidmeta.Option) fiber.Handler {
	set := rkmidmeta.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		beforeCtx := set.BeforeCtx(rkfiberctx.GetEvent(ctx))
		set.Before(beforeCtx)

		ctx.Set(rkmid.HeaderRequestId, beforeCtx.Output.RequestId)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Response().Header.Set(k, v)
		}

		return ctx.Next()
	}
}
