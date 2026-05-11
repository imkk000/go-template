# gateway — Gin → gRPC reverse-proxy template

Gin-based HTTP/JSON API gateway that reverse-proxies into a gRPC backend.
Uses Go 1.26 `tool` directive for tool pinning.

## Scaffold a new project with `gonew`

```bash
gonew github.com/imkk000/go-template/gateway my.org/my-gateway
```

This clones the template and rewrites the `gateway` module name (and every
`gateway/internal/...` import) to `my.org/my-gateway`.

## Features

| Concern | Library |
| --- | --- |
| HTTP router | `github.com/gin-gonic/gin` |
| gRPC client | `google.golang.org/grpc` (+ `protojson` for JSON ↔ proto) |
| Logging | `github.com/rs/zerolog` (request log + structured app log) |
| CORS | `github.com/gin-contrib/cors` |
| Auth middleware | `github.com/golang-jwt/jwt/v5` (HS256) **or** remote auth API |
| Swagger UI | `github.com/swaggest/swgui` (embedded, no CDN) |
| Env loading | `github.com/joho/godotenv` |

## Layout

```
gateway/
├── go.mod                 # module + Go 1.26
├── main.go                # entrypoint, graceful shutdown
├── .env.example
├── Taskfile.yml
├── docs/
│   └── openapi.json       # placeholder spec — replace with your own
└── internal/
    ├── config/            # env-driven config
    ├── logger/            # zerolog setup
    ├── grpcclient/        # grpc.ClientConn dial helper
    ├── middleware/        # auth, cors, logger, recovery
    ├── handler/           # route registrations (add gRPC-backed handlers here)
    └── server/            # gin engine assembly + swagger mount
```

## Getting started

```bash
# pull deps and populate go.sum
task tidy

# copy and edit config
cp .env.example .env

# run
task run

# or with hot reload
task dev
```

Endpoints:

| Path | Notes |
| --- | --- |
| `GET /healthz` | Liveness probe |
| `GET /v1/ping` | Example route (replace with proxied gRPC handlers) |
| `GET /openapi.json` | Raw OpenAPI spec served from `SWAGGER_SPEC` |
| `GET /docs/` | Embedded Swagger UI |

## Adding a gRPC-backed route

Routes are driven directly by the `google.api.http` annotations in your `.proto` files via [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway). The gateway mounts a `runtime.ServeMux` as the Gin fallback handler, so no path is hardcoded in Go.

1. In your proto pipeline, run `protoc-gen-grpc-gateway` alongside `protoc-gen-go` / `protoc-gen-go-grpc`. This emits a `*.pb.gw.go` file containing `RegisterXxxServiceHandler(ctx, mux, conn)`.
2. Add one line per service in `internal/handler/handler.go`:
   ```go
   import (
       greeterv1 "<your-proto-module>/gen/go/greeter/v1"
       userv1    "<your-proto-module>/gen/go/user/v1"
   )

   func NewMux(ctx context.Context, conn *grpc.ClientConn) (*runtime.ServeMux, error) {
       mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(incomingHeaderMatcher))
       if err := greeterv1.RegisterGreeterServiceHandler(ctx, mux, conn); err != nil { return nil, err }
       if err := userv1.RegisterUserServiceHandler(ctx, mux, conn); err != nil { return nil, err }
       return mux, nil
   }
   ```
3. Every `option (google.api.http) = { post: "/v1/hello" body: "*" }` in your proto becomes a live route — no further wiring in Gin needed.

## Configuration (env)

See `.env.example` for the full list. Highlights:

- `GRPC_BACKEND_ADDR` — the gRPC service this gateway proxies to.
- `AUTH_ENABLED=true` + either `AUTH_JWT_KEY` (HS256) **or** `AUTH_API_URL` (token introspection).
- `CORS_ALLOWED_ORIGINS` — comma-separated; `*` allows everything (and disables credentials).
- `SWAGGER_SPEC` — path to the OpenAPI doc the UI loads; point at the `buf/gen/openapiv2/api.swagger.json` produced by your proto pipeline.

## Docker

The included `Dockerfile` is a multi-stage build that produces a hardened image aligned with OWASP container guidance:

- Builder is `golang:1.26-alpine`; the final stage is `gcr.io/distroless/static-debian12:nonroot` — no shell, no package manager, no setuid binaries.
- Static binary: `CGO_ENABLED=0`, `-trimpath`, `-buildvcs=false`, `-ldflags="-s -w -buildid="`.
- Runs as the unprivileged `nonroot` user (uid 65532).
- Bakes `docs/openapi.json` into `/app/docs/` and points `SWAGGER_SPEC` at it.

Build and run:

```bash
task docker                       # docker build -t gateway:dev .
docker run --rm -p 8080:8080 \
  --read-only --cap-drop=ALL --security-opt=no-new-privileges \
  gateway:dev
```

In Kubernetes pair it with a hardened `securityContext`:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65532
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop: ["ALL"]
  seccompProfile:
    type: RuntimeDefault
```

