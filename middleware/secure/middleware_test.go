// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfibersec

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/v2/middleware/secure"
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

	// without options
	app := fiber.New()
	app.Use(Middleware())
	app.Get("/ut-path", userHandler)
	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	containsHeader(t, resp,
		fiber.HeaderXXSSProtection,
		fiber.HeaderXContentTypeOptions,
		fiber.HeaderXFrameOptions)

	// with options
	app = fiber.New()
	app.Use(Middleware(
		rkmidsec.WithXSSProtection("ut-xss"),
		rkmidsec.WithContentTypeNosniff("ut-sniff"),
		rkmidsec.WithXFrameOptions("ut-frame"),
		rkmidsec.WithHSTSMaxAge(10),
		rkmidsec.WithHSTSExcludeSubdomains(true),
		rkmidsec.WithHSTSPreloadEnabled(true),
		rkmidsec.WithContentSecurityPolicy("ut-policy"),
		rkmidsec.WithCSPReportOnly(true),
		rkmidsec.WithReferrerPolicy("ut-ref"),
		rkmidsec.WithPathToIgnore("ut-prefix")))
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
