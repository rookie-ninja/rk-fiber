// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfibersec is a middleware of fiber framework for adding secure headers in RPC response
package rkfibersec

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-fiber/interceptor"
)

// Interceptor Add security interceptors.
//
// Mainly copied from bellow.
// https://github.com/labstack/echo/blob/master/middleware/secure.go
func Interceptor(opts ...Option) fiber.Handler {
	set := newOptionSet(opts...)

	return func(ctx *fiber.Ctx) error {
		ctx.SetUserContext(context.WithValue(ctx.UserContext(), rkfiberinter.RpcEntryNameKey, set.EntryName))

		if set.Skipper(ctx) {
			return ctx.Next()
		}

		req := ctx.Request()
		res := ctx.Response()

		// Add X-XSS-Protection header
		if set.XSSProtection != "" {
			res.Header.Set(fiber.HeaderXXSSProtection, set.XSSProtection)
		}

		// Add X-Content-Type-Options header
		if set.ContentTypeNosniff != "" {
			res.Header.Set(fiber.HeaderXContentTypeOptions, set.ContentTypeNosniff)
		}

		// Add X-Frame-Options header
		if set.XFrameOptions != "" {
			res.Header.Set(fiber.HeaderXFrameOptions, set.XFrameOptions)
		}

		// Add Strict-Transport-Security header
		if (ctx.Context().IsTLS() || (string(req.Header.Peek(fiber.HeaderXForwardedProto)) == "https")) && set.HSTSMaxAge != 0 {
			subdomains := ""
			if !set.HSTSExcludeSubdomains {
				subdomains = "; includeSubdomains"
			}
			if set.HSTSPreloadEnabled {
				subdomains = fmt.Sprintf("%s; preload", subdomains)
			}
			res.Header.Set(fiber.HeaderStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", set.HSTSMaxAge, subdomains))
		}

		// Add Content-Security-Policy-Report-Only or Content-Security-Policy header
		if set.ContentSecurityPolicy != "" {
			if set.CSPReportOnly {
				res.Header.Set(fiber.HeaderContentSecurityPolicyReportOnly, set.ContentSecurityPolicy)
			} else {
				res.Header.Set(fiber.HeaderContentSecurityPolicy, set.ContentSecurityPolicy)
			}
		}

		// Add Referrer-Policy header
		if set.ReferrerPolicy != "" {
			res.Header.Set(fiber.HeaderReferrerPolicy, set.ReferrerPolicy)
		}

		return ctx.Next()
	}
}
