// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiber an implementation of rkentry.Entry which could be used start restful server with fiber framework
package rkfiber

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-entry/middleware/auth"
	"github.com/rookie-ninja/rk-entry/middleware/cors"
	"github.com/rookie-ninja/rk-entry/middleware/csrf"
	"github.com/rookie-ninja/rk-entry/middleware/jwt"
	"github.com/rookie-ninja/rk-entry/middleware/log"
	"github.com/rookie-ninja/rk-entry/middleware/meta"
	"github.com/rookie-ninja/rk-entry/middleware/metrics"
	"github.com/rookie-ninja/rk-entry/middleware/panic"
	"github.com/rookie-ninja/rk-entry/middleware/ratelimit"
	"github.com/rookie-ninja/rk-entry/middleware/secure"
	"github.com/rookie-ninja/rk-entry/middleware/timeout"
	"github.com/rookie-ninja/rk-entry/middleware/tracing"
	"github.com/rookie-ninja/rk-fiber/interceptor/auth"
	rkfiberctx "github.com/rookie-ninja/rk-fiber/interceptor/context"
	"github.com/rookie-ninja/rk-fiber/interceptor/cors"
	"github.com/rookie-ninja/rk-fiber/interceptor/csrf"
	"github.com/rookie-ninja/rk-fiber/interceptor/jwt"
	"github.com/rookie-ninja/rk-fiber/interceptor/log/zap"
	"github.com/rookie-ninja/rk-fiber/interceptor/meta"
	"github.com/rookie-ninja/rk-fiber/interceptor/metrics/prom"
	"github.com/rookie-ninja/rk-fiber/interceptor/panic"
	"github.com/rookie-ninja/rk-fiber/interceptor/ratelimit"
	"github.com/rookie-ninja/rk-fiber/interceptor/secure"
	"github.com/rookie-ninja/rk-fiber/interceptor/timeout"
	"github.com/rookie-ninja/rk-fiber/interceptor/tracing/telemetry"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// FiberEntryType type of entry
	FiberEntryType = "FiberEntry"
	// FiberEntryDescription description of entry
	FiberEntryDescription = "Internal RK entry which helps to bootstrap with fiber framework."
)

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap fiber entry automatically from boot config file
func init() {
	rkentry.RegisterEntryRegFunc(RegisterFiberEntriesWithConfig)
}

