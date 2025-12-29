# Sistema de temperatura por CEP (Serviço A + Serviço B) com OTEL + Zipkin

## Visão geral

- **Serviço A (porta 8080)**: recebe `POST /cep` com JSON `{ "cep": "29902555" }`, valida e encaminha ao Serviço B via HTTP.
- **Serviço B (porta 8081)**: orquestra:
  1) busca cidade no ViaCEP
  2) busca temperatura atual na WeatherAPI
  3) retorna `{ "city": "...", "temp_C": ..., "temp_F": ..., "temp_K": ... }`

Observabilidade:
- Tracing distribuído com **OpenTelemetry** (OTLP/HTTP) e visualização no **Zipkin**.
- Spans dedicados para **ViaCEP** e **WeatherAPI**.

---

## Requisitos atendidos

### Serviço A
- `POST /cep` com body `{ "cep": "29902555" }`
- valida 8 dígitos e string
- inválido -> **422** `invalid zipcode`
- válido -> chama Serviço B e retorna o resultado

### Serviço B
- recebe CEP válido (8 dígitos)
- inválido -> **422** `invalid zipcode`
- não encontrado -> **404** `can not find zipcode`
- sucesso -> **200** JSON com cidade + temperaturas (C/F/K)

---

## Como rodar em dev (Docker)

1) Crie um arquivo `.env` na raiz (ou exporte a variável) com sua chave da WeatherAPI:

```env
cp .env.example .env
```

```env
WEATHER_API_KEY=coloque_sua_chave_aqui
```

Como conseguir essa chave (passo a passo)

- 1️⃣ Acesse o site oficial:
👉 https://www.weatherapi.com/

- 2️⃣ Clique em Sign Up (cadastro gratuito)

- 3️⃣ Crie uma conta (pode ser com e-mail)

- 4️⃣ Após logar, vá em Dashboard

- 5️⃣ Copie o valor chamado API Key

<br>

2) Suba tudo:

```bash
docker compose up --build
```

- Serviço A: http://localhost:8080
- Serviço B: http://localhost:8081
- Zipkin UI: http://localhost:9411

---

## Como testar

### Chamada principal (via Serviço A)

```bash
curl -i -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{ "cep": "01001000" }'
```

### CEP inválido (422)
```bash
curl -i -X POST http://localhost:8080/cep \
  -H "Content-Type: application/json" \
  -d '{ "cep": "123" }'
```

### Ver traces
Abra o Zipkin: http://localhost:9411 e procure por traces dos serviços `service-a` e `service-b`.

---

## Testes automatizados

Os testes mockam ViaCEP e WeatherAPI com `httptest` (não dependem de internet).

```bash
go test ./...
```
