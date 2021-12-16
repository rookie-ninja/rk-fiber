// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiberinter provides common utility functions for middleware of fiber framework
package rkfiberinter

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-common/common"
	"go.uber.org/zap"
	"net"
	"strings"
)

var (
	// Realm environment variable
	Realm = zap.String("realm", rkcommon.GetEnvValueOrDefault("REALM", "*"))
	// Region environment variable
	Region = zap.String("region", rkcommon.GetEnvValueOrDefault("REGION", "*"))
	// AZ environment variable
	AZ = zap.String("az", rkcommon.GetEnvValueOrDefault("AZ", "*"))
	// Domain environment variable
	Domain = zap.String("domain", rkcommon.GetEnvValueOrDefault("DOMAIN", "*"))
	// LocalIp read local IP from localhost
	LocalIp = zap.String("localIp", rkcommon.GetLocalIP())
	// LocalHostname read hostname from localhost
	LocalHostname = zap.String("localHostname", rkcommon.GetLocalHostname())
)

const (
	// RpcEntryNameKey entry name key
	RpcEntryNameKey = "fiberEntryName"
	// RpcEntryNameValue entry name
	RpcEntryNameValue = "fiber"
	// RpcEntryTypeValue entry type
	RpcEntryTypeValue = "fiber"
	// RpcEventKey event key
	RpcEventKey = "fiberEvent"
	// RpcLoggerKey logger key
	RpcLoggerKey = "fiberLogger"
	// RpcTracerKey tracer key
	RpcTracerKey = "fiberTracer"
	// RpcSpanKey span key
	RpcSpanKey = "fiberSpan"
	// RpcTracerProviderKey trace provider key
	RpcTracerProviderKey = "fiberTracerProvider"
	// RpcPropagatorKey propagator key
	RpcPropagatorKey = "fiberPropagator"
	// RpcAuthorizationHeaderKey auth key
	RpcAuthorizationHeaderKey = "authorization"
	// RpcApiKeyHeaderKey api auth key
	RpcApiKeyHeaderKey = "X-API-Key"
	// RpcJwtTokenKey key of jwt token in context
	RpcJwtTokenKey = "fiberJwt"
	// RpcCsrfTokenKey key of csrf token injected by csrf middleware
	RpcCsrfTokenKey = "fiberCsrfToken"
)

// GetRemoteAddressSet returns remote endpoint information set including IP, Port.
// We will do as best as we can to determine it.
// If fails, then just return default ones.
func GetRemoteAddressSet(ctx *fiber.Ctx) (remoteIp, remotePort string) {
	remoteIp, remotePort = "0.0.0.0", "0"

	if ctx == nil || ctx.Request() == nil {
		return
	}

	var err error
	if remoteIp, remotePort, err = net.SplitHostPort(ctx.Context().RemoteAddr().String()); err != nil {
		return
	}

	forwardedRemoteIp := ctx.IPs()

	// Deal with forwarded remote ip
	if len(forwardedRemoteIp) > 0 {
		if forwardedRemoteIp[0] == "::1" {
			forwardedRemoteIp[0] = "localhost"
		}

		remoteIp = forwardedRemoteIp[0]
	}

	if remoteIp == "::1" {
		remoteIp = "localhost"
	}

	return remoteIp, remotePort
}

// ShouldLog determines whether should log the RPC
func ShouldLog(ctx *fiber.Ctx) bool {
	if ctx == nil || ctx.Request() == nil {
		return false
	}

	// ignoring /rk/v1/assets, /rk/v1/tv and /sw/ path while logging since these are internal APIs.
	if strings.HasPrefix(ctx.Path(), "/rk/v1/assets") ||
		strings.HasPrefix(ctx.Path(), "/rk/v1/tv") ||
		strings.HasPrefix(ctx.Path(), "/sw/") {
		return false
	}

	return true
}
