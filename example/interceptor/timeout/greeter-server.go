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
	"github.com/rookie-ninja/rk-fiber/interceptor/log/zap"
	"github.com/rookie-ninja/rk-fiber/interceptor/panic"
	"github.com/rookie-ninja/rk-fiber/interceptor/timeout"
	"net/http"
	"time"
)

// In this example, we will start a new fiber server with rate limit interceptor enabled.
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
		rkfiberpanic.Interceptor(),
		rkfiberlog.Interceptor(),
		rkfibertimeout.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkfibertimeout.WithEntryNameAndType("greeter", "fiber"),
		//
		// Provide timeout and response handler, a default one would be assigned with http.StatusRequestTimeout
		// This option impact all routes
		//rkfibertimeout.WithTimeoutAndResp(time.Second, nil),
		//
		// Provide timeout and response handler by path, a default one would be assigned with http.StatusRequestTimeout
		//rkfibertimeout.WithTimeoutAndRespByPath("/rk/v1/healthy", time.Second, nil),
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

	// Set request id with X-Request-Id to outgoing headers.
	// rkfiberctx.SetHeaderToClient(ctx, rkfiberctx.RequestIdKey, "this-is-my-request-id-overridden")

	// Sleep for 5 seconds waiting to be timed out by interceptor
	time.Sleep(10 * time.Second)

	return ctx.JSON(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}
