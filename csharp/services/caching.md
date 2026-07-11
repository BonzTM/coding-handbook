# Caching

Caching defaults for repos that should add a cache only when a measured need exists, and then the cheapest layer that satisfies it.

## Default Approach

The default is **no cache.** A cache is not an optimization you reach for by reflex — it is new state with its own consistency, invalidation, eviction, and memory-bound problems, and it hides the original latency instead of fixing it. Add a cache only after a profile or load test shows a specific, repeated, expensive read that dominates a request path, and the source of truth cannot be made fast enough directly.

When a cache is justified, climb the ladder from cheapest to most operationally expensive and stop at the first layer that meets the measured need:

1. **Remove the work.** Memoize a pure computation, batch the N+1 query, add the missing index, or fix the slow query per [database.md](database.md). This is not a cache and has none of a cache's failure modes.
2. **In-process `HybridCache`.** The default cache: TTL-bounded, stampede-protected out of the box, no extra dependency, no cross-process consistency story. `IMemoryCache` remains for the narrow case where even `HybridCache` is overkill (see below).
3. **Distributed backend (Redis/Valkey) behind `HybridCache`.** Only when the working set exceeds one process's memory, must be shared across instances, or must survive restarts. Registering an `IDistributedCache` turns the same `HybridCache` calls two-level — no read-path rewrite — but the backend is a stateful dependency with its own SLO, failure modes, and operational cost: route the pick through [../decisions/framework-selection.md](../decisions/framework-selection.md) and treat its outage like any other external dependency (see [../operations/resilience.md](../operations/resilience.md)).

Reach for the external backend last, not first. Most "we need Redis" needs are met by steps 1–2.

### HybridCache: The Default Cache

`HybridCache` (Microsoft.Extensions.Caching.Hybrid) is the default because it solves, in one primitive, the two problems every hand-rolled cache gets wrong: stampedes (concurrent misses on one key are collapsed into a single loader execution) and the L1/L2 dance when a distributed backend arrives later.

```csharp
builder.Services.AddHybridCache(options =>
{
    options.MaximumPayloadBytes = 1024 * 1024;
    options.DefaultEntryOptions = new HybridCacheEntryOptions
    {
        Expiration = TimeSpan.FromMinutes(5),        // overall (and L2) lifetime
        LocalCacheExpiration = TimeSpan.FromMinutes(1), // in-process lifetime
    };
});

// Only when the ladder reaches step 3 — HybridCache picks it up automatically as L2:
// builder.Services.AddStackExchangeRedisCache(options =>
//     options.Configuration = builder.Configuration.GetConnectionString("cache"));
```

Keep the cache behind a typed accessor in `Orders.Infrastructure`, not a `HybridCache` poked from endpoints. The cache is an implementation detail of one loader, and the loader owns key format, TTL, and invalidation:

```csharp
public sealed class OrderSummaryReader(HybridCache cache, IOrderStore store)
{
    private static readonly HybridCacheEntryOptions SummaryOptions = new()
    {
        Expiration = TimeSpan.FromMinutes(5),
        LocalCacheExpiration = TimeSpan.FromMinutes(1),
    };

    public async ValueTask<OrderSummary?> GetAsync(OrderId id, CancellationToken cancellationToken) =>
        await cache.GetOrCreateAsync(
            $"orders:summary:v1:{id.Value}",
            async token => await store.LoadSummaryAsync(id, token),
            SummaryOptions,
            tags: ["orders:summary"],
            cancellationToken: cancellationToken);
}
```

`GetOrCreateAsync` is the cache-aside pattern with the miss path done right: check, load once, populate, return — and every concurrent caller of the same key shares that one load. Values must round-trip serialization once L2 exists (System.Text.Json by default): cache small, immutable records, never live entity-tracker objects or anything holding a connection.

### IMemoryCache When HybridCache Is Overkill

`IMemoryCache` is acceptable for a tiny, synchronous, process-lifetime memoization with no stampede risk and no prospect of ever needing L2 — a parsed config artifact, a compiled regex table. Two rules are non-negotiable:

- **Bound it or it is a memory leak with a friendly name.** Set `MemoryCacheOptions.SizeLimit` and a `Size` on every entry (the cache throws if `SizeLimit` is set and an entry has none — that is a feature), plus an expiration on every entry.
- **Cached references are shared.** `IMemoryCache` hands back the same object to every caller; mutating it is a data race. Cache immutable records or return defensive copies, per [../foundations/data-modeling.md](../foundations/data-modeling.md).

If you find yourself adding locking or `Lazy<Task<T>>` around `IMemoryCache` to stop a stampede, you have re-derived `HybridCache` — use it instead.

### Key Design

