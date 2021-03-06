// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfibercsrf

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/csrf"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var userHandler = func(ctx *fiber.Ctx) error {
	return nil
}

func TestMiddleware(t *testing.T) {
	defer assertNotPanic(t)

	// match 2.1
	app := fiber.New()

	app.Use(Middleware())
	app.Get("/ut-path", userHandler)

	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "_csrf")

	// match 2.2
	app = fiber.New()

	app.Use(Middleware())
	app.Get("/ut-path", userHandler)

	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.AddCookie(&http.Cookie{
		Name:  "_csrf",
		Value: "ut-csrf-token",
	})
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "_csrf")

	// match 3.1
	app = fiber.New()

	app.Use(Middleware())
	app.Get("/ut-path", userHandler)

	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// match 3.2
	app = fiber.New()

	app.Use(Middleware())
	app.Post("/ut-path", userHandler)

	req = httptest.NewRequest(http.MethodPost, "/ut-path", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// match 3.3
	app = fiber.New()

	app.Use(Middleware())
	app.Post("/ut-path", userHandler)

	req = httptest.NewRequest(http.MethodPost, "/ut-path", nil)
	req.Header.Set(rkmid.HeaderXCSRFToken, "ut-csrf-token")
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	// match 4.1
	app = fiber.New()

	app.Use(Middleware(rkmidcsrf.WithCookiePath("ut-path")))
	app.Get("/ut-path", userHandler)

	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "ut-path")

	// match 4.2
	app = fiber.New()

	app.Use(Middleware(rkmidcsrf.WithCookieDomain("ut-domain")))
	app.Get("/ut-path", userHandler)

	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "ut-domain")

	// match 4.3
	app = fiber.New()

	app.Use(Middleware(rkmidcsrf.WithCookieSameSite(http.SameSiteStrictMode)))
	app.Get("/ut-path", userHandler)

	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Set-Cookie"), "Strict")
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
