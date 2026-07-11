# Recipe: Add HTTP Middleware

Use this when a cross-cutting concern (auth, rate limiting, request shaping) must run for many routes and does not belong in any single handler.

Governing doc: [`csharp/services/http-services.md`](../services/http-services.md) (Middleware Order). This recipe is the missing half of [`add-http-endpoint.md`](add-http-endpoint.md), which assumes the pipeline already exists.

## Files To Touch

- `src/Orders.Api/Middleware/<Name>Middleware.cs` — the middleware class, or `src/Orders.Api/Endpoints/<Name>Filter.cs` if it is group-scoped (see step 1)
- `src/Orders.Api/Program.cs` — insert it at the correct pipeline position, or attach the filter to the API route group
- `src/Orders.Api/Options/*.cs` plus `appsettings.json` — only if it needs a tunable (timeout, rate, key); follow [add-config-key.md](add-config-key.md)
- `tests/Orders.UnitTests/Middleware/<Name>MiddlewareTests.cs` — unit test plus one negative test
- `tests/Orders.IntegrationTests` — a probe-exclusion test if the concern is heavy

## Steps

1. **Choose middleware vs endpoint filter first.** Middleware runs for the whole pipeline (or a branch), sees the raw `HttpContext`, and runs before routing — use it for concerns that must cover every request including unmatched routes: request ID, exception-to-`ProblemDetails`, security headers, host-level rate limiting. An `IEndpointFilter` runs after routing and binding, scoped to a route group or endpoint, with access to endpoint metadata and typed handler arguments — use it for API-group concerns: per-group validation, resource-level authorization checks, response shaping. If it must fire on a 404 or a probe, it cannot be a filter.
2. Write middleware as a class taking `RequestDelegate next` in the constructor with `public async Task InvokeAsync(HttpContext context)`. Take singleton dependencies (logger, options, a limiter) via the constructor; take scoped/transient dependencies as extra `InvokeAsync` parameters so DI resolves them per request. Never read dependencies from statics or a service locator.
3. Inside `InvokeAsync`, do the cross-cutting work, then `await next(context)`. Short-circuit (write the status via the ProblemDetails service, `return`, do not call `next`) only on a hard reject such as failed auth or an exceeded rate limit. An endpoint filter short-circuits by returning `TypedResults.Problem(...)` / `TypedResults.Unauthorized()` instead of `await next(context)`.
4. Propagate any request-scoped value you derive (caller identity, tenant) through the request, never a global. For the request ID, set `HttpContext.TraceIdentifier`, echo the response header, and open a logger scope. For other values, set a typed feature (`context.Features.Set<ITenantFeature>(...)`) or a typed entry — never a magic string in `HttpContext.Items`, never an `AsyncLocal` static. Downstream code reads the feature, not a global.
5. Register it in `Program.cs` at the documented position. Registration order **is** execution order — the middleware registered **first** runs **first** (the opposite of Go's inside-out wrapping). Place it by responsibility:
   - **request ID / correlation** — outermost, registered first; accept `X-Request-Id` or mint one, so every downstream log line, metric, reject, and the `ProblemDetails` `requestId` extension can be correlated
   - **exception handling** — next (`UseExceptionHandler` + `AddProblemDetails`); it must be unbypassable for everything below it so any unhandled exception becomes a `500` `application/problem+json` that already carries the request ID
   - **routing** — `UseRouting` (framework contract: it must precede authorization)
   - **auth / rate limiting** — `UseAuthentication`/`UseAuthorization`/`UseRateLimiter`, before logging and metrics, so the access log records only requests that were actually allowed to proceed and a flood is rejected cheaply before metrics and Core run
   - **access logging / HTTP metrics** — one line and one measurement per admitted request
   - **endpoint filters → handler** — innermost, scoped to the API route group
6. **Exclude health probes from heavy middleware.** Map `/livez` and `/readyz` outside the API route group and keep auth, rate limiting, access logging, and per-request metrics attached to the API group (`group.RequireAuthorization()`, `group.RequireRateLimiting(...)`, group filters) rather than global `Use...` calls. Probes see only request ID + exception handling. See [../operations/observability.md](../operations/observability.md) and the probe contract in [../operations/operability.md](../operations/operability.md).
7. Keep it cheap and allocation-aware: it runs on every request. Avoid per-request closures over large state, gate optional work, and do not buffer bodies unless the concern requires it. Put no business logic here — middleware shapes the request/response envelope; domain decisions stay in `Orders.Core`.

## Invariants To Preserve

- deterministic, documented order: request ID → exception handling → routing → auth/rate-limit → logging/metrics → endpoint filters → handler
- exception handling cannot be bypassed by anything below it; every unhandled exception becomes a `500` `ProblemDetails`
- request-scoped data flows through `HttpContext` features or `TraceIdentifier` — never statics, never `AsyncLocal` globals, never string-keyed `Items`
- no business logic in middleware; it never calls `Orders.Core` to make domain decisions
- health probes (`/livez`, `/readyz`) are excluded from auth, rate limiting, access logging, and per-request metrics
- a rejecting middleware or filter short-circuits with a mapped `ProblemDetails` status (`401`/`429`) and does not call `next`
- metric/log labels stay low-cardinality (route pattern, status class — never request/user/tenant IDs)

## Proof

- Unit test: construct the middleware around a fake `RequestDelegate` that records whether it ran, drive it with a `DefaultHttpContext`, and assert both the middleware's effect (header set, feature readable) and ordering (the fake `next` observes the request ID the middleware set).
- Negative test: assert the short-circuit path — a missing/invalid credential yields `401` and the fake `next` was never invoked (`called == false`); for rate limiting, the (N+1)th request yields `429`.
- Probe-exclusion test: via `WebApplicationFactory<Program>`, drive `/readyz` and assert the heavy middleware did not fire (e.g. a counting fake records zero, or no auth challenge occurred).
- run `pwsh ./verify.ps1`
- Targeted run: `dotnet test tests/Orders.UnitTests` after touching only the middleware and its tests
