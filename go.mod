module github.com/rookie-ninja/rk-fiber

go 1.16

require (
	github.com/gofiber/adaptor/v2 v2.1.15
	github.com/gofiber/fiber/v2 v2.23.0
	github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/juju/ratelimit v1.0.1
	github.com/markbates/pkger v0.17.1
	github.com/prometheus/client_golang v1.10.0
	github.com/rookie-ninja/rk-common v1.2.3
	github.com/rookie-ninja/rk-entry v1.0.5-0.20220115151807-5a8f2f3818c2
	github.com/rookie-ninja/rk-logger v1.2.3
	github.com/rookie-ninja/rk-prom v1.1.4
	github.com/rookie-ninja/rk-query v1.2.4
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/valyala/fasthttp v1.31.0
	go.opentelemetry.io/contrib v1.3.0
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/exporters/jaeger v1.3.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.3.0
	go.opentelemetry.io/otel/sdk v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
	go.uber.org/ratelimit v0.2.0
	go.uber.org/zap v1.19.1
)
