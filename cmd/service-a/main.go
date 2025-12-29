package main

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
    "os"
    "time"

    "github.com/mateus/cep-weather-otel/internal/otelx"
    "github.com/mateus/cep-weather-otel/internal/validation"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
    "go.opentelemetry.io/otel"
)

type cepRequest struct {
    Cep any `json:"cep"`
}

func main() {
    port := getenv("PORT", "8080")
    serviceB := getenv("SERVICE_B_URL", "http://localhost:8081")
    serviceName := getenv("OTEL_SERVICE_NAME", "service-a")

    ctx := context.Background()
    shutdown, err := otelx.InitTracer(ctx, serviceName)
    if err != nil {
        panic(err)
    }
    defer func() { _ = shutdown(context.Background()) }()

    mux := http.NewServeMux()
    mux.Handle("/health", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    }), "health"))

    mux.Handle("/cep", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }

        body, err := io.ReadAll(r.Body)
        if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        var req cepRequest
        if err := json.Unmarshal(body, &req); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }

        // valida: string com 8 dígitos
        cepStr, ok := req.Cep.(string)
        if !ok || !validation.IsValidZipcode(cepStr) {
            w.WriteHeader(http.StatusUnprocessableEntity)
            _, _ = w.Write([]byte("invalid zipcode"))
            return
        }

        tracer := otel.Tracer("service-a")
        ctx, span := tracer.Start(r.Context(), "call_service_b")
        defer span.End()

        client := &http.Client{Timeout: 10 * time.Second}
        httpClient := otelhttp.NewTransport(http.DefaultTransport)

        bReq, err := http.NewRequestWithContext(ctx, http.MethodGet, serviceB+"/weather?cep="+cepStr, nil)
        if err != nil {
            w.WriteHeader(http.StatusBadGateway)
            return
        }
        bReq.Header.Set("Accept", "application/json")
        // instrument client
        bResp, err := (&http.Client{Timeout: 10 * time.Second, Transport: httpClient}).Do(bReq)
        if err != nil {
            w.WriteHeader(http.StatusBadGateway)
            return
        }
        defer bResp.Body.Close()

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(bResp.StatusCode)
        _, _ = io.Copy(w, bResp.Body)
        _ = client
    }), "post_cep"))

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
