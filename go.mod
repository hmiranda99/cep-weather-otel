module github.com/mateus/cep-weather-otel

go 1.22

require (
    go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.59.0
    go.opentelemetry.io/otel v1.35.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.35.0
    go.opentelemetry.io/otel/sdk v1.35.0
)
