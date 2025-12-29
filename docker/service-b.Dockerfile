# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /app
RUN apk add --no-cache ca-certificates git
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service-b ./cmd/service-b

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/service-b /app/service-b
EXPOSE 8081
ENTRYPOINT ["/app/service-b"]
