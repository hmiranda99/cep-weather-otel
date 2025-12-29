# syntax=docker/dockerfile:1
FROM golang:1.22-alpine AS build
WORKDIR /app
RUN apk add --no-cache ca-certificates git
COPY go.mod ./
RUN go mod download
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/service-a ./cmd/service-a

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/service-a /app/service-a
EXPOSE 8080
ENTRYPOINT ["/app/service-a"]