- Keys are structured, ordinal strings with a stable scheme: `resource:projection:version:id` (`orders:summary:v1:42`). Format machine segments with invariant culture — a key must not change with the server's locale.
- The `v1` segment is the shape version: bump it when the cached type changes so old serialized entries in L2 become unreachable instead of undeserializable. There is no "flush the cache on deploy" step.
- **Key cardinality is the memory bound.** Total memory ≈ (distinct keys) × (entry size). Cache per-tenant or per-category, never per-entity unless the entity set is genuinely small and bounded. A per-arbitrary-ID key space makes the cache an unbounded map — all the cost, none of the hit rate.
- Never build a key from raw user input without normalization — `Orders:` and `orders:` as distinct entries is a subtle hit-rate bug, and unnormalized input is a key-space blowup.

### TTL Discipline

Every entry expires. TTL caps staleness and reclaims memory for keys that go cold. Choose the TTL against a **staleness budget tied to your SLOs**, not a round number: "reads may be up to N seconds stale" is a product decision, write it down. `HybridCache` gives you two dials — `Expiration` is the overall budget; `LocalCacheExpiration` is the shorter in-process budget and, once multiple instances exist, the *cross-instance* staleness bound (another instance's L1 copy lives until it lapses). Keys populated together in batches should get jittered TTLs so they do not all expire on the same tick and stampede the source of truth — the same full-jitter principle as retries in [../operations/resilience.md](../operations/resilience.md).

### Invalidation: The Hard Part

> There are only two hard things in computer science: cache invalidation and naming things.

State the invalidation strategy **before** adding the cache. There are exactly two, and you must pick deliberately:

- **TTL (expiry-based).** The entry is correct for at most the TTL, then refreshed on next miss. Simple, no coupling to writes, bounded staleness. Correct when a staleness budget is acceptable. This is the default.
- **Explicit (invalidate-on-write).** Writes to the source of truth remove the cached entry: `RemoveAsync(key)` for a known key, or **tags** — every entry declares tags at write (`tags: ["orders:summary"]`, or a per-tenant tag) and a write path calls `RemoveByTagAsync` to invalidate the whole family without enumerating keys. Stronger consistency, but every write path must know about the cache — coupling that grows quietly and is the source of "stale forever" bugs when one write path forgets. Keep TTL on explicitly-invalidated entries anyway; it is the safety net.

**Never cache what you cannot invalidate or bound.** If you cannot articulate when an entry becomes wrong and how it gets evicted, you do not have a cache, you have a correctness bug on a timer. And treat removal as authoritative for L2 and the local instance only: a peer instance's L1 copy can outlive the removal until its `LocalCacheExpiration` passes — keep that window short, and if the staleness budget cannot tolerate it, the projection was not cacheable in-process.

### Stampede And Thundering-Herd Protection

A cache makes the system *more* fragile at the moment entries expire if the miss path is undefended. `HybridCache` defends it by construction: concurrent `GetOrCreateAsync` calls for one key run the loader once per process and share the result — the counterpart of a hand-rolled request-coalescing layer, with no code. Two residual duties remain yours:

- **Jittered TTLs** for entries populated in batches (above), so expiry does not synchronize.
- **N instances still means up to N concurrent loads** for one hot key. If the source of truth cannot take instance-count concurrent loads, that is the measured signal to add the L2 backend, which absorbs most instance misses.

### Consistency With The Source Of Truth

The cache is never the source of truth. It is a hint that may be wrong by up to your staleness budget. Design every read so a cache miss, a cache outage, or a stale entry is **correct, just slower** — the read falls through to the source of truth. A request path that fails (rather than degrades) when the Redis backend is unavailable has made the cache a hard dependency; that is a design error, not a caching strategy.

### What Must Never Be Cached

Some values are wrong the moment they are stale, and staleness there is a security incident, not a performance trade:

- **Authorization and authentication decisions.** Cache the data an authz check reads if it is measured-hot; never cache the decision. A revoked permission that keeps working for a TTL is a breach window.
- **Per-user secrets and credentials** — tokens, API keys, password material, antiforgery tokens, one-time codes. Secrets are runtime-only values, never cache entries (see [../operations/security.md](../operations/security.md)).
- **Regulated personal data in shared caches** without the same protection and retention rules as the source store — a cache is a data store and [../operations/data-handling.md](../operations/data-handling.md) applies to it in full.

### HTTP-Level Caching: OutputCache

For whole-response caching of anonymous, safe requests, use the OutputCache middleware instead of hand-caching inside endpoints:

```csharp
builder.Services.AddOutputCache(options =>
    options.AddPolicy("catalog", policy =>
        policy.Expire(TimeSpan.FromSeconds(30)).SetVaryByQuery("page").Tag("catalog")));

app.UseOutputCache(); // after UseRouting and UseCors, per the middleware order

group.MapGet("/catalog", GetCatalog).CacheOutput("catalog");
```

Its defaults are the safety story: only successful GET/HEAD responses are cached, and authenticated requests and responses that set cookies are **not** cached — do not fight those rules; anything per-user belongs behind the data cache, not the response cache. Evict by tag (`IOutputCacheStore.EvictByTagAsync`) when the underlying data changes. OutputCache is server-side and additive to the client-facing `Cache-Control` discipline in [web-apps.md](web-apps.md).

### Observability

A cache you cannot measure is a cache you cannot tune or trust. Instrument the typed accessor with a `Meter` and emit, with **low-cardinality** tags only (cache name, never the key):

- **hit / miss / hit-ratio** — the headline number that justifies the cache's existence.
- **eviction count and current entry count** — proves the bound is working (`IMemoryCache` exposes these via `GetCurrentStatistics()` when `TrackStatistics` is on; for `HybridCache`, count hits/misses in the accessor around `GetOrCreateAsync`).
- **load latency and load errors** on the miss path — the latency the cache is hiding.

See [../operations/observability.md](../operations/observability.md) for the metrics seam and [../operations/operability.md](../operations/operability.md) for wiring these into dashboards and runbooks. Cache key, user ID, or tenant ID as a metric tag is a cardinality-explosion outage — see Common Mistakes.

## Common Mistakes And Forbidden Patterns

- **Premature or unmeasured caching.** Adding a cache because reads "feel slow" without a profile or load test pinning the hot path. You add invalidation, staleness, and memory risk to solve a problem you have not measured.
- **Unbounded in-memory state as a cache.** A `static ConcurrentDictionary<string, T>` that only ever grows, or `IMemoryCache` with no `SizeLimit` and no expirations. It works in dev, OOMs in production weeks later.
- **Caching with no invalidation story.** Adding the cache before deciding how entries become wrong and get evicted. If you cannot answer "when is this entry stale and what removes it," do not cache it.
- **Caching auth decisions or secrets.** A permission check or token that survives revocation for a TTL is a security hole with a timer, not a cache.
- **Per-entity cache-key cardinality blowups.** Keying on a high-cardinality or unbounded value (request ID, full query text, per-user ID for a huge user set) so the key space — and thus memory — is effectively unbounded.
- **Mutating a cached object.** `IMemoryCache` returns shared references; an in-process `HybridCache` hit can too. Cache immutable records.
- **High-cardinality cache metric tags.** Putting the cache key, tenant, or user on hit/miss metrics. This blows up the metrics backend; tag by cache *name* only.
- **Cache as a hard dependency.** Read paths that error instead of falling through to the source of truth when the cache misses or the Redis backend is down.
- **Reaching for Redis first.** Adding an external stateful dependency before trying query fixes, in-process `HybridCache`, and its built-in stampede protection.
- **Stale-forever from a forgotten write path.** Choosing explicit invalidation, then adding a new write that does not call `RemoveAsync`/`RemoveByTagAsync`. TTL is the safety net; an explicit-only cache with no TTL has none.
- **Hand-rolling stampede protection** (`SemaphoreSlim`-per-key, `Lazy<Task<T>>` in `IMemoryCache`) instead of using `HybridCache`, which does it correctly by construction.

## Verification And Proof

- **A load test proves the cache helps.** Show p99/throughput on the hot path with and without the cache. A cache that does not move the measured number is dead weight — remove it.
- **A load test proves memory is bounded.** Drive the cache past its intended key space under sustained load and show working-set memory plateaus (no unbounded growth, no OOM).
- **Hit/miss metrics exist and are exercised.** A test asserts a miss-then-hit sequence flips the counters, and the dashboard shows hit-ratio in production.
- **Invalidation is tested.** Write to the source of truth, then assert the read returns the new value after `RemoveAsync`/`RemoveByTagAsync` (explicit) or returns stale up to the TTL and fresh after (TTL) — drive time with `FakeTimeProvider` where the seam allows, short TTLs where it does not.
- **Stampede protection is tested.** Fire N concurrent `GetAsync` calls for one key against a counting fake store and assert the store was hit exactly once per process.
- **Miss/outage degrades, not fails.** A test forces a cache miss and a cache-backend error and asserts the read still returns the correct value from the source of truth.
- run `pwsh ./verify.ps1` — restore (locked), format-check, build (warnings-as-errors), test, audit

## Related

- [../operations/resilience.md](../operations/resilience.md) — jitter, thundering-herd, treating the cache backend as a fallible dependency.
- [../operations/operability.md](../operations/operability.md) — dashboards, SLOs, and the staleness budget.
- [../foundations/data-modeling.md](../foundations/data-modeling.md) — immutable records, bounded collections, returning copies not shared state.
- [../decisions/framework-selection.md](../decisions/framework-selection.md) — picking the distributed backend when the ladder reaches it.
