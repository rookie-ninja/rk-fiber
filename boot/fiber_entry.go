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
	"github.com/markbates/pkger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rookie-ninja/rk-common/common"
	"github.com/rookie-ninja/rk-entry/entry"
	"github.com/rookie-ninja/rk-fiber/interceptor/auth"
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
	"github.com/rookie-ninja/rk-prom"
	"github.com/rookie-ninja/rk-query"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path"
	"path/filepath"
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

// BootConfigFiber boot config which is for fiber entry.
//
// 1: Fiber.Enabled: Enable fiber entry, default is true.
// 2: Fiber.Name: Name of fiber entry, should be unique globally.
// 3: Fiber.Port: Port of fiber entry.
// 4: Fiber.Cert.Ref: Reference of rkentry.CertEntry.
// 5: Fiber.SW: See BootConfigSW for details.
// 6: Fiber.CommonService: See BootConfigCommonService for details.
// 7: Fiber.TV: See BootConfigTv for details.
// 8: Fiber.Prom: See BootConfigProm for details.
// 9: Fiber.Interceptors.LoggingZap.Enabled: Enable zap logging interceptor.
// 10: Fiber.Interceptors.MetricsProm.Enable: Enable prometheus interceptor.
// 11: Fiber.Interceptors.auth.Enabled: Enable basic auth.
// 12: Fiber.Interceptors.auth.Basic: Credential for basic auth, scheme: <user:pass>
// 13: Fiber.Interceptors.auth.ApiKey: Credential for X-API-Key.
// 14: Fiber.Interceptors.auth.igorePrefix: List of paths that will be ignored.
// 15: Fiber.Interceptors.Extension.Enabled: Enable extension interceptor.
// 16: Fiber.Interceptors.Extension.Prefix: Prefix of extension header key.
// 17: Fiber.Interceptors.TracingTelemetry.Enabled: Enable tracing interceptor with opentelemetry.
// 18: Fiber.Interceptors.TracingTelemetry.Exporter.File.Enabled: Enable file exporter which support type of stdout and local file.
// 19: Fiber.Interceptors.TracingTelemetry.Exporter.File.OutputPath: Output path of file exporter, stdout and file path is supported.
// 20: Fiber.Interceptors.TracingTelemetry.Exporter.Jaeger.Enabled: Enable jaeger exporter.
// 21: Fiber.Interceptors.TracingTelemetry.Exporter.Jaeger.AgentEndpoint: Specify jeager agent endpoint, localhost:6832 would be used by default.
// 22: Fiber.Interceptors.RateLimit.Enabled: Enable rate limit interceptor.
// 23: Fiber.Interceptors.RateLimit.Algorithm: Algorithm of rate limiter.
// 24: Fiber.Interceptors.RateLimit.ReqPerSec: Request per second.
// 25: Fiber.Interceptors.RateLimit.Paths.path: Name of full path.
// 26: Fiber.Interceptors.RateLimit.Paths.ReqPerSec: Request per second by path.
// 27: Fiber.Interceptors.Timeout.Enabled: Enable timeout interceptor.
// 28: Fiber.Interceptors.Timeout.TimeoutMs: Timeout in milliseconds.
// 29: Fiber.Interceptors.Timeout.Paths.path: Name of full path.
// 30: Fiber.Interceptors.Timeout.Paths.TimeoutMs: Timeout in milliseconds by path.
// 31: Fiber.Logger.ZapLogger.Ref: Zap logger reference, see rkentry.ZapLoggerEntry for details.
// 32: Fiber.Logger.EventLogger.Ref: Event logger reference, see rkentry.EventLoggerEntry for details.
type BootConfigFiber struct {
	Fiber []struct {
		Enabled     bool   `yaml:"enabled" json:"enabled"`
		Name        string `yaml:"name" json:"name"`
		Port        uint64 `yaml:"port" json:"port"`
		Description string `yaml:"description" json:"description"`
		Cert        struct {
			Ref string `yaml:"ref" json:"ref"`
		} `yaml:"cert" json:"cert"`
		SW            BootConfigSw            `yaml:"sw" json:"sw"`
		CommonService BootConfigCommonService `yaml:"commonService" json:"commonService"`
		TV            BootConfigTv            `yaml:"tv" json:"tv"`
		Prom          BootConfigProm          `yaml:"prom" json:"prom"`
		Static        BootConfigStaticHandler `yaml:"static" json:"static"`
		Interceptors  struct {
			LoggingZap struct {
				Enabled                bool     `yaml:"enabled" json:"enabled"`
				ZapLoggerEncoding      string   `yaml:"zapLoggerEncoding" json:"zapLoggerEncoding"`
				ZapLoggerOutputPaths   []string `yaml:"zapLoggerOutputPaths" json:"zapLoggerOutputPaths"`
				EventLoggerEncoding    string   `yaml:"eventLoggerEncoding" json:"eventLoggerEncoding"`
				EventLoggerOutputPaths []string `yaml:"eventLoggerOutputPaths" json:"eventLoggerOutputPaths"`
			} `yaml:"loggingZap" json:"loggingZap"`
			MetricsProm struct {
				Enabled bool `yaml:"enabled" json:"enabled"`
			} `yaml:"metricsProm" json:"metricsProm"`
			Auth struct {
				Enabled      bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				Basic        []string `yaml:"basic" json:"basic"`
				ApiKey       []string `yaml:"apiKey" json:"apiKey"`
			} `yaml:"auth" json:"auth"`
			Cors struct {
				Enabled          bool     `yaml:"enabled" json:"enabled"`
				AllowOrigins     []string `yaml:"allowOrigins" json:"allowOrigins"`
				AllowCredentials bool     `yaml:"allowCredentials" json:"allowCredentials"`
				AllowHeaders     []string `yaml:"allowHeaders" json:"allowHeaders"`
				AllowMethods     []string `yaml:"allowMethods" json:"allowMethods"`
				ExposeHeaders    []string `yaml:"exposeHeaders" json:"exposeHeaders"`
				MaxAge           int      `yaml:"maxAge" json:"maxAge"`
			} `yaml:"cors" json:"cors"`
			Meta struct {
				Enabled bool   `yaml:"enabled" json:"enabled"`
				Prefix  string `yaml:"prefix" json:"prefix"`
			} `yaml:"meta" json:"meta"`
			Jwt struct {
				Enabled      bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				SigningKey   string   `yaml:"signingKey" json:"signingKey"`
				SigningKeys  []string `yaml:"signingKeys" json:"signingKeys"`
				SigningAlgo  string   `yaml:"signingAlgo" json:"signingAlgo"`
				TokenLookup  string   `yaml:"tokenLookup" json:"tokenLookup"`
				AuthScheme   string   `yaml:"authScheme" json:"authScheme"`
			} `yaml:"jwt" json:"jwt"`
			Secure struct {
				Enabled               bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix          []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				XssProtection         string   `yaml:"xssProtection" json:"xssProtection"`
				ContentTypeNosniff    string   `yaml:"contentTypeNosniff" json:"contentTypeNosniff"`
				XFrameOptions         string   `yaml:"xFrameOptions" json:"xFrameOptions"`
				HstsMaxAge            int      `yaml:"hstsMaxAge" json:"hstsMaxAge"`
				HstsExcludeSubdomains bool     `yaml:"hstsExcludeSubdomains" json:"hstsExcludeSubdomains"`
				HstsPreloadEnabled    bool     `yaml:"hstsPreloadEnabled" json:"hstsPreloadEnabled"`
				ContentSecurityPolicy string   `yaml:"contentSecurityPolicy" json:"contentSecurityPolicy"`
				CspReportOnly         bool     `yaml:"cspReportOnly" json:"cspReportOnly"`
				ReferrerPolicy        string   `yaml:"referrerPolicy" json:"referrerPolicy"`
			} `yaml:"secure" json:"secure"`
			Csrf struct {
				Enabled        bool     `yaml:"enabled" json:"enabled"`
				IgnorePrefix   []string `yaml:"ignorePrefix" json:"ignorePrefix"`
				TokenLength    int      `yaml:"tokenLength" json:"tokenLength"`
				TokenLookup    string   `yaml:"tokenLookup" json:"tokenLookup"`
				CookieName     string   `yaml:"cookieName" json:"cookieName"`
				CookieDomain   string   `yaml:"cookieDomain" json:"cookieDomain"`
				CookiePath     string   `yaml:"cookiePath" json:"cookiePath"`
				CookieMaxAge   int      `yaml:"cookieMaxAge" json:"cookieMaxAge"`
				CookieHttpOnly bool     `yaml:"cookieHttpOnly" json:"cookieHttpOnly"`
				CookieSameSite string   `yaml:"cookieSameSite" json:"cookieSameSite"`
			} `yaml:"csrf" yaml:"csrf"`
			RateLimit struct {
				Enabled   bool   `yaml:"enabled" json:"enabled"`
				Algorithm string `yaml:"algorithm" json:"algorithm"`
				ReqPerSec int    `yaml:"reqPerSec" json:"reqPerSec"`
				Paths     []struct {
					Path      string `yaml:"path" json:"path"`
					ReqPerSec int    `yaml:"reqPerSec" json:"reqPerSec"`
				} `yaml:"paths" json:"paths"`
			} `yaml:"rateLimit" json:"rateLimit"`
			Timeout struct {
				Enabled   bool `yaml:"enabled" json:"enabled"`
				TimeoutMs int  `yaml:"timeoutMs" json:"timeoutMs"`
				Paths     []struct {
					Path      string `yaml:"path" json:"path"`
					TimeoutMs int    `yaml:"timeoutMs" json:"timeoutMs"`
				} `yaml:"paths" json:"paths"`
			} `yaml:"timeout" json:"timeout"`
			TracingTelemetry struct {
				Enabled  bool `yaml:"enabled" json:"enabled"`
				Exporter struct {
					File struct {
						Enabled    bool   `yaml:"enabled" json:"enabled"`
						OutputPath string `yaml:"outputPath" json:"outputPath"`
					} `yaml:"file" json:"file"`
					Jaeger struct {
						Agent struct {
							Enabled bool   `yaml:"enabled" json:"enabled"`
							Host    string `yaml:"host" json:"host"`
							Port    int    `yaml:"port" json:"port"`
						} `yaml:"agent" json:"agent"`
						Collector struct {
							Enabled  bool   `yaml:"enabled" json:"enabled"`
							Endpoint string `yaml:"endpoint" json:"endpoint"`
							Username string `yaml:"username" json:"username"`
							Password string `yaml:"password" json:"password"`
						} `yaml:"collector" json:"collector"`
					} `yaml:"jaeger" json:"jaeger"`
				} `yaml:"exporter" json:"exporter"`
			} `yaml:"tracingTelemetry" json:"tracingTelemetry"`
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
	EntryName          string                    `json:"entryName" yaml:"entryName"`
	EntryType          string                    `json:"entryType" yaml:"entryType"`
	EntryDescription   string                    `json:"-" yaml:"-"`
	ZapLoggerEntry     *rkentry.ZapLoggerEntry   `json:"-" yaml:"-"`
	EventLoggerEntry   *rkentry.EventLoggerEntry `json:"-" yaml:"-"`
	Port               uint64                    `json:"port" yaml:"port"`
	CertEntry          *rkentry.CertEntry        `json:"-" yaml:"-"`
	SwEntry            *SwEntry                  `json:"-" yaml:"-"`
	CommonServiceEntry *CommonServiceEntry       `json:"-" yaml:"-"`
	App                *fiber.App                `json:"-" yaml:"-"`
	FiberConfig        *fiber.Config             `json:"-" yaml:"-"`
	Interceptors       []fiber.Handler           `json:"-" yaml:"-"`
	PromEntry          *PromEntry                `json:"-" yaml:"-"`
	StaticFileEntry    *StaticFileHandlerEntry   `json:"-" yaml:"-"`
	TvEntry            *TvEntry                  `json:"-" yaml:"-"`
}

// FiberEntryOption Fiber entry option.
type FiberEntryOption func(*FiberEntry)

// WithNameFiber provide name.
func WithNameFiber(name string) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.EntryName = name
	}
}

