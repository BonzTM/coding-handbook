# HTTP Services

HTTP service defaults for repos that want predictable endpoints, cheap debugging, and minimal framework ceremony.

## Default Approach

- Use ASP.NET Core Minimal APIs with route groups and `TypedResults` first. MVC controllers require an ADR per [../decisions/framework-selection.md](../decisions/framework-selection.md).
- Keep transport code in `Orders.Api` and core behavior in `Orders.Core`. `Orders.Api` references `Orders.Core`; core never references ASP.NET.
- Treat endpoints as translation layers: bind, validate, call core, map errors, encode response.

### Minimal Layout

```text
src/Orders.Api/
  Program.cs                     # composition root: builder, middleware order, endpoint mapping
  Endpoints/
    OrderEndpoints.cs            # one endpoint group per resource
  Contracts/
    CreateOrderRequest.cs        # wire DTOs with validation attributes
    OrderResponse.cs
  Middleware/
    RequestIdMiddleware.cs       # request id, other cross-cutting transport concerns
  Telemetry/
    ServiceTelemetryExtensions.cs  # AddServiceTelemetry() wiring
```

### Routing Pattern

One static class per resource, one `MapXxxEndpoints` extension, registered from `Program.cs`. The canonical route-registration unit is the **endpoint group**:

```csharp
namespace Orders.Api.Endpoints;

public static class OrderEndpoints
{
    public static IEndpointRouteBuilder MapOrderEndpoints(this IEndpointRouteBuilder routes)
    {
        var group = routes.MapGroup("/orders").WithTags("Orders");

        group.MapGet("/{id:guid}", GetOrder).WithName("GetOrder");
        group.MapPost("/", CreateOrder);
        group.MapGet("/", ListOrders);

        return routes;
    }

    private static async Task<Results<Ok<OrderResponse>, NotFound>> GetOrder(
        Guid id, IOrderService orders, CancellationToken cancellationToken)
    {
        var order = await orders.GetAsync(new OrderId(id), cancellationToken);
        return order is null
            ? TypedResults.NotFound()
            : TypedResults.Ok(OrderResponse.From(order));
    }
}
```

`TypedResults` with a `Results<...>` union return type is mandatory: the compiler enforces the set of statuses an endpoint can produce, and the union is the endpoint's documented contract. Bare `Results.Ok(...)` (untyped `IResult`) hides the contract and is forbidden in new endpoints. Route-group-level filters and metadata (`RequireAuthorization`, `RequireRateLimiting`, `WithRequestTimeout`) belong on the group, not repeated per route.

### Endpoint Contract

Each endpoint handler should usually do five things in order:

1. accept bound route/query/body parameters and a `CancellationToken` (always last)
2. rely on validation having already rejected malformed input (see Request Validation)
3. call one core service method
4. map domain errors to a typed result / ProblemDetails
5. return a `TypedResults` value; telemetry is emitted by middleware, not hand-rolled per endpoint

If the HTTP surface is consumed outside one codebase, define and review the payload contract explicitly — OpenAPI generated from the endpoint metadata (`AddOpenApi`) or a documented DTO set in `Contracts/`, but one source of truth.

HTTP paths are unversioned by default — serve `/orders`, not `/v1/orders`. Evolve the contract additively per [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md); a breaking change means a new resource or an explicitly negotiated new surface, never a silent mutation of shapes existing clients depend on.

### Request Validation

Use the built-in minimal-API validation: DataAnnotations on request DTOs, enabled once at startup. Invalid requests are rejected with a `400` `ValidationProblemDetails` before the handler runs.

```csharp
builder.Services.AddValidation();
```

```csharp
namespace Orders.Api.Contracts;

public sealed record CreateOrderRequest(
    [property: Required, StringLength(64, MinimumLength = 1)] string CustomerId,
    [property: Range(1, 1_000)] int Quantity);
```

