# Caching

Caching defaults that require measured value, bounded state, explicit invalidation, and stampede control.

## Default Approach

The default is no cache. Fix the query, batching, indexing, serialization, or repeated computation first. Add the cheapest cache only after a profile/load test identifies a repeated expensive read and names an acceptable staleness budget.

### Escalation Ladder

1. Remove or batch the work.
2. Memoize a pure synchronous function with bounded `functools.lru_cache`.
3. Use a bounded in-process TTL/LRU cache plus per-key single-flight control.
4. Escalate to Redis/Valkey through an ADR only for cross-instance sharing, restart survival, coordination, or a working set that cannot fit one process.

Stop at the first layer meeting the measured requirement. An external cache is a stateful dependency with its own timeout, pool, security, observability, and outage behavior.

### functools Is For Pure Synchronous Work

`functools.lru_cache(maxsize=<positive bound>)` is acceptable for a deterministic, side-effect-free synchronous function whose arguments and results are safely retained. It caches by call arguments, may execute duplicate concurrent misses, and exposes `cache_info()`/`cache_clear()`; see the [Python 3.11 reference](https://docs.python.org/3.11/library/functools.html#functools.lru_cache).

Never decorate `async def`: it caches coroutine objects, not awaited results. Never use unbounded `functools.cache` for request-shaped or tenant-shaped data. Do not decorate instance methods because the cache retains `self`; Ruff `B019` documents the resulting [instance-retention risk](https://docs.astral.sh/ruff/rules/cached-instance-method/).

### Bounded TTL And LRU

Use `cachetools.TTLCache` as the default in-process data cache when entries need expiry; use `LRUCache` only when explicit invalidation makes TTL unnecessary and the correctness argument is written. Both receive a positive `maxsize`. `TTLCache` combines TTL expiry with LRU eviction and defaults to monotonic time, as documented by [cachetools](https://cachetools.readthedocs.io/en/latest/#cachetools.TTLCache).

`cachetools` cache objects are not thread-safe. An asyncio-owned cache is accessed only on its event-loop thread and protected for multi-step operations by the loader's `asyncio.Lock`; any thread access requires a separate thread lock or ownership redesign. Store immutable domain snapshots or copies, never live SQLAlchemy/Pydantic objects or mutable values shared with callers.

Bound both entry count and retention time unless an explicit invalidation proof makes one unnecessary. If entry sizes vary materially, supply and test `getsizeof`; `maxsize` otherwise counts entries, not bytes. Expired entries may retain memory until mutation/`expire()`, so low-write caches need deliberate expiry maintenance.

### Cache Keys

A key includes every input that changes the result: cache namespace/version, tenant or authorization identity, stable domain identifier, locale/variant, and relevant query/options. Normalize once before key construction. Never include secrets, raw credentials, full unbounded URLs, request IDs, or mutable object identity.

Cross-tenant caching is forbidden unless tenant/authorization scope is part of the key and tests prove isolation. Hashing or shortening a key does not remove the need to avoid sensitive material in logs and telemetry.

### Async Single-Flight

Protect the miss path with a per-key `asyncio.Lock` owned by the loader:

1. check the cache
2. acquire the key's lock
3. check the cache again
4. load under the caller's deadline
5. cache only a successful valid result
6. release and retire the lock entry

The lock registry is itself bounded/cleaned so arbitrary keys cannot become a memory leak. Do not hold a global lock during I/O. Decide cancellation semantics: a cancelled leader must not strand followers or publish a partial result. Followers retain their own deadline and may stop waiting without cancelling an independently owned load.

Do not cache exceptions by default. Negative caching requires an explicit short TTL, authorization-safe key, and proof that absence is stable enough to cache.

### Invalidation And Consistency

Choose before implementation:

- TTL expiry is the default for process-local caches; staleness is bounded by a documented product/SLO budget.
- Invalidate/update on write is allowed only when every write path is owned and tested; keep a TTL safety bound.
- Versioned keys handle representation/meaning changes and permit bounded old-key expiry.

The source of truth remains authoritative. A cache miss, eviction, or cache outage returns correct behavior through the source, only slower. Do not make TTL the only correctness mechanism when stale data can violate authorization, financial, or safety invariants.

Jitter TTLs when many entries are populated together. Serve-stale-while-revalidate is an explicit product decision with an owned background task, maximum stale window, and failure telemetry; it is not fire-and-forget.

### Redis Escalation

Adopt redis-py/Valkey only when multiple instances must share entries/invalidation, the working set exceeds a process, or restart survival is required. Record an ADR covering topology, persistence expectations, eviction, TLS/auth, cluster key behavior, timeouts, pool limits, serialization/versioning, outage degradation, and rollback.

Use the async client in async services, create it once in lifespan, and close it on shutdown. Redis is not promoted to the system of record by convenience. Distributed locking, rate limiting, sessions, or coordination are separate correctness designs, not ordinary caching.

### Observability

Emit low-cardinality metrics for hits, misses, stale serves, evictions, current size, load latency/errors, and single-flight followers. Label by cache name/result only; never key, tenant, user, or request ID. Log configuration and exceptional transitions, not every hit.

Remove a cache whose hit rate, latency benefit, or source-load reduction does not justify its complexity.

## Common Mistakes And Forbidden Patterns

- Cache added without a measured hot path or written staleness budget.
- Unbounded dict, `functools.cache`, missing TTL/size bound, or unbounded per-key lock registry.
- `lru_cache` on `async def` or instance methods.
- Mutable, ORM, Pydantic, secret-bearing, or authorization-sensitive values retained unsafely.
- Tenant/identity omitted from a key, or cache key used as a metric label.
- One global asyncio lock held across the source load.
- Cache errors failing a read that could use the source of truth.
- Explicit invalidation with an unowned write path and no TTL safety net.
- Redis added first, treated as infallible, or used for distributed correctness without a separate design.

## Verification And Proof

```bash
uv run pytest -k "cache or ttl or single_flight"
make verify
```

Prove miss-then-hit behavior, TTL expiry with an injected monotonic clock, size eviction, tenant isolation, invalidation, duplicate-load collapse, leader failure/cancellation, follower timeout, lock-registry cleanup, cache-outage fallback, immutable returns, and accurate metrics. Load-test beyond key capacity and show RSS plateaus; compare latency/source load with and without the cache and remove it if the result is immaterial.

Related: [time](../foundations/time.md), [resilience](../operations/resilience.md), [observability](../operations/observability.md), and [framework selection](../decisions/framework-selection.md).
