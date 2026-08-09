# Resilience

The bounded failure policy for every dependency, queue, and concurrency boundary.

## Default Approach

Timeouts are mandatory; retries, concurrency limits, queues, and shedding fit inside the caller's total budget. Keep policy explicit in the adapter. The default retry loop is small and hand-written with injected monotonic clock, sleep, and random seams; Tenacity and circuit-breaker libraries require an ADR under [framework selection](../decisions/framework-selection.md).

### Failure-Mode Inventory

Record this table per external dependency before implementation:

| Failure mode | Bound | Classification | Service behavior | Proof |
|---|---|---|---|---|
| connection/DNS/TLS failure | connect and total deadline | transient or configuration | bounded retry or fail closed | stub/network fault test |
| slow or hung response | read/write/pool and total deadline | transient | cancel, release resource, degrade/fail | timeout and leak test |
| `429`/overload | attempt and age budget | transient when instructed | honor bounded `Retry-After`, then shed | response matrix test |
| dependency `5xx` | attempt and total deadline | transient by allowlist | jittered retry, then mapped failure | exhaustion test |
| validation/auth/semantic `4xx` | one attempt | permanent | no retry; stable failure | classification test |
| database checkout/statement timeout | pool and statement deadline | capacity/transient | rollback, map, shed | real-Postgres test |
| broker redelivery/handler timeout | message age/attempt deadline | classified per outcome | settle, retry, or dead-letter | real-broker test |
| local saturation | semaphore/queue capacity | overload | reject before expensive work | load test |

### Timeouts Are The Foundation

Every HTTP, database, broker, subprocess, lock, queue, and external operation has a deadline. The entry boundary establishes one monotonic total budget; child calls receive the remaining budget and retries consume it.

Construct the lifespan-owned client with an explicit `httpx.Timeout`, including connect, read, write, and pool values; HTTPX documents those four timeout classes in its [timeout guide](https://www.python-httpx.org/advanced/timeouts/). Scope the logical call with `asyncio.timeout()` when a stricter total budget is needed. Configure database connect, checkout, statement, and command timeouts independently per [database](../services/database.md). A worker derives a per-message deadline from configured processing and message-age limits.

Timeout handling cancels/awaits owned work and releases responses, sessions, locks, and permits. A transport default is a backstop, not the service contract.

### Retries Are Bounded And Idempotent

Retry only explicitly transient outcomes and operations that are idempotent or protected by a durable idempotency key/inbox. A timeout after sending a write is ambiguous; without dedupe proof, retrying can duplicate the effect.

Use exponential backoff with full jitter: choose a delay uniformly from zero to `min(cap, base * 2**attempt)`. Bound by both a small configured total-attempt cap and the remaining total deadline. Honor a valid `Retry-After` only inside both bounds. Cancellation interrupts injected sleep immediately.

Classify exceptions and responses in one focused function. Retry selected network failures, `429`, and selected `502`/`503`/`504` responses when budget remains. Do not retry validation, authentication, authorization, not-found, programming, serialization, or other stable failures. Emit low-cardinality attempt, retry, exhaustion, and duration telemetry per dependency/operation/outcome.

### Backpressure And Bulkheads

Bound work before creating it. Each dependency receives its own `asyncio.Semaphore`, HTTP/DB connection pool, queue, and concurrency setting sized from dependency capacity and replica count. TaskGroup membership does not bound task creation by itself; acquire admission before creating a task or consume from a bounded queue with a fixed worker set.

Queues declare maximum size, enqueue deadline, overflow policy, and shutdown drain. Producers block only within budget or receive an explicit rejection. Never let a slow dependency consume every task, connection, or memory allocation needed by healthy paths.

### Load Shedding

When admission capacity is exhausted, reject fast before parsing expensive bodies or starting downstream work. HTTP returns the published `429` or `503` problem and a bounded `Retry-After` where useful. Consumers pause intake or use broker prefetch instead of accumulating local tasks. Scheduled work skips or coalesces according to its contract; it does not overlap without a bound.

Prefer preserving admitted work over extending an unbounded queue. A degraded response or stale cache is allowed only when product semantics, data sensitivity, and observability explicitly permit it.

### Circuit Breaker Posture

No circuit breaker by default. Deadlines, per-dependency bulkheads, bounded retries, and shedding are simpler and remain authoritative. Add a breaker only after measured sustained failure shows those controls cannot prevent resource exhaustion or retry pressure.

The ADR defines closed/open/half-open transitions, rolling failure window, minimum sample size, cooldown, probe concurrency, fallback semantics, per-process versus distributed meaning, telemetry, and tests. A breaker never replaces deadlines or idempotency and never presents process-local state as global truth.

### Composition Order

For one call: total budget -> bulkhead/admission -> breaker if approved -> retry loop -> attempt timeout -> transport. The total budget owns all sleeps and attempts. Map the final classified result once at the adapter boundary and close every resource before returning.

## Common Mistakes And Forbidden Patterns

- Any external call, queue wait, lock, subprocess, statement, or drain without a timeout.
- Fresh total timeout per retry, so attempts exceed the caller's budget.
- Retrying non-idempotent or ambiguous writes, permanent `4xx`, parsing failures, or programming errors.
- Fixed/no-jitter or unbounded retries; retry count reset by a nested helper or process restart without an age cap.
- `TaskGroup` or `gather` populated from unbounded external input; semaphore acquired after task creation.
- Unbounded queues, pools, response bodies, subprocess output, or shutdown drains.
- Shared concurrency pool across unrelated dependencies, allowing one failure to starve all work.
- Circuit breaker added by reflex or hidden in a decorator with invisible state transitions.
- Sleep/random/clock hard-coded so retry and deadline behavior cannot be deterministic in tests.

## Verification And Proof

```bash
uv run pytest -k "timeout or retry or backoff or overload or idempot"
uv run pytest -m integration
make verify
```

Prove every inventory row. A hung dependency fails inside budget and leaves no task, connection, response, transaction, or permit. Tests force every retryable and permanent outcome, both caps, full-jitter bounds, cancellation during sleep, and ambiguous-write exclusion. Load tests show bounded memory/concurrency/queue depth, fast shedding, stable admitted latency, and database/downstream headroom at planned replica count.

Related: [concurrency and asyncio](../foundations/concurrency-and-asyncio.md), [eventing and messaging](../services/eventing-and-messaging.md), [observability](observability.md), and [add external client](../recipes/add-external-client.md).
