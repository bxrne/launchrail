# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o launchrail ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/launchrail ./launchrail
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/internal ./internal
COPY --from=builder /app/pkg ./pkg
COPY --from=builder /app/plugins ./plugins
COPY --from=builder /app/config.yaml ./config.yaml
EXPOSE 8080
ENTRYPOINT ["./launchrail"]
