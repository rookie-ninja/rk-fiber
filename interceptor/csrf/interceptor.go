// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibercsrf is a middleware for fiber framework which validating csrf token for RPC
package rkfibercsrf

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"net/http"
	"time"
)

// Interceptor Add csrf interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/csrf.go
func Interceptor(opts ...Option) fiber.Handler {
	set := newOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, set.EntryName))

		// 1: skip csrf check based on skipper
		if set.Skipper(ctx) {
			return ctx.Next()
		}

		k := ctx.Cookies(set.CookieName)
		token := ""

		// 2.1: generate token if failed to get cookie from context
		if len(k) < 1 {
			token = rkcommon.RandString(set.TokenLength)
		} else {
			// 2.2: reuse token if exists
			token = k
		}

		// 3.1: do not check http methods of GET, HEAD, OPTIONS and TRACE
		switch ctx.Method() {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			// 3.2: validate token only for requests which are not defined as 'safe' by RFC7231
			clientToken, err := set.extractor(ctx)
			if err != nil {
				ctx.JSON(rkerror.New(
					rkerror.WithHttpCode(http.StatusBadRequest),
					rkerror.WithMessage("failed to extract client token"),
					rkerror.WithDetails(err)))
				return fiber.ErrBadRequest
			}

			// 3.3: return 403 to client if token is not matched
			if !isValidToken(token, clientToken) {
				ctx.JSON(rkerror.New(
					rkerror.WithHttpCode(http.StatusForbidden),
					rkerror.WithMessage("invalid csrf token"),
					rkerror.WithDetails(err)))
				return fiber.ErrForbidden
			}
		}

		// set CSRF cookie
		cookie := &fiber.Cookie{}
		cookie.Name = set.CookieName
		cookie.Value = token

		// 4.1
		if set.CookiePath != "" {
			cookie.Path = set.CookiePath
		}
		// 4.2
		if set.CookieDomain != "" {
			cookie.Domain = set.CookieDomain
		}
		// 4.3
		if set.CookieSameSite != http.SameSiteDefaultMode {
			switch set.CookieSameSite {
			case http.SameSiteNoneMode:
				cookie.SameSite = fiber.CookieSameSiteNoneMode
			case http.SameSiteLaxMode:
				cookie.SameSite = fiber.CookieSameSiteLaxMode
			case http.SameSiteStrictMode:
				cookie.SameSite = fiber.CookieSameSiteStrictMode
			}
		}
		cookie.Expires = time.Now().Add(time.Duration(set.CookieMaxAge) * time.Second)
		cookie.Secure = set.CookieSameSite == http.SameSiteNoneMode
		cookie.HTTPOnly = set.CookieHTTPOnly
		ctx.Cookie(cookie)

		// store token in the context
		ctx.Set(rkfiberinter.RpcCsrfTokenKey, token)

		// protect clients from caching the response
		ctx.Response().Header.Add(headerVary, headerCookie)

		return ctx.Next()
	}
}
