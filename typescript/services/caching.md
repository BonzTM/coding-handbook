# Caching

Cache ownership, key, expiry, invalidation, and failure rules for TypeScript services.

## Default Approach

Add a cache only after measuring a latency, capacity, or dependency-protection problem and defining acceptable staleness.

### Cache Contract

For each cached value, document owner, source of truth, key schema, tenant scope, value schema, TTL, stale tolerance, invalidation trigger, size bound, failure behavior, and privacy classification.

The caller owns cache policy; a generic cache adapter owns storage mechanics. Core behavior must remain correct on miss, stale entry, eviction, and cache outage.

### Keys And Values

Version key namespaces and include every input that changes meaning: tenant, locale, authorization scope, and schema version where applicable. Hash sensitive or oversized components only with an understood collision and observability strategy.

Never place secrets, tokens, raw PII, or authorization decisions in a cache unless the security and data-handling design explicitly approves storage, encryption, retention, and invalidation.

Parse cached values with Zod. Treat corruption or old schema as a miss, record bounded telemetry, and repair from the source of truth. Do not deserialize arbitrary executable objects.

### Expiry And Invalidation

TTL limits staleness; it does not make invalidation correct. Prefer explicit invalidation or write-through behavior for correctness-sensitive data. Define ordering when source update succeeds but invalidation fails.

Add randomized TTL jitter when synchronized expiry could cause a stampede. Negative caching is short-lived and only for stable not-found outcomes; never cache transient dependency or authorization failures as absence.

### Stampede Protection

Coalesce concurrent fills per key within a bounded process and cap total fill concurrency. Distributed locking is an escalation that requires lease, fencing, timeout, crash, and split-brain analysis.

Stale-while-revalidate is allowed only when stale content is explicitly acceptable. One owner refreshes; callers receive bounded stale data or a documented failure.

### Failure And Availability

Cache calls have short timeouts, bounded retries, and circuit/load policy. Decide whether outage fails open to the source, serves stale data, or fails closed. Prevent fallback from overwhelming the source dependency.

Local in-memory caches have explicit item/byte bounds and are per-process. They cannot provide cross-instance invalidation or durable semantics. TanStack Query browser caching follows [frontend-data-and-state.md](frontend-data-and-state.md), not backend cache policy.

### Observability

Measure hit, miss, stale, error, eviction, fill latency, and fill concurrency by bounded cache name/result. Never label metrics by raw key. Log a hashed or safe resource reference only when diagnosis requires it.

## Common Mistakes And Forbidden Patterns

- Caching introduced without a measured problem or staleness contract.
- Keys missing tenant, authorization scope, locale, or schema version.
- TTL presented as complete invalidation.
- Unbounded in-memory maps or cached value sizes.
- Stampede on expiry or cache failure overwhelming the source.
- Cached values trusted without schema validation.
- Secrets, raw PII, or bearer tokens in keys, values, logs, or metrics.
- Cache required for correctness while its outage behavior is undefined.

## Verification And Proof

- Tests cover hit, miss, expiry, stale, invalidation, corruption, and schema-version mismatch.
- Tenant and authorization-scope tests prove keys cannot collide across principals.
- Concurrent miss tests prove one bounded fill and no source stampede.
- Cache-outage tests prove fail-open, stale, or fail-closed behavior and source protection.
- Memory and value-size bounds are exercised under representative load.
- Metrics have bounded attributes and contain no raw keys or sensitive values.

Related: [frontend-data-and-state.md](frontend-data-and-state.md), [../operations/resilience.md](../operations/resilience.md), and [../operations/data-handling.md](../operations/data-handling.md).
