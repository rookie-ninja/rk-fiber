// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/interceptor/auth"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"github.com/rookie-ninja/rk-fiber/interceptor/log/zap"
	"github.com/valyala/fasthttp"
	"net/http"
)

// In this example, we will start a new fiber server with auth interceptor enabled.
// Listen on port of 8080 with GET /rk/v1/greeter?name=<xxx>.
func main() {
	// ********************************************
	// ********** Enable interceptors *************
	// ********************************************
	interceptors := []fiber.Handler{
		rkfiberlog.Interceptor(),
		rkfiberauth.Interceptor(
			// rkfiberauth.WithIgnorePrefix("/rk/v1/greeter"),
			rkfiberauth.WithBasicAuth("", "rk-user:rk-pass"),
			rkfiberauth.WithApiKeyAuth("rk-api-key"),
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
	validateCtx(ctx)

	ctx.JSON(&GreeterResponse{
		Message: fmt.Sprintf("Hello %s!", ctx.Query("name")),
	})

	return nil
}

func validateCtx(ctx *fiber.Ctx) {
	// 1: get incoming headers
	printIndex("[1]: get incoming headers")
	prettyHeader(rkfiberctx.GetIncomingHeaders(ctx))

	// 2: add header to client
	printIndex("[2]: add header to client")
	rkfiberctx.AddHeaderToClient(ctx, "add-key", "add-value")

	// 3: set header to client
	printIndex("[3]: set header to client")
	rkfiberctx.SetHeaderToClient(ctx, "set-key", "set-value")

	// 4: get event
	printIndex("[4]: get event")
	rkfiberctx.GetEvent(ctx).SetCounter("my-counter", 1)

	// 5: get logger
	printIndex("[5]: get logger")
	rkfiberctx.GetLogger(ctx).Info("error msg")

	// 6: get request id
	printIndex("[6]: get request id")
	fmt.Println(rkfiberctx.GetRequestId(ctx))

	// 7: get trace id
	printIndex("[7]: get trace id")
	fmt.Println(rkfiberctx.GetTraceId(ctx))

	// 8: get entry name
	printIndex("[8]: get entry name")
	fmt.Println(rkfiberctx.GetEntryName(ctx))

	// 9: get trace span
	printIndex("[9]: get trace span")
	fmt.Println(rkfiberctx.GetTraceSpan(ctx))

	// 10: get tracer
	printIndex("[10]: get tracer")
	fmt.Println(rkfiberctx.GetTracer(ctx))

	// 11: get trace provider
	printIndex("[11]: get trace provider")
	fmt.Println(rkfiberctx.GetTracerProvider(ctx))

	// 12: get tracer propagator
	printIndex("[12]: get tracer propagator")
	fmt.Println(rkfiberctx.GetTracerPropagator(ctx))

	// 13: inject span
	printIndex("[13]: inject span")
	req := &http.Request{}
	rkfiberctx.InjectSpanToHttpRequest(ctx, req)

	// 14: new trace span
	printIndex("[14]: new trace span")
	fmt.Println(rkfiberctx.NewTraceSpan(ctx, "my-span"))

	// 15: end trace span
	printIndex("[15]: end trace span")
	rkfiberctx.EndTraceSpan(ctx, rkfiberctx.NewTraceSpan(ctx, "my-span"), true)
}

func printIndex(key string) {
	fmt.Println(fmt.Sprintf("%s", key))
}

func prettyHeader(header *fasthttp.RequestHeader) {
	header.VisitAll(func(k, v []byte) {
		fmt.Println(fmt.Sprintf("%s:%s", k, v))
	})
}
