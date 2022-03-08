// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkfiber an implementation of rkentry.Entry which could be used start restful server with fiber framework
package rkfiber

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-entry/v2/entry"
	"github.com/rookie-ninja/rk-entry/v2/middleware"
	"github.com/rookie-ninja/rk-entry/v2/middleware/auth"
	"github.com/rookie-ninja/rk-entry/v2/middleware/cors"
	"github.com/rookie-ninja/rk-entry/v2/middleware/csrf"
	"github.com/rookie-ninja/rk-entry/v2/middleware/jwt"
	"github.com/rookie-ninja/rk-entry/v2/middleware/log"
	"github.com/rookie-ninja/rk-entry/v2/middleware/meta"
	"github.com/rookie-ninja/rk-entry/v2/middleware/panic"
	"github.com/rookie-ninja/rk-entry/v2/middleware/prom"
	"github.com/rookie-ninja/rk-entry/v2/middleware/ratelimit"
	"github.com/rookie-ninja/rk-entry/v2/middleware/secure"
	"github.com/rookie-ninja/rk-entry/v2/middleware/timeout"
	"github.com/rookie-ninja/rk-entry/v2/middleware/tracing"
	"github.com/rookie-ninja/rk-fiber/middleware/auth"
	"github.com/rookie-ninja/rk-fiber/middleware/cors"
	"github.com/rookie-ninja/rk-fiber/middleware/csrf"
	"github.com/rookie-ninja/rk-fiber/middleware/jwt"
	"github.com/rookie-ninja/rk-fiber/middleware/log"
	"github.com/rookie-ninja/rk-fiber/middleware/meta"
	"github.com/rookie-ninja/rk-fiber/middleware/panic"
	rkfiberprom "github.com/rookie-ninja/rk-fiber/middleware/prom"
	"github.com/rookie-ninja/rk-fiber/middleware/ratelimit"
	"github.com/rookie-ninja/rk-fiber/middleware/secure"
	"github.com/rookie-ninja/rk-fiber/middleware/timeout"
	"github.com/rookie-ninja/rk-fiber/middleware/tracing"
	"github.com/rookie-ninja/rk-query"
	"go.uber.org/zap"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// FiberEntryType type of entry
	FiberEntryType = "FiberEntry"
)

// This must be declared in order to register registration function into rk context
// otherwise, rk-boot won't able to bootstrap fiber entry automatically from boot config file
func init() {
	rkentry.RegisterEntryRegFunc(RegisterFiberEntryYAML)
}

// BootFiber boot config which is for fiber entry.
type BootFiber struct {
	Fiber []struct {
		Enabled       bool                          `yaml:"enabled" json:"enabled"`
		Name          string                        `yaml:"name" json:"name"`
		Port          uint64                        `yaml:"port" json:"port"`
		Description   string                        `yaml:"description" json:"description"`
		CertEntry     string                        `yaml:"certEntry" json:"certEntry"`
		LoggerEntry   string                        `yaml:"loggerEntry" json:"loggerEntry"`
		EventEntry    string                        `yaml:"eventEntry" json:"eventEntry"`
		SW            rkentry.BootSW                `yaml:"sw" json:"sw"`
		Docs          rkentry.BootDocs              `yaml:"docs" json:"docs"`
		CommonService rkentry.BootCommonService     `yaml:"commonService" json:"commonService"`
		Prom          rkentry.BootProm              `yaml:"prom" json:"prom"`
		Static        rkentry.BootStaticFileHandler `yaml:"static" json:"static"`
		Middleware    struct {
			Ignore    []string                `yaml:"ignore" json:"ignore"`
			Logging   rkmidlog.BootConfig     `yaml:"logging" json:"logging"`
			Prom      rkmidprom.BootConfig    `yaml:"prom" json:"prom"`
			Auth      rkmidauth.BootConfig    `yaml:"auth" json:"auth"`
			Cors      rkmidcors.BootConfig    `yaml:"cors" json:"cors"`
			Meta      rkmidmeta.BootConfig    `yaml:"meta" json:"meta"`
			Jwt       rkmidjwt.BootConfig     `yaml:"jwt" json:"jwt"`
			Secure    rkmidsec.BootConfig     `yaml:"secure" json:"secure"`
			Csrf      rkmidcsrf.BootConfig    `yaml:"csrf" yaml:"csrf"`
			RateLimit rkmidlimit.BootConfig   `yaml:"rateLimit" json:"rateLimit"`
			Timeout   rkmidtimeout.BootConfig `yaml:"timeout" json:"timeout"`
			Trace     rkmidtrace.BootConfig   `yaml:"trace" json:"trace"`
		} `yaml:"middleware" json:"middleware"`
	} `yaml:"fiber" json:"fiber"`
}

