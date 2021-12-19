// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/interceptor/log/zap"
	"github.com/rookie-ninja/rk-fiber/interceptor/panic"
	"net/http"
)

// In this example, we will start a new fiber server with panic interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []fiber.Handler{
		rkfiberlog.Interceptor(),
		rkfiberpanic.Interceptor(),
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
	// All bellow panic case should return same error response.
	// {"error":{"code":500,"status":"Internal Server Error","message":"Panic manually!","details":[]}}

	// Panic interceptor will wrap error with standard RK style error.
	// Please refer to rkerror.ErrorResp.
	// panic(errors.New("Panic manually!"))

	// Please refer to rkerror.ErrorResp.
	// panic(rkerror.FromError(errors.New("Panic manually!")))

	// Please refer to rkerror.ErrorResp.
	panic("Panic manually!")

	return ctx.JSON(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})
}
