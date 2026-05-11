# Go Template

A collection of Go module templates designed to be scaffolded with [`gonew`](https://pkg.go.dev/golang.org/x/tools/cmd/gonew).

Install gonew once:

```bash
go install golang.org/x/tools/cmd/gonew@latest
```

Each template lives in its own subdirectory with a short, generic `module` name (e.g. `module poc`, `module buf`, `module gateway`) so `gonew` cleanly rewrites the path of your new project.

## Templates

| Template | Module | Purpose |
| --- | --- | --- |
| [`poc/`](./poc) | `poc` | Minimal Go binary with zerolog — handy for quick experiments |
| [`buf/`](./buf) | `buf` | Proto + OpenAPI v2 pipeline using buf with Go 1.26 `tool` directive (no global tool installs) |
| [`gateway/`](./gateway) | `gateway` | Gin HTTP/JSON API gateway that reverse-proxies into a gRPC backend via `grpc-gateway` (auth middleware, CORS, zerolog, Swagger UI, Docker) |
| [`grpc-server/`](./grpc-server) | `grpc-server` | gRPC service template — one directory per service (`repository.go` / `client.go` / `service.go` / `handler.go`) with interface-driven layers and testify unit tests |

## Scaffold

```bash
# Minimal POC
gonew github.com/imkk000/go-template/poc go.play/play-go-poc

# Proto + OpenAPI v2 pipeline
gonew github.com/imkk000/go-template/buf my.org/my-proto

# Gin → gRPC reverse-proxy gateway
gonew github.com/imkk000/go-template/gateway my.org/my-gateway

# gRPC service (repo + service + handler + client per service dir)
gonew github.com/imkk000/go-template/grpc-server my.org/my-server
```

`gonew` clones the chosen template, rewrites the `module` directive in `go.mod`, and updates all matching internal imports to the new destination path.

After scaffolding, run `task tidy` (or `go mod tidy`) inside the new module to populate `go.sum`.
