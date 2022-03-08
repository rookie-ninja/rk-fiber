// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiberauth

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/auth"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_WithIgnoringPath(t *testing.T) {
	defer assertNotPanic(t)

	app := fiber.New()

	handler := Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidauth.WithBasicAuth("ut-realm", "user:pass"),
		rkmidauth.WithApiKeyAuth("ut-api-key"),
		rkmidauth.WithPathToIgnore("/ut-ignore-path"))

	app.Use(handler)
	app.Get("/ut-ignore-path", func(ctx *fiber.Ctx) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ut-ignore-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMiddleware_WithBasicAuth_Invalid(t *testing.T) {
	defer assertNotPanic(t)

	app := fiber.New()

	handler := Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidauth.WithBasicAuth("ut-realm", "user:pass"))

	app.Use(handler)
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderAuthorization, "invalid")
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMiddleware_WithBasicAuth_InvalidBasicAuth(t *testing.T) {
	defer assertNotPanic(t)

	app := fiber.New()

	handler := Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidauth.WithBasicAuth("ut-realm", "user:pass"))

	app.Use(handler)
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderAuthorization, fmt.Sprintf("%s invalid", "Basic"))
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMiddleware_WithApiKey_Invalid(t *testing.T) {
	defer assertNotPanic(t)

	app := fiber.New()

	handler := Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidauth.WithApiKeyAuth("ut-api-key"))

	app.Use(handler)
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderApiKey, "invalid")
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMiddleware_MissingAuth(t *testing.T) {
	defer assertNotPanic(t)

	app := fiber.New()

	handler := Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidauth.WithApiKeyAuth("ut-api-key"))

	app.Use(handler)
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMiddleware_HappyCase(t *testing.T) {
	defer assertNotPanic(t)

	app := fiber.New()

	handler := Middleware(
		rkmidauth.WithEntryNameAndType("ut-entry", "ut-type"),
		//WithBasicAuth("ut-realm", "user:pass"),
		rkmidauth.WithApiKeyAuth("ut-api-key"))

	app.Use(handler)
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderApiKey, "ut-api-key")
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
