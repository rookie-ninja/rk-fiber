// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibercors is a CORS middleware for fiber framework
package rkfibercors

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"net/http"
	"strconv"
	"strings"
)

// Interceptor Add cors interceptors.
//
// Mainly copied and modified from bellow.
// https://github.com/labstack/echo/blob/master/middleware/cors.go
func Interceptor(opts ...Option) fiber.Handler {
	set := newOptionSet(opts...)

	allowMethods := strings.Join(set.AllowMethods, ",")
	allowHeaders := strings.Join(set.AllowHeaders, ",")
	exposeHeaders := strings.Join(set.ExposeHeaders, ",")
	maxAge := strconv.Itoa(set.MaxAge)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, set.EntryName))

		if set.Skipper(ctx) {
			return ctx.Next()
		}

		originHeader := string(ctx.Request().Header.Peek(fiber.HeaderOrigin))
		preflight := string(ctx.Context().Method()) == http.MethodOptions

		// 1: if no origin header was provided, we will return 204 if request is not a OPTION method
		if originHeader == "" {
			// 1.1: if not a preflight request, then pass through
			if !preflight {
				return ctx.Next()
			}

			// 1.2: if it is a preflight request, then return with 204
			return fiber.NewError(http.StatusNoContent)
		}

		// 2: origin not allowed, we will return 204 if request is not a OPTION method
		if !set.isOriginAllowed(originHeader) {
			return fiber.NewError(http.StatusNoContent)
		}

		// 3: not a OPTION method
		if !preflight {
			ctx.Response().Header.Set(fiber.HeaderAccessControlAllowOrigin, originHeader)
			// 3.1: add Access-Control-Allow-Credentials
			if set.AllowCredentials {
				ctx.Response().Header.Set(fiber.HeaderAccessControlAllowCredentials, "true")
			}
			// 3.2: add Access-Control-Expose-Headers
			if exposeHeaders != "" {
				ctx.Response().Header.Set(fiber.HeaderAccessControlExposeHeaders, exposeHeaders)
			}
			return ctx.Next()
		}

		// 4: preflight request, return 204
		// add related headers including:
		//
		// - Vary
		// - Access-Control-Allow-Origin
		// - Access-Control-Allow-Methods
		// - Access-Control-Allow-Credentials
		// - Access-Control-Allow-Headers
		// - Access-Control-Max-Age
		ctx.Response().Header.Add(fiber.HeaderVary, fiber.HeaderAccessControlRequestMethod)
		ctx.Response().Header.Add(fiber.HeaderVary, fiber.HeaderAccessControlRequestHeaders)
		ctx.Response().Header.Set(fiber.HeaderAccessControlAllowOrigin, originHeader)
		ctx.Response().Header.Set(fiber.HeaderAccessControlAllowMethods, allowMethods)

		// 4.1: Access-Control-Allow-Credentials
		if set.AllowCredentials {
			ctx.Response().Header.Set(fiber.HeaderAccessControlAllowCredentials, "true")
		}

		// 4.2: Access-Control-Allow-Headers
		if allowHeaders != "" {
			ctx.Response().Header.Set(fiber.HeaderAccessControlAllowHeaders, allowHeaders)
		} else {
			h := string(ctx.Request().Header.Peek(fiber.HeaderAccessControlRequestHeaders))
			if h != "" {
				ctx.Response().Header.Set(fiber.HeaderAccessControlAllowHeaders, h)
			}
		}
		if set.MaxAge > 0 {
			// 4.3: Access-Control-Max-Age
			ctx.Response().Header.Set(fiber.HeaderAccessControlMaxAge, maxAge)
		}

		return fiber.NewError(http.StatusNoContent)
	}
}
