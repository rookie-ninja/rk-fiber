// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiberpanic

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/v2/middleware/panic"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	defer assertNotPanic(t)

	app := fiber.New()

	handler := Middleware(
		rkmidpanic.WithEntryNameAndType("ut-entry", "ut-type"))

	app.Use(handler)
	app.Get("/ut-path", func(ctx *fiber.Ctx) error {
		panic(errors.New("ut panic"))
	})

	req := httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
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