// FiberEntry implements rkentry.Entry interface.
type FiberEntry struct {
	entryName          string                          `json:"-" yaml:"-"`
	entryType          string                          `json:"-" yaml:"-"`
	entryDescription   string                          `json:"-" yaml:"-"`
	LoggerEntry        *rkentry.LoggerEntry            `json:"-" yaml:"-"`
	EventEntry         *rkentry.EventEntry             `json:"-" yaml:"-"`
	CertEntry          *rkentry.CertEntry              `json:"-" yaml:"-"`
	Port               uint64                          `json:"-" yaml:"-"`
	SwEntry            *rkentry.SWEntry                `json:"-" yaml:"-"`
	CommonServiceEntry *rkentry.CommonServiceEntry     `json:"-" yaml:"-"`
	App                *fiber.App                      `json:"-" yaml:"-"`
	FiberConfig        *fiber.Config                   `json:"-" yaml:"-"`
	Middlewares        []fiber.Handler                 `json:"-" yaml:"-"`
	PromEntry          *rkentry.PromEntry              `json:"-" yaml:"-"`
	StaticFileEntry    *rkentry.StaticFileHandlerEntry `json:"-" yaml:"-"`
	DocsEntry          *rkentry.DocsEntry              `json:"-" yaml:"-"`
	bootstrapLogOnce   sync.Once                       `json:"-" yaml:"-"`
}