// WithDescriptionFiber provide name.
func WithDescriptionFiber(description string) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.EntryDescription = description
	}
}

// WithPortFiber provide port.
func WithPortFiber(port uint64) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.Port = port
	}
}

// WithZapLoggerEntryFiber provide rkentry.ZapLoggerEntry.
func WithZapLoggerEntryFiber(zapLogger *rkentry.ZapLoggerEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.ZapLoggerEntry = zapLogger
	}
}

// WithEventLoggerEntryFiber provide rkentry.EventLoggerEntry.
func WithEventLoggerEntryFiber(eventLogger *rkentry.EventLoggerEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.EventLoggerEntry = eventLogger
	}
}

// WithCertEntryFiber provide rkentry.CertEntry.
func WithCertEntryFiber(certEntry *rkentry.CertEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.CertEntry = certEntry
	}
}

// WithSwEntryFiber provide SwEntry.
func WithSwEntryFiber(sw *SwEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.SwEntry = sw
	}
}

// WithCommonServiceEntryFiber provide CommonServiceEntry.
func WithCommonServiceEntryFiber(commonServiceEntry *CommonServiceEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.CommonServiceEntry = commonServiceEntry
	}
}

// WithInterceptorsFiber provide user interceptors.
func WithInterceptorsFiber(inters ...fiber.Handler) FiberEntryOption {
	return func(entry *FiberEntry) {
		if entry.Interceptors == nil {
			entry.Interceptors = make([]fiber.Handler, 0)
		}

		entry.Interceptors = append(entry.Interceptors, inters...)
	}
}

