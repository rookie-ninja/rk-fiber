// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.
package rkfibercors

import (
	"github.com/gofiber/fiber/v2"
	rkmidcors "github.com/rookie-ninja/rk-entry/middleware/cors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const originHeaderValue = "http://ut-origin"

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// with empty option, all request will be passed
	app := fiber.New()
	app.Use(Interceptor())
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(fiber.HeaderOrigin, originHeaderValue)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// match 1.1
	app = fiber.New()
	app.Use(Interceptor())
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// match 1.2
	app = fiber.New()
	app.Use(Interceptor())
	app.Options("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req = httptest.NewRequest(http.MethodOptions, "/ut-path", nil)
	req.Header.Set(fiber.HeaderOrigin, originHeaderValue)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// match 2
	app = fiber.New()
	app.Use(Interceptor(rkmidcors.WithAllowOrigins("http://do-not-pass-through")))
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(fiber.HeaderOrigin, originHeaderValue)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// match 3
	app = fiber.New()
	app.Use(Interceptor())
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(fiber.HeaderOrigin, originHeaderValue)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, originHeaderValue, resp.Header.Get(fiber.HeaderAccessControlAllowOrigin))

	// match 3.1
	app = fiber.New()
	app.Use(Interceptor(rkmidcors.WithAllowCredentials(true)))
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(fiber.HeaderOrigin, originHeaderValue)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, originHeaderValue, resp.Header.Get(fiber.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", resp.Header.Get(fiber.HeaderAccessControlAllowCredentials))

	// match 3.2
	app = fiber.New()
	app.Use(Interceptor(
		rkmidcors.WithAllowCredentials(true),
		rkmidcors.WithExposeHeaders("expose")))
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(fiber.HeaderOrigin, originHeaderValue)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, originHeaderValue, resp.Header.Get(fiber.HeaderAccessControlAllowOrigin))
	assert.Equal(t, "true", resp.Header.Get(fiber.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "expose", resp.Header.Get(fiber.HeaderAccessControlExposeHeaders))

	// match 4
	app = fiber.New()
	app.Use(Interceptor())
	app.Options("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req = httptest.NewRequest(http.MethodOptions, "/ut-path", nil)
	req.Header.Set(fiber.HeaderOrigin, originHeaderValue)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		fiber.HeaderAccessControlRequestMethod,
		fiber.HeaderAccessControlRequestHeaders,
	}, resp.Header.Values(fiber.HeaderVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(fiber.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(fiber.HeaderAccessControlAllowMethods))

	// match 4.1
	app = fiber.New()
	app.Use(Interceptor(rkmidcors.WithAllowCredentials(true)))
	app.Options("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req = httptest.NewRequest(http.MethodOptions, "/ut-path", nil)
	req.Header.Set(fiber.HeaderOrigin, originHeaderValue)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		fiber.HeaderAccessControlRequestMethod,
		fiber.HeaderAccessControlRequestHeaders,
	}, resp.Header.Values(fiber.HeaderVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(fiber.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(fiber.HeaderAccessControlAllowMethods))
	assert.Equal(t, "true", resp.Header.Get(fiber.HeaderAccessControlAllowCredentials))

	// match 4.2
	app = fiber.New()
	app.Use(Interceptor(rkmidcors.WithAllowHeaders("ut-header")))
	app.Options("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req = httptest.NewRequest(http.MethodOptions, "/ut-path", nil)
	req.Header.Set(fiber.HeaderOrigin, originHeaderValue)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		fiber.HeaderAccessControlRequestMethod,
		fiber.HeaderAccessControlRequestHeaders,
	}, resp.Header.Values(fiber.HeaderVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(fiber.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(fiber.HeaderAccessControlAllowMethods))
	assert.Equal(t, "ut-header", resp.Header.Get(fiber.HeaderAccessControlAllowHeaders))

	// match 4.3
	app = fiber.New()
	app.Use(Interceptor(rkmidcors.WithMaxAge(1)))
	app.Options("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})
	req = httptest.NewRequest(http.MethodOptions, "/ut-path", nil)
	req.Header.Set(fiber.HeaderOrigin, originHeaderValue)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{
		fiber.HeaderAccessControlRequestMethod,
		fiber.HeaderAccessControlRequestHeaders,
	}, resp.Header.Values(fiber.HeaderVary))
	assert.Equal(t, originHeaderValue, resp.Header.Get(fiber.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, resp.Header.Get(fiber.HeaderAccessControlAllowMethods))
	assert.Equal(t, "1", resp.Header.Get(fiber.HeaderAccessControlMaxAge))
}

func assertNotPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, false)
	} else {
		// This should never be called in case of a bug
		assert.True(t, true)
	}
}
