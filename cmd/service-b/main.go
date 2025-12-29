package main

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "os"
    "time"

    "github.com/mateus/cep-weather-otel/internal/otelx"
    "github.com/mateus/cep-weather-otel/internal/services"
    "github.com/mateus/cep-weather-otel/internal/validation"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
    "go.opentelemetry.io/otel"
)

func main() {
    port := getenv("PORT", "8081")
    serviceName := getenv("OTEL_SERVICE_NAME", "service-b")

    ctx := context.Background()
    shutdown, err := otelx.InitTracer(ctx, serviceName)
    if err != nil {
        panic(err)
    }
    defer func() { _ = shutdown(context.Background()) }()

    viaBase := getenv("VIA_CEP_BASE_URL", "https://viacep.com.br")
    weatherBase := getenv("WEATHER_API_BASE_URL", "https://api.weatherapi.com")
    weatherKey := os.Getenv("WEATHER_API_KEY")

    deps := services.Deps{
        ViaCEPBaseURL:   viaBase,
        WeatherBaseURL:  weatherBase,
        WeatherAPIKey:   weatherKey,
        HTTPClient:      &http.Client{Timeout: 10 * time.Second, Transport: otelhttp.NewTransport(http.DefaultTransport)},
        Tracer:          otel.Tracer("service-b"),
    }

    mux := http.NewServeMux()
    mux.Handle("/health", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    }), "health"))

    mux.Handle("/weather", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cep := r.URL.Query().Get("cep")
        if !validation.IsValidZipcode(cep) {
            w.WriteHeader(http.StatusUnprocessableEntity)
            _, _ = w.Write([]byte("invalid zipcode"))
            return
        }

        resp, code, err := services.HandleWeatherByCEP(r.Context(), deps, cep)
        if err != nil {
            w.WriteHeader(code)
            _, _ = w.Write([]byte(err.Error()))
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _ = json.NewEncoder(w).Encode(resp)
    }), "get_weather"))

    srv := &http.Server{
        Addr:              ":" + port,
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,
    }
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        panic(err)
    }
}

func getenv(k, def string) string {
    v := os.Getenv(k)
    if v == "" {
        return def
    }
    return v
}

var _ = errors.New
var _ = io.Copy