// RegisterFiberEntryYAML register fiber entries with provided config file (Must YAML file).
//
// Currently, support two ways to provide config file path.
// 1: With function parameters
// 2: With command line flag "--rkboot" described in rkentry.BootConfigPathFlagKey (Will override function parameter if exists)
// Command line flag has high priority which would override function parameter
//
// Error handling:
// Process will shutdown if any errors occur with rkentry.ShutdownWithError function
//
// Override elements in config file:
// We learned from HELM source code which would override elements in YAML file with "--set" flag followed with comma
// separated key/value pairs.
//
// We are using "--rkset" described in rkentry.BootConfigOverrideKey in order to distinguish with user flags
// Example of common usage: ./binary_file --rkset "key1=val1,key2=val2"
// Example of nested map:   ./binary_file --rkset "outer.inner.key=val"
// Example of slice:        ./binary_file --rkset "outer[0].key=val"
func RegisterFiberEntryYAML(raw []byte) map[string]rkentry.Entry {
	res := make(map[string]rkentry.Entry)

	// 1: Decode config map into boot config struct
	config := &BootFiber{}
	rkentry.UnmarshalBootYAML(raw, config)

	// 2: Init fiber entries with boot config
	for i := range config.Fiber {
		element := config.Fiber[i]
		if !element.Enabled {
			continue
		}

		name := element.Name

		// logger entry
		loggerEntry := rkentry.GlobalAppCtx.GetLoggerEntry(element.LoggerEntry)
		if loggerEntry == nil {
			loggerEntry = rkentry.LoggerEntryStdout
		}

		// event entry
		eventEntry := rkentry.GlobalAppCtx.GetEventEntry(element.EventEntry)
		if eventEntry == nil {
			eventEntry = rkentry.EventEntryStdout
		}

		// cert entry
		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.CertEntry)

		// Register swagger entry
		swEntry := rkentry.RegisterSWEntry(&element.SW, rkentry.WithNameSWEntry(element.Name))

		// Register docs entry
		docsEntry := rkentry.RegisterDocsEntry(&element.Docs, rkentry.WithNameDocsEntry(element.Name))

		// Register prometheus entry
		promRegistry := prometheus.NewRegistry()
		promEntry := rkentry.RegisterPromEntry(&element.Prom, rkentry.WithRegistryPromEntry(promRegistry))

		// Register common service entry
		commonServiceEntry := rkentry.RegisterCommonServiceEntry(&element.CommonService)

		// Register static file handler
		staticEntry := rkentry.RegisterStaticFileHandlerEntry(&element.Static, rkentry.WithNameStaticFileHandlerEntry(element.Name))

		inters := make([]fiber.Handler, 0)

		// add global path ignorance
		rkmid.AddPathToIgnoreGlobal(element.Middleware.Ignore...)

		// logging middlewares
		if element.Middleware.Logging.Enabled {
			inters = append(inters, rkfiberlog.Middleware(
				rkmidlog.ToOptions(&element.Middleware.Logging, element.Name, FiberEntryType,
					loggerEntry, eventEntry)...))
		}

		// insert panic interceptor
		inters = append(inters, rkfiberpanic.Middleware(
			rkmidpanic.WithEntryNameAndType(element.Name, FiberEntryType)))

		// metrics middleware
		if element.Middleware.Prom.Enabled {
			inters = append(inters, rkfiberprom.Middleware(
				rkmidprom.ToOptions(&element.Middleware.Prom, element.Name, FiberEntryType,
					promRegistry, rkmidprom.LabelerTypeHttp)...))
		}

		// tracing middleware
		if element.Middleware.Trace.Enabled {
			inters = append(inters, rkfibertrace.Middleware(
				rkmidtrace.ToOptions(&element.Middleware.Trace, element.Name, FiberEntryType)...))
		}

		// jwt middleware
		if element.Middleware.Jwt.Enabled {
			inters = append(inters, rkfiberjwt.Middleware(
				rkmidjwt.ToOptions(&element.Middleware.Jwt, element.Name, FiberEntryType)...))
		}

		// secure middleware
		if element.Middleware.Secure.Enabled {
			inters = append(inters, rkfibersec.Middleware(
				rkmidsec.ToOptions(&element.Middleware.Secure, element.Name, FiberEntryType)...))
		}

		// csrf middleware
		if element.Middleware.Csrf.Enabled {
			inters = append(inters, rkfibercsrf.Middleware(
				rkmidcsrf.ToOptions(&element.Middleware.Csrf, element.Name, FiberEntryType)...))
		}

		// cors middleware
		if element.Middleware.Cors.Enabled {
			inters = append(inters, rkfibercors.Middleware(
				rkmidcors.ToOptions(&element.Middleware.Cors, element.Name, FiberEntryType)...))
		}

		// meta middleware
		if element.Middleware.Meta.Enabled {
			inters = append(inters, rkfibermeta.Middleware(
				rkmidmeta.ToOptions(&element.Middleware.Meta, element.Name, FiberEntryType)...))
		}

		// auth middlewares
		if element.Middleware.Auth.Enabled {
			inters = append(inters, rkfiberauth.Middleware(
				rkmidauth.ToOptions(&element.Middleware.Auth, element.Name, FiberEntryType)...))
		}

		// timeout middlewares
		if element.Middleware.Timeout.Enabled {
			inters = append(inters, rkfibertimeout.Middleware(
				rkmidtimeout.ToOptions(&element.Middleware.Timeout, element.Name, FiberEntryType)...))
		}

		// rate limit middleware
		if element.Middleware.RateLimit.Enabled {
			inters = append(inters, rkfiberlimit.Middleware(
				rkmidlimit.ToOptions(&element.Middleware.RateLimit, element.Name, FiberEntryType)...))
		}

		entry := RegisterFiberEntry(
			WithName(name),
			WithDescription(element.Description),
			WithPort(element.Port),
			WithLoggerEntry(loggerEntry),
			WithEventEntry(eventEntry),
			WithCertEntry(certEntry),
			WithPromEntry(promEntry),
			WithDocsEntry(docsEntry),
			WithCommonServiceEntry(commonServiceEntry),
			WithSwEntry(swEntry),
			WithStaticFileHandlerEntry(staticEntry),
			WithMiddleware(inters...))

		res[name] = entry
	}

	return res
}

