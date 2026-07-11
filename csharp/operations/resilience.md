# Resilience

The standardized outbound resilience policy a service applies to every dependency it calls so that a struggling upstream degrades gracefully instead of taking the caller down with it.

## Default Approach

Resilience is layered. Timeouts are the foundation; retries, breakers, and shedding are bounded controls stacked on top. Every outbound call (HTTP, gRPC, database, broker) carries a deadline and a failure classification. Unlike a stance-only policy, .NET ships the default implementation: `Microsoft.Extensions.Http.Resilience` and its standard handler on every typed client. Polly is the engine underneath — you configure it through the standard options, not by hand-rolling pipelines. A different resilience library, or a bespoke pipeline replacing the standard handler, is a [framework-selection](../decisions/framework-selection.md) exception that must justify itself.

Apply these controls at the client boundary, the same place you instrument calls in [../recipes/add-external-client.md](../recipes/add-external-client.md). Cancellation-token propagation and task-ownership rules from [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md) are prerequisites, not optional add-ons.

### The Standard Pipeline

Every typed outbound HTTP client gets the standard handler; it composes five strategies, outermost first:

```csharp
builder.Services.AddHttpClient<PaymentsClient>(c => c.BaseAddress = paymentsOptions.BaseAddress)
    .AddStandardResilienceHandler(o =>
    {
        o.Retry.DisableForUnsafeHttpMethods();   // opt back in per-call with an idempotency key
    });
```

| Order | Strategy | What it does |
|---|---|---|
| 1 | rate limiter | caps concurrent requests to the dependency; excess is rejected fast |
| 2 | total timeout | the overall budget for the call including all retries |
| 3 | retry | bounded attempts with exponential backoff and jitter; honors `Retry-After` |
| 4 | circuit breaker | stops sending to a dependency that is already failing |
| 5 | attempt timeout | per-attempt deadline inside the retry loop |

- Accept the shipped defaults first; tune per dependency only from measured latency, and read the current defaults off `HttpStandardResilienceOptions` rather than a blog post.
- Customize via the options lambda (per-client) or named options (fleet-wide) — never by stacking a second retry layer on top, which multiplies attempts.
- For latency-critical idempotent reads, the standard hedging handler is the sanctioned alternative; adopting it is a per-client decision recorded like any policy change.
- Non-HTTP dependencies (database, broker) do not get this handler but must satisfy the same layers: EF Core/Npgsql command timeouts and cancellation tokens per [../services/database.md](../services/database.md); consumer retry and dead-letter policy per [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).

### Timeouts Are The Foundation

- Every external call has a deadline. A call with no timeout is a latent outage: one hung upstream parks a request, then a connection, then the pool, then the process.
- Every async call path takes and forwards a `CancellationToken`, starting from `HttpContext.RequestAborted` (minimal APIs bind it as a parameter) or the worker's `stoppingToken`. A token that stops flowing is a deadline that stops working.
- Enforce a **total request budget** at the entry boundary with the request-timeouts middleware (`AddRequestTimeouts` + `UseRequestTimeouts`, per-group `WithRequestTimeout`). Per-dependency budgets must fit inside it; the standard handler's total timeout is the per-dependency slice, not the whole meal.
- Size attempt timeouts to the realistic latency of that dependency, not a round number copied from another service.
- Never rely on `HttpClient.Timeout` as the only control. It is a backstop; the pipeline's timeout strategies plus the flowing `CancellationToken` are the contract that cancels work everywhere, including in-flight downstream calls.

### Retries: Bounded, Idempotent-Only, Jittered

Retries trade latency for success probability. They are only safe when retrying cannot cause duplicate effects and only useful when the failure is transient.

