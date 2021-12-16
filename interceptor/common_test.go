// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiberinter

import (
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"net"
	"testing"
)

func newCtx(uri string) (*fiber.Ctx, *fasthttp.RequestCtx) {
	app := fiber.New()
	reqCtx := &fasthttp.RequestCtx{
		Request: fasthttp.Request{},
	}
	reqCtx.Request.SetRequestURI(uri)
	ctx := app.AcquireCtx(reqCtx)
	return ctx, reqCtx
}

func TestGetRemoteAddressSet(t *testing.T) {
	// With nil context
	ip, port := GetRemoteAddressSet(nil)
	assert.Equal(t, "0.0.0.0", ip)
	assert.Equal(t, "0", port)

	// With nil Request in context
	ctx, _ := newCtx("")
	assert.Equal(t, "0.0.0.0", ip)
	assert.Equal(t, "0", port)

	// With x-forwarded-for equals to ::1
	remoteAddr, err := net.ResolveTCPAddr("", "1.1.1.1:1")
	assert.Nil(t, err)

	ctx, reqCtx := newCtx("")
	reqCtx.SetRemoteAddr(remoteAddr)
	ctx.Request().Header.Set("x-forwarded-for", "::1")
	ip, port = GetRemoteAddressSet(ctx)

	assert.Equal(t, "localhost", ip)
	assert.Equal(t, "1", port)

	// Happy case
	remoteAddr, err = net.ResolveTCPAddr("", "1.1.1.1:1")
	assert.Nil(t, err)

	ctx, reqCtx = newCtx("")
	reqCtx.SetRemoteAddr(remoteAddr)
	ip, port = GetRemoteAddressSet(ctx)

	assert.Equal(t, "1.1.1.1", ip)
	assert.Equal(t, "1", port)
}

func TestShouldLog(t *testing.T) {
	// With nil context
	assert.False(t, ShouldLog(nil))

	// With ignoring path
	ctx, _ := newCtx("/rk/v1/assets")
	assert.False(t, ShouldLog(ctx))

	ctx, _ = newCtx("/rk/v1/tv")
	assert.False(t, ShouldLog(ctx))

	ctx, _ = newCtx("/sw/")
	assert.False(t, ShouldLog(ctx))

	// Expect true
	ctx, _ = newCtx("ut-path")
	assert.True(t, ShouldLog(ctx))
}
