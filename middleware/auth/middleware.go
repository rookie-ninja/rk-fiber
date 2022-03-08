// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberauth is auth middleware for fiber framework
package rkfiberauth

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/auth"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
)

// Middleware validate bellow authorization.
//
// 1: Basic Auth: The client sends HTTP requests with the Authorization header that contains the word Basic, followed by a space and a base64-encoded(non-encrypted) string username: password.
// 2: Bearer Token: Commonly known as token authentication. It is an HTTP authentication scheme that involves security tokens called bearer tokens.
// 3: API key: An API key is a token that a client provides when making API calls. With API key auth, you send a key-value pair to the API in the request headers.
func Middleware(opts ...rkmidauth.Option) fiber.Handler {
	set := rkmidauth.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		// case 1: return to user if error occur
		beforeCtx := set.BeforeCtx(req)
		set.Before(beforeCtx)

		if beforeCtx.Output.ErrResp != nil {
			for k, v := range beforeCtx.Output.HeadersToReturn {
				ctx.Response().Header.Set(k, v)
			}
			ctx.Response().SetStatusCode(beforeCtx.Output.ErrResp.Err.Code)
			ctx.JSON(beforeCtx.Output.ErrResp)
		}

		return ctx.Next()
	}
}
