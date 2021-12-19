# rk-fiber
[![build](https://github.com/rookie-ninja/rk-fiber/actions/workflows/ci.yml/badge.svg)](https://github.com/rookie-ninja/rk-fiber/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/rookie-ninja/rk-fiber/branch/master/graph/badge.svg?token=Y1HM9UQBX6)](https://codecov.io/gh/rookie-ninja/rk-fiber)
[![Go Report Card](https://goreportcard.com/badge/github.com/rookie-ninja/rk-fiber)](https://goreportcard.com/report/github.com/rookie-ninja/rk-fiber)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

> Under testing stage!

Interceptor & bootstrapper designed for [fiber](https://github.com/gofiber/fiber) framework. Currently, supports bellow functionalities.

| Name | Description |
| ---- | ---- |
| Start with YAML | Start service with YAML config. |
| Start with code | Start service from code. |
| Fiber Service | Fiber service. |
| Swagger Service | Swagger UI. |
| Common Service | List of common API available on Fiber. |
| TV Service | A Web UI shows application and environment information. |
| Static file handler | A Web UI shows files could be downloaded from server, currently support source of local and pkger. |
| Metrics interceptor | Collect RPC metrics and export as prometheus client. |
| Log interceptor | Log every RPC requests as event with rk-query. |
| Trace interceptor | Collect RPC trace and export it to stdout, file or jaeger. |
| Panic interceptor | Recover from panic for RPC requests and log it. |
| Meta interceptor | Send application metadata as header to client. |
| Auth interceptor | Support [Basic Auth] and [API Key] authorization types. |
| RateLimit interceptor | Limiting RPC rate |
| Timeout interceptor | Timing out request by configuration. |
| Gzip interceptor | Compress and Decompress message body based on request header. |
| CORS interceptor | Server side CORS interceptor. |
| JWT interceptor | Server side JWT interceptor. |
| Secure interceptor | Server side secure interceptor. |
| CSRF interceptor | Server side csrf interceptor. |

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Installation](#installation)
- [Quick Start](#quick-start)
  - [Start fiber Service](#start-fiber-service)
  - [Output](#output)
    - [fiber Service](#fiber-service)
    - [Swagger Service](#swagger-service)
    - [TV Service](#tv-service)
    - [Metrics](#metrics)
    - [Logging](#logging)
    - [Meta](#meta)
- [YAML Config](#yaml-config)
  - [fiber Service](#fiber-service-1)
  - [Common Service](#common-service)
  - [Swagger Service](#swagger-service-1)
  - [Prom Client](#prom-client)
  - [TV Service](#tv-service-1)
  - [Static file handler Service](#static-file-handler-service)
  - [Interceptors](#interceptors)
    - [Log](#log)
    - [Metrics](#metrics-1)
    - [Auth](#auth)
    - [Meta](#meta-1)
    - [Tracing](#tracing)
    - [RateLimit](#ratelimit)
    - [Timeout](#timeout)
    - [CORS](#cors)
    - [JWT](#jwt)
    - [Secure](#secure)
    - [CSRF](#csrf)
  - [Development Status: Testing](#development-status-testing)
  - [Contributing](#contributing)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Installation
`go get -u github.com/rookie-ninja/rk-fiber`

## Quick Start
Bootstrapper can be used with YAML config. In the bellow example, we will start bellow services automatically.
- Fiber Service
- Swagger Service
- Common Service
- TV Service
- Metrics
- Logging
- Meta

Please refer example at [example/boot/simple](example/boot/simple).

### Start fiber Service

- [boot.yaml](example/boot/simple/boot.yaml)

```yaml
---
fiber:
  - name: greeter                     # Required
    port: 8080                        # Required
    enabled: true                     # Required
    tv:
      enabled: true                   # Optional, default: false
    prom:
      enabled: true                   # Optional, default: false
    sw:                               # Optional
      enabled: true                   # Optional, default: false
    commonService:                    # Optional
      enabled: true                   # Optional, default: false
    interceptors:
      loggingZap:
        enabled: true
      metricsProm:
        enabled: true
      meta:
        enabled: true
```

- [main.go](example/boot/simple/main.go)

```go
func main() {
	// Bootstrap basic entries from boot config.
	rkentry.RegisterInternalEntriesFromConfig("example/boot/simple/boot.yaml")

	// Bootstrap fiber entry from boot config
	res := rkfiber.RegisterFiberEntriesWithConfig("example/boot/simple/boot.yaml")

	// Bootstrap fiber entry
	res["greeter"].Bootstrap(context.Background())

	// Wait for shutdown signal
	rkentry.GlobalAppCtx.WaitForShutdownSig()

	// Interrupt fiber entry
	res["greeter"].Interrupt(context.Background())
}
```

```go
$ go run main.go
```

### Output
#### fiber Service
Try to test fiber Service with [curl](https://curl.se/)
```shell script
# Curl to common service
$ curl localhost:8080/rk/v1/healthy
{"healthy":true}
```

#### Swagger Service
By default, we could access swagger UI at [/sw].
- http://localhost:8080/sw

![sw](docs/img/simple-sw.png)

#### TV Service
By default, we could access TV at [/tv].

![tv](docs/img/simple-tv.png)

#### Metrics
By default, we could access prometheus client at [/metrics]
- http://localhost:8080/metrics

![prom](docs/img/simple-prom.png)

#### Logging
By default, we enable zap logger and event logger with console encoding type.
```shell script
2021-11-01T23:28:14.555+0800    INFO    boot/sw_entry.go:201    Bootstrapping SwEntry.  {"eventId": "29b411e0-7e44-4e67-aeb3-95682d2b0a2d", "entryName": "greeter-sw", "entryType": "SwEntry", "jsonPath": "", "path": "/sw/", "port": 8080}
2021-11-01T23:28:14.555+0800    INFO    boot/prom_entry.go:207  Bootstrapping promEntry.        {"eventId": "29b411e0-7e44-4e67-aeb3-95682d2b0a2d", "entryName": "greeter-prom", "entryType": "PromEntry", "entryDescription": "Internal RK entry which implements prometheus client with Fiber framework.", "path": "/metrics", "port": 8080}
2021-11-01T23:28:14.555+0800    INFO    boot/common_service_entry.go:156        Bootstrapping CommonServiceEntry.       {"eventId": "29b411e0-7e44-4e67-aeb3-95682d2b0a2d", "entryName": "greeter-commonService", "entryType": "CommonServiceEntry"}
2021-11-01T23:28:14.557+0800    INFO    boot/tv_entry.go:213    Bootstrapping tvEntry.  {"eventId": "29b411e0-7e44-4e67-aeb3-95682d2b0a2d", "entryName": "greeter-tv", "entryType": "TvEntry", "path": "/rk/v1/tv/*item"}
2021-11-01T23:28:14.557+0800    INFO    boot/fiber_entry.go:694  Bootstrapping FiberEntry.        {"eventId": "29b411e0-7e44-4e67-aeb3-95682d2b0a2d", "entryName": "greeter", "entryType": "FiberEntry", "port": 8080}
```
```shell script
------------------------------------------------------------------------
endTime=2021-11-01T23:28:14.555288+08:00
startTime=2021-11-01T23:28:14.555234+08:00
elapsedNano=54231
timezone=CST
ids={"eventId":"29b411e0-7e44-4e67-aeb3-95682d2b0a2d"}
app={"appName":"rk-fiber","appVersion":"master-e4538d7","entryName":"greeter-sw","entryType":"SwEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.1.104","os":"darwin","realm":"*","region":"*"}
payloads={"entryName":"greeter-sw","entryType":"SwEntry","jsonPath":"","path":"/sw/","port":8080}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost
operation=bootstrap
resCode=OK
eventStatus=Ended
EOE
...
------------------------------------------------------------------------
endTime=2021-11-01T23:28:14.55714+08:00
startTime=2021-11-01T23:28:14.555172+08:00
elapsedNano=1968399
timezone=CST
ids={"eventId":"29b411e0-7e44-4e67-aeb3-95682d2b0a2d"}
app={"appName":"rk-fiber","appVersion":"master-e4538d7","entryName":"greeter","entryType":"FiberEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.1.104","os":"darwin","realm":"*","region":"*"}
payloads={"entryName":"greeter","entryType":"FiberEntry","port":8080}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost
operation=bootstrap
resCode=OK
eventStatus=Ended
EOE
```

#### Meta
By default, we will send back some metadata to client including gateway with headers.
```shell script
$ curl -vs localhost:8080/rk/v1/healthy
...
< HTTP/1.1 200 OK
< Content-Type: application/json; charset=utf-8
< X-Request-Id: 3332e575-43d8-4bfe-84dd-45b5fc5fb104
< X-Rk-App-Name: rk-fiber
< X-Rk-App-Unix-Time: 2021-06-25T01:30:45.143869+08:00
< X-Rk-App-Version: master-xxx
< X-Rk-Received-Time: 2021-06-25T01:30:45.143869+08:00
< X-Trace-Id: 65b9aa7a9705268bba492fdf4a0e5652
< Date: Thu, 24 Jun 2021 17:30:45 GMT
...
```

## YAML Config
Available configuration
User can start multiple fiber servers at the same time. Please make sure use different port and name.

### fiber Service
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.name | The name of fiber server | string | N/A |
| fiber.port | The port of fiber server | integer | nil, server won't start |
| fiber.enabled | Enable fiber entry or not | bool | false |
| fiber.description | Description of fiber entry. | string | "" |
| fiber.cert.ref | Reference of cert entry declared in [cert entry](https://github.com/rookie-ninja/rk-entry#certentry) | string | "" |
| fiber.logger.zapLogger.ref | Reference of zapLoggerEntry declared in [zapLoggerEntry](https://github.com/rookie-ninja/rk-entry#zaploggerentry) | string | "" |
| fiber.logger.eventLogger.ref | Reference of eventLoggerEntry declared in [eventLoggerEntry](https://github.com/rookie-ninja/rk-entry#eventloggerentry) | string | "" |

### Common Service
| Path | Description |
| ---- | ---- |
| /rk/v1/apis | List APIs in current FiberEntry. |
| /rk/v1/certs | List CertEntry. |
| /rk/v1/configs | List ConfigEntry. |
| /rk/v1/deps | List dependencies related application, entire contents of go.mod file would be returned. |
| /rk/v1/entries | List all Entries. |
| /rk/v1/gc | Trigger GC |
| /rk/v1/healthy | Get application healthy status. |
| /rk/v1/info | Get application and process info. |
| /rk/v1/license | Get license related application, entire contents of LICENSE file would be returned. |
| /rk/v1/logs | List logger related entries. |
| /rk/v1/git | Get git information. |
| /rk/v1/readme | Get contents of README file. |
| /rk/v1/req | List prometheus metrics of requests. |
| /rk/v1/sys | Get OS stat. |
| /rk/v1/tv | Get HTML page of /tv. |

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.commonService.enabled | Enable embedded common service | boolean | false |

### Swagger Service
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.sw.enabled | Enable swagger service over fiber server | boolean | false |
| fiber.sw.path | The path access swagger service from web | string | /sw |
| fiber.sw.jsonPath | Where the swagger.json files are stored locally | string | "" |
| fiber.sw.headers | Headers would be sent to caller as scheme of [key:value] | []string | [] |

### Prom Client
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.prom.enabled | Enable prometheus | boolean | false |
| fiber.prom.path | Path of prometheus | string | /metrics |
| fiber.prom.pusher.enabled | Enable prometheus pusher | bool | false |
| fiber.prom.pusher.jobName | Job name would be attached as label while pushing to remote pushgateway | string | "" |
| fiber.prom.pusher.remoteAddress | PushGateWay address, could be form of http://x.x.x.x or x.x.x.x | string | "" |
| fiber.prom.pusher.intervalMs | Push interval in milliseconds | string | 1000 |
| fiber.prom.pusher.basicAuth | Basic auth used to interact with remote pushgateway, form of [user:pass] | string | "" |
| fiber.prom.pusher.cert.ref | Reference of rkentry.CertEntry | string | "" |

### TV Service
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.tv.enabled | Enable RK TV | boolean | false |

### Static file handler Service
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.static.enabled | Optional, Enable static file handler | boolean | false |
| fiber.static.path | Optional, path of static file handler | string | /rk/v1/static |
| fiber.static.sourceType | Required, local and pkger supported | string | "" |
| fiber.static.sourcePath | Required, full path of source directory | string | "" |

- About [pkger](https://github.com/markbates/pkger)
User can use pkger command line tool to embed static files into .go files.

Please use sourcePath like: github.com/rookie-ninja/rk-fiber:/boot/assets

### Interceptors
#### Log
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.loggingZap.enabled | Enable log interceptor | boolean | false |
| fiber.interceptors.loggingZap.zapLoggerEncoding | json or console | string | console |
| fiber.interceptors.loggingZap.zapLoggerOutputPaths | Output paths | []string | stdout |
| fiber.interceptors.loggingZap.eventLoggerEncoding | json or console | string | console |
| fiber.interceptors.loggingZap.eventLoggerOutputPaths | Output paths | []string | false |

We will log two types of log for every RPC call.
- zapLogger

Contains user printed logging with requestId or traceId.

- eventLogger

Contains per RPC metadata, response information, environment information and etc.

| Field | Description |
| ---- | ---- |
| endTime | As name described |
| startTime | As name described |
| elapsedNano | Elapsed time for RPC in nanoseconds |
| timezone | As name described |
| ids | Contains three different ids(eventId, requestId and traceId). If meta interceptor was enabled or event.SetRequestId() was called by user, then requestId would be attached. eventId would be the same as requestId if meta interceptor was enabled. If trace interceptor was enabled, then traceId would be attached. |
| app | Contains [appName, appVersion](https://github.com/rookie-ninja/rk-entry#appinfoentry), entryName, entryType. |
| env | Contains arch, az, domain, hostname, localIP, os, realm, region. realm, region, az, domain were retrieved from environment variable named as REALM, REGION, AZ and DOMAIN. "*" means empty environment variable.|
| payloads | Contains RPC related metadata |
| error | Contains errors if occur |
| counters | Set by calling event.SetCounter() by user. |
| pairs | Set by calling event.AddPair() by user. |
| timing | Set by calling event.StartTimer() and event.EndTimer() by user. |
| remoteAddr |  As name described |
| operation | RPC method name |
| resCode | Response code of RPC |
| eventStatus | Ended or InProgress |

- example

```shell script
------------------------------------------------------------------------
endTime=2021-11-01T23:31:01.706614+08:00
startTime=2021-11-01T23:31:01.706335+08:00
elapsedNano=278966
timezone=CST
ids={"eventId":"61cae46e-ea98-47b5-8a39-1090d015e09a","requestId":"61cae46e-ea98-47b5-8a39-1090d015e09a"}
app={"appName":"rk-fiber","appVersion":"master-e4538d7","entryName":"greeter","entryType":"FiberEntry"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.1.104","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/healthy","apiProtocol":"HTTP/1.1","apiQuery":"","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=localhost:54376
operation=/rk/v1/healthy
resCode=200
eventStatus=Ended
EOE
```

#### Metrics
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.metricsProm.enabled | Enable metrics interceptor | boolean | false |

#### Auth
Enable the server side auth. codes.Unauthenticated would be returned to client if not authorized with user defined credential.

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.auth.enabled | Enable auth interceptor | boolean | false |
| fiber.interceptors.auth.basic | Basic auth credentials as scheme of <user:pass> | []string | [] |
| fiber.interceptors.auth.apiKey | API key auth | []string | [] |
| fiber.interceptors.auth.ignorePrefix | The paths of prefix that will be ignored by interceptor | []string | [] |

#### Meta
Send application metadata as header to client.

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.meta.enabled | Enable meta interceptor | boolean | false |
| fiber.interceptors.meta.prefix | Header key was formed as X-<Prefix>-XXX | string | RK |

#### Tracing
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.tracingTelemetry.enabled | Enable tracing interceptor | boolean | false |
| fiber.interceptors.tracingTelemetry.exporter.file.enabled | Enable file exporter | boolean | RK |
| fiber.interceptors.tracingTelemetry.exporter.file.outputPath | Export tracing info to files | string | stdout |
| fiber.interceptors.tracingTelemetry.exporter.jaeger.agent.enabled | Export tracing info to jaeger agent | boolean | false |
| fiber.interceptors.tracingTelemetry.exporter.jaeger.agent.host | As name described | string | localhost |
| fiber.interceptors.tracingTelemetry.exporter.jaeger.agent.port | As name described | int | 6831 |
| fiber.interceptors.tracingTelemetry.exporter.jaeger.collector.enabled | Export tracing info to jaeger collector | boolean | false |
| fiber.interceptors.tracingTelemetry.exporter.jaeger.collector.endpoint | As name described | string | http://localhost:16368/api/trace |
| fiber.interceptors.tracingTelemetry.exporter.jaeger.collector.username | As name described | string | "" |
| fiber.interceptors.tracingTelemetry.exporter.jaeger.collector.password | As name described | string | "" |

#### RateLimit
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.rateLimit.enabled | Enable rate limit interceptor | boolean | false |
| fiber.interceptors.rateLimit.algorithm | Provide algorithm, tokenBucket and leakyBucket are available options | string | tokenBucket |
| fiber.interceptors.rateLimit.reqPerSec | Request per second globally | int | 0 |
| fiber.interceptors.rateLimit.paths.path | Full path | string | "" |
| fiber.interceptors.rateLimit.paths.reqPerSec | Request per second by full path | int | 0 |

#### Timeout
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.timeout.enabled | Enable timeout interceptor | boolean | false |
| fiber.interceptors.timeout.timeoutMs | Global timeout in milliseconds. | int | 5000 |
| fiber.interceptors.timeout.paths.path | Full path | string | "" |
| fiber.interceptors.timeout.paths.timeoutMs | Timeout in milliseconds by full path | int | 5000 |

#### CORS
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.cors.enabled | Enable cors interceptor | boolean | false |
| fiber.interceptors.cors.allowOrigins | Provide allowed origins with wildcard enabled. | []string | * |
| fiber.interceptors.cors.allowMethods | Provide allowed methods returns as response header of OPTIONS request. | []string | All http methods |
| fiber.interceptors.cors.allowHeaders | Provide allowed headers returns as response header of OPTIONS request. | []string | Headers from request |
| fiber.interceptors.cors.allowCredentials | Returns as response header of OPTIONS request. | bool | false |
| fiber.interceptors.cors.exposeHeaders | Provide exposed headers returns as response header of OPTIONS request. | []string | "" |
| fiber.interceptors.cors.maxAge | Provide max age returns as response header of OPTIONS request. | int | 0 |

#### JWT
In order to make swagger UI and RK tv work under JWT without JWT token, we need to ignore prefixes of paths as bellow.

```yaml
jwt:
  ...
  ignorePrefix:
   - "/rk/v1/tv"
   - "/sw"
   - "/rk/v1/assets"
```

| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.jwt.enabled | Enable JWT interceptor | boolean | false |
| fiber.interceptors.jwt.signingKey | Required, Provide signing key. | string | "" |
| fiber.interceptors.jwt.ignorePrefix | Provide ignoring path prefix. | []string | [] |
| fiber.interceptors.jwt.signingKeys | Provide signing keys as scheme of <key>:<value>. | []string | [] |
| fiber.interceptors.jwt.signingAlgo | Provide signing algorithm. | string | HS256 |
| fiber.interceptors.jwt.tokenLookup | Provide token lookup scheme, please see bellow description. | string | "header:Authorization" |
| fiber.interceptors.jwt.authScheme | Provide auth scheme. | string | Bearer |

The supported scheme of **tokenLookup** 

```
// Optional. Default value "header:Authorization".
// Possible values:
// - "header:<name>"
// - "query:<name>"
// - "cookie:<name>"
// Multiply sources example:
// - "header: Authorization,cookie: myowncookie"
```

#### Secure
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.secure.enabled | Enable secure interceptor | boolean | false |
| fiber.interceptors.secure.xssProtection | X-XSS-Protection header value. | string | "1; mode=block" |
| fiber.interceptors.secure.contentTypeNosniff | X-Content-Type-Options header value. | string | nosniff |
| fiber.interceptors.secure.xFrameOptions | X-Frame-Options header value. | string | SAMEORIGIN |
| fiber.interceptors.secure.hstsMaxAge | Strict-Transport-Security header value. | int | 0 |
| fiber.interceptors.secure.hstsExcludeSubdomains | Excluding subdomains of HSTS. | bool | false |
| fiber.interceptors.secure.hstsPreloadEnabled | Enabling HSTS preload. | bool | false |
| fiber.interceptors.secure.contentSecurityPolicy | Content-Security-Policy header value. | string | "" |
| fiber.interceptors.secure.cspReportOnly | Content-Security-Policy-Report-Only header value. | bool | false |
| fiber.interceptors.secure.referrerPolicy | Referrer-Policy header value. | string | "" |
| fiber.interceptors.secure.ignorePrefix | Ignoring path prefix. | []string | [] |

#### CSRF
| name | description | type | default value |
| ------ | ------ | ------ | ------ |
| fiber.interceptors.csrf.enabled | Enable csrf interceptor | boolean | false |
| fiber.interceptors.csrf.tokenLength | Provide the length of the generated token. | int | 32 |
| fiber.interceptors.csrf.tokenLookup | Provide csrf token lookup rules, please see code comments for details. | string | "header:X-CSRF-Token" |
| fiber.interceptors.csrf.cookieName | Provide name of the CSRF cookie. This cookie will store CSRF token. | string | _csrf |
| fiber.interceptors.csrf.cookieDomain | Domain of the CSRF cookie. | string | "" |
| fiber.interceptors.csrf.cookiePath | Path of the CSRF cookie. | string | "" |
| fiber.interceptors.csrf.cookieMaxAge | Provide max age (in seconds) of the CSRF cookie. | int | 86400 |
| fiber.interceptors.csrf.cookieHttpOnly | Indicates if CSRF cookie is HTTP only. | bool | false |
| fiber.interceptors.csrf.cookieSameSite | Indicates SameSite mode of the CSRF cookie. Options: lax, strict, none, default | string | default |
| fiber.interceptors.csrf.ignorePrefix | Ignoring path prefix. | []string | [] |

### Development Status: Testing

### Contributing
We encourage and support an active, healthy community of contributors &mdash;
including you! Details are in the [contribution guide](CONTRIBUTING.md) and
the [code of conduct](CODE_OF_CONDUCT.md). The rk maintainers keep an eye on
issues and pull requests, but you can also report any negative conduct to
lark@rkdev.info.

Released under the [Apache 2.0 License](LICENSE).

