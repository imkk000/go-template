# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.20
ARG DISTROLESS_IMAGE=gcr.io/distroless/static-debian12:nonroot

# --- builder ---------------------------------------------------------------
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

RUN apk add --no-cache ca-certificates tzdata
WORKDIR /src

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
      -o /out/server .

# --- runtime ---------------------------------------------------------------
FROM ${DISTROLESS_IMAGE}

COPY --from=builder /out/server /usr/local/bin/server

USER nonroot:nonroot

ENV GRPC_ADDR=:9090

EXPOSE 9090

ENTRYPOINT ["/usr/local/bin/server"]
