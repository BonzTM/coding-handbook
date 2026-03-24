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

## Common Mistakes And Forbidden Patterns

- Handlers querying the database directly instead of going through core or repository seams.
- Using a framework that hides `http.Handler` semantics without a strong reason.
- Logging full request or response bodies by default.
- Returning transport-specific DTOs from core packages.
- Changing public request or response shapes without compatibility review.
- Treating readiness as "process is running" rather than "dependencies needed for traffic are ready".

## Verification And Proof

- handler tests with `httptest.NewRecorder` and `httptest.NewRequest`
- negative tests for validation, auth, and error mapping
- one smoke test against `/livez` and `/readyz`
- audit that timeout, body-size, and middleware defaults are actually wired in `server.go`