- Validation attributes live on the wire DTO in `Contracts/`, never on `Orders.Core` domain types.
- Wire DTOs must be `public`: the validation source generator behind `AddValidation()` silently skips `internal` types — no error, no validation. This is why the canonical [.editorconfig](../templates/.editorconfig) disables CA1515 under `Contracts/`; do not "fix" a DTO to `internal`.
- Validation checks shape and range. Business rules (inventory, tenancy, state transitions) live in core and surface as domain errors, not annotations.
- Opt an endpoint out only with `.DisableValidation()` plus a comment saying why (e.g., raw-stream ingest). FluentValidation is an ADR, not a default.

### Error Responses

The wire error contract is RFC 9457 `ProblemDetails` (`application/problem+json`) — ASP.NET Core's native shape. Wire it once:

```csharp
builder.Services.AddProblemDetails(options =>
    options.CustomizeProblemDetails = context =>
        context.ProblemDetails.Extensions["requestId"] = context.HttpContext.TraceIdentifier);
builder.Services.AddExceptionHandler<DomainExceptionHandler>(); // domain error → status mapping
```

```csharp
app.UseExceptionHandler();
app.UseStatusCodePages();
```

Map each domain error to a `(status, type)` pair in one place — an `IExceptionHandler` for thrown exceptions plus a small result-mapping helper for returned domain errors. Validation failures carry the field→messages map in the standard `errors` extension; every problem response carries `requestId`. The envelope rules are defined once in [../foundations/serialization.md](../foundations/serialization.md); the domain-error-to-status mapping lives in [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md). Never return a bare `{"error":"..."}` string, and never leak exception messages, stack traces, or connection strings in a 5xx body — the default `UseExceptionHandler` + `AddProblemDetails` pipeline already withholds detail outside Development; keep it that way.

### Server Hardening Defaults

- a default request timeout via the RequestTimeouts middleware, with named policies for known-slow endpoints
- a request body size cap sized to the real payloads, not the 30 MB Kestrel default
- separate liveness and readiness endpoints (see Health Endpoints)
- a bounded shutdown drain (see Graceful Shutdown And Drain)

```csharp
builder.Services.AddRequestTimeouts(options =>
{
    options.DefaultPolicy = new RequestTimeoutPolicy { Timeout = TimeSpan.FromSeconds(10) };
    options.AddPolicy("slow-report", TimeSpan.FromSeconds(60));
});
builder.WebHost.ConfigureKestrel(kestrel =>
    kestrel.Limits.MaxRequestBodySize = 1 * 1024 * 1024); // per-endpoint overrides for uploads
```

The timeout middleware only cancels the request's `CancellationToken` — it protects nothing unless every endpoint and every downstream call actually honors the token, per [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md). An endpoint that ignores its token has no timeout, whatever the policy says.

### Transport Security

The default for an HTTP service is TLS terminated at the platform edge — load balancer, ingress, or mesh sidecar — with Kestrel listening plaintext inside the pod boundary via `ASPNETCORE_URLS=http://+:8080`. Do not double-terminate: when the platform already owns the edge, an in-app TLS listener only adds cert-rotation surface and a second failure mode. Behind a proxy, wire `UseForwardedHeaders` (with an explicit known-proxy list) so scheme and client IP survive; never trust forwarded headers from arbitrary sources.

If the service is standalone or edge-exposed and must terminate TLS itself, configure the Kestrel HTTPS endpoint from config (cert path + key path), and gate it: empty cert config selects the plaintext local/dev listener, and a *set but unloadable* cert fails fast at startup — never a silent downgrade to plaintext. Only a host that terminates TLS itself enables `UseHsts()` and `UseHttpsRedirection()`; behind an edge they are the platform's job. See [../operations/security.md](../operations/security.md) for the full stance and [grpc-services.md](grpc-services.md) for the gRPC equivalent.

### Middleware Order

Use a stable order so failures are observable and safe. `Program.cs` is the single place this order lives:

