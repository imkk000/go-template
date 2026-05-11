# buf — Go-tool based proto/swagger pipeline

Generates Go code and OpenAPI v2 (Swagger) from `.proto` sources using **local Go tools** declared in `go.mod` (Go 1.24+ `tool` directive) — no globally installed binaries, no remote buf plugin runs.

## Layout

```
buf/
├── buf.yaml          # module + lint + breaking config (v2)
├── buf.gen.yaml      # plugin config — uses `local: [go, tool, ...]`
├── go.mod            # `tool` directive pins plugin versions
├── proto/
│   └── greeter/v1/greeter.proto
├── gen/              # output (gitignored)
│   ├── go/
│   └── openapiv2/
└── Taskfile.yml
```

## Tools (declared in `go.mod`)

| Tool | Purpose |
| --- | --- |
| `github.com/bufbuild/buf/cmd/buf` | The buf CLI itself |
| `google.golang.org/protobuf/cmd/protoc-gen-go` | Go message types |
| `google.golang.org/grpc/cmd/protoc-gen-go-grpc` | Go gRPC server/client stubs |
| `github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2` | OpenAPI v2 / Swagger |

Each plugin is invoked by buf as `go tool <name>`, so versions are reproducible from `go.mod` / `go.sum`.

## Usage

Requires [Task](https://taskfile.dev) (`go install github.com/go-task/task/v3/cmd/task@latest`).

```bash
task tidy        # one-time: pull tool binaries into the module cache
task generate    # produce gen/go/... and gen/openapiv2/api.swagger.json
task lint        # buf lint .proto sources
task format      # buf format -w
task clean       # remove gen/
```

Run `task` with no args to list every task with its description.

For less common workflows call buf directly: `go tool buf dep update`,
`go tool buf breaking --against '.git#branch=main,subdir=buf'`. Bump pinned
tools with `go get -tool <module>@latest && go mod tidy`.

Output:

- `gen/go/greeter/v1/greeter.pb.go` — Go message types
- `gen/go/greeter/v1/greeter_grpc.pb.go` — gRPC server + client stubs
- `gen/openapiv2/api.swagger.json` — single merged swagger doc

## Adding more plugins

Same pattern for `protoc-gen-grpc-gateway`, `protoc-gen-connect-go`, etc:

```bash
go get -tool <module-path>@latest
```

```yaml
# buf.gen.yaml
  - local: ["go", "tool", "<binary-name>"]
    out: gen/go
    opt: [paths=source_relative]
```

## Notes

- `buf.gen.yaml` uses `local:` (not `remote:`), so generation runs entirely offline once dependencies are cached.
- `go.mod` `tool` directive requires **Go 1.24+**; this template targets **Go 1.26**.
- The sample proto uses `google.api.http` + `openapiv2_swagger` annotations from the BSR deps in `buf.yaml`.
- Tool versions are pinned in `go.mod`; bump with `go get -tool <module>@latest` then commit `go.mod` / `go.sum`.
