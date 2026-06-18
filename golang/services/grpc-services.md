# gRPC Services

gRPC defaults for repos that need strong contracts, typed APIs, and predictable transport boundaries.

## Default Approach

- Use gRPC when the repo benefits from strong schemas, streaming, or polyglot clients.
- Put wire contracts under `api/<service>/v1/`.
- Keep generated transport code and handwritten server logic separate from core logic.

### Protocol Layout

```text
api/orders/v1/
  orders.proto
internal/api/grpc/
  server.go
  errors.go
  interceptors.go
```

- Always version the package and directory, even at `v1`.
- Keep messages and services as small and explicit as possible.
- Use `buf` or deterministic `protoc` invocations; do not rely on undocumented local generation steps.

Proto package versioning is separate from release versioning for the repo. Keep versioned API packages like `v1` where they clarify compatibility. Repo release tags still follow the canonical Go module form `v1.2.3` (the `v` prefix is required for module versions); only changelog or display strings may render the version without the `v`.

### Generated Code Policy

Pick one rule and automate it:

- default for new repos: commit generated Go stubs if normal contributors are expected to run `go build ./...` without extra proto tooling
- acceptable alternative: generate in CI only, if the repo already standardizes the toolchain and verifies generated output deterministically

### Server Rules

- server methods should decode transport data, call core services, and map errors to `codes.*`
- interceptors own cross-cutting behavior such as recovery, auth, tracing, and access logs
- clients should set deadlines; servers should still respect context cancellation and protect themselves from unbounded work
- expose the gRPC health service for deployability

### Error Details

A bare `status.New(code, msg)` tells a client *that* the request failed but not *what* to fix. For a validation failure, attach a machine-readable detail so the client can map errors back to fields instead of scraping a message string.

- Map the domain error to `codes.InvalidArgument`, then attach a `google.rpc.BadRequest` carrying one `FieldViolation{Field, Description}` per failed field via `status.WithDetails`. `WithDetails` returns a new `*status.Status`; it only errors if the detail cannot be marshaled, so fall back to the plain status rather than dropping the error.
- Carry the violations *structurally* out of core (a typed `FieldViolation` slice the validation error exposes), and build the proto detail once at the transport boundary — the same place the sentinel becomes a `codes.*`. Core must not import `errdetails`; the proto type is a transport concern.
- Use the standard `google.rpc.*` detail types (`BadRequest`, `ErrorInfo`, `RetryInfo`, `QuotaFailure`) so any gRPC client can decode them, rather than inventing a bespoke detail message.
- The reference wires exactly this: [`reference/examplegrpc/internal/api/grpc/errors.go`](../reference/examplegrpc/internal/api/grpc/errors.go) renders `core.ErrInvalidWidget` to `InvalidArgument` + `BadRequest`, with the structured carrier in `internal/core/widget.go` (`FieldViolations`).

### Transport Security

gRPC over HTTP/2 must run on TLS in production; plaintext (`h2c`) is for local development only.

- **TLS by default in production.** The server presents a certificate/key and refuses to start if TLS was requested but the key pair fails to load — fail-fast at startup, never a silent downgrade to plaintext. Enforce a `MinVersion` of TLS 1.2.
- **mTLS for internal service-to-service traffic** unless a service mesh terminates TLS for you. When a client-CA bundle is configured the server requires and verifies a client certificate (`tls.RequireAndVerifyClientCert`); mTLS requires the server cert/key to be set too, validated together. If a mesh (Istio/Linkerd) already provides mutual TLS at the sidecar, do not double-terminate — run the app listener plaintext *inside* the pod and let the mesh own the edge.
- Keep it config-gated: empty cert/key selects the insecure local/dev listener so the service boots offline, the same way auth and tracing are gated. The reference implements this in [`reference/examplegrpc/internal/api/grpc/tls.go`](../reference/examplegrpc/internal/api/grpc/tls.go) (`ServerTransportCredentials`), with the `GRPC_TLS_CERT_FILE` / `GRPC_TLS_KEY_FILE` / `GRPC_TLS_CLIENT_CA_FILE` keys validated together in `internal/config`.

## Common Mistakes And Forbidden Patterns

- domain logic living in proto-generated types or server methods
- missing versioned proto packages
- returning raw internal errors to clients
- ignoring deadlines and stream cancellation
- proto files that mirror database tables instead of transport contracts
- flattening a field-validation failure into a bare `InvalidArgument` message with no `google.rpc.BadRequest` detail, so clients cannot map errors to fields
- building `errdetails` proto types inside core instead of at the transport boundary, coupling domain logic to the wire format
- serving plaintext gRPC in production, or silently downgrading to plaintext when a configured key pair fails to load instead of failing fast
- double-terminating TLS when a mesh already provides mTLS, or skipping client-cert verification on internal service-to-service links a mesh does not cover

## Verification And Proof

- proto lint or generation check (`buf lint`, `buf generate`, or equivalent)
- service tests for status mapping and interceptor behavior
- a test that asserts a validation failure returns `InvalidArgument` *and* a `google.rpc.BadRequest` detail with the expected `{field, description}` violations (decode the detail, do not match the message string)
- a TLS/mTLS test: the server rejects a missing/invalid client cert when a client CA is configured, and fails fast when a configured key pair cannot load
- `grpcurl` smoke test against a local server
- backward-compatibility review before breaking a public proto contract
