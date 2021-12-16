// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibermetrics is a middleware for fiber framework which record prometheus metrics for RPC
package rkfibermetrics

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"time"
)

// Interceptor create a new prometheus metrics interceptor with options.
func Interceptor(opts ...Option) fiber.Handler {
	set := newOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, set.EntryName))

		// start timer
		startTime := time.Now()

		err := ctx.Next()

		// end timer
		elapsed := time.Now().Sub(startTime)

		// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
		if rkfiberinter.ShouldLog(ctx) {
			if durationMetrics := GetServerDurationMetrics(ctx); durationMetrics != nil {
				durationMetrics.Observe(float64(elapsed.Nanoseconds()))
			}

			if resCodeMetrics := GetServerResCodeMetrics(ctx); resCodeMetrics != nil {
				resCodeMetrics.Inc()
			}
		}

		return err
	}
}
