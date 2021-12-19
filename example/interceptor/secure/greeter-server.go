// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"github.com/rookie-ninja/rk-fiber/interceptor/secure"
	"net/http"
)

// In this example, we will start a new fiber server with secure interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter.
func main() {
	// ******************************************************
	// ********** Override App name and version *************
	// ******************************************************
	//
	// rkentry.GlobalAppCtx.GetAppInfoEntry().AppName = "demo-app"
	// rkentry.GlobalAppCtx.GetAppInfoEntry().Version = "demo-version"

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []fiber.Handler{
		rkfibersec.Interceptor(
			// Required, entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkfibersec.WithEntryNameAndType("greeter", "fiber"),
			//
			// X-XSS-Protection header value.
			// Optional. Default value "1; mode=block".
			//rkfibersec.WithXSSProtection("my-value"),
			//
			// X-Content-Type-Options header value.
			// Optional. Default value "nosniff".
			//rkfibersec.WithContentTypeNosniff("my-value"),
			//
			// X-Frame-Options header value.
			// Optional. Default value "SAMEORIGIN".
			//rkfibersec.WithXFrameOptions("my-value"),
			//
			// Optional, Strict-Transport-Security header value.
			//rkfibersec.WithHSTSMaxAge(1),
			//
			// Optional, excluding subdomains of HSTS, default is false
			//rkfibersec.WithHSTSExcludeSubdomains(true),
			//
			// Optional, enabling HSTS preload, default is false
			//rkfibersec.WithHSTSPreloadEnabled(true),
			//
			// Content-Security-Policy header value.
			// Optional. Default value "".
			//rkfibersec.WithContentSecurityPolicy("my-value"),
			//
			// Content-Security-Policy-Report-Only header value.
			// Optional. Default value false.
			//rkfibersec.WithCSPReportOnly(true),
			//
			// Referrer-Policy header value.
			// Optional. Default value "".
			//rkfibersec.WithReferrerPolicy("my-value"),
			//
			// Ignoring path prefix.
			//rkfibersec.WithIgnorePrefix("/rk/v1"),
		),
	}

	// 1: Create fiber server
	server := startGreeterServer(interceptors...)
	defer server.Shutdown()

	// 2: Wait for ctrl-C to shutdown server
	rkentry.GlobalAppCtx.WaitForShutdownSig()
}

// Start fiber server.
func startGreeterServer(interceptors ...fiber.Handler) *fiber.App {
	app := fiber.New()
	for _, v := range interceptors {
		app.Use(v)
	}
	app.Get("/rk/v1/greeter", Greeter)

	go func() {
		if err := app.Listen(":8080"); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	return app
}

// GreeterResponse Response of Greeter.
type GreeterResponse struct {
	Message string
}

// Greeter Handler.
func Greeter(ctx *fiber.Ctx) error {
	// ******************************************
	// ********** rpc-scoped logger *************
	// ******************************************
	rkfiberctx.GetLogger(ctx).Info("Received request from client.")

	return ctx.JSON(&GreeterResponse{
		Message: "Received message!",
	})
}
