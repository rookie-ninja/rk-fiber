// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibercsrf is a middleware for fiber framework which validating csrf token for RPC
package rkfibercsrf

import (
	"context"
	"github.com/gofiber/fiber/v2"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidcsrf "github.com/rookie-ninja/rk-entry/middleware/csrf"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
)

// Interceptor Add csrf interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/csrf.go
func Interceptor(opts ...rkmidcsrf.Option) fiber.Handler {
	set := rkmidcsrf.NewOptionSet(opts...)

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

		for _, v := range beforeCtx.Output.VaryHeaders {
			ctx.Response().Header.Add(rkmid.HeaderVary, v)
		}

		if beforeCtx.Output.Cookie != nil {
			cookie := &fiber.Cookie{
				Name:     beforeCtx.Output.Cookie.Name,
				Value:    beforeCtx.Output.Cookie.Value,
				Path:     beforeCtx.Output.Cookie.Path,
				Domain:   beforeCtx.Output.Cookie.Domain,
				Expires:  beforeCtx.Output.Cookie.Expires,
				Secure:   beforeCtx.Output.Cookie.Secure,
				HTTPOnly: beforeCtx.Output.Cookie.HttpOnly,
			}

			switch beforeCtx.Output.Cookie.SameSite {
			case http.SameSiteNoneMode:
				cookie.SameSite = fiber.CookieSameSiteNoneMode
			case http.SameSiteLaxMode:
				cookie.SameSite = fiber.CookieSameSiteLaxMode
			case http.SameSiteStrictMode:
				cookie.SameSite = fiber.CookieSameSiteStrictMode
			}

			ctx.Cookie(cookie)
		}

		// store token in the context
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.CsrfTokenKey, beforeCtx.Input.Token))

		return ctx.Next()
	}
}