// RegisterFiberEntry register FiberEntry with options.
func RegisterFiberEntry(opts ...FiberEntryOption) *FiberEntry {
	entry := &FiberEntry{
		entryType:        FiberEntryType,
		entryDescription: "Internal RK entry which helps to bootstrap with fiber framework.",
		Port:             80,
		LoggerEntry:      rkentry.NewLoggerEntryStdout(),
		EventEntry:       rkentry.NewEventEntryStdout(),
		Middlewares:      make([]fiber.Handler, 0),
	}

	for i := range opts {
		opts[i](entry)
	}

	if len(entry.entryName) < 1 {
		entry.entryName = "fiber-" + strconv.FormatUint(entry.Port, 10)
	}

	// add entry name and entry type into loki syncer if enabled
	entry.LoggerEntry.AddEntryLabelToLokiSyncer(entry)
	entry.EventEntry.AddEntryLabelToLokiSyncer(entry)

	rkentry.GlobalAppCtx.AddEntry(entry)

	return entry
}

// Bootstrap FiberEntry.
func (entry *FiberEntry) Bootstrap(ctx context.Context) {
	event, logger := entry.logBasicInfo("Bootstrap", ctx)

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
	for _, v := range entry.Middlewares {
		entry.App.Use(v)
	}

	// Is common service enabled?
	if entry.IsCommonServiceEnabled() {
		// Register common service path into Router.
		entry.App.Get(entry.CommonServiceEntry.ReadyPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Ready))
		entry.App.Get(entry.CommonServiceEntry.GcPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Gc))
		entry.App.Get(entry.CommonServiceEntry.InfoPath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Info))
		entry.App.Get(entry.CommonServiceEntry.AlivePath, adaptor.HTTPHandlerFunc(entry.CommonServiceEntry.Alive))

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	// Is swagger enabled?
	if entry.IsSwEnabled() {
		entry.App.Get(path.Join(entry.SwEntry.Path, "*"), adaptor.HTTPHandler(entry.SwEntry.ConfigFileHandler()))
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

	// Is Docs enabled?
	if entry.IsDocsEnabled() {
		// Bootstrap TV entry.
		entry.App.Get(path.Join(entry.DocsEntry.Path, "*"), adaptor.HTTPHandlerFunc(entry.DocsEntry.ConfigFileHandler()))
		entry.DocsEntry.Bootstrap(ctx)
	}

	go entry.startServer(event, logger)

	entry.bootstrapLogOnce.Do(func() {
		// Print link and logging message
		scheme := "http"
		if entry.IsTlsEnabled() {
			scheme = "https"
		}

		if entry.IsSwEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("SwaggerEntry: %s://localhost:%d%s", scheme, entry.Port, entry.SwEntry.Path))
		}
		if entry.IsDocsEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("DocsEntry: %s://localhost:%d%s", scheme, entry.Port, entry.DocsEntry.Path))
		}
		if entry.IsPromEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("PromEntry: %s://localhost:%d%s", scheme, entry.Port, entry.PromEntry.Path))
		}
		if entry.IsStaticFileHandlerEnabled() {
			entry.LoggerEntry.Info(fmt.Sprintf("StaticFileHandlerEntry: %s://localhost:%d%s", scheme, entry.Port, entry.StaticFileEntry.Path))
		}
		if entry.IsCommonServiceEnabled() {
			handlers := []string{
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.ReadyPath),
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.AlivePath),
				fmt.Sprintf("%s://localhost:%d%s", scheme, entry.Port, entry.CommonServiceEntry.InfoPath),
			}

			entry.LoggerEntry.Info(fmt.Sprintf("CommonSreviceEntry: %s", strings.Join(handlers, ", ")))
		}
		entry.EventEntry.Finish(event)
	})
}

