package services

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    "time"

    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

type Deps struct {
    ViaCEPBaseURL  string
    WeatherBaseURL string
    WeatherAPIKey  string
    HTTPClient     *http.Client
    Tracer         trace.Tracer
}

type Response struct {
    City  string  `json:"city"`
    TempC float64 `json:"temp_C"`
    TempF float64 `json:"temp_F"`
    TempK float64 `json:"temp_K"`
}

var (
    ErrNotFoundZip = errors.New("can not find zipcode")
)

func HandleWeatherByCEP(ctx context.Context, deps Deps, cep string) (Response, int, error) {
    // span: viacep
    cctx, span := deps.Tracer.Start(ctx, "viacep_lookup")
    span.SetAttributes(attribute.String("cep", cep))
    city, err := lookupCityViaCEP(cctx, deps, cep)
    span.End()
    if err != nil {
        if errors.Is(err, ErrNotFoundZip) {
            return Response{}, http.StatusNotFound, ErrNotFoundZip
        }
        return Response{}, http.StatusBadGateway, err
    }

    // span: weatherapi
    wctx, span2 := deps.Tracer.Start(ctx, "weatherapi_lookup")
    span2.SetAttributes(attribute.String("city", city))
    tempC, err := lookupTempC(wctx, deps, city)
    span2.End()
    if err != nil {
        return Response{}, http.StatusBadGateway, err
    }

    return Response{
        City:  city,
        TempC: tempC,
        TempF: cToF(tempC),
        TempK: cToK(tempC),
    }, http.StatusOK, nil
}

func lookupCityViaCEP(ctx context.Context, deps Deps, cep string) (string, error) {
    u := fmt.Sprintf("%s/ws/%s/json/", deps.ViaCEPBaseURL, cep)
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
    if err != nil {
        return "", err
    }
    resp, err := deps.HTTPClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        // ViaCEP retorna 200 com {"erro": true} quando não encontra
        return "", ErrNotFoundZip
    }

    var payload struct {
        Localidade string `json:"localidade"`
        Erro       bool   `json:"erro"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
        return "", err
    }
    if payload.Erro || payload.Localidade == "" {
        return "", ErrNotFoundZip
    }
    return payload.Localidade, nil
}

func lookupTempC(ctx context.Context, deps Deps, city string) (float64, error) {
    if deps.WeatherAPIKey == "" {
        return 0, errors.New("WEATHER_API_KEY não configurada")
    }

    q := url.QueryEscape(city)
    u := fmt.Sprintf("%s/v1/current.json?key=%s&q=%s&aqi=no", deps.WeatherBaseURL, deps.WeatherAPIKey, q)

    // timeout por chamada, além do timeout do client
    cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
    defer cancel()

    req, err := http.NewRequestWithContext(cctx, http.MethodGet, u, nil)
    if err != nil {
        return 0, err
    }
    resp, err := deps.HTTPClient.Do(req)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return 0, fmt.Errorf("weatherapi status %d", resp.StatusCode)
    }

    var payload struct {
        Current struct {
            TempC float64 `json:"temp_c"`
        } `json:"current"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
        return 0, err
    }
    return payload.Current.TempC, nil
}

func cToF(c float64) float64 { return c*1.8 + 32 }
func cToK(c float64) float64 { return c + 273 }

