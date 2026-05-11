# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.20
ARG DISTROLESS_IMAGE=gcr.io/distroless/static-debian12:nonroot

# --- builder ---------------------------------------------------------------
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /src

# Cache module downloads first
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
      -trimpath \
      -buildvcs=false \
      -ldflags="-s -w -buildid= -X main.version=${VERSION}" \
      -o /out/gateway .

# --- runtime ---------------------------------------------------------------
# distroless/static:nonroot — no shell, no package manager, runs as uid 65532
FROM ${DISTROLESS_IMAGE}

WORKDIR /app

COPY --from=builder /out/gateway       /app/gateway
COPY --from=builder /src/docs          /app/docs

USER nonroot:nonroot

ENV HTTP_ADDR=:8080 \
    GIN_MODE=release \
    SWAGGER_SPEC=/app/docs/openapi.json

EXPOSE 8080

ENTRYPOINT ["/app/gateway"]
