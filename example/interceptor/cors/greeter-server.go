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
	"github.com/rookie-ninja/rk-fiber/interceptor/cors"
	"net/http"
)

// In this example, we will start a new fiber server with cors interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
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
		rkfibercors.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkfibercors.WithEntryNameAndType("greeter", "fiber"),
			// Provide skipper function
			// rkfibercors.WithSkipper(func(e *fiber.Ctx) bool {
			//     return false
			// }),
			// Bellow section is for CORS policy.
			// Please refer https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS for details.
			// Provide allowed origins
			rkfibercors.WithAllowOrigins("http://localhost:8080"),
			// Whether to allow credentials
			// rkfibercors.WithAllowCredentials(true),
			// Provide expose headers
			// rkfibercors.WithExposeHeaders(""),
			// Provide max age
			// rkfibercors.WithMaxAge(1),
			// Provide allowed headers
			// rkfibercors.WithAllowHeaders(""),
			// Provide allowed headers
			// rkfibercors.WithAllowMethods(""),
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
	//
	// RequestId will be printed if enabled by bellow codes.
	// 1: Enable rkfibermeta.Interceptor() in server side.
	// 2: rkfiberctx.SetHeaderToClient(ctx, rkfiberctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkfiberctx.GetLogger(ctx).Info("Received request from client.")

	return ctx.JSON(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}
