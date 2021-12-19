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
	"github.com/rookie-ninja/rk-fiber/interceptor/ratelimit"
	"net/http"
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
		rkfiberlog.Interceptor(),
		rkfiberlimit.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		//rkfiberlimit.WithEntryNameAndType("greeter", "fiber"),
		//
		// Provide algorithm, rkfiberlimit.LeakyBucket and rkfiberlimit.TokenBucket was available, default is TokenBucket.
		//rkfiberlimit.WithAlgorithm(rkfiberlimit.LeakyBucket),
		//
		// Provide request per second, if provide value of zero, then no requests will be pass through and user will receive an error with
		// resource exhausted.
		//rkfiberlimit.WithReqPerSec(10),
		//
		// Provide request per second with path name.
		// The name should be full path name. if provide value of zero,
		// then no requests will be pass through and user will receive an error with resource exhausted.
		//rkfiberlimit.WithReqPerSecByPath("/rk/v1/greeter", 10),
		//
		// Provide user function of limiter. Returns error if you want to limit the request.
		// Please do not try to set response code since it will be overridden by middleware.
		//rkfiberlimit.WithGlobalLimiter(func(ctx *fiber.Ctx) error {
		//	return fmt.Errorf("limited by custom limiter")
		//}),
		//
		// Provide user function of limiter by path name.
		// The name should be full path name.
		//rkfiberlimit.WithLimiterByPath("/rk/v1/greeter", func(ctx *fiber.Ctx) error {
		//	 return nil
		//}),
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
	// 2: rkfiberctx.AddHeaderToClient(ctx, rkfiberctx.RequestIdKey, rkcommon.GenerateRequestId())
	//
	rkfiberctx.GetLogger(ctx).Info("Received request from client.")

	// Set request id with X-Request-Id to outgoing headers.
	// rkfiberctx.SetHeaderToClient(ctx, rkfiberctx.RequestIdKey, "this-is-my-request-id-overridden")

	return ctx.JSON(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}
