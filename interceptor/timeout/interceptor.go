// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibertimeout is a middleware of fiber framework for timing out request in RPC response
package rkfibertimeout

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-fiber/interceptor"
)

// Interceptor Add timeout interceptors.
func Interceptor(opts ...Option) fiber.Handler {
	set := newOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, set.EntryName))

		return set.Tick(ctx)
	}
}
