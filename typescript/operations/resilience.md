# Resilience

Timeout, retry, overload, isolation, and degradation rules for networked TypeScript systems.

## Default Approach

Bound every external operation and design failure policy from an end-to-end budget.

### Timeout Budgets

The inbound request, job, or shutdown deadline owns the total budget. Each dependency call receives the caller `AbortSignal` plus a shorter timeout that leaves time for mapping and cleanup.

Configure connect, response, statement, lock, and overall deadlines where supported. Timeout values are validated, observable, documented, and tested. Longer timeouts are not a substitute for capacity or dependency repair.

### Retries

Retry only classified transient failures and only when the operation is idempotent or protected by a durable idempotency contract. Use a fixed maximum attempt count, exponential backoff, jitter, and the remaining caller deadline.

Never retry validation, authentication, authorization, most conflicts, caller abort, or arbitrary unknown failures. Honor safe server retry guidance within the caller budget. Record final outcome and attempt count without multiplying error logs per layer.

### Concurrency And Backpressure

Bound in-flight requests, dependency calls, queued jobs, pages, batches, streams, and worker threads. When capacity is exhausted, reject early, shed optional work, or apply bounded backpressure.

Do not accept unbounded work into memory. Queue capacity and overflow behavior are part of the service contract. Protect scarce pools from one tenant, route, or job class when fairness matters.

### Circuit Breaking And Isolation

Add a circuit breaker only after simple timeout, retry, and concurrency policy is insufficient. Define rolling failure window, open threshold, cool-down, half-open probes, and observability. The fallback must be safer than the failed call.

Use separate pools or concurrency limits for dependencies and workloads whose failures must not cascade. Isolation has a resource cost and requires load evidence.

### Rate Limiting And Load Shedding

Rate limits state identity, scope, algorithm, burst, response, retry guidance, and distributed consistency. Apply authentication and trusted proxy policy before selecting a client identity.

Load shedding uses measured saturation signals such as queue depth, event-loop delay, pool wait, or memory pressure. Reject before expensive parsing or downstream work when safe.

### Degradation

Fallbacks are explicit product behavior: stale cache, partial response, deferred durable work, or a clear unavailable result. Do not return empty or fabricated success after a dependency failure.

Define recovery and consistency after dependency return. A fallback path receives the same security, validation, telemetry, and test scrutiny as the primary path.

### Fault And Capacity Testing

Inject dependency timeout, reset, malformed response, saturation, retry storm, cache failure, broker redelivery, and shutdown during work. Load tests establish sustainable throughput, queue bounds, tail latency, and recovery.

## Common Mistakes And Forbidden Patterns

- No timeout, or nested timeouts exceeding the caller budget.
- Retries around non-idempotent work or at multiple layers.
- Infinite retry, fixed-delay synchronized retry, or retry after caller cancellation.
- Unbounded queues and promise fan-out.
- Circuit breaker or fallback added without explicit semantics and metrics.
- Rate limiting based on spoofable headers.
- Returning empty success during dependency failure.

## Verification And Proof

- Tests cover timeout, abort, transient retry, permanent failure, exhaustion, and cleanup.
- Attempt count and elapsed time remain within configured and caller bounds.
- Load tests demonstrate concurrency/queue caps and intentional overload response.
- Idempotency tests prove retries cannot duplicate durable effects.
- Breaker tests cover closed, open, half-open, recovery, and fallback failure.
- Dashboards expose saturation, retries, shed work, fallback use, and dependency outcomes.

Related: [../foundations/async-and-cancellation.md](../foundations/async-and-cancellation.md), [../services/caching.md](../services/caching.md), and [operability.md](operability.md).