- Retry **only idempotent or provably safe** operations: GETs, reads, and writes carrying an idempotency key the upstream honors (see [../recipes/add-idempotent-write.md](../recipes/add-idempotent-write.md)). The standard handler retries unsafe methods by default — call `Retry.DisableForUnsafeHttpMethods()` (or `DisableFor(...)`) unless every write through that client is idempotency-keyed. A retry after a timeout you cannot interpret may double a charge.
- **What is retryable:** transient network errors, `502`/`503`/`504`, `429` (the handler honors `Retry-After`), and timeouts that still leave budget. **What is not:** `4xx` validation/auth/not-found, request-shaping errors, and anything semantically terminal. A `400` will fail identically every attempt; retrying it only adds load.
- The default backoff is exponential with jitter — keep it. Jitter desynchronizes retrying clients and prevents the thundering-herd retry wave that turns a brief blip into an outage; never override to a fixed schedule.
- Every retry loop is bounded by **both** the max-attempts cap and the total timeout above it. Whichever fires first wins. An unbounded retry loop is an outage amplifier — this applies doubly to any hand-written loop outside the handler, which must check the `CancellationToken` before every sleep and attempt.
- Emit retry counters and exhaustion events per dependency — the resilience handler publishes them as metrics automatically; make sure they are exported (see [observability.md](observability.md)). A rising retry rate is an early outage signal.

For asynchronous message consumers, retry/backoff and dead-letter behavior follow the consumer policy in [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md); do not invent a second backoff model for the same broker.

### Circuit Breakers: When They Earn Their Place

A breaker stops sending traffic to a dependency that is already failing, giving it room to recover and failing the caller fast instead of piling on timeouts.

- The standard pipeline includes one; the design question is not "add a breaker?" but "are its thresholds right for this dependency, and what does the caller do when it opens?"
- States: **closed** (traffic flows, failures counted), **open** (calls fail fast without hitting the upstream, for a break window), **half-open** (limited probes test recovery; success closes, failure re-opens). An open breaker surfaces as a `BrokenCircuitException` — map it to the same fast-fail/degraded path the caller uses for load shedding (fast `503`, fallback, or degraded response), never a generic 500 after a full timeout.
- The breaker complements retries; it does not replace them. Retries handle a single blip; the breaker handles a sustained failure that retries would only make worse.
- Decide the degraded behavior per dependency: serve stale from `HybridCache` (see [../services/caching.md](../services/caching.md)), return a partial response with the failing section omitted, or fail fast — an explicit choice in the endpoint, not an accident of exception propagation.
- Breaker state transitions and trip counts are observable per dependency through the handler's built-in metrics; alert on a breaker that stays open.

### Rate Limiting And Throttling

Limit the rate at which the service issues calls to a dependency (client-side) or accepts work (server-side) so it stays inside a known-safe envelope.

- Client-side: the standard pipeline's rate limiter caps concurrent calls per dependency. Size it to the upstream's published or measured capacity, not to your own ambition — a client-side limit that protects an upstream from your own fan-out is cheaper than discovering its limit in production.
- Server-side: the built-in `RateLimiter` middleware (`AddRateLimiter` + `UseRateLimiter`, policies attached per endpoint group with `RequireRateLimiting`). Token bucket — steady refill plus bounded burst — is the default model; a pure concurrency limiter is the load-shedding variant below.
- Partition inbound limits per authenticated principal (the token subject) or per tenant via `RateLimitPartition`; fall back to per-IP only at anonymous edges where no identity exists. Partitioned limiters are bounded and evicted by the framework — or let the platform edge (gateway/ingress) own global limits and keep only coarse in-process admission control.
- Rejections return `429` with a `Retry-After` header (`RejectionStatusCode = StatusCodes.Status429TooManyRequests` plus an `OnRejected` callback that sets the header) so clients back off in a coordinated way.

### Load Shedding

When demand exceeds capacity, shed work fast and cheaply instead of queueing it until everything collapses.

