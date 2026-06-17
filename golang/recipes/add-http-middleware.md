# Recipe: Add HTTP Middleware

Use this when a cross-cutting concern (auth, rate limiting, request shaping) must run for many routes and does not belong in any single handler.

Governing doc: [`golang/services/http-services.md`](../services/http-services.md) (Middleware Order). Working example: [`golang/reference/exampleservice/internal/api/http/middleware.go`](../reference/exampleservice/internal/api/http/middleware.go) and [`server.go`](../reference/exampleservice/internal/api/http/server.go). This recipe is the missing half of [`add-http-endpoint.md`](add-http-endpoint.md), which assumes middleware already exists.

## Files To Touch

- `internal/api/http/middleware.go` — the new `func(http.Handler) http.Handler` wrapper
- `internal/api/http/server.go` — insert the wrapper at the correct position in `routes()`
- `internal/api/http/middleware_test.go` — unit test plus one negative test
- `internal/config/*.go` — only if the middleware needs a tunable (timeout, rate, key); route the specific library pick to [`decisions/framework-selection.md`](../decisions/framework-selection.md)

## Steps

1. Write the wrapper as a `func(http.Handler) http.Handler`. If it needs dependencies (logger, config, a limiter), make it a constructor `func(deps) func(http.Handler) http.Handler` and close over them — matching `recoverMiddleware(logger)` and `loggingMiddleware(logger, metrics)` in the reference. Do not read dependencies from package globals.
2. Inside the returned handler, do the cross-cutting work, then call `next.ServeHTTP(w, r)`. Short-circuit (write status, `return`, do not call `next`) only on a hard reject such as failed auth or an exceeded rate limit.
3. Propagate any request-scoped value you derive (caller identity, tenant) via `context.WithValue(r.Context(), key, v)` and `next.ServeHTTP(w, r.WithContext(ctx))`. Use an **unexported typed key** (`type ctxKey int`), never a string, never a global variable — see `requestIDKey` in `middleware.go`. Expose a typed `fooFrom(ctx)` accessor; downstream code reads the value from the context, never from a global.
4. Wire it into `routes()` at the documented position. The chain is built inside-out, so the wrapper applied **last** runs **first**. Place it by responsibility:
   - **recovery** — outermost, applied last, runs first; it must be unbypassable so any panic below it becomes a 500
   - **request ID / trace** — next, so every downstream log line and reject can be correlated
   - **auth / rate limit** — before logging, so the access log records the body of work only for requests that were actually allowed to proceed, and so a flood is rejected cheaply before metrics and core run
   - **access logging / metrics** — wraps the handler; emits one line and one metric per admitted request
   - **handler** — innermost

   In the reference, `recoverMiddleware` and `requestIDMiddleware` wrap the **root** mux while `loggingMiddleware` wraps only the **API** mux. New auth/rate-limit middleware almost always belongs on the API mux too: probes must skip it. Add it like `apiHandler := authMiddleware(...)(loggingMiddleware(...)(apiMux))` — note auth is outside logging.
5. Exclude health probes from heavy middleware. The reference splits `apiMux` from `probeMux` precisely so `/livez` and `/readyz` are not access-logged, not metered, and not subject to auth or rate limiting (see the `routes()` comment in `server.go`). Add new heavy middleware to the API branch only; leave the probe branch on recovery + request ID.
6. Keep it cheap and allocation-aware: it runs on every request. Reuse buffers, avoid per-request closures over large state, and gate optional work (do not allocate a `statusRecorder`-style wrapper unless you need it). Put no business logic here — middleware shapes the request/response envelope; domain decisions stay in `internal/core`.

## Invariants To Preserve

- deterministic, documented order: recovery → request ID/trace → auth/rate-limit → logging/metrics → handler
- recovery stays outermost and cannot be bypassed; every panic below it becomes a 500
- request-scoped data flows through `context` with an unexported typed key — never package globals, never a `string` key
- no business logic in middleware; it never calls `internal/core` to make domain decisions
- health probes (`/livez`, `/readyz`) are excluded from auth, rate limiting, access logging, and metrics
- a rejecting middleware short-circuits with a mapped status (`401`/`429`) and does not call `next`
- metric/log labels stay low-cardinality (route pattern, status class — never request/user/tenant IDs)

## Proof

- Unit test: wrap a fake `http.HandlerFunc` that records whether it ran, drive it with `httptest.NewRecorder` + `httptest.NewRequest`, and assert both the wrapper's effect (header set, context value readable via the accessor) and ordering (e.g. the fake handler observes the request ID the request-ID middleware added).
- Negative test: assert the short-circuit path. For recovery, a handler that `panic`s yields `500` and `next`'s post-panic code never runs; for auth, a missing/invalid credential yields `401` and the fake handler is never invoked (`got.called == false`).
- Probe-exclusion test: model on `probes_test.go` — drive `/readyz` through `routes()` and assert the heavy middleware did not fire (e.g. `countingMetrics.requestCount() == 0`, or the auth check was not reached).
- `make verify`
- Targeted run: `go test ./internal/api/http/ -run Middleware -count=1`