// WithPromEntryFiber provide PromEntry.
func WithPromEntryFiber(prom *PromEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.PromEntry = prom
	}
}

// WithStaticFileHandlerEntryFiber provide StaticFileHandlerEntry.
func WithStaticFileHandlerEntryFiber(staticEntry *StaticFileHandlerEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.StaticFileEntry = staticEntry
	}
}

// WithTVEntryFiber provide TvEntry.
func WithTVEntryFiber(tvEntry *TvEntry) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.TvEntry = tvEntry
	}
}

// WithFiberConfigFiber provide fiber.Config.
func WithFiberConfigFiber(conf *fiber.Config) FiberEntryOption {
	return func(entry *FiberEntry) {
		entry.FiberConfig = conf
	}
}

// GetFiberEntry Get FiberEntry from rkentry.GlobalAppCtx.
func GetFiberEntry(name string) *FiberEntry {
	entryRaw := rkentry.GlobalAppCtx.GetEntry(name)
	if entryRaw == nil {
		return nil
	}

	entry, _ := entryRaw.(*FiberEntry)
	return entry
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
	config := &BootConfigFiber{}
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

		promRegistry := prometheus.NewRegistry()
		// Did we enabled swagger?
		var swEntry *SwEntry
		if element.SW.Enabled {
			// Init swagger custom headers from config
			headers := make(map[string]string, 0)
			for i := range element.SW.Headers {
				header := element.SW.Headers[i]
				tokens := strings.Split(header, ":")
				if len(tokens) == 2 {
					headers[tokens[0]] = tokens[1]
				}
			}

			swEntry = NewSwEntry(
				WithNameSw(fmt.Sprintf("%s-sw", element.Name)),
				WithZapLoggerEntrySw(zapLoggerEntry),
				WithEventLoggerEntrySw(eventLoggerEntry),
				WithEnableCommonServiceSw(element.CommonService.Enabled),
				WithPortSw(element.Port),
				WithPathSw(element.SW.Path),
				WithJsonPathSw(element.SW.JsonPath),
				WithHeadersSw(headers))
		}

		// Did we enabled prometheus?
		var promEntry *PromEntry
		if element.Prom.Enabled {
			var pusher *rkprom.PushGatewayPusher
			if element.Prom.Pusher.Enabled {
				certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Prom.Pusher.Cert.Ref)
				var certStore *rkentry.CertStore

				if certEntry != nil {
					certStore = certEntry.Store
				}

				pusher, _ = rkprom.NewPushGatewayPusher(
					rkprom.WithIntervalMSPusher(time.Duration(element.Prom.Pusher.IntervalMs)*time.Millisecond),
					rkprom.WithRemoteAddressPusher(element.Prom.Pusher.RemoteAddress),
					rkprom.WithJobNamePusher(element.Prom.Pusher.JobName),
					rkprom.WithBasicAuthPusher(element.Prom.Pusher.BasicAuth),
					rkprom.WithZapLoggerEntryPusher(zapLoggerEntry),
					rkprom.WithEventLoggerEntryPusher(eventLoggerEntry),
					rkprom.WithCertStorePusher(certStore))
			}

			promRegistry.Register(prometheus.NewGoCollector())
			promEntry = NewPromEntry(
				WithNameProm(fmt.Sprintf("%s-prom", element.Name)),
				WithPortProm(element.Port),
				WithPathProm(element.Prom.Path),
				WithZapLoggerEntryProm(zapLoggerEntry),
				WithPromRegistryProm(promRegistry),
				WithEventLoggerEntryProm(eventLoggerEntry),
				WithPusherProm(pusher))

			if promEntry.Pusher != nil {
				promEntry.Pusher.SetGatherer(promEntry.Gatherer)
			}
		}

		inters := make([]fiber.Handler, 0)

		// Did we enabled logging interceptor?
		if element.Interceptors.LoggingZap.Enabled {
			opts := []rkfiberlog.Option{
				rkfiberlog.WithEntryNameAndType(element.Name, FiberEntryType),
				rkfiberlog.WithEventLoggerEntry(eventLoggerEntry),
				rkfiberlog.WithZapLoggerEntry(zapLoggerEntry),
			}

			if strings.ToLower(element.Interceptors.LoggingZap.ZapLoggerEncoding) == "json" {
				opts = append(opts, rkfiberlog.WithZapLoggerEncoding(rkfiberlog.ENCODING_JSON))
			}

			if strings.ToLower(element.Interceptors.LoggingZap.EventLoggerEncoding) == "json" {
				opts = append(opts, rkfiberlog.WithEventLoggerEncoding(rkfiberlog.ENCODING_JSON))
			}

			if len(element.Interceptors.LoggingZap.ZapLoggerOutputPaths) > 0 {
				opts = append(opts, rkfiberlog.WithZapLoggerOutputPaths(element.Interceptors.LoggingZap.ZapLoggerOutputPaths...))
			}

			if len(element.Interceptors.LoggingZap.EventLoggerOutputPaths) > 0 {
				opts = append(opts, rkfiberlog.WithEventLoggerOutputPaths(element.Interceptors.LoggingZap.EventLoggerOutputPaths...))
			}

			inters = append(inters, rkfiberlog.Interceptor(opts...))
		}

		// Did we enabled metrics interceptor?
		if element.Interceptors.MetricsProm.Enabled {
			opts := []rkfibermetrics.Option{
				rkfibermetrics.WithRegisterer(promRegistry),
				rkfibermetrics.WithEntryNameAndType(element.Name, FiberEntryType),
			}

			inters = append(inters, rkfibermetrics.Interceptor(opts...))
		}

		// Did we enabled tracing interceptor?
		if element.Interceptors.TracingTelemetry.Enabled {
			var exporter trace.SpanExporter

			if element.Interceptors.TracingTelemetry.Exporter.File.Enabled {
				exporter = rkfibertrace.CreateFileExporter(element.Interceptors.TracingTelemetry.Exporter.File.OutputPath)
			}

			if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Enabled {
				opts := make([]jaeger.AgentEndpointOption, 0)
				if len(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Host) > 0 {
					opts = append(opts,
						jaeger.WithAgentHost(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Host))
				}
				if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Port > 0 {
					opts = append(opts,
						jaeger.WithAgentPort(
							fmt.Sprintf("%d", element.Interceptors.TracingTelemetry.Exporter.Jaeger.Agent.Port)))
				}

				exporter = rkfibertrace.CreateJaegerExporter(jaeger.WithAgentEndpoint(opts...))
			}

			if element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Enabled {
				opts := []jaeger.CollectorEndpointOption{
					jaeger.WithUsername(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Username),
					jaeger.WithPassword(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Password),
				}

				if len(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Endpoint) > 0 {
					opts = append(opts, jaeger.WithEndpoint(element.Interceptors.TracingTelemetry.Exporter.Jaeger.Collector.Endpoint))
				}

				exporter = rkfibertrace.CreateJaegerExporter(jaeger.WithCollectorEndpoint(opts...))
			}

			opts := []rkfibertrace.Option{
				rkfibertrace.WithEntryNameAndType(element.Name, FiberEntryType),
				rkfibertrace.WithExporter(exporter),
			}

			inters = append(inters, rkfibertrace.Interceptor(opts...))
		}

		// Did we enabled jwt interceptor?
		if element.Interceptors.Jwt.Enabled {
			var signingKey []byte
			if len(element.Interceptors.Jwt.SigningKey) > 0 {
				signingKey = []byte(element.Interceptors.Jwt.SigningKey)
			}

			opts := []rkfiberjwt.Option{
				rkfiberjwt.WithEntryNameAndType(element.Name, FiberEntryType),
				rkfiberjwt.WithSigningKey(signingKey),
				rkfiberjwt.WithSigningAlgorithm(element.Interceptors.Jwt.SigningAlgo),
				rkfiberjwt.WithTokenLookup(element.Interceptors.Jwt.TokenLookup),
				rkfiberjwt.WithAuthScheme(element.Interceptors.Jwt.AuthScheme),
				rkfiberjwt.WithIgnorePrefix(element.Interceptors.Jwt.IgnorePrefix...),
			}

			for _, v := range element.Interceptors.Jwt.SigningKeys {
				tokens := strings.SplitN(v, ":", 2)
				if len(tokens) == 2 {
					opts = append(opts, rkfiberjwt.WithSigningKeys(tokens[0], tokens[1]))
				}
			}

			inters = append(inters, rkfiberjwt.Interceptor(opts...))
		}

		// Did we enabled secure interceptor?
		if element.Interceptors.Secure.Enabled {
			opts := []rkfibersec.Option{
				rkfibersec.WithEntryNameAndType(element.Name, FiberEntryType),
				rkfibersec.WithXSSProtection(element.Interceptors.Secure.XssProtection),
				rkfibersec.WithContentTypeNosniff(element.Interceptors.Secure.ContentTypeNosniff),
				rkfibersec.WithXFrameOptions(element.Interceptors.Secure.XFrameOptions),
				rkfibersec.WithHSTSMaxAge(element.Interceptors.Secure.HstsMaxAge),
				rkfibersec.WithHSTSExcludeSubdomains(element.Interceptors.Secure.HstsExcludeSubdomains),
				rkfibersec.WithHSTSPreloadEnabled(element.Interceptors.Secure.HstsPreloadEnabled),
				rkfibersec.WithContentSecurityPolicy(element.Interceptors.Secure.ContentSecurityPolicy),
				rkfibersec.WithCSPReportOnly(element.Interceptors.Secure.CspReportOnly),
				rkfibersec.WithReferrerPolicy(element.Interceptors.Secure.ReferrerPolicy),
				rkfibersec.WithIgnorePrefix(element.Interceptors.Secure.IgnorePrefix...),
			}

			inters = append(inters, rkfibersec.Interceptor(opts...))
		}

		// Did we enabled csrf interceptor?
		if element.Interceptors.Csrf.Enabled {
			opts := []rkfibercsrf.Option{
				rkfibercsrf.WithEntryNameAndType(element.Name, FiberEntryType),
				rkfibercsrf.WithTokenLength(element.Interceptors.Csrf.TokenLength),
				rkfibercsrf.WithTokenLookup(element.Interceptors.Csrf.TokenLookup),
				rkfibercsrf.WithCookieName(element.Interceptors.Csrf.CookieName),
				rkfibercsrf.WithCookieDomain(element.Interceptors.Csrf.CookieDomain),
				rkfibercsrf.WithCookiePath(element.Interceptors.Csrf.CookiePath),
				rkfibercsrf.WithCookieMaxAge(element.Interceptors.Csrf.CookieMaxAge),
				rkfibercsrf.WithCookieHTTPOnly(element.Interceptors.Csrf.CookieHttpOnly),
				rkfibercsrf.WithIgnorePrefix(element.Interceptors.Csrf.IgnorePrefix...),
			}

			// convert to string to cookie same sites
			sameSite := http.SameSiteDefaultMode

			switch strings.ToLower(element.Interceptors.Csrf.CookieSameSite) {
			case "lax":
				sameSite = http.SameSiteLaxMode
			case "strict":
				sameSite = http.SameSiteStrictMode
			case "none":
				sameSite = http.SameSiteNoneMode
			default:
				sameSite = http.SameSiteDefaultMode
			}

			opts = append(opts, rkfibercsrf.WithCookieSameSite(sameSite))

			inters = append(inters, rkfibercsrf.Interceptor(opts...))
		}

		// Did we enabled cors interceptor?
		if element.Interceptors.Cors.Enabled {
			opts := []rkfibercors.Option{
				rkfibercors.WithEntryNameAndType(element.Name, FiberEntryType),
				rkfibercors.WithAllowOrigins(element.Interceptors.Cors.AllowOrigins...),
				rkfibercors.WithAllowCredentials(element.Interceptors.Cors.AllowCredentials),
				rkfibercors.WithExposeHeaders(element.Interceptors.Cors.ExposeHeaders...),
				rkfibercors.WithMaxAge(element.Interceptors.Cors.MaxAge),
				rkfibercors.WithAllowHeaders(element.Interceptors.Cors.AllowHeaders...),
				rkfibercors.WithAllowMethods(element.Interceptors.Cors.AllowMethods...),
			}

			inters = append(inters, rkfibercors.Interceptor(opts...))
		}

		// Did we enabled meta interceptor?
		if element.Interceptors.Meta.Enabled {
			opts := []rkfibermeta.Option{
				rkfibermeta.WithEntryNameAndType(element.Name, FiberEntryType),
				rkfibermeta.WithPrefix(element.Interceptors.Meta.Prefix),
			}

			inters = append(inters, rkfibermeta.Interceptor(opts...))
		}

		// Did we enabled auth interceptor?
		if element.Interceptors.Auth.Enabled {
			opts := make([]rkfiberauth.Option, 0)
			opts = append(opts,
				rkfiberauth.WithEntryNameAndType(element.Name, FiberEntryType),
				rkfiberauth.WithBasicAuth(element.Name, element.Interceptors.Auth.Basic...),
				rkfiberauth.WithApiKeyAuth(element.Interceptors.Auth.ApiKey...))

			// Add exceptional path
			if swEntry != nil {
				opts = append(opts, rkfiberauth.WithIgnorePrefix(strings.TrimSuffix(swEntry.Path, "/")))
			}

			opts = append(opts, rkfiberauth.WithIgnorePrefix("/rk/v1/assets"))
			opts = append(opts, rkfiberauth.WithIgnorePrefix(element.Interceptors.Auth.IgnorePrefix...))

			inters = append(inters, rkfiberauth.Interceptor(opts...))
		}

		// Did we enabled timeout interceptor?
		// This should be in front of rate limit interceptor since rate limit may block over the threshold of timeout.
		if element.Interceptors.Timeout.Enabled {
			opts := make([]rkfibertimeout.Option, 0)
			opts = append(opts,
				rkfibertimeout.WithEntryNameAndType(element.Name, FiberEntryType))

			timeout := time.Duration(element.Interceptors.Timeout.TimeoutMs) * time.Millisecond
			opts = append(opts, rkfibertimeout.WithTimeoutAndResp(timeout, nil))

			for i := range element.Interceptors.Timeout.Paths {
				e := element.Interceptors.Timeout.Paths[i]
				timeout := time.Duration(e.TimeoutMs) * time.Millisecond
				opts = append(opts, rkfibertimeout.WithTimeoutAndRespByPath(e.Path, timeout, nil))
			}

			inters = append(inters, rkfibertimeout.Interceptor(opts...))
		}

		// Did we enabled rate limit interceptor?
		if element.Interceptors.RateLimit.Enabled {
			opts := make([]rkfiberlimit.Option, 0)
			opts = append(opts,
				rkfiberlimit.WithEntryNameAndType(element.Name, FiberEntryType))

			if len(element.Interceptors.RateLimit.Algorithm) > 0 {
				opts = append(opts, rkfiberlimit.WithAlgorithm(element.Interceptors.RateLimit.Algorithm))
			}
			opts = append(opts, rkfiberlimit.WithReqPerSec(element.Interceptors.RateLimit.ReqPerSec))

			for i := range element.Interceptors.RateLimit.Paths {
				e := element.Interceptors.RateLimit.Paths[i]
				opts = append(opts, rkfiberlimit.WithReqPerSecByPath(e.Path, e.ReqPerSec))
			}

			inters = append(inters, rkfiberlimit.Interceptor(opts...))
		}

		// Did we enabled common service?
		var commonServiceEntry *CommonServiceEntry
		if element.CommonService.Enabled {
			commonServiceEntry = NewCommonServiceEntry(
				WithNameCommonService(fmt.Sprintf("%s-commonService", element.Name)),
				WithZapLoggerEntryCommonService(zapLoggerEntry),
				WithEventLoggerEntryCommonService(eventLoggerEntry))
		}

		// Did we enabled tv?
		var tvEntry *TvEntry
		if element.TV.Enabled {
			tvEntry = NewTvEntry(
				WithNameTv(fmt.Sprintf("%s-tv", element.Name)),
				WithZapLoggerEntryTv(zapLoggerEntry),
				WithEventLoggerEntryTv(eventLoggerEntry))
		}

		// DId we enabled static file handler?
		var staticEntry *StaticFileHandlerEntry
		if element.Static.Enabled {
			var fs http.FileSystem
			switch element.Static.SourceType {
			case "pkger":
				fs = pkger.Dir(element.Static.SourcePath)
			case "local":
				if !filepath.IsAbs(element.Static.SourcePath) {
					wd, _ := os.Getwd()
					element.Static.SourcePath = path.Join(wd, element.Static.SourcePath)
				}
				fs = http.Dir(element.Static.SourcePath)
			}

			staticEntry = NewStaticFileHandlerEntry(
				WithZapLoggerEntryStatic(zapLoggerEntry),
				WithEventLoggerEntryStatic(eventLoggerEntry),
				WithNameStatic(fmt.Sprintf("%s-static", element.Name)),
				WithPathStatic(element.Static.Path),
				WithFileSystemStatic(fs))
		}

		certEntry := rkentry.GlobalAppCtx.GetCertEntry(element.Cert.Ref)

		entry := RegisterFiberEntry(
			WithNameFiber(name),
			WithDescriptionFiber(element.Description),
			WithPortFiber(element.Port),
			WithZapLoggerEntryFiber(zapLoggerEntry),
			WithEventLoggerEntryFiber(eventLoggerEntry),
			WithCertEntryFiber(certEntry),
			WithPromEntryFiber(promEntry),
			WithTVEntryFiber(tvEntry),
			WithCommonServiceEntryFiber(commonServiceEntry),
			WithSwEntryFiber(swEntry),
			WithStaticFileHandlerEntryFiber(staticEntry),
			WithInterceptorsFiber(inters...))

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
		rkfiberpanic.WithEntryNameAndType(entry.EntryName, entry.EntryType)))

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
		// Register swagger path into Router.
		//entry.App.Get(strings.TrimSuffix(entry.SwEntry.Path, "/"), func(ctx *fiber.Ctx) error {
		//	return ctx.Redirect(entry.SwEntry.Path, http.StatusTemporaryRedirect)
		//})
		entry.App.Get(path.Join(entry.SwEntry.Path, "*"), entry.SwEntry.ConfigFileHandler())
		entry.App.Get("/rk/v1/assets/sw/*", entry.SwEntry.AssetsFileHandler())

		// Bootstrap swagger entry.
		entry.SwEntry.Bootstrap(ctx)
	}

	// Is static file handler enabled?
	if entry.IsStaticFileHandlerEnabled() {
		// Register path into Router.
		entry.App.Get(strings.TrimSuffix(entry.StaticFileEntry.Path, "/"), func(ctx *fiber.Ctx) error {
			return ctx.Redirect(entry.StaticFileEntry.Path, http.StatusTemporaryRedirect)
		})

		// Register path into Router.
		entry.App.Get(path.Join(entry.StaticFileEntry.Path, "*"), entry.StaticFileEntry.GetFileHandler())

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
		entry.App.Get("/rk/v1/healthy", entry.CommonServiceEntry.Healthy)
		entry.App.Get("/rk/v1/gc", entry.CommonServiceEntry.Gc)
		entry.App.Get("/rk/v1/info", entry.CommonServiceEntry.Info)
		entry.App.Get("/rk/v1/configs", entry.CommonServiceEntry.Configs)
		entry.App.Get("/rk/v1/apis", entry.CommonServiceEntry.Apis)
		entry.App.Get("/rk/v1/sys", entry.CommonServiceEntry.Sys)
		entry.App.Get("/rk/v1/req", entry.CommonServiceEntry.Req)
		entry.App.Get("/rk/v1/entries", entry.CommonServiceEntry.Entries)
		entry.App.Get("/rk/v1/certs", entry.CommonServiceEntry.Certs)
		entry.App.Get("/rk/v1/logs", entry.CommonServiceEntry.Logs)
		entry.App.Get("/rk/v1/deps", entry.CommonServiceEntry.Deps)
		entry.App.Get("/rk/v1/license", entry.CommonServiceEntry.License)
		entry.App.Get("/rk/v1/readme", entry.CommonServiceEntry.Readme)
		entry.App.Get("/rk/v1/git", entry.CommonServiceEntry.Git)

		// Bootstrap common service entry.
		entry.CommonServiceEntry.Bootstrap(ctx)
	}

	// Is TV enabled?
	if entry.IsTvEnabled() {
		// Bootstrap TV entry.
		entry.App.Get("/rk/v1/tv/*", entry.TvEntry.TV)
		entry.App.Get("/rk/v1/assets/tv/*", entry.TvEntry.AssetsFileHandler())

		entry.TvEntry.Bootstrap(ctx)
	}

	go func(fiberEntry *FiberEntry) {
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
	}(entry)

	entry.EventLoggerEntry.GetEventHelper().Finish(event)
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

// AddInterceptor Add interceptors.
// This function should be called before Bootstrap() called.
func (entry *FiberEntry) AddInterceptor(inters ...fiber.Handler) {
	entry.Interceptors = append(entry.Interceptors, inters...)
}

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
		zap.String("entryName", entry.EntryName),
		zap.String("entryType", entry.EntryType),
		zap.Uint64("entryPort", entry.Port),
	)

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
