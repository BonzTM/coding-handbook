# examplegrpc

A compiling gRPC reference service for the Go engineering handbook. It is the
gRPC analog of `exampleservice` (the HTTP exemplar) and mirrors its structure,
conventions, and infrastructure: a thin `cmd/examplegrpc/main.go`, a `core`
domain with a consumer-defined `Store` seam, config-gated telemetry, an
injectable clock, and a single `make verify` safety gate.

Governing docs: [grpc-services.md](../../services/grpc-services.md),
[add-grpc-method.md](../../recipes/add-grpc-method.md).

## Layout

```text
api/widget/v1/widget.proto   versioned wire contract (widget.v1)
api/widget/v1/*.pb.go        committed generated stubs (buf + protoc-gen-go*)
internal/api/grpc/           transport adapter
  server.go                  thin WidgetServiceServer over internal/core
  errors.go                  domain error -> codes.* mapping (mapped once at the boundary)
  interceptors.go            chained unary stack: recovery, request-id, access log, deadline guard, auth
  service.go                 grpc.Server wiring: otelgrpc stats handler, health, reflection
  sidecar.go                 HTTP sidecar: /metrics, /livez, /readyz
internal/core/               widgets domain + Store interface + in-memory store (keyset-paginated)
internal/config/             env+flag load, fail-fast Validate
internal/telemetry/          slog logger, readiness, Prometheus metrics, config-gated OTel tracing
internal/buildinfo/          -ldflags build metadata
internal/testutil/           FakeClock
cmd/examplegrpc/main.go      signal.NotifyContext + errgroup + ordered bounded shutdown
```

## Run

```sh
make run            # in-memory store, local/dev auth (synthetic principal)
```

The gRPC server listens on `:9090`; the metrics/probes HTTP sidecar on `:8080`
(`/metrics`, `/livez`, `/readyz`). Server reflection and the standard gRPC health
service are registered, so `grpcurl` works out of the box:

```sh
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext -d '{"id":"w1","name":"One"}' localhost:9090 widget.v1.WidgetService/CreateWidget
grpcurl -plaintext -d '{"id":"w1"}' localhost:9090 widget.v1.WidgetService/GetWidget
grpcurl -plaintext -d '{"page_size":2}' localhost:9090 widget.v1.WidgetService/ListWidgets
```

## Codegen

Generated stubs are committed. They are produced deterministically with `buf` and
the pinned `protoc-gen-go` / `protoc-gen-go-grpc` plugins (module `tool` deps):

```sh
make gen        # regenerate api/widget/v1/*.pb.go
make gen-check  # regenerate and fail if the result differs (CI gate)
```

## Verify

`make verify` is the ordered safety gate (also run in CI):

```text
tidy -> fmt-check -> buf-lint -> gen-check -> lint -> vet -> test -> race -> vuln -> build
```

All builds are pure Go (`CGO_ENABLED=0`).
