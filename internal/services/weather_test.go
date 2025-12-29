package services

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "go.opentelemetry.io/otel/trace/noop"
)

func TestHandleWeatherByCEP_Success(t *testing.T) {
    // mock ViaCEP
    via := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"localidade":"São Paulo"}`))
    }))
    defer via.Close()

    // mock WeatherAPI
    weather := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"current":{"temp_c":28.5}}`))
    }))
    defer weather.Close()

    deps := Deps{
        ViaCEPBaseURL:  via.URL,
        WeatherBaseURL: weather.URL,
        WeatherAPIKey:  "test-key",
        HTTPClient:     &http.Client{Timeout: 2 * time.Second},
        Tracer:         noop.NewTracerProvider().Tracer("test"),
    }

    got, code, err := HandleWeatherByCEP(context.Background(), deps, "01001000")
    if err != nil || code != http.StatusOK {
        t.Fatalf("expected success, got code=%d err=%v", code, err)
    }
    if got.City != "São Paulo" {
        t.Fatalf("expected city São Paulo, got %q", got.City)
    }
    if got.TempC != 28.5 {
        t.Fatalf("expected tempC 28.5, got %v", got.TempC)
    }
    if got.TempF != 28.5*1.8+32 {
        t.Fatalf("tempF mismatch")
    }
    if got.TempK != 28.5+273 {
        t.Fatalf("tempK mismatch")
    }
}

func TestHandleWeatherByCEP_ZipNotFound(t *testing.T) {
    via := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"erro":true}`))
    }))
    defer via.Close()

    deps := Deps{
        ViaCEPBaseURL:  via.URL,
        WeatherBaseURL: "http://example",
        WeatherAPIKey:  "x",
        HTTPClient:     &http.Client{Timeout: 2 * time.Second},
        Tracer:         noop.NewTracerProvider().Tracer("test"),
    }

    _, code, err := HandleWeatherByCEP(context.Background(), deps, "01001000")
    if err == nil || code != http.StatusNotFound {
        t.Fatalf("expected not found, got code=%d err=%v", code, err)
    }
}