// BootConfig boot config which is for fiber entry.
type BootConfig struct {
	Fiber []struct {
		Enabled     bool   `yaml:"enabled" json:"enabled"`
		Name        string `yaml:"name" json:"name"`
		Port        uint64 `yaml:"port" json:"port"`
		Description string `yaml:"description" json:"description"`
		Cert        struct {
			Ref string `yaml:"ref" json:"ref"`
		} `yaml:"cert" json:"cert"`
		SW            rkentry.BootConfigSw            `yaml:"sw" json:"sw"`
		CommonService rkentry.BootConfigCommonService `yaml:"commonService" json:"commonService"`
		TV            rkentry.BootConfigTv            `yaml:"tv" json:"tv"`
		Prom          rkentry.BootConfigProm          `yaml:"prom" json:"prom"`
		Static        rkentry.BootConfigStaticHandler `yaml:"static" json:"static"`
		Interceptors  struct {
			LoggingZap       rkmidlog.BootConfig     `yaml:"loggingZap" json:"loggingZap"`
			MetricsProm      rkmidmetrics.BootConfig `yaml:"metricsProm" json:"metricsProm"`
			Auth             rkmidauth.BootConfig    `yaml:"auth" json:"auth"`
			Cors             rkmidcors.BootConfig    `yaml:"cors" json:"cors"`
			Meta             rkmidmeta.BootConfig    `yaml:"meta" json:"meta"`
			Jwt              rkmidjwt.BootConfig     `yaml:"jwt" json:"jwt"`
			Secure           rkmidsec.BootConfig     `yaml:"secure" json:"secure"`
			Csrf             rkmidcsrf.BootConfig    `yaml:"csrf" yaml:"csrf"`
			RateLimit        rkmidlimit.BootConfig   `yaml:"rateLimit" json:"rateLimit"`
			Timeout          rkmidtimeout.BootConfig `yaml:"timeout" json:"timeout"`
			TracingTelemetry rkmidtrace.BootConfig   `yaml:"tracingTelemetry" json:"tracingTelemetry"`
		} `yaml:"interceptors" json:"interceptors"`
		Logger struct {
			ZapLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"zapLogger" json:"zapLogger"`
			EventLogger struct {
				Ref string `yaml:"ref" json:"ref"`
			} `yaml:"eventLogger" json:"eventLogger"`
		} `yaml:"logger" json:"logger"`
	} `yaml:"fiber" json:"fiber"`
}

// FiberEntry implements rkentry.Entry interface.
type FiberEntry struct {
	EntryName          string                          `json:"entryName" yaml:"entryName"`
	EntryType          string                          `json:"entryType" yaml:"entryType"`
	EntryDescription   string                          `json:"-" yaml:"-"`
	ZapLoggerEntry     *rkentry.ZapLoggerEntry         `json:"-" yaml:"-"`
	EventLoggerEntry   *rkentry.EventLoggerEntry       `json:"-" yaml:"-"`
	Port               uint64                          `json:"port" yaml:"port"`
	CertEntry          *rkentry.CertEntry              `json:"-" yaml:"-"`
	SwEntry            *rkentry.SwEntry                `json:"-" yaml:"-"`
	CommonServiceEntry *rkentry.CommonServiceEntry     `json:"-" yaml:"-"`
	App                *fiber.App                      `json:"-" yaml:"-"`
	FiberConfig        *fiber.Config                   `json:"-" yaml:"-"`
	Interceptors       []fiber.Handler                 `json:"-" yaml:"-"`
	PromEntry          *rkentry.PromEntry              `json:"-" yaml:"-"`
	StaticFileEntry    *rkentry.StaticFileHandlerEntry `json:"-" yaml:"-"`
	TvEntry            *rkentry.TvEntry                `json:"-" yaml:"-"`
}

// RegisterFiberEntriesWithConfig register fiber entries with provided config file (Must YAML file).
//
// Currently, support two ways to provide config file path.
// 1: With function parameters
// 2: With command line flag "--rkboot" described in rkcommon.BootConfigPathFlagKey (Will override function parameter if exists)
// Command line flag has high priority which would override function parameter
//
// Error handling:
// Process will shutdown if any errors occur with rkcommon.ShutdownWithError function
//
// Override elements in config file:
// We learned from HELM source code which would override elements in YAML file with "--set" flag followed with comma
// separated key/value pairs.
//
// We are using "--rkset" described in rkcommon.BootConfigOverrideKey in order to distinguish with user flags
// Example of common usage: ./binary_file --rkset "key1=val1,key2=val2"
// Example of nested map:   ./binary_file --rkset "outer.inner.key=val"
// Example of slice:        ./binary_file --rkset "outer[0].key=val"
func RegisterFiberEntriesWithConfig(configFilePath string) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: Decode config map into boot config struct
	config := &BootConfig{}
	rkcommon.UnmarshalBootConfig(configFilePath, config)

	// 2: Init fiber entries with boot config
	for i := range config.Fiber {
		element := config.Fiber[i]
		if !element.Enabled {
			continue
		}

		name := element.Name

		zapLoggerEntry := rkentry.GlobalAppCtx.GetZapLoggerEntry(element.Logger.ZapLogger.Ref)
		if zapLoggerEntry == nil {
			zapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
		}

		eventLoggerEntry := rkentry.GlobalAppCtx.GetEventLoggerEntry(element.Logger.EventLogger.Ref)
		if eventLoggerEntry == nil {
			eventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
		}

		// Register swagger entry
		swEntry := rkentry.RegisterSwEntryWithConfig(&element.SW, element.Name, element.Port,
			zapLoggerEntry, eventLoggerEntry, element.CommonService.Enabled)

		// Register prometheus entry
		promRegistry := prometheus.NewRegistry()
		promEntry := rkentry.RegisterPromEntryWithConfig(&element.Prom, element.Name, element.Port,
			zapLoggerEntry, eventLoggerEntry, promRegistry)

		// Register common service entry
		commonServiceEntry := rkentry.RegisterCommonServiceEntryWithConfig(&element.CommonService, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		// Register TV entry
		tvEntry := rkentry.RegisterTvEntryWithConfig(&element.TV, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		// Register static file handler
		staticEntry := rkentry.RegisterStaticFileHandlerEntryWithConfig(&element.Static, element.Name,
			zapLoggerEntry, eventLoggerEntry)

		inters := make([]fiber.Handler, 0)

		// logging middlewares
		if element.Interceptors.LoggingZap.Enabled {
			inters = append(inters, rkfiberlog.Interceptor(
				rkmidlog.ToOptions(&element.Interceptors.LoggingZap, element.Name, FiberEntryType,
					zapLoggerEntry, eventLoggerEntry)...))
		}

		// metrics middleware
		if element.Interceptors.MetricsProm.Enabled {
			inters = append(inters, rkfibermetrics.Interceptor(
				rkmidmetrics.ToOptions(&element.Interceptors.MetricsProm, element.Name, FiberEntryType,
					promRegistry, rkmidmetrics.LabelerTypeHttp)...))
		}

		// tracing middleware
		if element.Interceptors.TracingTelemetry.Enabled {
			inters = append(inters, rkfibertrace.Interceptor(
				rkmidtrace.ToOptions(&element.Interceptors.TracingTelemetry, element.Name, FiberEntryType)...))
		}

		// jwt middleware
		if element.Interceptors.Jwt.Enabled {
			inters = append(inters, rkfiberjwt.Interceptor(
				rkmidjwt.ToOptions(&element.Interceptors.Jwt, element.Name, FiberEntryType)...))
		}

		// secure middleware
		if element.Interceptors.Secure.Enabled {
			inters = append(inters, rkfibersec.Interceptor(
				rkmidsec.ToOptions(&element.Interceptors.Secure, element.Name, FiberEntryType)...))
		}

		// csrf middleware
		if element.Interceptors.Csrf.Enabled {
			inters = append(inters, rkfibercsrf.Interceptor(
				rkmidcsrf.ToOptions(&element.Interceptors.Csrf, element.Name, FiberEntryType)...))
		}

		// cors middleware
		if element.Interceptors.Cors.Enabled {
			inters = append(inters, rkfibercors.Interceptor(
				rkmidcors.ToOptions(&element.Interceptors.Cors, element.Name, FiberEntryType)...))
		}

		// meta middleware
		if element.Interceptors.Meta.Enabled {
			inters = append(inters, rkfibermeta.Interceptor(
				rkmidmeta.ToOptions(&element.Interceptors.Meta, element.Name, FiberEntryType)...))
		}

		// auth middlewares
		if element.Interceptors.Auth.Enabled {
			inters = append(inters, rkfiberauth.Interceptor(
				rkmidauth.ToOptions(&element.Interceptors.Auth, element.Name, FiberEntryType)...))
		}

		// timeout middlewares
		if element.Interceptors.Timeout.Enabled {
			inters = append(inters, rkfibertimeout.Interceptor(
				rkmidtimeout.ToOptions(&element.Interceptors.Timeout, element.Name, FiberEntryType)...))
		}

		// rate limit middleware
		if element.Interceptors.RateLimit.Enabled {
			inters = append(inters, rkfiberlimit.Interceptor(
				rkmidlimit.ToOptions(&element.Interceptors.RateLimit, element.Name, FiberEntryType)...))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Cert.Ref)

		entry := RegisterFiberEntry(
			WithName(name),
			WithDescription(element.Description),
			WithPort(element.Port),
			WithZapLoggerEntry(zapLoggerEntry),
			WithEventLoggerEntry(eventLoggerEntry),
			WithCertEntry(certEntry),
			WithPromEntry(promEntry),
			WithTvEntry(tvEntry),
			WithCommonServiceEntry(commonServiceEntry),
			WithSwEntry(swEntry),
			WithStaticFileHandlerEntry(staticEntry),
			WithInterceptors(inters...))

		res[name] = entry
	}

	return res
}

// RegisterFiberEntry register FiberEntry with options.
func RegisterFiberEntry(opts ...FiberEntryOption) *FiberEntry {
	entry := &FiberEntry{
		EntryType:        FiberEntryType,
		EntryDescription: FiberEntryDescription,
		Port:             8080,
	}

	for i := range opts {
		opts[i](entry)
	}

	// insert panic interceptor
	entry.Interceptors = append(entry.Interceptors, rkfiberpanic.Interceptor(
		rkmidpanic.WithEntryNameAndType(entry.EntryName, entry.EntryType)))

	if entry.ZapLoggerEntry == nil {
		entry.ZapLoggerEntry = rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()
	}

	if entry.EventLoggerEntry == nil {
		entry.EventLoggerEntry = rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()
	}

	if len(entry.EntryName) < 1 {
		entry.EntryName = "FiberServer-" + strconv.FormatUint(entry.Port, 10)
	}

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// Bootstrap FiberEntry.
func (entry *FiberEntry) Bootstrap(ctx context.Context) {
	event, logger := entry.logBasicInfo("Bootstrap")

	if entry.App == nil {
		if entry.FiberConfig != nil {
			entry.FiberConfig.DisableStartupMessage = true
		} else {
			entry.FiberConfig = &fiber.Config{
				DisableStartupMessage: true,
				ReadTimeout:           5 * time.Second,
				IdleTimeout:           5 * time.Second,
			}
		}

		entry.App = fiber.New(*entry.FiberConfig)
	}

	// Default interceptor should be at front
	for _, v := range entry.Interceptors {
		entry.App.Use(v)
	}

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		entry.App.Get(path.Join(entry.SwEntry.Path, "*"), adaptor.HTTPHandler(entry.SwEntry.ConfigFileHandler()))
		entry.App.Get(path.Join(entry.SwEntry.AssetsFilePath, "*"), adaptor.HTTPHandler(entry.SwEntry.AssetsFileHandler()))
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is static file handler enabled?
	if entry.IsStaticFileHandlerEnabled() {
		// Register path into Router.
		entry.App.Get(strings.TrimSuffix(entry.StaticFileEntry.Path, "/"), func(ctx *fiber.Ctx) error {
			return ctx.Redirect(entry.StaticFileEntry.Path, http.StatusTemporaryRedirect)
		})

		// Register path into Router.
		entry.App.Get(path.Join(entry.StaticFileEntry.Path, "*"), adaptor.HTTPHandler(entry.StaticFileEntry.GetFileHandler()))

		// Bootstrap entry.
		entry.StaticFileEntry.Bootstrap(ctx)
	}

	// Is prometheus enabled?
	if entry.IsPromEnabled() {
		// Register prom path into Router.
		entry.App.Get(entry.PromEntry.Path, adaptor.HTTPHandler(promhttp.HandlerFor(entry.PromEntry.Gatherer, promhttp.HandlerOpts{})))

		// don't start with http handler, we will handle it by ourselves
		entry.PromEntry.Bootstrap(ctx)
	}

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.App.Get(entry.CommonServiceEntry.HealthyPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Healthy))
		entry.App.Get(entry.CommonServiceEntry.GcPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Gc))
		entry.App.Get(entry.CommonServiceEntry.InfoPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Info))
		entry.App.Get(entry.CommonServiceEntry.ConfigsPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Configs))
		entry.App.Get(entry.CommonServiceEntry.SysPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Sys))
		entry.App.Get(entry.CommonServiceEntry.EntriesPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Entries))
		entry.App.Get(entry.CommonServiceEntry.CertsPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Certs))
		entry.App.Get(entry.CommonServiceEntry.LogsPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Logs))
		entry.App.Get(entry.CommonServiceEntry.DepsPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Deps))
		entry.App.Get(entry.CommonServiceEntry.LicensePath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.License))
		entry.App.Get(entry.CommonServiceEntry.ReadmePath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Readme))
		entry.App.Get(entry.CommonServiceEntry.GitPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Git))

		entry.App.Get(entry.CommonServiceEntry.ApisPath, entry.Apis)
		entry.App.Get(entry.CommonServiceEntry.ReqPath, entry.Req)

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	// Is TV enabled?
	if entry.IsTvEnabled() {
		// Bootstrap TV entry.
		entry.App.Get(path.Join(entry.TvEntry.BasePath, "*"), entry.TV)
		entry.App.Get(path.Join(entry.TvEntry.AssetsFilePath, "*"), adaptor.HTTPHandlerFunc(entry.TvEntry.AssetsFileHandler()))

		entry.TvEntry.Bootstrap(ctx)
	}

	go entry.startServer(event, logger)

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// Start server
// We move the code here for testability
func (entry *FiberEntry) startServer(event rkquery.Event, logger *zap.Logger) {
	if entry.App != nil {
		// If TLS was enabled, we need to load server certificate and key and start http server with ListenAndServeTLS()
		if entry.IsTlsEnabled() {
			err := entry.App.Server().ListenAndServeTLSEmbed(
				":"+strconv.FormatUint(entry.Port, 10),
				entry.CertEntry.Store.ServerCert,
				entry.CertEntry.Store.ServerKey)

			if err != nil && err != http.ErrServerClosed {
				event.AddErr(err)
				logger.Error("Error occurs while starting fiber server with tls.", event.ListPayloads()...)
				rkcommon.ShutdownWithError(err)
			}
		} else {
			err := entry.App.Listen(":" + strconv.FormatUint(entry.Port, 10))

			if err != nil && err != http.ErrServerClosed {
				event.AddErr(err)
				logger.Error("Error occurs while starting fiber server.", event.ListPayloads()...)
				rkcommon.ShutdownWithError(err)
			}
		}
	}
}

// Interrupt FiberEntry.
func (entry *FiberEntry) Interrupt(ctx context.Context) {
	event, logger := entry.logBasicInfo("Interrupt")

	if entry.IsSwEnabled() {
		// Interrupt swagger entry
		entry.SwEntry.Interrupt(ctx)
	}

	if entry.IsStaticFileHandlerEnabled() {
		// Interrupt entry
		entry.StaticFileEntry.Interrupt(ctx)
	}

	if entry.IsPromEnabled() {
		// Interrupt prometheus entry
		entry.PromEntry.Interrupt(ctx)
	}

	if entry.IsCommonServiceEnabled() {
		// Interrupt common service entry
		entry.CommonServiceEntry.Interrupt(ctx)
	}

	if entry.IsTvEnabled() {
		// Interrupt common service entry
		entry.TvEntry.Interrupt(ctx)
	}

	if entry.App != nil {
		if err := entry.App.Shutdown(); err != nil && err != http.ErrServerClosed {
			event.AddErr(err)
			logger.Warn("Error occurs while stopping fiber-server.", event.ListPayloads()...)
		}
	}

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
}

// GetName Get entry name.
func (entry *FiberEntry) GetName() string {
	return entry.EntryName
}

// GetType Get entry type.
func (entry *FiberEntry) GetType() string {
	return entry.EntryType
}

// GetDescription Get description of entry.
func (entry *FiberEntry) GetDescription() string {
	return entry.EntryDescription
}

// String Stringfy entry.
func (entry *FiberEntry) String() string {
	bytes, _ := json.Marshal(entry)
	return string(bytes)
}

// ***************** Stringfy *****************

// MarshalJSON Marshal entry.
func (entry *FiberEntry) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"entryName":          entry.EntryName,
		"entryType":          entry.EntryType,
		"entryDescription":   entry.EntryDescription,
		"eventLoggerEntry":   entry.EventLoggerEntry.GetName(),
		"zapLoggerEntry":     entry.ZapLoggerEntry.GetName(),
		"port":               entry.Port,
		"swEntry":            entry.SwEntry,
		"commonServiceEntry": entry.CommonServiceEntry,
		"promEntry":          entry.PromEntry,
		"tvEntry":            entry.TvEntry,
	}

	if entry.CertEntry != nil {
		m["certEntry"] = entry.CertEntry.GetName()
	}

	interceptorsStr := make([]string, 0)
	m["interceptors"] = &interceptorsStr

	for i := range entry.Interceptors {
		element := entry.Interceptors[i]
		interceptorsStr = append(interceptorsStr,
			path.Base(runtime.FuncForPC(reflect.ValueOf(element).Pointer()).Name()))
	}

	return json.Marshal(&m)
}

// UnmarshalJSON Not supported.
func (entry *FiberEntry) UnmarshalJSON([]byte) error {
	return nil
}

// ***************** Public functions *****************

// GetFiberEntry Get FiberEntry from rkentry.GlobalAppCtx.
func GetFiberEntry(name string) *FiberEntry {
	entryRaw := rkentry.GlobalAppCtx.GetEntry(name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*FiberEntry)
	return entry
}

// AddInterceptor Add interceptors.
// This function should be called before Bootstrap() called.
func (entry *FiberEntry) AddInterceptor(inters ...fiber.Handler) {
	entry.Interceptors = append(entry.Interceptors, inters...)
}

// SetFiberConfig override fiber config
func (entry *FiberEntry) SetFiberConfig(conf *fiber.Config) {
	entry.FiberConfig = conf
}

// IsTlsEnabled Is TLS enabled?
func (entry *FiberEntry) IsTlsEnabled() bool {
	return entry.CertEntry != nil && entry.CertEntry.Store != nil
}

// IsSwEnabled Is swagger entry enabled?
func (entry *FiberEntry) IsSwEnabled() bool {
	return entry.SwEntry != nil
}

// IsCommonServiceEnabled Is common service entry enabled?
func (entry *FiberEntry) IsCommonServiceEnabled() bool {
	return entry.CommonServiceEntry != nil
}

// IsTvEnabled Is TV entry enabled?
func (entry *FiberEntry) IsTvEnabled() bool {
	return entry.TvEntry != nil
}

// IsPromEnabled Is prometheus entry enabled?
func (entry *FiberEntry) IsPromEnabled() bool {
	return entry.PromEntry != nil
}

// IsStaticFileHandlerEnabled Is static file handler entry enabled?
func (entry *FiberEntry) IsStaticFileHandlerEnabled() bool {
	return entry.StaticFileEntry != nil
}

// RefreshFiberRoutes will rebuild fiber app tree, this is required!!!
// Why not create fiber.App before bootstrap?
//
// This is because we hope to provide user specified fiber.Config which can override our custom settings.
func (entry *FiberEntry) RefreshFiberRoutes() {
	entry.App.Handler()
}

type Route struct {
	Method string `yaml:"method" json:"method"`
	Path   string `yaml:"path" json:"path"`
	Port   uint64 `yaml:"port" json:"port"`
}

// ListRoutes returns all user routes except interceptors
func (entry *FiberEntry) ListRoutes() []*Route {
	res := make([]*Route, 0)

	stacks := entry.App.Stack()
	for _, i := range stacks {
		for _, j := range i {
			if !entry.isInterceptor(j) {
				res = append(res, &Route{
					Method: j.Method,
					Path:   j.Path,
					Port:   entry.Port,
				})
			}
		}
	}

	return res
}

// ***************** Helper function *****************

func (entry *FiberEntry) isInterceptor(r *fiber.Route) bool {
	for _, i := range entry.Interceptors {
		pt := reflect.ValueOf(i).Pointer()

		for _, j := range r.Handlers {
			if pt == reflect.ValueOf(j).Pointer() {
				return true
			}
		}
	}

	return false
}

// Add basic fields into event.
func (entry *FiberEntry) logBasicInfo(operation string) (rkquery.Event, *zap.Logger) {
	event := entry.EventLoggerEntry.GetEventHelper().Start(
		operation,
		rkquery.WithEntryName(entry.GetName()),
		rkquery.WithEntryType(entry.GetType()))
	logger := entry.ZapLoggerEntry.GetLogger().With(
		zap.String("eventId", event.GetEventId()),
		zap.String("entryName", entry.EntryName))

	// add FiberEntry info
	event.AddPayloads(
		zap.Uint64("fiberPort", entry.Port))

	// add SwEntry info
	if entry.IsSwEnabled() {
		event.AddPayloads(
			zap.Bool("swEnabled", true),
			zap.String("swPath", entry.SwEntry.Path))
	}

	// add CommonServiceEntry info
	if entry.IsCommonServiceEnabled() {
		event.AddPayloads(
			zap.Bool("commonServiceEnabled", true),
			zap.String("commonServicePathPrefix", "/rk/v1/"))
	}

	// add TvEntry info
	if entry.IsTvEnabled() {
		event.AddPayloads(
			zap.Bool("tvEnabled", true),
			zap.String("tvPath", "/rk/v1/tv/"))
	}

	// add PromEntry info
	if entry.IsPromEnabled() {
		event.AddPayloads(
			zap.Bool("promEnabled", true),
			zap.Uint64("promPort", entry.PromEntry.Port),
			zap.String("promPath", entry.PromEntry.Path))
	}

	// add StaticFileHandlerEntry info
	if entry.IsStaticFileHandlerEnabled() {
		event.AddPayloads(
			zap.Bool("staticFileHandlerEnabled", true),
			zap.String("staticFileHandlerPath", entry.StaticFileEntry.Path))
	}

	// add tls info
	if entry.IsTlsEnabled() {
		event.AddPayloads(
			zap.Bool("tlsEnabled", true))
	}

	logger.Info(fmt.Sprintf("%s fiberEntry", operation))

	return event, logger
}

// ***************** Common Service Extension API *****************

// Apis list apis
func (entry *FiberEntry) Apis(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Access-Control-Allow-Origin", "*")

	ctx.Context().SetStatusCode(http.StatusOK)
	return ctx.JSON(entry.doApis(ctx))
}

// Req handler
func (entry *FiberEntry) Req(ctx *fiber.Ctx) error {
	ctx.Context().SetStatusCode(http.StatusOK)
	return ctx.JSON(entry.doReq(ctx))
}

// TV handler
func (entry *FiberEntry) TV(ctx *fiber.Ctx) error {
	logger := rkfiberctx.GetLogger(ctx)

	ctx.Context().SetContentType("text/html; charset=utf-8")

	switch item := ctx.Params("*"); item {
	case "apis":
		buf := entry.TvEntry.ExecuteTemplate("apis", entry.doApis(ctx), logger)
		ctx.Context().SetBodyStream(buf, -1)
	default:
		buf := entry.TvEntry.Action(item, logger)
		ctx.Context().SetBodyStream(buf, -1)
	}

	return nil
}

// Construct swagger URL based on IP and scheme
func (entry *FiberEntry) constructSwUrl(ctx *fiber.Ctx) string {
	if !entry.IsSwEnabled() {
		return "N/A"
	}

	originalURL := fmt.Sprintf("localhost:%d", entry.Port)
	if ctx != nil && ctx.Request() != nil && len(ctx.Hostname()) > 0 {
		originalURL = ctx.Hostname()
	}

	scheme := "http"
	if ctx != nil && ctx.Request() != nil && ctx.Context().IsTLS() {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s%s", scheme, originalURL, entry.SwEntry.Path)
}

// Helper function for APIs call
func (entry *FiberEntry) doApis(ctx *fiber.Ctx) *rkentry.ApisResponse {
	res := &rkentry.ApisResponse{
		Entries: make([]*rkentry.ApisResponseElement, 0),
	}

	routes := entry.ListRoutes()
	for j := range routes {
		info := routes[j]

		element := &rkentry.ApisResponseElement{
			EntryName: entry.GetName(),
			Method:    info.Method,
			Path:      info.Path,
			Port:      entry.Port,
			SwUrl:     entry.constructSwUrl(ctx),
		}
		res.Entries = append(res.Entries, element)
	}

	return res
}

// Is metrics from prometheus contains particular api?
func (entry *FiberEntry) containsMetrics(api string, metrics []*rkentry.ReqMetricsRK) bool {
	for i := range metrics {
		if metrics[i].RestPath == api {
			return true
		}
	}

	return false
}

// Helper function for Req call
func (entry *FiberEntry) doReq(ctx *fiber.Ctx) *rkentry.ReqResponse {
	metricsSet := rkmidmetrics.GetServerMetricsSet(entry.GetName())
	if metricsSet == nil {
		return &rkentry.ReqResponse{
			Metrics: make([]*rkentry.ReqMetricsRK, 0),
		}
	}

	vector := metricsSet.GetSummary(rkmidmetrics.MetricsNameElapsedNano)
	if vector == nil {
		return &rkentry.ReqResponse{
			Metrics: make([]*rkentry.ReqMetricsRK, 0),
		}
	}

	reqMetrics := rkentry.NewPromMetricsInfo(vector)

	// Fill missed metrics
	apis := make([]string, 0)

	routes := entry.ListRoutes()
	for j := range routes {
		info := routes[j]
		apis = append(apis, info.Path)
	}

	// Add empty metrics into result
	for i := range apis {
		if !entry.containsMetrics(apis[i], reqMetrics) {
			reqMetrics = append(reqMetrics, &rkentry.ReqMetricsRK{
				RestPath: apis[i],
				ResCode:  make([]*rkentry.ResCodeRK, 0),
			})
		}
	}

	return &rkentry.ReqResponse{
		Metrics: reqMetrics,
	}
}

// ***************** Options *****************

// FiberEntryOption Fiber entry option.
type FiberEntryOption func(*FiberEntry)

// WithName provide name.
func WithName(name string) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.EntryName = name
	}
}

// WithDescription provide name.
func WithDescription(description string) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.EntryDescription = description
	}
}

// WithPort provide port.
func WithPort(port uint64) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.Port = port
	}
}

// WithZapLoggerEntry provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntry(zapLogger *rkentry.ZapLoggerEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.ZapLoggerEntry = zapLogger
	}
}

// WithEventLoggerEntry provide rkentry.EventLoggerEntry.
func WithEventLoggerEntry(eventLogger *rkentry.EventLoggerEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.EventLoggerEntry = eventLogger
	}
}

// WithCertEntry provide rkentry.CertEntry.
func WithCertEntry(certEntry *rkentry.CertEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntry provide SwEntry.
func WithSwEntry(sw *rkentry.SwEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.SwEntry = sw
	}
}

// WithCommonServiceEntry provide CommonServiceEntry.
func WithCommonServiceEntry(commonServiceEntry *rkentry.CommonServiceEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.CommonServiceEntry = commonServiceEntry
	}
}

// WithInterceptors provide user interceptors.
func WithInterceptors(inters ...fiber.Handler) FiberEntryOption {
	return func(entry *FiberEntry) {
		if entry.Interceptors == nil {
			entry.Interceptors = make([]fiber.Handler, 0)
		}

		entry.Interceptors = append(entry.Interceptors, inters...)
	}
}

// WithPromEntry provide PromEntry.
func WithPromEntry(prom *rkentry.PromEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.PromEntry = prom
	}
}

// WithStaticFileHandlerEntry provide StaticFileHandlerEntry.
func WithStaticFileHandlerEntry(staticEntry *rkentry.StaticFileHandlerEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.StaticFileEntry = staticEntry
	}
}

// WithTvEntry provide TvEntry.
func WithTvEntry(tvEntry *rkentry.TvEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.TvEntry = tvEntry
	}
}

// WithFiberConfig provide fiber.Config.
func WithFiberConfig(conf *fiber.Config) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.FiberConfig = conf
	}
}
