// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibermeta is a middleware of fiber framework for adding metadata in RPC response
package rkfibermeta

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/meta"
	"github.com/rookie-ninja/rk-fiber/middleware/context"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
)

// Middleware will add common headers as extension style in http response.
func Middleware(opts ...rkmidmeta.Option) fiber.Handler {
	set := rkmidmeta.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		beforeCtx := set.BeforeCtx(req, rkfiberctx.GetEvent(ctx))
		set.Before(beforeCtx)

		ctx.Set(rkmid.HeaderRequestId, beforeCtx.Output.RequestId)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Response().Header.Set(k, v)
		}

		return ctx.Next()
	}
}
