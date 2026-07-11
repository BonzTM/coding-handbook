# Recipe: Add HTTP Endpoint

Use this when a feature adds or changes one HTTP route.

Governing doc: [`csharp/services/http-services.md`](../services/http-services.md).

## Files To Touch

- `src/Orders.Api/Endpoints/<Feature>Endpoints.cs` — the endpoint group: route registration plus handler methods
- `src/Orders.Api/Endpoints/<Feature>Contracts.cs` (or the same file) — request/response DTO records
- the `JsonSerializerContext` for the Api project if a new DTO crosses the wire (see [../foundations/serialization.md](../foundations/serialization.md))
- `src/Orders.Core/...` for business logic
- `src/Orders.Api/Program.cs` only if a brand-new route group must be mapped
- unit tests in `tests/Orders.UnitTests` and endpoint tests in `tests/Orders.IntegrationTests`

## Steps

1. Define request and response DTOs as records in `Orders.Api`, not in `Orders.Core`. Annotate the request DTO with DataAnnotations; the built-in minimal-API validation (`AddValidation`) rejects invalid bodies with a `400` validation problem before the handler runs. Register both DTOs in the source-generated `JsonSerializerContext`.
2. Add or update the `Orders.Core` method that owns the behavior. The handler translates; Core decides.
3. Implement the handler as a static method whose return type is the `Results<...>` union of every outcome (e.g. `Results<Ok<OrderResponse>, NotFound, ValidationProblem>`) and construct each with `TypedResults` — never bare `Results.Ok(object)`. Decode/bind, call Core, map domain failures to statuses, encode the response.
4. Register the route on the feature's endpoint group (`MapGroup("/orders")`) with a stable method and path pattern, and name it (`.WithName(...)`) so links and telemetry stay stable.
5. Confirm the pipeline already covers request ID, exception-to-`ProblemDetails` conversion, access logging, and auth for this group — if not, follow [add-http-middleware.md](add-http-middleware.md) first.
6. Add telemetry for the route through the existing HTTP instrumentation, or an explicit metric per [add-metric.md](add-metric.md).

## Invariants To Preserve

- no `DbContext` or database access directly in the handler — call the Core service or port
- the request's `CancellationToken` flows through the full call chain
- every error leaves as RFC 9457 `ProblemDetails` (`application/problem+json`) with the `requestId` extension; status mapping is consistent with existing transport behavior
- request bodies are size-limited and validated before Core logic runs

## Proof

- endpoint tests in-process via `WebApplicationFactory<Program>` asserting status, content type, and body
- one negative test for validation (`400` with the field→messages `errors` extension) or auth (`401`)
- one smoke request against a locally running host
- log or metric review showing the new route is observable
