// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibermetrics is a middleware for fiber framework which record prometheus metrics for RPC
package rkfibermetrics

import (
	"context"
	"github.com/gofiber/fiber/v2"
	rkmid "github.com/rookie-ninja/rk-entry/middleware"
	rkmidmetrics "github.com/rookie-ninja/rk-entry/middleware/metrics"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"net/http"
	"strconv"
)

// Interceptor create a new prometheus metrics interceptor with options.
func Interceptor(opts ...rkmidmetrics.Option) fiber.Handler {
	set := rkmidmetrics.NewOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkmid.EntryNameKey, set.GetEntryName()))

		req := &http.Request{}
		fasthttpadaptor.ConvertRequest(ctx.Context(), req, true)

		beforeCtx := set.BeforeCtx(req)
		set.Before(beforeCtx)

		err := ctx.Next()

		afterCtx := set.AfterCtx(strconv.Itoa(ctx.Response().StatusCode()))
		set.After(beforeCtx, afterCtx)

		return err
	}
}
