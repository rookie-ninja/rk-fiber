// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiberjwt

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
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
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// with parse token error
	parseTokenErrFunc := func(auth string, c *fiber.Ctx) (*jwt.Token, error) {
		return nil, errors.New("ut-error")
	}

	app = fiber.New()
	app.Use(Interceptor(
		WithParseTokenFunc(parseTokenErrFunc)))
	app.Get("/ut-path", userHandler)
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(headerAuthorization, strings.Join([]string{"Bearer", "ut-auth"}, " "))
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// happy case
	parseTokenErrFunc = func(auth string, c *fiber.Ctx) (*jwt.Token, error) {
		return &jwt.Token{}, nil
	}

	app = fiber.New()
	app.Use(Interceptor(
		WithParseTokenFunc(parseTokenErrFunc)))
	app.Get("/ut-path", userHandler)
	req = httptest.NewRequest(http.MethodGet, "/ut-path", nil)
	req.Header.Set(headerAuthorization, strings.Join([]string{"Bearer", "ut-auth"}, " "))
	resp, err = app.Test(req)
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
