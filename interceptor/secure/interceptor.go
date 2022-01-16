// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibersec is a middleware of fiber framework for adding secure headers in RPC response
package rkfibersec

import (
	"context"
	"github.com/gofiber/fiber/v2"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidsec "github.com/rookie-ninja/rk-entry/middleware/secure"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
)

// Interceptor Add security interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/secure.go
func Interceptor(opts ...rkmidsec.Option) fiber.Handler {
	set := rkmidsec.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		// case 1: return to user if error occur
		beforeCtx := set.BeforeCtx(req)
		set.Before(beforeCtx)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Response().Header.Set(k, v)
		}

		return ctx.Next()
	}
}
