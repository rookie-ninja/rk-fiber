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
	"net/http"
)

// In this example, we will start a new fiber server with log interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []fiber.Handler{
		//rkfibermeta.Interceptor(),
		rkfiberlog.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		// rkfiberlog.WithEntryNameAndType("greeter", "fiber"),
		//
		// Zap logger would be logged as JSON format.
		//rkfiberlog.WithZapLoggerEncoding(rkfiberlog.ENCODING_JSON),
		//
		// Event logger would be logged as JSON format.
		//rkfiberlog.WithEventLoggerEncoding(rkfiberlog.ENCODING_JSON),
		//
		// Zap logger would be logged to specified path.
		//rkfiberlog.WithZapLoggerOutputPaths("logs/server-zap.log"),
		//
		// Event logger would be logged to specified path.
		//rkfiberlog.WithEventLoggerOutputPaths("logs/server-event.log"),
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

// Greeter Handler for greeter.
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

	// *******************************************
	// ********** rpc-scoped event  *************
	// *******************************************
	//
	// Get rkquery.Event which would be printed as soon as request finish.
	// User can call any Add/Set/Get functions on rkquery.Event
	//
	// rkfiberctx.GetEvent(ctx).AddPair("rk-key", "rk-value")

	// *********************************************
	// ********** Get incoming headers *************
	// *********************************************
	//
	// Read headers sent from client.
	//
	//for k, v := range rkfiberctx.GetIncomingHeaders(ctx) {
	//	 fmt.Println(fmt.Sprintf("%s: %s", k, v))
	//}

	// *********************************************************
	// ********** Add headers will send to client **************
	// *********************************************************
	//
	// Send headers to client with this function
	//
	//rkfiberctx.AddHeaderToClient(ctx, "from-server", "value")

	// ***********************************************
	// ********** Get and log request id *************
	// ***********************************************
	//
	// RequestId will be printed on both client and server side.
	//
	//rkfiberctx.SetHeaderToClient(ctx, rkfiberctx.RequestIdKey, rkcommon.GenerateRequestId())

	return ctx.JSON(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}
