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

### Error Responses

Map each domain error to a `(status, code)` at the boundary (the `errors.go` helper) and encode it with the single structured error envelope the repo uses — a machine-readable `code`, a human `message`, and an optional `fields` array for validation failures. The envelope shape and rules are defined once in [foundations/serialization.md](../foundations/serialization.md#error-responses); the domain-error-to-status mapping lives in [foundations/errors-and-logging.md](../foundations/errors-and-logging.md). Never return a bare `{"error":"..."}` string, and never leak internal detail in a 5xx body.

### Server Hardening Defaults

- set `ReadHeaderTimeout`
- set `IdleTimeout`
- use `WriteTimeout` unless the endpoint intentionally streams
- cap request body size for non-streaming endpoints
- expose separate liveness and readiness endpoints

### Middleware Order

Use a stable order so failures are observable and safe:

1. panic recovery
2. request ID and trace context
3. auth and rate limiting
4. access logging and metrics
5. handler

Authentication stays outermost (before logging) so unauthenticated requests are rejected early. Per-route *authorization*, by contrast, MAY run just inside logging/metrics so the authz decision and the matched route land in the access log — [reference/exampleservice/](../reference/exampleservice/) does exactly this, keeping authN at the edge while running RBAC/tenancy checks after the logging middleware.

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

## Verification And Proof

- handler tests with `httptest.NewRecorder` and `httptest.NewRequest`
- negative tests for validation, auth, and error mapping
- one smoke test against `/livez` and `/readyz`
- audit that timeout, body-size, and middleware defaults are actually wired in `server.go`
