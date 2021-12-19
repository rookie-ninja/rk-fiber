# Panic interceptor
In this example, we will try to create fiber server with panic interceptor enabled.

Panic interceptor will add do the bellow actions.
- Recover from panic
- Convert interface to standard rkerror.ErrorResp style of error
- Set resCode to 500
- Print stacktrace  
- Set [panic:1] into event as counters
- Add error into event

**Please make sure panic interceptor to be added at last in chain of interceptors.**

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick start](#quick-start)
  - [Code](#code)
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
import     "github.com/rookie-ninja/rk-fiber/interceptor/panic"
```
```go
    // ********************************************
    // ********** Enable interceptors *************
    // ********************************************
	interceptors := []fiber.Handler{
        rkfiberpanic.Interceptor(),
    }
```

## Example
We will enable log interceptor to monitor RPC.

### Start server
```shell script
$ go run greeter-server.go
```

### Output
- Server side log (zap & event)
```shell script
2021-12-20T04:05:20.386+0800    ERROR   panic/interceptor.go:42 panic occurs:
goroutine 69 [running]:
...
created by github.com/valyala/fasthttp.(*workerPool).getCh
        /Users/dongxuny/go/pkg/mod/github.com/valyala/fasthttp@v1.31.0/workerpool.go:194 +0x11f
        {"error": "[Internal Server Error] Panic manually!"}

```
```shell script
------------------------------------------------------------------------
endTime=2021-12-20T04:05:20.387055+08:00
startTime=2021-12-20T04:05:20.386341+08:00
elapsedNano=713907
timezone=CST
ids={"eventId":"df7befe8-1d0d-4352-ad69-b3d5fbcc69d7"}
app={"appName":"rk","appVersion":"","entryName":"fiber","entryType":"fiber"}
env={"arch":"amd64","az":"*","domain":"*","hostname":"lark.local","localIP":"192.168.101.5","os":"darwin","realm":"*","region":"*"}
payloads={"apiMethod":"GET","apiPath":"/rk/v1/greeter","apiProtocol":"http","apiQuery":"name=rk-dev","userAgent":"curl/7.64.1"}
error={"[Internal Server Error] Panic manually!":1}
counters={"panic":1}
pairs={}
timing={}
remoteAddr=127.0.0.1:59658
operation=/rk/v1/greeter
resCode=500
eventStatus=Ended
EOE
```
- Client side
```shell script
$ curl "localhost:8080/rk/v1/greeter?name=rk-dev"
{"error":{"code":500,"status":"Internal Server Error","message":"Panic manually!","details":[]}}
```

### Code
- [greeter-server.go](greeter-server.go)
