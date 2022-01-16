// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiberlimit

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	rkmidlimit "github.com/rookie-ninja/rk-entry/middleware/ratelimit"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInterceptor_WithoutOptions(t *testing.T) {
	app := fiber.New()
	app.Use(Interceptor())
	app.Get("/ut-path", func(*fiber.Ctx) error {
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInterceptor_WithTokenBucket(t *testing.T) {
	app := fiber.New()
	app.Use(Interceptor(
		rkmidlimit.WithAlgorithm(rkmidlimit.TokenBucket),
		rkmidlimit.WithReqPerSec(1),
		rkmidlimit.WithReqPerSecByPath("ut-path", 1)))
	app.Get("/ut-path", func(*fiber.Ctx) error {
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInterceptor_WithLeakyBucket(t *testing.T) {
	app := fiber.New()
	app.Use(Interceptor(
		rkmidlimit.WithAlgorithm(rkmidlimit.LeakyBucket),
		rkmidlimit.WithReqPerSec(1),
		rkmidlimit.WithReqPerSecByPath("ut-path", 1)))
	app.Get("/ut-path", func(*fiber.Ctx) error {
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestInterceptor_WithUserLimiter(t *testing.T) {
	app := fiber.New()
	app.Use(Interceptor(
		rkmidlimit.WithGlobalLimiter(func() error {
			return fmt.Errorf("ut-error")
		}),
		rkmidlimit.WithLimiterByPath("/ut-path", func() error {
			return fmt.Errorf("ut-error")
		})))
	app.Get("/ut-path", func(*fiber.Ctx) error {
		return nil
	})
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
