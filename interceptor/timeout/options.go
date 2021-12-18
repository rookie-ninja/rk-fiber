// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkfibertimeout

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/rookie-ninja/rk-common/error"
	"github.com/rookie-ninja/rk-fiber/interceptor"
	"github.com/rookie-ninja/rk-fiber/interceptor/context"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

const global = "rk-global"

var (
	defaultResponse = func(ctx *fiber.Ctx) error {
		return nil
	}
	defaultTimeout  = 5 * time.Second
	globalTimeoutRk = &timeoutRk{
		timeout:  defaultTimeout,
		response: defaultResponse,
	}
)

type timeoutRk struct {
	timeout  time.Duration
	response fiber.Handler
}

// Interceptor would distinguish auth set based on.
var optionsMap = make(map[string]*optionSet)

// Create new optionSet with rpc type nad options.
func newOptionSet(opts ...Option) *optionSet {
	set := &optionSet{
		EntryName: rkfiberinter.RpcEntryNameValue,
		EntryType: rkfiberinter.RpcEntryTypeValue,
		timeouts:  make(map[string]*timeoutRk),
	}

	for i := range opts {
		opts[i](set)
	}

	// add global timeout
	set.timeouts[global] = &timeoutRk{
		timeout:  globalTimeoutRk.timeout,
		response: globalTimeoutRk.response,
	}

	if _, ok := optionsMap[set.EntryName]; !ok {
		optionsMap[set.EntryName] = set
	}

	return set
}

// Tick will continue request it rest of handlers.
// If timeout triggered, then return http.StatusRequestTimeout back to client.
//
// Mainly copied from https://github.com/gofiber/fiber/tree/master/middleware/timeout
func (set *optionSet) Tick(ctx *fiber.Ctx) error {
	var err error

	rk := set.getTimeoutRk(ctx.Path())

	event := rkfiberctx.GetEvent(ctx)

	finishChan := make(chan struct{}, 1)
	panicChan := make(chan interface{}, 1)

	go func() {
		defer func() {
			if recv := recover(); recv != nil {
				panicChan <- recv
			}
		}()
		err = ctx.Next()
		finishChan <- struct{}{}
	}()

	select {
	case <-finishChan:
	case recv := <-panicChan:
		rkfiberctx.GetEvent(ctx).SetCounter("panic", 1)
		rkfiberctx.GetEvent(ctx).AddErr(errors.New(fmt.Sprintf("%v", recv)))
		rkfiberctx.GetLogger(ctx).Error(fmt.Sprintf("panic occurs:\n%s", string(debug.Stack())), zap.Any("panic", recv))

		ctx.Response().SetStatusCode(http.StatusInternalServerError)
		res := rkerror.New(
			rkerror.WithHttpCode(http.StatusInternalServerError),
			rkerror.WithMessage(fmt.Sprintf("%v", recv)))
		ctx.JSON(res)
	case <-time.After(rk.timeout):
		event.SetCounter("timeout", 1)
		details := make([]interface{}, 0)
		if err != nil {
			details = append(details, err)
		}
		if customErr := rk.response(ctx); customErr != nil {
			details = append(details, customErr)
		}

		ctx.JSON(rkerror.New(
			rkerror.WithHttpCode(http.StatusRequestTimeout),
			rkerror.WithMessage("Request timed out!"),
			rkerror.WithDetails(details...)))
		ctx.Context().SetStatusCode(http.StatusRequestTimeout)
		return nil
	}

	return err
}

// Get timeout instance with path.
// Global one will be returned if no not found.
func (set *optionSet) getTimeoutRk(path string) *timeoutRk {
	if v, ok := set.timeouts[path]; ok {
		return v
	}

	return set.timeouts[global]
}

// Options which is used while initializing extension interceptor
type optionSet struct {
	EntryName string
	EntryType string
	timeouts  map[string]*timeoutRk
}

// Option if for middleware options while creating middleware
type Option func(*optionSet)

// WithEntryNameAndType provide entry name and entry type.
func WithEntryNameAndType(entryName, entryType string) Option {
	return func(opt *optionSet) {
		opt.EntryName = entryName
		opt.EntryType = entryType
	}
}

// WithTimeoutAndResp Provide global timeout and response handler.
// If response is nil, default globalResponse will be assigned
func WithTimeoutAndResp(timeout time.Duration, resp fiber.Handler) Option {
	return func(set *optionSet) {
		if resp == nil {
			resp = defaultResponse
		}

		if timeout == 0 {
			timeout = defaultTimeout
		}

		globalTimeoutRk.timeout = timeout
		globalTimeoutRk.response = resp
	}
}

// WithTimeoutAndRespByPath Provide timeout and response handler by path.
// If response is nil, default globalResponse will be assigned
func WithTimeoutAndRespByPath(path string, timeout time.Duration, resp fiber.Handler) Option {
	return func(set *optionSet) {
		path = normalisePath(path)

		if resp == nil {
			resp = defaultResponse
		}

		if timeout == 0 {
			timeout = defaultTimeout
		}

		set.timeouts[path] = &timeoutRk{
			timeout:  timeout,
			response: resp,
		}
	}
}

func normalisePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return path
}
