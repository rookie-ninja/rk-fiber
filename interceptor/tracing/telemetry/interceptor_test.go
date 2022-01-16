// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfibertrace

import (
	"github.com/gofiber/fiber/v2"
	rkmidtrace "github.com/rookie-ninja/rk-entry/middleware/tracing"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInterceptor(t *testing.T) {
	app := fiber.New()

	handler := Interceptor(
		rkmidtrace.WithEntryNameAndType("ut-entry", "ut-type"),
		rkmidtrace.WithExporter(&rkmidtrace.NoopExporter{}))

	app.Use(handler)
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
