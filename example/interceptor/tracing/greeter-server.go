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
	"github.com/rookie-ninja/rk-fiber/interceptor/tracing/telemetry"
	"net/http"
)

// In this example, we will start a new fiber server with tracing interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ****************************************
	// ********** Create Exporter *************
	// ****************************************

	// Export trace to stdout
	exporter := rkfibertrace.CreateFileExporter("stdout")

	// Export trace to local file system
	//exporter := rkfibertrace.CreateFileExporter("logs/trace.log")

	// Export trace to jaeger agent
	//exporter := rkfibertrace.CreateJaegerExporter(jaeger.WithAgentEndpoint())

	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []fiber.Handler{
		rkfiberlog.Interceptor(),
		rkfibertrace.Interceptor(
			// Entry name and entry type will be used for distinguishing interceptors. Recommended.
			//rkfibertrace.WithEntryNameAndType("greeter", "fiber"),
			//
			// Provide an exporter.
			rkfibertrace.WithExporter(exporter),
			//
			// Provide propagation.TextMapPropagator
			// rkfibertrace.WithPropagator(<propagator>),
			//
			// Provide SpanProcessor
			// rkfibertrace.WithSpanProcessor(<span processor>),
			//
			// Provide TracerProvider
			// rkfibertrace.WithTracerProvider(<trace provider>),
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
	rkfiberctx.GetLogger(ctx).Info("Received request from client.")

	return ctx.JSON(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}
