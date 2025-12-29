package otelx

import (
    "context"
    "errors"
    "os"
    "time"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type ShutdownFunc func(context.Context) error

func InitTracer(ctx context.Context, serviceName string) (ShutdownFunc, error) {
    endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
    if endpoint == "" {
        return nil, errors.New("OTEL_EXPORTER_OTLP_ENDPOINT não definido")
    }

    // OTLP/HTTP exporter (env OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf já está no compose)
    exp, err := otlptracehttp.New(ctx,
        otlptracehttp.WithEndpointURL(endpoint),
        otlptracehttp.WithTimeout(5*time.Second),
    )
    if err != nil {
        return nil, err
    }

    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName(serviceName),
        ),
    )
    if err != nil {
        return nil, err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithResource(res),
        sdktrace.WithBatcher(exp),
    )
    otel.SetTracerProvider(tp)

    return tp.Shutdown, nil
}
