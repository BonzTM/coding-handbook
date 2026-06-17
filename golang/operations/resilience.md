# Resilience

The standardized outbound resilience policy a service applies to every dependency it calls so that a struggling upstream degrades gracefully instead of taking the caller down with it.

## Default Approach

Resilience is layered. Timeouts are the foundation; retries, breakers, and shedding are bounded controls stacked on top. Every outbound call (HTTP, gRPC, database, broker) carries a deadline and a failure classification. The concrete retry/breaker/rate-limit library is a [framework-selection](../decisions/framework-selection.md) decision: this doc fixes the stance and the defaults, and the library choice is routed there as an exception that must justify itself.

Apply these controls at the client boundary, the same place you instrument calls in [recipes/add-external-client.md](../recipes/add-external-client.md). Context propagation and goroutine ownership rules from [foundations/context-and-concurrency.md](../foundations/context-and-concurrency.md) are prerequisites, not optional add-ons.

### Timeouts Are The Foundation

- Every external call has a deadline. A call with no timeout is a latent outage: one hung upstream blocks a goroutine, then a connection, then the pool, then the process.
- Use a **per-call** context deadline (`context.WithTimeout`) sized to the realistic latency of that dependency, not a round number copied from another service.
- Enforce a **total request budget** at the entry boundary (handler, consumer loop iteration, job). Per-call deadlines must fit inside it; deduct elapsed time as you fan out so the last downstream call cannot exceed the budget that remains.
- Never set transport-level timeouts (`http.Client.Timeout`, dial/TLS/read timeouts) as the only control. They are a backstop; the context deadline is the contract that flows through the stack and cancels work everywhere.
- The deadline propagates: a cancelled or expired caller context must abort in-flight downstream calls, not leak them.

### Retries: Bounded, Idempotent-Only, Jittered

Retries trade latency for success probability. They are only safe when retrying cannot cause duplicate effects and only useful when the failure is transient.

- Retry **only idempotent or provably safe** operations: GETs, reads, and writes carrying an idempotency key the upstream honors. Never auto-retry a non-idempotent write (a POST that creates, a charge, a non-keyed publish) — a retry after a timeout you cannot interpret may double the effect.
- **What is retryable:** transient network errors (connection refused, reset, DNS blip), `503`/`502`/`504`, `429` (honor `Retry-After`), and context-deadline failures that still leave budget. **What is not:** `4xx` validation/auth/not-found, request-shaping errors, and anything semantically terminal. A `400` will fail identically every attempt; retrying it only adds load.
- Use **exponential backoff with full jitter**: `sleep = random_between(0, min(cap, base * 2^attempt))`. Full jitter (not "equal jitter", not fixed backoff) is the default because it desynchronizes retrying clients and prevents the thundering-herd retry wave that turns a brief blip into an outage.
- Bound every retry loop with **both** a max-attempts cap (e.g. 3 total attempts) **and** an overall deadline derived from the request budget. Whichever fires first wins. An unbounded retry loop is an outage amplifier.
- Respect cancellation between attempts: check `ctx.Err()` before sleeping and before the next try, and use a timer that the context can cancel.
- Emit retry counters and exhaustion events per dependency (see [operations/observability.md](observability.md)). A rising retry rate is an early outage signal.

For asynchronous message consumers, retry/backoff and dead-letter behavior follow the consumer policy in [services/eventing-and-messaging.md](../services/eventing-and-messaging.md); do not invent a second backoff model for the same broker.

### Circuit Breakers: When They Earn Their Place

A breaker stops sending traffic to a dependency that is already failing, giving it room to recover and failing the caller fast instead of piling on timeouts.

- Add a breaker when a dependency can get **overloaded** and retries alone would keep it down (retry storms), or when slow failures from one upstream would exhaust the caller's resources. A breaker is not justified for every client; a dependency that fails cleanly and rarely does not need one.
- States: **closed** (traffic flows, failures counted), **open** (calls fail fast without hitting the upstream, for a cooldown window), **half-open** (a limited probe of calls tests recovery; success closes, failure re-opens).
- The breaker complements retries; it does not replace them. Retries handle a single blip; the breaker handles a sustained failure that retries would only make worse.
- A breaker that trips must surface the same fast-failure path the caller uses for load shedding (fast `503`, fallback, or degraded response) rather than a generic timeout.
- The concrete breaker library is deferred to [framework-selection](../decisions/framework-selection.md). State the stance: prefer a small, transparent breaker over a framework that hides state transitions; whatever is chosen must expose state and trip metrics.