// Start server
// We move the code here for testability
func (entry *FiberEntry) startServer(event rkquery.Event, logger *zap.Logger) {
	if entry.App != nil {
		// If TLS was enabled, we need to load server certificate and key and start http server with ListenAndServeTLS()
		if entry.IsTlsEnabled() {
			conn, err := net.Listen("tcp4", ":"+strconv.FormatUint(entry.Port, 10))
			if err != nil && err != http.ErrServerClosed {
				event.AddErr(err)
				logger.Error("Error occurs while starting fiber server with tls.", event.ListPayloads()...)
				rkentry.ShutdownWithError(err)
			}

			err = entry.App.Server().Serve(tls.NewListener(conn, &tls.Config{
				Certificates: []tls.Certificate{*entry.CertEntry.Certificate},
			}))

			if err != nil && err != http.ErrServerClosed {
				event.AddErr(err)
				logger.Error("Error occurs while starting fiber server with tls.", event.ListPayloads()...)
				rkentry.ShutdownWithError(err)
			}
		} else {
			err := entry.App.Listen(":" + strconv.FormatUint(entry.Port, 10))

			if err != nil && err != http.ErrServerClosed {
				event.AddErr(err)
				logger.Error("Error occurs while starting fiber server.", event.ListPayloads()...)
				rkentry.ShutdownWithError(err)
			}
		}
	}
}

// Interrupt FiberEntry.
func (entry *FiberEntry) Interrupt(ctx context.Context) {
	event, logger := entry.logBasicInfo("Interrupt", ctx)

	if entry.IsSwEnabled() {
		entry.SwEntry.Interrupt(ctx)
	}

	if entry.IsStaticFileHandlerEnabled() {
		entry.StaticFileEntry.Interrupt(ctx)
	}

	if entry.IsPromEnabled() {
		entry.PromEntry.Interrupt(ctx)
	}

	if entry.IsCommonServiceEnabled() {
		entry.CommonServiceEntry.Interrupt(ctx)
	}

	if entry.IsDocsEnabled() {
		entry.DocsEntry.Interrupt(ctx)
	}

	if entry.App != nil {
		if err := entry.App.Shutdown(); err != nil && err != http.ErrServerClosed {
			event.AddErr(err)
			logger.Warn("Error occurs while stopping fiber-server.", event.ListPayloads()...)
		}
	}

	entry.EventEntry.Finish(event)
}

// GetName Get entry name.
func (entry *FiberEntry) GetName() string {
	return entry.entryName
}

// GetType Get entry type.
func (entry *FiberEntry) GetType() string {
	return entry.entryType
}

// GetDescription Get description of entry.
func (entry *FiberEntry) GetDescription() string {
	return entry.entryDescription
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
		"name":                   entry.entryName,
		"type":                   entry.entryType,
		"description":            entry.entryDescription,
		"port":                   entry.Port,
		"swEntry":                entry.SwEntry,
		"docsEntry":              entry.DocsEntry,
		"commonServiceEntry":     entry.CommonServiceEntry,
		"promEntry":              entry.PromEntry,
		"staticFileHandlerEntry": entry.StaticFileEntry,
	}

	if entry.CertEntry != nil {
		m["certEntry"] = entry.CertEntry.GetName()
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
	entryRaw := rkentry.GlobalAppCtx.GetEntry(FiberEntryType, name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*FiberEntry)
	return entry
}

