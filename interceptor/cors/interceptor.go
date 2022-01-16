// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibercors is a CORS middleware for fiber framework
package rkfibercors

import (
	"context"
	"github.com/gofiber/fiber/v2"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidcors "github.com/rookie-ninja/rk-entry/middleware/cors"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
)

// Interceptor Add cors interceptors.
//
// Mainly copied and modified from bellow.
// https://github.com/labstack/echo/blob/master/middleware/cors.go
func Interceptor(opts ...rkmidcors.Option) fiber.Handler {
	set := rkmidcors.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		beforeCtx := set.BeforeCtx(req)
		set.Before(beforeCtx)

		for k, v := range beforeCtx.Output.HeadersToReturn {
			ctx.Response().Header.Set(k, v)
		}

		for _, v := range beforeCtx.Output.HeaderVary {
			ctx.Response().Header.Add(rkmid.HeaderVary, v)
		}

		// case 1: with abort
		if beforeCtx.Output.Abort {
			return fiber.NewError(http.StatusNoContent)
		}

		return ctx.Next()
	}
}