```csharp
var app = builder.Build();

app.UseMiddleware<RequestIdMiddleware>(); // 1. request ID — outermost
app.UseExceptionHandler();                // 2. unhandled exception → ProblemDetails
app.UseStatusCodePages();
// app.UseHsts();                         //    only when this host terminates TLS
app.UseRouting();                         // 3. route match available downstream
// app.UseCors("browser-clients");        // 4. only when spec intake finds browser clients
app.UseAuthentication();                  // 5. authN before logging/limiting
app.UseAuthorization();
app.UseRateLimiter();                     // 6. after auth so partitions can key on identity
app.UseRequestTimeouts();                 // 7. innermost before endpoints

app.MapOrderEndpoints();
```

Request ID sits outside the exception handler so the id is on the context before the handler runs — an unhandled-exception 500 still carries `requestId` in its ProblemDetails body. Authentication stays ahead of request logging so unauthenticated requests are rejected early; per-route *authorization* metadata (`RequireAuthorization` on the group) runs at the endpoint so the authz decision and the matched route land in the access log together. New cross-cutting behavior goes through [../recipes/add-http-middleware.md](../recipes/add-http-middleware.md), which preserves this order.

### CORS

The default is **no CORS middleware**. CORS is a browser mechanism; a service-to-service API has no browser clients and needs no CORS layer — adding one by reflex only widens the attack surface. Enable it only when [spec intake](../checklists/spec-intake.md) identifies browser clients.

When it is enabled, the policy is:

- a named policy (`AddCors` + `AddPolicy`) with an **explicit origin allowlist** driven from config — never `AllowAnyOrigin`, and never any wildcard-equivalent combined with `AllowCredentials`; ASP.NET Core refuses that combination at runtime, do not work around it with `SetIsOriginAllowed(_ => true)`
- correct preflight handling comes with the middleware: `OPTIONS` is answered with the allowed methods/headers and a sane `SetPreflightMaxAge`, without invoking the real endpoint
- `UseCors` sits after `UseRouting` and **before authentication**, so preflights (which carry no credentials) are answered instead of rejected

### Rate Limiting

Use the built-in RateLimiter middleware. Two non-defaults are mandatory:

```csharp
builder.Services.AddRateLimiter(options =>
{
    options.RejectionStatusCode = StatusCodes.Status429TooManyRequests; // default is 503 — wrong
    options.OnRejected = (context, _) =>
    {
        if (context.Lease.TryGetMetadata(MetadataName.RetryAfter, out var retryAfter))
        {
            context.HttpContext.Response.Headers.RetryAfter =
                ((int)retryAfter.TotalSeconds).ToString(CultureInfo.InvariantCulture);
        }
        return ValueTask.CompletedTask;
    };
    options.AddPolicy("per-client", context =>
        RateLimitPartition.GetTokenBucketLimiter(
            context.User.Identity?.Name
                ?? context.Connection.RemoteIpAddress?.ToString()
                ?? "anonymous",
            _ => new TokenBucketRateLimiterOptions
            {
                TokenLimit = 100,
                TokensPerPeriod = 100,
                ReplenishmentPeriod = TimeSpan.FromSeconds(1),
                QueueLimit = 0,
            }));
});
```

Reject with `429` plus `Retry-After` — a well-behaved client backs off exactly as [../operations/resilience.md](../operations/resilience.md) tells our own clients to. Partition by authenticated identity first, falling back to client address; a single global bucket punishes every tenant for one noisy one. Apply the policy at the group (`group.RequireRateLimiting("per-client")`), and keep limits in config, not literals, per [../foundations/configuration.md](../foundations/configuration.md).

### List Endpoints And Pagination

Every endpoint that returns a collection is paginated from the start — an unbounded list is a latency and memory incident waiting for the table to grow.

- Default to keyset/cursor pagination (an opaque `nextCursor` over a stable sort key) for large or append-heavy sets; offset/limit is acceptable only for small, bounded, or admin lists where deep pages are rare.
- Enforce a default and a maximum page size on the server; clamp an oversized request rather than rejecting it, and never honor an unbounded one.
- Sort on a total, stable order (the cursor key plus a tiebreaker such as the primary key) so pages neither overlap nor skip rows under concurrent writes. The EF Core query shape lives in [database.md](database.md).
- Return exact totals only when they are cheap; for large tables prefer a `hasMore` flag or `nextCursor` over a full count.
- Keep one consistent list envelope (an `Items` array plus pagination metadata) across endpoints, and treat it as a wire contract per [../foundations/serialization.md](../foundations/serialization.md).