// AddMiddleware Add middleware.
// This function should be called before Bootstrap() called.
func (entry *FiberEntry) AddMiddleware(inters ...fiber.Handler) {
	entry.Middlewares = append(entry.Middlewares, inters...)
}

// SetFiberConfig override fiber config
func (entry *FiberEntry) SetFiberConfig(conf *fiber.Config) {
	entry.FiberConfig = conf
}

// IsTlsEnabled Is TLS enabled?
func (entry *FiberEntry) IsTlsEnabled() bool {
	return entry.CertEntry != nil && entry.CertEntry.Certificate != nil
}

// IsSwEnabled Is swagger entry enabled?
func (entry *FiberEntry) IsSwEnabled() bool {
	return entry.SwEntry != nil
}

// IsCommonServiceEnabled Is common service entry enabled?
func (entry *FiberEntry) IsCommonServiceEnabled() bool {
	return entry.CommonServiceEntry != nil
}

// IsDocsEnabled Is TV entry enabled?
func (entry *FiberEntry) IsDocsEnabled() bool {
	return entry.DocsEntry != nil
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

// Add basic fields into event.
func (entry *FiberEntry) logBasicInfo(operation string, ctx context.Context) (rkquery.Event, *zap.Logger) {
	event := entry.EventEntry.Start(
		operation,
		rkquery.WithEntryName(entry.GetName()),
		rkquery.WithEntryType(entry.GetType()))

	// extract eventId if exists
	if val := ctx.Value("eventId"); val != nil {
		if id, ok := val.(string); ok {
			event.SetEventId(id)
		}
	}

	logger := entry.LoggerEntry.With(
		zap.String("eventId", event.GetEventId()),
		zap.String("entryName", entry.entryName),
		zap.String("entryType", entry.entryType))

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

	// add DocsEntry info
	if entry.IsDocsEnabled() {
		event.AddPayloads(
			zap.Bool("docsEnabled", true),
			zap.String("docsPath", entry.DocsEntry.Path))
	}

	// add PromEntry info
	if entry.IsPromEnabled() {
		event.AddPayloads(
			zap.Bool("promEnabled", true),
			zap.Uint64("promPort", entry.Port),
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

// ***************** Options *****************

// FiberEntryOption Fiber entry option.
type FiberEntryOption func(*FiberEntry)

// WithName provide name.
func WithName(name string) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.entryName = name
	}
}

// WithDescription provide name.
func WithDescription(description string) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.entryDescription = description
	}
}

// WithPort provide port.
func WithPort(port uint64) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.Port = port
	}
}

// WithLoggerEntry provide rkentry.LoggerEntry.
func WithLoggerEntry(zapLogger *rkentry.LoggerEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.LoggerEntry = zapLogger
	}
}

// WithEventEntry provide rkentry.EventEntry.
func WithEventEntry(eventLogger *rkentry.EventEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.EventEntry = eventLogger
	}
}

// WithCertEntry provide rkentry.CertEntry.
func WithCertEntry(certEntry *rkentry.CertEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntry provide SwEntry.
func WithSwEntry(sw *rkentry.SWEntry) FiberEntryOption {
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

// WithMiddleware provide user interceptors.
func WithMiddleware(inters ...fiber.Handler) FiberEntryOption {
	return func(entry *FiberEntry) {
		if entry.Middlewares == nil {
			entry.Middlewares = make([]fiber.Handler, 0)
		}

		entry.Middlewares = append(entry.Middlewares, inters...)
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

// WithDocsEntry provide DocsEntry.
func WithDocsEntry(tvEntry *rkentry.DocsEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.DocsEntry = tvEntry
	}
}

// WithFiberConfig provide fiber.Config.
func WithFiberConfig(conf *fiber.Config) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.FiberConfig = conf
	}
}
