# Recipe: Add HTTP Endpoint

Use this when a feature adds or changes one HTTP route.

## Files To Touch

- `internal/api/http/handlers.go` or a route-specific handler file
- `internal/api/http/server.go` or router wiring
- `internal/api/http/errors.go` if status mapping changes
- `internal/core/...` for business logic
- transport and core tests

## Steps

1. Define request and response DTOs in the HTTP package, not in core.
2. Add or update the core method that owns the behavior.
3. Implement the handler: decode, validate, call core, map errors, encode response.
4. Register the route on the mux using a stable method and path pattern.
5. Ensure middleware already covers request ID, recovery, access logging, and auth if needed.
6. Add telemetry for the route through existing middleware or explicit metrics hooks.

## Invariants To Preserve

- no database calls directly from the handler
- request context flows through the full call chain
- response status mapping is consistent with existing transport behavior
- request bodies are size-limited and validated before core logic runs

## Proof

- handler tests with `httptest`
- one negative test for validation or auth
- one smoke request against a local server
- log or metric review showing the new route is observable
