// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfibersec

import (
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var userHandler = func(ctx *fiber.Ctx) error {
	return nil
}

func TestInterceptor(t *testing.T) {
	defer assertNotPanic(t)

	// with skipper
	app := fiber.New()
	app.Use(Interceptor(WithSkipper(func(*fiber.Ctx) bool {
		return true
	})))
	app.Get("/ut-path", userHandler)
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// without options
	app = fiber.New()
	app.Use(Interceptor())
	app.Get("/ut-path", userHandler)
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	containsHeader(t, resp,
		fiber.HeaderXXSSProtection,
		fiber.HeaderXContentTypeOptions,
		fiber.HeaderXFrameOptions)

	// with options
	app = fiber.New()
	app.Use(Interceptor(
		WithXSSProtection("ut-xss"),
		WithContentTypeNosniff("ut-sniff"),
		WithXFrameOptions("ut-frame"),
		WithHSTSMaxAge(10),
		WithHSTSExcludeSubdomains(true),
		WithHSTSPreloadEnabled(true),
		WithContentSecurityPolicy("ut-policy"),
		WithCSPReportOnly(true),
		WithReferrerPolicy("ut-ref"),
		WithIgnorePrefix("ut-prefix")))
	app.Get("/ut-path", userHandler)
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(fiber.HeaderXForwardedProto, "https")
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	containsHeader(t, resp,
		fiber.HeaderXXSSProtection,
		fiber.HeaderXContentTypeOptions,
		fiber.HeaderXFrameOptions,
		fiber.HeaderStrictTransportSecurity,
		fiber.HeaderContentSecurityPolicyReportOnly,
		fiber.HeaderReferrerPolicy)
}

func containsHeader(t *testing.T, resp *http.Response, headers ...string) {
	for _, v := range headers {
		assert.NotEmpty(t, resp.Header.Get(v))
	}
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
