// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"github.com/rookie-ninja/rk-fiber/interceptor/csrf"
	"net/http"
)

// In this example, we will start a new fiber server with csrf interceptor enabled.
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
		rkfibercsrf.Interceptor(
			// Required, entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkfibercsrf.WithEntryNameAndType("greeter", "fiber"),
			//
			// Optional, provide skipper function
			//rkfibercsrf.WithSkipper(func(e *fiber.Ctx) bool {
			//	return true
			//}),
			//
			// WithTokenLength the length of the generated token.
			// Optional. Default value 32.
			//rkfibercsrf.WithTokenLength(10),
			//
			// WithTokenLookup a string in the form of "<source>:<key>" that is used
			// to extract token from the request.
			// Optional. Default value "header:X-CSRF-Token".
			// Possible values:
			// - "header:<name>"
			// - "query:<name>"
			// Optional. Default value "header:X-CSRF-Token".
			//rkfibercsrf.WithTokenLookup("header:X-CSRF-Token"),
			//
			// WithCookieName provide name of the CSRF cookie. This cookie will store CSRF token.
			// Optional. Default value "csrf".
			//rkfibercsrf.WithCookieName("csrf"),
			//
			// WithCookieDomain provide domain of the CSRF cookie.
			// Optional. Default value "".
			//rkfibercsrf.WithCookieDomain(""),
			//
			// WithCookiePath provide path of the CSRF cookie.
			// Optional. Default value "".
			//rkfibercsrf.WithCookiePath(""),
			//
			// WithCookieMaxAge provide max age (in seconds) of the CSRF cookie.
			// Optional. Default value 86400 (24hr).
			//rkfibercsrf.WithCookieMaxAge(10),
			//
			// WithCookieHTTPOnly indicates if CSRF cookie is HTTP only.
			// Optional. Default value false.
			//rkfibercsrf.WithCookieHTTPOnly(false),
			//
			// WithCookieSameSite indicates SameSite mode of the CSRF cookie.
			// Optional. Default value SameSiteDefaultMode.
			//rkfibercsrf.WithCookieSameSite(http.SameSiteStrictMode),
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
	app.Post("/rk/v1/greeter", Greeter)

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
	//
	// RequestId will be printed if enabled by bellow codes.
	// 1: Enable rkfibermeta.Interceptor() in server side.
	// 2: rkfiberctx.SetHeaderToClient(ctx, rkfiberctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkfiberctx.GetLogger(ctx).Info("Received request from client.")

	return ctx.JSON(&GreeterResponse{
		Message: fmt.Sprintf("CSRF token:%v", rkfiberctx.GetCsrfToken(ctx)),
	})
}