- Bound **in-flight concurrency** per server with a concurrency-limiter policy (`AddConcurrencyLimiter`) sized to measured capacity. When the bound is reached, reject immediately.
- Reject with a fast `429 Too Many Requests` or `503 Service Unavailable` plus `Retry-After`.
- **Drop before queues grow unbounded.** Keep the limiter's `QueueLimit` at zero or very small: an unbounded queue under overload just converts a fast failure into a slow one while latency climbs for everyone. A bounded queue with fast rejection keeps healthy requests fast.
- Prefer shedding the newest/excess work; protect the requests already in flight and within budget. Kestrel's connection limits are the outermost backstop, not the strategy.

### Bulkheads

Isolate resources per dependency so one slow upstream cannot starve calls to healthy ones.

- Give each dependency its **own** typed client. Each named/typed client owns its own handler chain, connection pool, and resilience pipeline — the bulkhead falls out of the registration pattern; sharing one `HttpClient` across dependencies collapses them into one pool and one failure domain.
- The per-dependency rate limiter in the standard pipeline is the concurrency bulkhead: a slow upstream saturates its own limiter and its callers fail fast, while other dependencies proceed untouched.
- Size each pool and limiter to that dependency's role; see [../services/database.md](../services/database.md) for connection-pool sizing on the DB side.
- Combine bulkheads with the per-dependency breaker: the blast radius of a failing dependency is then confined to its own bulkhead.

### Composition Order

Per outbound call, the controls nest from outside in: **bulkhead/rate limiter** (am I allowed to send) → **total timeout** (the budget) → **retry loop** (bounded, jittered) → **circuit breaker** (is this dependency healthy) → **attempt timeout** → the actual call. This is exactly the standard handler's order — another reason not to hand-assemble it. Failures and tripped controls map to the fast-fail/degraded path, never to an unbounded wait. Deployment of new policy values (timeouts, attempt caps, breaker thresholds) follows [deployment.md](deployment.md) like any config change.

## Common Mistakes And Forbidden Patterns

- An outbound call that ignores or fails to forward the `CancellationToken`, so a hung upstream holds a request slot forever.
- `new HttpClient()` per call or a raw singleton without the factory — no pipeline, no bulkhead, and socket exhaustion under load.
- Hand-rolled retry loops with `Task.Delay` and no jitter, no attempt cap, or no cancellation check — use the standard handler.
- Retrying non-idempotent writes (un-keyed POSTs, charges, publishes) because the default was never audited with `DisableForUnsafeHttpMethods()`.
- Retrying `4xx`/validation errors that will fail identically every attempt.
- Stacking a custom Polly pipeline on top of the standard handler, silently multiplying retries and stretching the real budget.
- A breaker used as a substitute for, rather than a complement to, bounded retries and timeouts — or `BrokenCircuitException` surfacing as a generic 500.
- Unbounded in-flight concurrency or a rate limiter with an unbounded queue, turning overload into a cascading collapse instead of fast rejection.
- A single shared `HttpClient`/pool across all dependencies, so one slow upstream exhausts everything (no bulkhead).
- Adopting a different resilience library without routing it through [framework-selection](../decisions/framework-selection.md).
- High-cardinality retry/breaker metric labels (per-host, per-request-id) instead of per-dependency labels.

## Verification And Proof

- Every external call path forwards a `CancellationToken`; a test asserts a hung upstream fails fast within the configured budget, not after a transport default.
- Retry policy proven bounded: a test shows max-attempts and total-timeout caps both fire, and backoff carries jitter (not a fixed schedule).
- Non-idempotent operations are excluded from automatic retry — a test proves a failing POST through the client is attempted exactly once (do not trust the configuration by inspection).
- A load test under overload shows graceful shedding (fast `429`/`503` with `Retry-After`, stable latency for admitted requests) rather than collapse or unbounded latency growth.
- Breaker open/half-open/closed transitions and retry counts are observable per dependency in exported metrics (see [observability.md](observability.md)).
- run `pwsh ./verify.ps1`; resilience-policy tests use `FakeTimeProvider` so backoff and breaker windows run deterministically without real sleeps.
