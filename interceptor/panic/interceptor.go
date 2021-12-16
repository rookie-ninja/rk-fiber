// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberpanic is a middleware of fiber framework for recovering from panic
package rkfiberpanic

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
)

// Interceptor returns a fiber.Handler(middleware)
func Interceptor(opts ...Option) fiber.Handler {
	set := newOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, set.EntryName))

		defer func() {
			if recv := recover(); recv != nil {
				var res *rkerror.ErrorResp

				if se, ok := recv.(*rkerror.ErrorResp); ok {
					res = se
				} else if re, ok := recv.(error); ok {
					res = rkerror.FromError(re)
				} else {
					res = rkerror.New(rkerror.WithMessage(fmt.Sprintf("%v", recv)))
				}

				rkfiberctx.GetEvent(ctx).SetCounter("panic", 1)
				rkfiberctx.GetEvent(ctx).AddErr(res.Err)
				rkfiberctx.GetLogger(ctx).Error(fmt.Sprintf("panic occurs:\n%s", string(debug.Stack())), zap.Error(res.Err))

				ctx.Response().SetStatusCode(http.StatusInternalServerError)
				ctx.JSON(res)
			}
		}()

		return ctx.Next()
	}
}
