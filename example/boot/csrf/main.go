// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-fiber/boot"
	"github.com/rookie-ninja/rk-fiber/middleware/context"
)

//go:embed boot.yaml
var boot []byte

func main() {
	// Bootstrap preload entries
	rkentry.BootstrapBuiltInEntryFromYAML(boot)
	rkentry.BootstrapPluginEntryFromYAML(boot)

	// Bootstrap gin entry from boot config
	res := rkfiber.RegisterFiberEntryYAML(boot)

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