### Rate Limiting And Throttling

Limit the rate at which the service issues calls to a dependency (client-side) or accepts work (server-side) so it stays inside a known-safe envelope.

- Default model is a **token bucket**: a steady refill rate plus a bounded burst. The conventional Go choice is `golang.org/x/time/rate` (`rate.NewLimiter(r, b)` with `Wait`/`Allow`/`Reserve`); it is stdlib-adjacent and transparent. Adopting it is still recorded through [framework-selection](../decisions/framework-selection.md) like any dependency.
- Use `Wait(ctx)` for outbound client throttling so the limiter respects the caller's deadline; use `Allow()` for server-side admission where you reject rather than block.
- Size the limiter to the upstream's published or measured capacity, not to your own ambition. A client-side limit that protects an upstream from your own fan-out is cheaper than discovering its limit in production.

### Load Shedding

When demand exceeds capacity, shed work fast and cheaply instead of queueing it until everything collapses.

- Bound **in-flight concurrency** per server with a semaphore (a buffered channel of `struct{}` or `golang.org/x/sync/semaphore`). When the bound is reached, reject immediately.
- Reject with a fast `429 Too Many Requests` or `503 Service Unavailable` plus a `Retry-After` header so clients back off in a coordinated way.
- **Drop before queues grow unbounded.** An unbounded queue under overload just converts a fast failure into a slow one while latency climbs for everyone. A bounded queue with fast rejection keeps healthy requests fast.
- Prefer shedding the newest/excess work; protect the requests already in flight and within budget.

### Bulkheads

Isolate resources per dependency so one slow upstream cannot starve calls to healthy ones.

- Give each dependency its **own** client, connection pool, and concurrency limit. A single shared pool means one slow upstream consumes every connection and every caller queues behind it — a cascading failure.
- Size each pool to that dependency's role; see [services/database.md](../services/database.md) for connection-pool sizing on the DB side.
- Combine bulkheads with per-dependency breakers and limiters: the blast radius of a failing dependency is then confined to its own bulkhead.

### Composition Order

Per outbound call, the controls nest from outside in: **bulkhead** (which pool) → **rate limiter** (am I allowed to send) → **circuit breaker** (is this dependency healthy) → **retry loop** (bounded, jittered) → **per-call timeout** (context deadline) → the actual call. Failures and tripped controls map to the fast-fail/degraded path, never to an unbounded wait. Deployment of new policy values (timeouts, attempt caps, breaker thresholds) follows [operations/deployment.md](deployment.md) like any config change.

## Common Mistakes And Forbidden Patterns

- No client timeout, so a hung upstream blocks a goroutine and pool slot forever.
- Unbounded or zero-jitter retries that synchronize clients and amplify a blip into an outage.
- Retrying non-idempotent writes (un-keyed POSTs, charges, publishes) and causing duplicate effects.
- Retrying `4xx`/validation errors that will fail identically every attempt.
- Retry storms against a struggling dependency with no circuit breaker to stop the pile-on.
- A breaker used as a substitute for, rather than a complement to, bounded retries and timeouts.
- Unbounded in-flight concurrency or unbounded queues that turn overload into a cascading collapse instead of fast rejection.
- A single shared connection pool across all dependencies, so one slow upstream exhausts everything (no bulkhead).
- Hard-coding a specific retry/breaker library without routing it through [framework-selection](../decisions/framework-selection.md).
- High-cardinality retry/breaker metric labels (per-host, per-request-id) instead of per-dependency labels.

## Verification And Proof

- Every external call has a context deadline; a test asserts a hung upstream fails fast within budget, not after a transport default.
- Retry policy proven bounded: a test shows max-attempts and total-deadline caps both fire, and backoff carries jitter (not a fixed schedule).
- Non-idempotent operations are excluded from automatic retry; a test or review note documents the classification.
- A load test under overload shows graceful shedding (fast `429`/`503` with `Retry-After`, stable latency for admitted requests) rather than collapse or unbounded latency growth.
- Breaker open/half-open/closed transitions and retry counts are observable per dependency in metrics (see [operations/observability.md](observability.md)).
- `make race` passes; concurrency controls (semaphores, limiters, breaker state) are race-free under the detector.
