# Log interceptor
In this example, we will try to create fiber server with log interceptor enabled.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick start](#quick-start)
  - [Code](#code)
- [Options](#options)
  - [Encoding](#encoding)
  - [OutputPath](#outputpath)
  - [Context Usage](#context-usage)
- [Example](#example)
    - [Start server](#start-server)
    - [Output](#output)
  - [Code](#code-1)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Quick start
Get rk-fiber package from the remote repository.

```go
go get -u github.com/rookie-ninja/rk-fiber
```

### Code
```go
import     "github.com/rookie-ninja/rk-fiber/interceptor/log/zap"
```

```go
	interceptors := []fiber.Handler{
        rkfiberlog.Interceptor(),
    }
```

## Options
Log interceptor will init rkquery.Event, zap.Logger and entryName which will be injected into request context before user function.
As soon as user function returns, interceptor will write the event into files.

![arch](img/arch.png)

| Name | Default | Description |
| ---- | ---- | ---- |
| WithEntryNameAndType(entryName, entryType string) | entryName=gf, entryType=gf | entryName and entryType will be used to distinguish options if there are multiple interceptors in single process. |
| WithZapLoggerEntry(zapLoggerEntry *rkentry.ZapLoggerEntry) | [rkentry.GlobalAppCtx.GetZapLoggerEntryDefault()](https://github.com/rookie-ninja/rk-entry/blob/master/entry/context.go) | Zap logger would print to stdout with console encoding type. |
| WithEventLoggerEntry(eventLoggerEntry *rkentry.EventLoggerEntry) | [rkentry.GlobalAppCtx.GetEventLoggerEntryDefault()](https://github.com/rookie-ninja/rk-entry/blob/master/entry/context.go) | Event logger would print to stdout with console encoding type. |
| WithZapLoggerEncoding(ec int) | rkfiberlog.ENCODING_CONSOLE | rkfiberlog.ENCODING_CONSOLE and rkfiberlog.ENCODING_JSON are available options. |
| WithZapLoggerOutputPaths(path ...string) | stdout | Both absolute path and relative path is acceptable. Current working directory would be used if path is relative. |
| WithEventLoggerEncoding(ec int) | rkfiberlog.ENCODING_CONSOLE | rkfiberlog.ENCODING_CONSOLE and rkfiberlog.ENCODING_JSON are available options. |
| WithEventLoggerOutputPaths(path ...string) | stdout | Both absolute path and relative path is acceptable. Current working directory would be used if path is relative. |

```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
	interceptors := []fiber.Handler{
		rkfiberlog.Interceptor(
		// Entry name and entry type will be used for distinguishing interceptors. Recommended.
		// rkfiberlog.WithEntryNameAndType("greeter", "fiber"),
		//
		// Zap logger would be logged as JSON format.
		// rkfiberlog.WithZapLoggerEncoding(rkfiberlog.ENCODING_JSON),
		//
		// Event logger would be logged as JSON format.
		// rkfiberlog.WithEventLoggerEncoding(rkfiberlog.ENCODING_JSON),
		//
		// Zap logger would be logged to specified path.
		// rkfiberlog.WithZapLoggerOutputPaths("logs/server-zap.log"),
		//
		// Event logger would be logged to specified path.
		// rkfiberlog.WithEventLoggerOutputPaths("logs/server-event.log"),
		),
	}
```

### Encoding
- CONSOLE
No options needs to be provided. 
```shell script
2021-12-20T03:30:51.108+0800    INFO    log/greeter-server.go:83        Received request from client.
```

```shell script
------------------------------------------------------------------------
endTime=2021-12-20T03:30:51.10878+08:00
startTime=2021-12-20T03:30:51.108506+08:00
elapsedNano=274229
timezone=CST
ids={"eventId":"b37dfd6d-8fa5-453e-b514-58fd12116b9f"}
app={"appName":"rk","appVersion":"","entryName":"fiber","entryType":"fiber"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.101.5","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"http","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=127.0.0.1:58400
operation=/rk/v1/greeter
resCode=200
eventStatus=Ended
EOE
```

- JSON
```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
	interceptors := []fiber.Handler{
        rkfiberlog.Interceptor(
            // Zap logger would be logged as JSON format.
            rkfiberlog.WithZapLoggerEncoding(rkfiberlog.ENCODING_JSON),
            //
            // Event logger would be logged as JSON format.
            rkfiberlog.WithEventLoggerEncoding(rkfiberlog.ENCODING_JSON),
        ),
    }
```
```json
{"level":"INFO","ts":"2021-12-20T03:31:34.755+0800","msg":"Received request from client."}
```
```json
{"endTime": "2021-12-20T03:31:34.755+0800", "startTime": "2021-12-20T03:31:34.755+0800", "elapsedNano": 238313, "timezone": "CST", "ids": {"eventId":"afd8969b-2fe4-41d6-9f91-05da691e8c4c"}, "app": {"appName":"rk","appVersion":"","entryName":"fiber","entryType":"fiber"}, "env": {"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.101.5","os":"darwin","realm":"*","region":"*"}, "payloads": {"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"http","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}, "error": {}, "counters": {}, "pairs": {}, "timing": {}, "remoteAddr": "127.0.0.1:61165", "operation": "/rk/v1/greeter", "eventStatus": "Ended", "resCode": "200"}
```

### OutputPath
- Stdout
No options needs to be provided. 

- Files
```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
	interceptors := []fiber.Handler{
        rkfiberlog.Interceptor(
            // Zap logger would be logged to specified path.
            rkfiberlog.WithZapLoggerOutputPaths("logs/server-zap.log"),
            //
            // Event logger would be logged to specified path.
            rkfiberlog.WithEventLoggerOutputPaths("logs/server-event.log"),
        ),
    }
```

### Context Usage
| Name | Functionality |
| ------ | ------ |
| rkfiberctx.GetLogger(*fiber.Ctx) | Get logger generated by log interceptor. If there are X-Request-Id or X-Trace-Id as headers in incoming and outgoing metadata, then loggers will has requestId and traceId attached by default. |
| rkfiberctx.GetEvent(*fiber.Ctx) | Get event generated by log interceptor. Event would be printed as soon as RPC finished. |
| rkfiberctx.GetIncomingHeaders(*fiber.Ctx) | Get incoming header. |
| rkfiberctx.AddHeaderToClient(ctx, "k", "v") | Add k/v to headers which would be sent to client. This is append operation. |
| rkfiberctx.SetHeaderToClient(ctx, "k", "v") | Set k/v to headers which would be sent to client. |

## Example
In this example, we enable log interceptor.

#### Start server
```shell script
$ go run greeter-server.go
```

#### Output
- Server side (zap & event)
```shell script
2021-12-20T03:32:17.949+0800	INFO	Received request from client.
```

```shell script
------------------------------------------------------------------------
endTime=2021-12-20T03:32:17.949939+08:00
startTime=2021-12-20T03:32:17.949164+08:00
elapsedNano=774449
timezone=CST
ids={"eventId":"98bd3d60-bf68-4420-8c2f-66e92bb130ed"}
app={"appName":"rk","appVersion":"","entryName":"fiber","entryType":"fiber"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.101.5","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"http","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={}
counters={}
pairs={}
timing={}
remoteAddr=127.0.0.1:63930
operation=/rk/v1/greeter
resCode=200
eventStatus=Ended
EOE
```

- Client side
```shell script
$ curl "localhost:8080/rk/v1/greeter?name=rk-dev"
{"Message":"Hello rk-dev!"}
```

### Code
- [greeter-server.go](greeter-server.go)
