// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberjwt is a middleware for fiber framework which validating jwt token for RPC
package rkfiberjwt

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"net/http"
)

// Interceptor Add jwt interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/jwt.go
func Interceptor(opts ...Option) fiber.Handler {
	set := newOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, set.EntryName))

		if set.Skipper(ctx) {
			return ctx.Next()
		}

		// extract token from extractor
		var auth string
		var err error
		for _, extractor := range set.extractors {
			// Extract token from extractor, if it's not fail break the loop and
			// set auth
			auth, err = extractor(ctx)
			if err == nil {
				break
			}
		}

		if err != nil {
			ctx.JSON(rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("invalid or expired jwt"),
				rkerror.WithDetails(err)))
			return fiber.ErrUnauthorized
		}

		// parse token
		token, err := set.ParseTokenFunc(auth, ctx)

		if err != nil {
			ctx.JSON(rkerror.New(
				rkerror.WithHttpCode(http.StatusUnauthorized),
				rkerror.WithMessage("invalid or expired jwt"),
				rkerror.WithDetails(err)))
			return fiber.ErrUnauthorized
		}

		// insert into context
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcJwtTokenKey, token))

		return ctx.Next()
	}
}
