# Recipe: Add gRPC Method

Use this when a feature adds or changes one RPC on a gRPC service.

## Files To Touch

- `api/<svc>/v<N>/<svc>.proto` for the contract change
- generated stubs (commit them, or rely on the CI generation check, per the repo's [generated code policy](../services/grpc-services.md#generated-code-policy))
- `internal/api/grpc/server.go` for the method implementation
- `internal/api/grpc/errors.go` if domain-to-`codes.*` mapping changes
- `internal/api/grpc/interceptors.go` only if cross-cutting behavior changes
- `internal/core/...` for business logic
- transport and core tests

## Steps

1. Edit the proto under the versioned package: add the RPC and its request/response messages, give every field an explicit number, and `reserve` the numbers and names of any removed fields. Do not renumber existing fields.
2. Regenerate deterministically with `buf generate` (or the repo's pinned `protoc` invocation). Commit the stubs or let the CI gen check enforce them, never both ad hoc.
3. Add or update the core method that owns the behavior. Core takes and returns domain types, not generated proto types.
4. Implement the server method in `server.go`: decode the request message, call exactly one core method, map domain errors to `codes.*` via `errors.go`, and build the response message.
5. Confirm interceptors already cover recovery, auth, access logging, tracing, and metrics; add a metric hook only if the existing interceptor does not capture the new RPC. Use low-cardinality labels (method name, code) and never request, user, or tenant IDs.
6. Verify the server honors `ctx` cancellation and deadlines, and self-protects against unbounded work even if a client sets no deadline.

## Invariants To Preserve

- proto package and directory stay versioned (`v<N>`), even at `v1`
- no domain logic in generated types or in the server method; the server is a thin transport adapter
- request context, deadlines, and cancellation flow into the core call
- domain errors are wrapped with `%w` in core and mapped once to `codes.*` at the server boundary, not logged twice
- the proto change is backward compatible (additive fields, new RPC); any wire break is an explicit, reviewed decision recorded against the contract
- metric labels stay low cardinality

## Proof

- `buf lint` passes and `buf generate` produces no diff (or `make verify` runs the repo's gen check)
- table test for status mapping in `internal/api/grpc/errors.go` covering each mapped `codes.*`, run with `go test ./internal/api/grpc/ -run TestStatusMapping`
- interceptor test proving recovery, auth rejection, and metric emission on the new method: `go test ./internal/api/grpc/ -run TestInterceptors`
- core unit test for the new behavior plus a context-cancellation test: `go test ./internal/core/... -run <Method>`
- `grpcurl -plaintext -d '{...}' localhost:<port> <svc>.v<N>.<Service>/<Method>` smoke call against a local server
- `make verify`

Governing doc: [grpc-services.md](../services/grpc-services.md). The HTTP analog is [add-http-endpoint.md](add-http-endpoint.md). If the method calls another service, see [add-external-client.md](add-external-client.md). To retire an RPC or field, see [deprecate-and-remove-contract.md](deprecate-and-remove-contract.md).