### Idempotent Writes

Any write a client may retry — every `POST` behind a queue, a mobile client, or a payment flow — carries a client-supplied idempotency key and is deduplicated server-side. The full pattern (key header, uniqueness constraint, replay response) is [../recipes/add-idempotent-write.md](../recipes/add-idempotent-write.md); the storage side is in [database.md](database.md). `PUT`/`DELETE` are idempotent by contract — keep them that way.

### Graceful Shutdown And Drain

On SIGTERM the host stops accepting new connections and Kestrel drains in-flight requests until `HostOptions.ShutdownTimeout` expires. Set the timeout below the orchestrator's kill grace period:

```csharp
builder.Services.Configure<HostOptions>(options =>
    options.ShutdownTimeout = TimeSpan.FromSeconds(25)); // pod grace period is 30s
```

Background work observes shutdown through the `stoppingToken` in `BackgroundService.ExecuteAsync` — never `Environment.Exit`, never a hand-rolled signal handler. Readiness flips unhealthy during drain so the load balancer stops routing before connections close. See [../operations/operability.md](../operations/operability.md) for the rollout contract.

### Health Endpoints

Map liveness and readiness separately, from `Microsoft.Extensions.Diagnostics.HealthChecks`:

```csharp
app.MapHealthChecks("/livez", new HealthCheckOptions { Predicate = _ => false });
app.MapHealthChecks("/readyz", new HealthCheckOptions { Predicate = r => r.Tags.Contains("ready") });
```

`/livez` runs no checks — it answers "is the process serving" and must stay cheap and dependency-free. `/readyz` runs the checks tagged `ready` (database, broker, downstream dependencies needed for traffic). Readiness is "dependencies needed for traffic are ready", never "process is running". Registration details live in [../operations/observability.md](../operations/observability.md).

## Common Mistakes And Forbidden Patterns

- Endpoints injecting `DbContext` and querying directly instead of going through core or repository seams.
- Untyped `Results.Ok(...)`/`IResult` returns that hide the endpoint's status contract, or business logic living in lambda bodies in `Program.cs` instead of endpoint groups.
- Logging full request or response bodies by default.
- Returning `Orders.Core` domain types on the wire instead of `Contracts/` DTOs — every core refactor becomes a silent breaking change.
- Changing public request or response shapes without compatibility review.
- Treating readiness as "process is running" rather than "dependencies needed for traffic are ready".
- Returning an unbounded or unpaginated collection, or deep offset pagination that scans the whole prefix of a large table.
- An endpoint without a `CancellationToken` parameter, or a timeout policy over handlers that ignore the token — the timeout exists only on paper.
- Rate limiting left on its `503` rejection default, or without `Retry-After` — clients cannot distinguish "slow down" from "I am down".
- Double-terminating TLS behind an edge that already owns it — or, when the service must terminate TLS itself, silently falling back to plaintext when a configured cert fails to load instead of failing fast.
- A CORS policy that reflects arbitrary origins or pairs a wildcard origin with credentials — either one hands any website credentialed access to the API.
- Exception details, stack traces, or `DeveloperExceptionPage` behavior reaching production responses.

## Verification And Proof

- endpoint tests through `WebApplicationFactory<Program>` (in-proc, real pipeline, no network)
- negative tests for validation (`400` + `errors` map), auth (`401`/`403`), and domain-error mapping (`404`/`409` ProblemDetails with `requestId`)
- one smoke test against `/livez` and `/readyz`
- a rate-limit test asserting `429` and a `Retry-After` header once the bucket is exhausted
- audit that timeout, body-size, and middleware order are actually wired in `Program.cs`
- run `pwsh ./verify.ps1` — restore (locked), format-check, build (warnings-as-errors), test, audit
