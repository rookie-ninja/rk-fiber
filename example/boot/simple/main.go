// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/boot"
	"net/http"
)

func main() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/simple/boot.yaml")

	// Bootstrap fiber entry from boot config
	res := rkfiber.RegisterFiberEntriesWithConfig("example/boot/simple/boot.yaml")

	// Get rkfiber.FiberEntry
	fiberEntry := res["greeter"].(*rkfiber.FiberEntry)

	// Bootstrap fiber entry
	fiberEntry.Bootstrap(context.Background())

	// Routes must be registered after Bootstrap()
	fiberEntry.App.Get("/v1/greeter", func(ctx *fiber.Ctx) error {
		ctx.Response().SetStatusCode(http.StatusOK)
		return ctx.JSON(map[string]string{
			"message": "Hello!",
		})
	})

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt fiber entry
	fiberEntry.Interrupt(context.Background())
}
