# grpc-server — gRPC service template

Go 1.26 gRPC server with one directory per service. Each service directory
holds its **repository**, **client**, **service**, and **handler** side by
side, all wired through Go interfaces. Unit tests use
[`stretchr/testify`](https://github.com/stretchr/testify).

No proto pipeline is bundled — generate your `.proto` files elsewhere (e.g.
with the sibling `buf/` template) and `go get` the resulting module. The
greeter example ships with its generated stubs already committed in `gen/`
so the template compiles out of the box.

The server registers a unary interceptor chain out of the box:

| Order | Interceptor | Purpose |
| --- | --- | --- |
| 1 | `requestid.UnaryServerInterceptor()` | Pull / generate `x-request-id`, store on context |
| 2 | `recoveryInterceptor()` | Catch panics, log stack + request_id, return `Internal` |
| 3 | `loggingInterceptor()` | One zerolog line per RPC with method, status code, latency, request_id |
| 4 | `validateInterceptor(...)` | Run `buf.build/go/protovalidate` rules against every request |

The gRPC [health service](https://github.com/grpc/grpc/blob/master/doc/health-checking.md) is registered automatically and starts as `SERVING`. `GracefulStop()` flips it to `NOT_SERVING` first so readiness probes drain in-flight traffic.

## Scaffold a new project with `gonew`

```bash
gonew github.com/imkk000/go-template/grpc-server my.org/my-server
```

## Layout

```
grpc-server/
├── go.mod                          # module grpc-server, go 1.26
├── main.go                         # entrypoint; one `Register(s)` line per service
├── Dockerfile                      # multi-stage, distroless, non-root
├── gen/go/greeter/v1/              # committed proto-generated stubs (example)
└── internal/
    ├── config/                     # env-driven config
    ├── logger/                     # zerolog setup
    ├── requestid/                  # x-request-id propagation (server + client interceptors)
    ├── server/                     # grpc.Server bootstrap, health, recovery, validate, logging
    └── greeter/                    # example service — everything here
        ├── repository.go           # Repository interface + InMemory impl
        ├── client.go               # Client interface + Noop impl
        ├── service.go              # Service interface + impl (uses repo + client)
        ├── handler.go              # gRPC handler + one-line Register(s)
        ├── mocks_test.go           # hand-rolled mocks for the three interfaces
        ├── repository_test.go
        ├── service_test.go
        └── handler_test.go
```

## Layered interfaces

```
                 Handler ──▶ Service ──▶ Repository (data)
                                 │
                                 └──▶ Client (outbound API)
```

Every arrow is a Go interface, so each layer can be unit-tested with a
hand-rolled mock — no extra mocking library required. Concrete types live
next to the interface they implement.

## Adding a new service

1. Generate your proto somewhere (e.g. with the `buf/` template) and import
   the resulting module.
2. `mkdir internal/orders` and create the four files:
   `repository.go`, `client.go`, `service.go`, `handler.go`.
3. In `handler.go` add a `Register(s *grpc.Server)` function:
   ```go
   func Register(s *grpc.Server) {
       repo := NewRepository(/* deps */)
       client := NewClient(/* deps */)
       svc := NewService(repo, client)
       ordersv1.RegisterOrdersServiceServer(s, NewHandler(svc))
   }
   ```
4. Add one line in `main.go`:
   ```go
   greeter.Register(srv.GRPC())
   orders.Register(srv.GRPC())  // <- new
   ```

## Request-id propagation

`internal/requestid` carries an `x-request-id` value through the request
context and gRPC metadata.

**Incoming** — the server's unary chain pulls `x-request-id` from incoming
metadata (or generates one if missing), stores it on the context, and echoes
it back to the caller as a response header. Handlers and services read it
with:

```go
import "grpc-server/internal/requestid"

rid := requestid.FromContext(ctx)
```

The logging interceptor already includes it in every `grpc request` log line.

**Outgoing** — when this server calls *another* gRPC service, build that
client with the client interceptor so the same id flows downstream:

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    "grpc-server/internal/requestid"
)

conn, err := grpc.NewClient(addr,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithChainUnaryInterceptor(requestid.UnaryClientInterceptor()),
)
```

Any RPC made through `conn` automatically appends `x-request-id` to outgoing
metadata whenever the calling context carries one — so the gateway → server
→ downstream-server chain all share a single correlation id.

## Getting started

```bash
task tidy        # populate go.sum
task test        # go test -race -count=1 ./...
task run         # go run .
```

The server is reachable on `:9090` by default. With `GRPC_REFLECTION=true`
(the default), you can poke it with `grpcurl`:

```bash
grpcurl -plaintext -d '{"name":"Alice","lang":"en"}' \
  localhost:9090 greeter.v1.GreeterService/SayHello

# pass an upstream request id and watch it appear in the server logs +
# response header
grpcurl -plaintext \
  -H 'x-request-id: 0123456789abcdef' \
  -d '{"name":"Alice"}' \
  localhost:9090 greeter.v1.GreeterService/SayHello

# health check (also what K8s gRPC probes call)
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
```

## Request validation

`buf.build/go/protovalidate` runs CEL rules declared in your `.proto` file
against every incoming request. Add rules with the
[`buf.validate`](https://buf.build/bufbuild/protovalidate) field options:

```proto
import "buf/validate/validate.proto";

message SayHelloRequest {
  string name = 1 [(buf.validate.field).string = {min_len: 1, max_len: 64}];
  string lang = 2 [(buf.validate.field).string = {len: 2}];
}
```

Regenerate, redeploy, and violations come back as `codes.InvalidArgument`
without ever reaching your handler. The greeter example proto ships
without rules — the interceptor is a no-op until you add them.

## Docker

Multi-stage build, distroless final image, non-root user (uid 65532):

```bash
task docker                       # docker build -t grpc-server:dev .
docker run --rm -p 9090:9090 \
  --read-only --cap-drop=ALL --security-opt=no-new-privileges \
  grpc-server:dev
```

In Kubernetes, native gRPC probes pair with the bundled health service:

```yaml
livenessProbe:
  grpc:
    port: 9090
readinessProbe:
  grpc:
    port: 9090
```
