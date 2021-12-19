// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"github.com/rookie-ninja/rk-fiber/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-prom"
	"net/http"
)

// In this example, we will start a new fiber server with metrics interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// Override app name which would replace namespace value in prometheus.
	// rkentry.GlobalAppCtx.GetAppInfoEntry().AppName = "newApp"

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []fiber.Handler{
		rkfibermetrics.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			rkfibermetrics.WithEntryNameAndType("greeter", "fiber"),
			//
			// Provide new prometheus registerer.
			// Default value is prometheus.DefaultRegisterer
			//rkfibermetrics.WithRegisterer(prometheus.NewRegistry()),
		),
	}

	// 1: Start prometheus client
	// By default, we will start prometheus client at localhost:1608/metrics
	promEntry := rkprom.RegisterPromEntry()
	promEntry.Bootstrap(context.Background())
	defer promEntry.Interrupt(context.Background())

	// 2: Create fiber server
	server := startGreeterServer(interceptors...)
	defer server.Shutdown()

	// 3: Wait for ctrl-C to shutdown server
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
