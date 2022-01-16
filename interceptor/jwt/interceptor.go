// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberjwt is a middleware for fiber framework which validating jwt token for RPC
package rkfiberjwt

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/middleware"
	"github.com/rookie-ninja/rk-entry/middleware/jwt"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
)

// Interceptor Add jwt interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/jwt.go
func Interceptor(opts ...rkmidjwt.Option) fiber.Handler {
	set := rkmidjwt.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		beforeCtx := set.BeforeCtx(req, nil)
		set.Before(beforeCtx)

		// case 1: error response
		if beforeCtx.Output.ErrResp != nil {
			ctx.Response().SetStatusCode(beforeCtx.Output.ErrResp.Err.Code)
			return ctx.JSON(beforeCtx.Output.ErrResp)
		}

		// insert into context
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.JwtTokenKey, beforeCtx.Output.JwtToken))

		return ctx.Next()
	}
}
