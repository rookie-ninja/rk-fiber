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
	"github.com/rookie-ninja/rk-fiber/boot"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
)

func main() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/csrf/boot.yaml")

	// Bootstrap fiber entry from boot config
	res := rkfiber.RegisterFiberEntriesWithConfig("example/boot/csrf/boot.yaml")

	// Bootstrap echo entry
	res["greeter"].Bootstrap(context.Background())

	// Register GET and POST method of /rk/v1/greeter
	entry := res["greeter"].(*rkfiber.FiberEntry)
	entry.App.Get("/rk/v1/greeter", Greeter)
	entry.App.Post("/rk/v1/greeter", Greeter)

	// This is required!!!
	entry.RefreshFiberRoutes()

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt echo entry
	res["greeter"].Interrupt(context.Background())
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
		Message: fmt.Sprintf("CSRF token:%v", rkfiberctx.GetCsrfToken(ctx)),
	})
}
