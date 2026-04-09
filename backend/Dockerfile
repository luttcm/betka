## syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.23.8
ARG ALPINE_VERSION=3.21

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS build
WORKDIR /src

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker
RUN --mount=type=cache,target=/go/pkg/mod \
    go install github.com/pressly/goose/v3/cmd/goose@v3.24.1
RUN cp -r /src/migrations /out/migrations

FROM alpine:${ALPINE_VERSION} AS runtime
WORKDIR /app
RUN apk --no-cache upgrade && \
    addgroup -S app && adduser -S app -G app

COPY --from=build /out/api /app/api
COPY --from=build /out/worker /app/worker
COPY --from=build /go/bin/goose /app/goose
COPY --from=build /out/migrations /app/migrations

EXPOSE 3000
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:3000/health || exit 1

USER app:app
CMD ["/app/api"]
