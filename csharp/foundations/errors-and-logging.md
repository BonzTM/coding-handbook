# Errors and Logging

Error semantics and structured logging rules that keep failures actionable instead of noisy.

## Default Approach

### Exception Semantics

- Exceptions are the error channel. Throw the most specific type that names the failure; `throw new Exception(...)` is forbidden.
- Wrap with context when crossing a subsystem boundary, always preserving the cause: `throw new OrderStoreUnavailableException($"persisting order {id} failed", ex);`. Discarding the inner exception discards the diagnosis.
- Re-throw with `throw;`, never `throw ex;` — the latter resets the stack trace (CA2200).
- Add context at package or subsystem boundaries (repository, outbound client, consumer), not mechanically at every stack frame.
- Catch only where you can act: map to a wire response, retry with a bound, fall back, or enrich and rethrow. A `catch` that only logs and rethrows is a layer that cannot act — delete it.
- Use `catch (X ex) when (...)` filters to avoid catching what you cannot handle; the filter runs before the stack unwinds.
- `OperationCanceledException` from the caller's own token is not an error — let it propagate; see [cancellation-and-async.md](cancellation-and-async.md).

### Custom Exception Design

```csharp
public sealed class OrderNotFoundException(OrderId orderId)
    : Exception($"Order {orderId} was not found.")
{
    public OrderId OrderId { get; } = orderId;
}

public sealed class OrderStoreUnavailableException : Exception
{
    public OrderStoreUnavailableException(string message) : base(message) { }
    public OrderStoreUnavailableException(string message, Exception inner) : base(message, inner) { }
}
```

- Name ends in `Exception`; derive directly from `Exception` (never `ApplicationException`); `sealed` unless a hierarchy is genuinely needed.
- Carry structured data as properties so handlers branch on values, never by parsing `Message`.
- One type per condition callers branch on. Do not build speculative hierarchies; two or three domain exceptions per service is typical.
- Provide a `(string message, Exception inner)` constructor on any type used for wrapping.
- Domain exceptions live in `Orders.Core`; infrastructure wrappers live in `Orders.Infrastructure`.
- Analyzer interplay: CA1032 demands parameterless and message-only constructors — the exact constructors this design forbids (an `OrderNotFoundException` without an order id is useless). The canonical [.editorconfig](../templates/.editorconfig) disables CA1032 for this reason; wrapping constructors are enforced by review, not by the analyzer.

### When To Return Instead Of Throw

Exceptions are for exceptional flow. Expected, frequent outcomes — especially on hot paths — use result shapes, because a thrown exception costs orders of magnitude more than a return and hides the branch from the type system.

- **TryX pattern** for parse/lookup style operations, matching the framework's own idiom (`int.TryParse`, `TryGetValue`): `bool TryParseSku(string raw, out Sku sku)`.
- **Domain results** when every caller must branch on the outcome:

```csharp
public abstract record PlaceOrderResult
{
    public sealed record Placed(OrderId Id) : PlaceOrderResult;
    public sealed record OutOfStock(Sku Sku) : PlaceOrderResult;
    public sealed record CreditRejected(string Reason) : PlaceOrderResult;
}
```

- Rule of thumb: if every caller handles it, it is a result; if only a boundary handles it, it is an exception. Never expose the same condition through both channels.
- No general-purpose `Result<T>` library by default — model the specific domain outcomes. A generic result monad across the codebase requires an ADR ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).
- Analyzer interplay: the nested sealed case records are what seal the hierarchy, which trips CA1034 (no nested public types). The canonical [.editorconfig](../templates/.editorconfig) disables CA1034 so this pattern builds under warnings-as-errors.

### Error Categories

| Kind | Use for | Example handling |
|---|---|---|
| opaque internal failure | callers only need success or failure | caught once at the boundary, logged, mapped to a 500 ProblemDetails |
| framework exception (`ArgumentException`, `InvalidOperationException`) | programmer error, violated invariants | fix the calling code; never caught in normal flow |
| domain exception | a boundary branches on a stable condition | `catch (OrderNotFoundException)` → 404 ProblemDetails |
| domain result | expected business outcome every caller handles | pattern match, map to 2xx/4xx |

### Exception Middleware And ProblemDetails

Exactly one boundary owns exception-to-wire mapping. The wire error contract is RFC 9457 `ProblemDetails` (`application/problem+json`) via ASP.NET Core's native support — `AddProblemDetails` + `UseExceptionHandler` + `IExceptionHandler` implementations for domain mappings. Every response carries a `requestId` extension; validation failures carry an `errors` extension (field → messages map).

```csharp
builder.Services.AddProblemDetails(options =>
    options.CustomizeProblemDetails = ctx =>
        ctx.ProblemDetails.Extensions["requestId"] = ctx.HttpContext.TraceIdentifier);
builder.Services.AddExceptionHandler<DomainExceptionHandler>();
```

