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
	"github.com/rookie-ninja/rk-fiber/interceptor/jwt"
	"net/http"
)

// In this example, we will start a new fiber server with jwt interceptor enabled.
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
		//rkfiberlog.Interceptor(),
		rkfiberjwt.Interceptor(
			// Required, entry name and entry type will be used for distinguishing interceptors. Recommended.
			//rkechojwt.WithEntryNameAndType("greeter", "fiber"),
			//
			// Required, provide signing key.
			rkfiberjwt.WithSigningKey([]byte("my-secret")),
			//
			// Optional, provide skipper function
			//rkfiberjwt.WithSkipper(func(e *fiber.Ctx) bool {
			//	return true
			//}),
			//
			// Optional, provide token parse function, default one will be assigned.
			//rkfiberjwt.WithParseTokenFunc(func(auth string, ctx *fiber.Ctx) (*jwt.Token, error) {
			//	return nil, nil
			//}),
			//
			// Optional, provide key function, default one will be assigned.
			//rkfiberjwt.WithKeyFunc(func(token *jwt.Token) (interface{}, error) {
			//	return nil, nil
			//}),
			//
			// Optional, default is Bearer
			//rkfiberjwt.WithAuthScheme("Bearer"),
			//
			// Optional
			//rkfiberjwt.WithTokenLookup("header:my-jwt-header-key"),
			//
			// Optional, default is HS256
			//rkfiberjwt.WithSigningAlgorithm(rkfiberjwt.AlgorithmHS256),
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

	ctx.JSON(&GreeterResponse{
		Message: fmt.Sprintf("Is token valid:%v!", rkfiberctx.GetJwtToken(ctx).Valid),
	})

	return nil
}
