// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfibertimeout

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func sleepHandler(ctx *fiber.Ctx) error {
	time.Sleep(2 * time.Second)
	ctx.JSON(http.StatusOK)
	return nil
}

func panicHandler(ctx *fiber.Ctx) error {
	panic(fmt.Errorf("ut panic"))
}

func returnHandler(ctx *fiber.Ctx) error {
	ctx.JSON(http.StatusOK)
	return nil
}

var customResponse = func(ctx *fiber.Ctx) error {
	return fmt.Errorf("custom error")
}

func getFiberApp(userHandler fiber.Handler, interceptor fiber.Handler) *fiber.App {
	app := fiber.New()
	app.Use(interceptor)
	app.Get("/", userHandler)
	return app
}

func TestInterceptor_WithTimeout(t *testing.T) {
	// with global timeout response
	app := getFiberApp(sleepHandler, Interceptor(
		WithTimeoutAndResp(time.Nanosecond, nil)))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)

	// with path
	app = getFiberApp(sleepHandler, Interceptor(
		WithTimeoutAndRespByPath("/", time.Nanosecond, nil)))
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)

	// with custom response
	app = getFiberApp(sleepHandler, Interceptor(
		WithTimeoutAndRespByPath("/", time.Nanosecond, customResponse)))
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bodyBytes), customResponse(nil).Error())
}

func TestInterceptor_WithPanic(t *testing.T) {
	app := getFiberApp(panicHandler, Interceptor(
		WithTimeoutAndResp(time.Minute, nil)))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestInterceptor_HappyCase(t *testing.T) {
	// Let's add two routes /timeout and /happy
	// We expect interceptor acts as the name describes
	app := getFiberApp(panicHandler, Interceptor(
		WithTimeoutAndRespByPath("/timeout", time.Nanosecond, nil),
		WithTimeoutAndRespByPath("/happy", time.Minute, nil)))
	app.Get("/timeout", sleepHandler)
	app.Get("/happy", returnHandler)

	req := httptest.NewRequest(http.MethodGet, "/timeout", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)

	// OK on /happy
	req = httptest.NewRequest(http.MethodGet, "/happy", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func assertPanic(t *testing.T) {
	if r := recover(); r != nil {
		// Expect panic to be called with non nil error
		assert.True(t, true)
	} else {
		// This should never be called in case of a bug
		assert.True(t, false)
	}
}