- Unhandled exceptions map to an opaque 500: the body never contains the exception message, type, or stack trace — those leak internals. The middleware logs the exception once at `Error` with request context; that log is the diagnostic, the ProblemDetails `requestId` is the correlation key.
- Domain exceptions map to their status in one `IExceptionHandler`, not in scattered try/catch per endpoint. Full endpoint wiring: [../services/http-services.md](../services/http-services.md).

### Structured Logging

- `ILogger<T>` by constructor injection. No static logger, no ambient global.
- Every recurring log site uses the `[LoggerMessage]` source generator (CA1848 flags direct `LogX` calls as the slower path):

```csharp
internal static partial class OrderLog
{
    [LoggerMessage(Level = LogLevel.Information, Message = "Order {OrderId} placed for customer {CustomerId} with {LineCount} lines")]
    public static partial void OrderPlaced(ILogger logger, OrderId orderId, CustomerId customerId, int lineCount);

    [LoggerMessage(Level = LogLevel.Error, Message = "Persisting order {OrderId} failed")]
    public static partial void OrderPersistFailed(ILogger logger, OrderId orderId, Exception exception);
}
```

- Message templates use named placeholders. String interpolation or concatenation in a log call is forbidden — it destroys the structured fields and CA2254 rejects non-constant templates.
- JSON console logs in production, human-readable locally; formatter selection is configuration ([configuration.md](configuration.md)).
- Correlation fields (`RequestId`, `TraceId`, `SpanId`) come from ASP.NET Core and OpenTelemetry — do not hand-roll them. See [../operations/observability.md](../operations/observability.md).
- `Orders.Core` does not log. Domain code throws or returns; the hosts (Api, workers) decide what is log-worthy. This keeps Core free of infrastructure concerns and keeps log placement deliberate.

### Log Levels Contract

| Level | Meaning | Production |
|---|---|---|
| `Trace` | payload-level detail for local debugging | never enabled |
| `Debug` | diagnostic flow detail | off by default, enabled per-category during incidents |
| `Information` | business-meaningful events (order placed, worker started) | on |
| `Warning` | handled anomaly; degraded but self-correcting (retry succeeded, fallback used) | on |
| `Error` | an operation failed; an operator or alert may act | on, alertable |
| `Critical` | the process or a required dependency is unusable | on, pages |

### Scopes

- `BeginScope` attaches ambient fields to every log entry in a unit of work — use it where a unit spans many log sites (a message consumer attaching `MessageId`, a job attaching `JobRunId`):

```csharp
using (logger.BeginScope(new Dictionary<string, object> { ["OrderId"] = order.Id.Value }))
{
    // every log entry in here carries OrderId
}
```

- Enable `IncludeScopes` in the JSON formatter or scope fields silently vanish.
- HTTP requests already carry `RequestId`/`RequestPath` scope from the framework; do not duplicate them.

## Log Placement Rules

- Request lifecycle (status, latency) is logged once per request by the telemetry wiring, not per endpoint.
- The exception middleware logs unhandled exceptions once at `Error`. Layers below it do not log the same exception — wrap and rethrow instead. One failure, one `Error` entry.
- Background workers log lifecycle (started, stopping) and actionable failures.
- Repositories and outbound clients log retries, dependency failures, and unusual latency when that helps operators.
- `Orders.Core` and reusable libraries do not log; they throw or return and let the host decide.

## Common Mistakes And Forbidden Patterns

- `catch (Exception)` anywhere except the exception middleware and worker supervisors; empty catch blocks anywhere.
- `throw ex;` instead of `throw;` (CA2200).
- Control flow by exception on hot paths — catching `FormatException` in a parse loop instead of `TryParse`.
- Logging the same failure at every layer on its way up.
- `logger.LogInformation($"order {id} placed")` — interpolation in log calls (CA2254); templates are constants with named placeholders.
- Logging secrets, tokens, auth headers, connection strings, or PII; logging raw request/response bodies by default. See [../operations/data-handling.md](../operations/data-handling.md).
- Returning `exception.Message` or stack traces to clients in ProblemDetails `detail`.
- Catching `OperationCanceledException` during shutdown and logging it as an error.
- Inventing string error codes matched with `==` instead of exception types or result shapes.
- Deriving from `ApplicationException`, or one giant `OrdersException` for every failure kind.

## Verification And Proof

- Run `pwsh ./verify.ps1` — the analyzer stage enforces CA2200, CA2254, CA1848, and friends as errors ([../quality/linting.md](../quality/linting.md)).
- Unit tests prove domain exceptions and results stay matchable: assert the thrown type and its structured properties, not message substrings.
- Integration tests assert the ProblemDetails contract: a failing request returns `application/problem+json` with `requestId` and (for validation) `errors`, and no internal detail.
- Review one successful and one failing request path end to end: stable structured fields, exactly one `Error` entry for the failure, no sensitive data.
- Search the repo for `Console.WriteLine`, `catch { }`, and `$"` inside log calls before calling the logging shape consistent.
