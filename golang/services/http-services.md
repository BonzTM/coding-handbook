# HTTP Services

HTTP service defaults for repos that want predictable handlers, cheap debugging, and minimal framework lock-in.

## Default Approach

- Use `net/http` and Go 1.22+ `ServeMux` first.
- Keep transport code under `internal/api/http` and core behavior under `internal/core`.
- Treat handlers as translation layers: decode, validate, call core, map errors, encode response.

### Minimal Layout

```text
internal/api/http/
  server.go        # mux construction and middleware wiring
  handlers.go      # handler methods on a server struct
  errors.go        # HTTP error mapping helpers
  middleware.go    # recovery, logging, request id, auth, timeouts
```

### Routing Pattern

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", srv.handleGetUser)
mux.HandleFunc("POST /users", srv.handleCreateUser)
```

Prefer route patterns that make ownership obvious. If the route tree becomes hard to manage with stdlib primitives, document why a router like `chi` earns its place.

### Handler Contract

Each handler should usually do five things in order:

1. read request metadata from `r.Context()` and route params
2. decode and validate input
3. call one core service method
4. map domain errors to HTTP status codes
5. encode the response and emit request telemetry

If the HTTP surface is consumed outside one codebase, define and review the payload contract explicitly. That can be a stable JSON schema, OpenAPI description, or a well-documented response model in the transport package, but it should have one source of truth.

HTTP paths are unversioned by default — the reference serves `/widgets`, not `/v1/widgets`. Evolve the contract additively per [foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md); a breaking change means a new resource or an explicitly negotiated new surface, never a silent mutation of the shapes existing clients depend on.

### Error Responses

Map each domain error to a `(status, code)` at the boundary (the `errors.go` helper) and encode it with the single structured error envelope the repo uses — a machine-readable `code`, a human `message`, and an optional `fields` array for validation failures. The envelope shape and rules are defined once in [foundations/serialization.md](../foundations/serialization.md#error-responses); the domain-error-to-status mapping lives in [foundations/errors-and-logging.md](../foundations/errors-and-logging.md). Never return a bare `{"error":"..."}` string, and never leak internal detail in a 5xx body.

### Server Hardening Defaults

- set `ReadHeaderTimeout`
- set `IdleTimeout`
- use `WriteTimeout` unless the endpoint intentionally streams
- cap request body size for non-streaming endpoints
- expose separate liveness and readiness endpoints

### Transport Security

The default for an HTTP service is TLS terminated at the platform edge — load balancer, ingress, or mesh sidecar — with the app listening plaintext inside the pod boundary. Do not double-terminate: when the platform already owns the edge, an in-app TLS listener only adds cert-rotation surface and a second failure mode.

If the service is standalone or edge-exposed and must terminate TLS itself, use `http.Server` with a `tls.Config` enforcing `MinVersion: tls.VersionTLS12`, and gate the cert/key through config: empty keys select the plaintext local/dev listener, and a *set but unloadable* key pair fails fast at startup — never a silent downgrade to plaintext. This is the same gating pattern the gRPC reference implements in [`reference/examplegrpc/internal/api/grpc/tls.go`](../reference/examplegrpc/internal/api/grpc/tls.go); see [grpc-services.md](grpc-services.md#transport-security) for the full stance.

### Middleware Order

Use a stable order so failures are observable and safe:

1. request ID
2. panic recovery
3. trace context
4. auth and rate limiting
5. access logging and metrics
6. handler

Request ID sits outside recovery so the id is on the context before the recovery layer runs — a panic-recovered 500 can then still carry `request_id` in its error envelope; [reference/exampleservice/](../reference/exampleservice/) wires exactly this in `internal/api/http/server.go`.

Authentication stays ahead of logging so unauthenticated requests are rejected early. Per-route *authorization*, by contrast, MAY run just inside logging/metrics so the authz decision and the matched route land in the access log — [reference/exampleservice/](../reference/exampleservice/) does exactly this, keeping authN at the edge while running RBAC/tenancy checks after the logging middleware.

### CORS

The default is **no CORS middleware**. CORS is a browser mechanism; a service-to-service API has no browser clients and needs no CORS layer — adding one by reflex only widens the attack surface. Enable it only when [spec intake](../checklists/spec-intake.md) identifies browser clients.

When it is enabled, the policy is:

- an **explicit origin allowlist** driven from config — never `*`, and never any wildcard-equivalent combined with `Access-Control-Allow-Credentials: true`
- correct preflight handling: answer `OPTIONS` with the allowed methods/headers and a sane `Access-Control-Max-Age`, without invoking the real handler
- CORS sits with the edge middleware, **before auth** in the chain, so preflights (which carry no credentials) are answered instead of rejected

Default library per [framework-selection](../decisions/framework-selection.md) is `github.com/rs/cors` (pure Go); a small hand-rolled handler is fine for a trivial static allowlist.

### List Endpoints And Pagination

Every endpoint that returns a collection is paginated from the start — an unbounded list is a latency and memory incident waiting for the table to grow.

- Default to keyset/cursor pagination (an opaque `next_cursor` over a stable sort key) for large or append-heavy sets; offset/limit is acceptable only for small, bounded, or admin lists where deep pages are rare.
- Enforce a default and a maximum page size on the server; clamp an oversized request rather than rejecting it, and never honor an unbounded one.
- Sort on a total, stable order (the cursor key plus a tiebreaker such as the primary key) so pages neither overlap nor skip rows under concurrent writes.
- Return exact totals only when they are cheap; for large tables prefer a `has_more` flag or `next_cursor` over a full count.
- Keep one consistent list envelope (an `items` array plus pagination metadata) across endpoints, and treat it as a wire contract per [serialization.md](../foundations/serialization.md).

## Common Mistakes And Forbidden Patterns

- Handlers querying the database directly instead of going through core or repository seams.
- Using a framework that hides `http.Handler` semantics without a strong reason.
- Logging full request or response bodies by default.
- Returning transport-specific DTOs from core packages.
- Changing public request or response shapes without compatibility review.
- Treating readiness as "process is running" rather than "dependencies needed for traffic are ready".
- Returning an unbounded or unpaginated collection, or deep offset pagination that scans the whole prefix of a large table.
- Double-terminating TLS behind an edge or mesh that already owns it — or, when the service must terminate TLS itself, silently falling back to plaintext when a configured cert/key fails to load instead of failing fast.
- A CORS policy that reflects arbitrary `Origin` values or pairs a wildcard origin with `Allow-Credentials` — either one hands any website credentialed access to the API.

## Verification And Proof

- handler tests with `httptest.NewRecorder` and `httptest.NewRequest`
- negative tests for validation, auth, and error mapping
- one smoke test against `/livez` and `/readyz`
- audit that timeout, body-size, and middleware defaults are actually wired in `server.go`
