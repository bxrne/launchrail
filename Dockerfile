# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
# Only copy necessary source files, not .git, .env, testdata, or other sensitive/dev files
COPY ./*.go ./
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum
COPY ./templates ./templates
COPY ./plugins ./plugins
COPY ./config.yaml ./config.yaml
RUN CGO_ENABLED=0 GOOS=linux go build -o launchrail ./cmd/server

FROM alpine:3.19
WORKDIR /app
RUN addgroup -S launchrail && adduser -S launchrail -G launchrail
COPY --from=builder /app/launchrail ./launchrail
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/internal ./internal
COPY --from=builder /app/pkg ./pkg
COPY --from=builder /app/plugins ./plugins
COPY --from=builder /app/config.yaml ./config.yaml
RUN chown -R launchrail:launchrail /app
USER launchrail
EXPOSE 8080
ENTRYPOINT ["./launchrail"]
