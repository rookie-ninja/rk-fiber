// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberlog is a middleware for fiber framework for logging RPC.
package rkfiberlog

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// Interceptor returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
func Interceptor(opts ...Option) fiber.Handler {
	set := newOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, set.EntryName))

		before(ctx, set)

		err := ctx.Next()

		after(ctx)

		return err
	}
}

func before(ctx *fiber.Ctx, set *optionSet) {
	var event rkquery.Event
	if rkfiberinter.ShouldLog(ctx) {
		event = set.eventLoggerEntry.GetEventFactory().CreateEvent(
			rkquery.WithZapLogger(set.eventLoggerOverride),
			rkquery.WithEncoding(set.eventLoggerEncoding),
			rkquery.WithAppName(rkentry.GlobalAppCtx.GetAppInfoEntry().AppName),
			rkquery.WithAppVersion(rkentry.GlobalAppCtx.GetAppInfoEntry().Version),
			rkquery.WithEntryName(set.EntryName),
			rkquery.WithEntryType(set.EntryType))
	} else {
		event = set.eventLoggerEntry.GetEventFactory().CreateEventNoop()
	}

	event.SetStartTime(time.Now())

	remoteIp, remotePort := rkfiberinter.GetRemoteAddressSet(ctx)
	// handle remote address
	event.SetRemoteAddr(remoteIp + ":" + remotePort)

	payloads := []zap.Field{
		zap.String("apiPath", ctx.Path()),
		zap.String("apiMethod", ctx.Method()),
		zap.String("apiQuery", ctx.Context().QueryArgs().String()),
		zap.String("apiProtocol", ctx.Protocol()),
		zap.String("userAgent", string(ctx.Context().UserAgent())),
	}

	// handle payloads
	event.AddPayloads(payloads...)

	// handle operation
	event.SetOperation(ctx.Path())

	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEventKey, event))
	ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcLoggerKey, set.ZapLogger))
}

func after(ctx *fiber.Ctx) {
	event := rkfiberctx.GetEvent(ctx)

	if requestId := rkfiberctx.GetRequestId(ctx); len(requestId) > 0 {
		event.SetEventId(requestId)
		event.SetRequestId(requestId)
	}

	if traceId := rkfiberctx.GetTraceId(ctx); len(traceId) > 0 {
		event.SetTraceId(traceId)
	}

	event.SetResCode(strconv.Itoa(ctx.Response().StatusCode()))
	event.SetEndTime(time.Now())
	event.Finish()
}
