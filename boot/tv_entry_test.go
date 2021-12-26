// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiber

import (
	"context"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTvEntry(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	assert.Equal(t, TvEntryNameDefault, entry.GetName())
	assert.Equal(t, TvEntryType, entry.GetType())
	assert.Equal(t, TvEntryDescription, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.Nil(t, entry.UnmarshalJSON(nil))
}

func TestTvEntry_Bootstrap(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	ctx := context.Background()
	entry.Bootstrap(ctx)
}

func TestTvEntry_Interrupt(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))

	ctx := context.Background()
	entry.Interrupt(ctx)
}

func TestTvEntry_TV(t *testing.T) {
	entry := NewTvEntry(
		WithEventLoggerEntryTv(rkentry.NoopEventLoggerEntry()),
		WithZapLoggerEntryTv(rkentry.NoopZapLoggerEntry()))
	entry.Bootstrap(context.TODO())

	defer assertNotPanic(t)

	app, _ := newCtx()
	app.Get("/rk/v1/tv/*", entry.TV)

	// With nil context
	entry.TV(nil)

	// With all paths
	req := httptest.NewRequest(http.MethodGet, "/rk/v1/tv/", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// apis
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/apis", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// entries
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/entries", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// configs
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/configs", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// certs
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/certs", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// os
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/os", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// env
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/env", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// prometheus
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/prometheus", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// logs
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/logs", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// deps
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/deps", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// license
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/license", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// info
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/info", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// git
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/git", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))

	// unknown
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/tv/unknown", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, string(bytes))
}
