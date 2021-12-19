// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfiber

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
)

func TestNewStaticFileHandlerEntry(t *testing.T) {
	// without options
	entry := NewStaticFileHandlerEntry()
	assert.NotNil(t, entry)
	assert.NotNil(t, entry.ZapLoggerEntry)
	assert.NotNil(t, entry.EventLoggerEntry)
	assert.Equal(t, "/rk/v1/static/", entry.Path)
	assert.NotNil(t, entry.Fs)
	assert.NotNil(t, entry.Template)

	// with options
	utFs := http.Dir("")
	utPath := "/ut-path/"
	utZapLogger := rkentry.NoopZapLoggerEntry()
	utEventLogger := rkentry.NoopEventLoggerEntry()
	utName := "ut-entry"

	entry = NewStaticFileHandlerEntry(
		WithPathStatic(utPath),
		WithEventLoggerEntryStatic(utEventLogger),
		WithZapLoggerEntryStatic(utZapLogger),
		WithNameStatic(utName),
		WithFileSystemStatic(utFs))

	assert.NotNil(t, entry)
	assert.Equal(t, utZapLogger, entry.ZapLoggerEntry)
	assert.Equal(t, utEventLogger, entry.EventLoggerEntry)
	assert.Equal(t, utPath, entry.Path)
	assert.Equal(t, utFs, entry.Fs)
	assert.NotNil(t, entry.Template)
	assert.Equal(t, utName, entry.EntryName)
}

func TestStaticFileHandlerEntry_Bootstrap(t *testing.T) {
	defer assertNotPanic(t)

	// without eventId in context
	entry := NewStaticFileHandlerEntry()
	entry.Bootstrap(context.TODO())

	// with eventId in context
	entry.Bootstrap(context.WithValue(context.TODO(), bootstrapEventIdKey, "ut-event-id"))
}

func TestStaticFileHandlerEntry_Interrupt(t *testing.T) {
	defer assertNotPanic(t)

	// without eventId in context
	entry := NewStaticFileHandlerEntry()
	entry.Interrupt(context.TODO())

	// with eventId in context
	entry.Interrupt(context.WithValue(context.TODO(), bootstrapEventIdKey, "ut-event-id"))
}

func TestStaticFileHandlerEntry_EntryFunctions(t *testing.T) {
	entry := NewStaticFileHandlerEntry()
	assert.NotEmpty(t, entry.GetName())
	assert.NotEmpty(t, entry.GetType())
	assert.NotEmpty(t, entry.GetDescription())
	assert.NotEmpty(t, entry.String())
	assert.Nil(t, entry.UnmarshalJSON([]byte{}))
}

func TestStaticFileHandlerEntry_GetFileHandler(t *testing.T) {
	currDir := t.TempDir()
	os.MkdirAll(path.Join(currDir, "ut-dir"), os.ModePerm)
	os.WriteFile(path.Join(currDir, "ut-file"), []byte("ut content"), os.ModePerm)

	entry := NewStaticFileHandlerEntry(
		WithNameStatic("ut-static"),
		WithFileSystemStatic(http.Dir(currDir)))
	entry.Bootstrap(context.TODO())
	handler := entry.GetFileHandler()

	// expect to get list of files
	app := fiber.New()
	app.Get("/rk/v1/static/", handler)
	req := httptest.NewRequest(http.MethodGet, "/rk/v1/static/", nil)
	resp, err := app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ := ioutil.ReadAll(resp.Body)
	assert.Contains(t, string(bytes), "Index of")

	// expect to get files to download
	app = fiber.New()
	app.Get("/rk/v1/static/ut-file", handler)
	req = httptest.NewRequest(http.MethodGet, "/rk/v1/static/ut-file", nil)
	resp, err = app.Test(req)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	bytes, _ = ioutil.ReadAll(resp.Body)
	assert.NotEmpty(t, resp.Header.Get("Content-Disposition"))
	assert.NotEmpty(t, resp.Header.Get("Content-Type"))
	assert.Contains(t, string(bytes), "ut content")
}
