# Caching

Caching defaults for repos that should add a cache only when a measured need exists, and then the cheapest layer that satisfies it.

## Default Approach

The default is **no cache.** A cache is not an optimization you reach for by reflex — it is new state with its own consistency, invalidation, eviction, and memory-bound problems, and it hides the original latency instead of fixing it. Add a cache only after a profile or load test shows a specific, repeated, expensive read that dominates a request path, and the source of truth cannot be made fast enough directly.

When a cache is justified, climb the ladder from cheapest to most operationally expensive and stop at the first layer that meets the measured need:

1. **Remove the work.** Memoize a pure computation, batch N+1 queries, add the missing index, or fix the slow query. This is not a cache and has none of a cache's failure modes.
2. **Bounded in-process cache.** A size- and TTL-bounded map local to one process. No network hop, no extra dependency, no cross-process consistency story.
3. **`singleflight` to collapse duplicate concurrent loads.** Often the actual problem is a stampede of identical concurrent loads, not repeated loads over time — `golang.org/x/sync/singleflight` fixes that with no stored state at all.
4. **External cache (Redis/memcached).** Only when the working set exceeds one process's memory, must be shared across instances, or must survive restarts. This is a stateful dependency with its own SLO, failure modes, and operational cost — route the pick through [../decisions/framework-selection.md](../decisions/framework-selection.md) and treat its outage like any other external dependency (see [../operations/resilience.md](../operations/resilience.md)).

Reach for the external cache last, not first. Most "we need Redis" needs are met by steps 1–3.

### Bounded In-Process Cache

Every in-process cache MUST be bounded on **both** entry count and time, or it is a memory leak with a friendly name.

- **Size bound.** Cap the number of entries and evict (LRU/LFU) when full. An unbounded `map[K]V` used as a cache grows until the process OOMs — it is the single most common caching incident. Pick a library that enforces the bound (routed via [../decisions/framework-selection.md](../decisions/framework-selection.md)); do not hand-roll eviction.
- **Time bound (TTL).** Every entry expires. TTL caps staleness and reclaims memory for keys that go cold. Choose the TTL against a **staleness budget tied to your SLOs**, not a round number: "reads may be up to N seconds stale" is a product decision, write it down.
- **Key cardinality is the memory bound.** Total memory ≈ (distinct keys) × (entry size). Cache per-tenant or per-category, never per-entity unless the entity set is genuinely small and bounded. A per-user-request or per-arbitrary-ID key turns the cache into an unbounded map again — the size cap then just thrashes (evict-on-every-insert) and you pay all the cost for none of the hit rate.
- **Keep it behind a typed accessor**, not a package-global map poked from handlers. The cache is an implementation detail of one loader (`UserLoader.Get(ctx, id)`), and the loader owns invalidation. See bounded-map discipline in [../foundations/data-modeling.md](../foundations/data-modeling.md).

### Collapsing Duplicate Loads With singleflight

When many goroutines miss on the same key at the same instant (cold start, just-expired hot key, cache flush), they all stampede the source of truth simultaneously — a self-inflicted thundering herd. `singleflight.Group.Do` runs the load **once** per key and shares the result with every concurrent caller:

```go
type Loader struct {
	sf    singleflight.Group
	cache *lru.Cache[string, User] // size- and TTL-bounded
	db    UserStore
}

func (l *Loader) Get(ctx context.Context, id string) (User, error) {
	if u, ok := l.cache.Get(id); ok {
		return u, nil
	}
	v, err, _ := l.sf.Do(id, func() (any, error) {
		u, err := l.db.LoadUser(ctx, id) // context-first; see context-and-concurrency.md
		if err != nil {
			return User{}, err // %w mapped at the boundary, logged once
		}
		l.cache.Add(id, u)
		return u, nil
	})
	if err != nil {
		return User{}, err
	}
	return v.(User), nil
}
```

`singleflight` shares one caller's `ctx` and error with all waiters: a cancellation or failure in the leader propagates to every follower. Use `DoChan` with a per-caller `select` on `ctx.Done()` when followers must not be hostage to the leader's deadline, and forget the in-flight key (`Forget`) if you must not pin a failed result.

### Invalidation: The Hard Part

> There are only two hard things in computer science: cache invalidation and naming things.

State the invalidation strategy **before** adding the cache. There are exactly two, and you must pick deliberately:

- **TTL (expiry-based).** The entry is correct for at most the TTL, then refreshed on next miss. Simple, no coupling to writes, bounded staleness. Correct when a staleness budget is acceptable. This is the default.
- **Explicit (write-through / invalidate-on-write).** Writes to the source of truth delete or update the cached entry. Stronger consistency, but every write path must know about every cache — coupling that grows quietly and is the source of "stale forever" bugs when one write path forgets.

**Never cache what you cannot invalidate or bound.** If you cannot articulate when an entry becomes wrong and how it gets evicted, you do not have a cache, you have a correctness bug on a timer. In-process caches across N instances cannot be explicitly invalidated cheaply (you would need to fan out to every instance) — so an in-process cache almost always means TTL-based invalidation and an accepted staleness budget. If you need cross-instance explicit invalidation, that is a reason to move to a shared external cache, not to fake it.

### Stampede And Thundering-Herd Protection

A cache makes the system *more* fragile at the moment entries expire if you do not defend the miss path:

- **`singleflight` on the load** so a hot expired key triggers one reload, not thousands (above).
- **Jittered TTL.** Adding `±jitter` to each entry's TTL desynchronizes expiry so a batch of keys populated together does not all expire on the same tick and stampede simultaneously. This is the same full-jitter principle used for retries in [../operations/resilience.md](../operations/resilience.md).
- **Optional: serve-stale-while-revalidate.** For read paths that tolerate brief staleness, return the expired value and refresh asynchronously so a miss never blocks on the source of truth.

### Consistency With The Source Of Truth

The cache is never the source of truth. It is a hint that may be wrong by up to your staleness budget. Design every read so a cache miss, a cache outage, or a stale entry is **correct, just slower** — the read falls through to the source of truth. A request path that fails (rather than degrades) when the cache is unavailable has made the cache a hard dependency; that is a design error, not a caching strategy.

### Observability

A cache you cannot measure is a cache you cannot tune or trust. Emit, with **low-cardinality** labels only (cache name, never the key):

- **hit / miss / hit-ratio** — the headline number that justifies the cache's existence.
- **eviction count and current size / entry count** — proves the bound is working and shows when capacity is too small (eviction churn) or too large (wasted memory).
- **load latency and load errors** on the miss path — the latency the cache is hiding.
- **`singleflight` shared-call count** — how many duplicate loads were collapsed.

See [../operations/observability.md](../operations/observability.md) for the metrics seam and [../operations/operability.md](../operations/operability.md) for wiring these into dashboards and runbooks. Cache key, user ID, or tenant ID as a metric label is a cardinality-explosion outage — see Common Mistakes.

## Common Mistakes And Forbidden Patterns

- **Premature or unmeasured caching.** Adding a cache because reads "feel slow" without a profile or load test pinning the hot path. You add invalidation, staleness, and memory risk to solve a problem you have not measured.
- **Unbounded in-memory maps as caches.** A package-level `map[string]T` that only ever grows. It works in dev, OOMs in production weeks later. Both a size bound and a TTL are mandatory.
- **Caching with no invalidation story.** Adding the cache before deciding how entries become wrong and get evicted. If you cannot answer "when is this entry stale and what removes it," do not cache it.
- **Per-entity cache-key cardinality blowups.** Keying on a high-cardinality or unbounded value (request ID, full query, per-user ID for a huge user set) so the key space — and thus memory, or eviction thrash — is effectively unbounded.
- **High-cardinality cache metric labels.** Putting the cache key, tenant, or user on hit/miss metrics. This blows up the metrics backend; label by cache *name* only.
- **Cache as a hard dependency.** Read paths that error instead of falling through to the source of truth when the cache misses or is down.
- **Reaching for Redis first.** Adding an external stateful dependency before trying memoization, query fixes, a bounded in-process cache, or `singleflight`.
- **Stale-forever from a forgotten write path.** Choosing explicit invalidation, then adding a new write that does not invalidate. TTL is the safety net; an explicit-only cache with no TTL has none.

## Verification And Proof

- **A load test proves the cache helps.** Show p99/throughput on the hot path with and without the cache. A cache that does not move the measured number is dead weight — remove it.
- **A load test proves memory is bounded.** Drive the cache past capacity with diverse keys under sustained load and show RSS plateaus (no unbounded growth, no OOM). Confirm eviction metrics increment.
- **Hit/miss/eviction metrics exist and are exercised.** A test asserts a miss-then-hit sequence flips the counters, and the dashboard shows hit-ratio in production.
- **Invalidation is tested.** A table-driven test: write to the source of truth, then assert the cache returns the new value (explicit) or returns stale up to the TTL and fresh after (TTL). For `singleflight`, a test fires N concurrent `Get`s for one key against a counting fake store and asserts the store was hit exactly once.
- **Miss/outage degrades, not fails.** A test forces a cache miss and a cache error and asserts the read still returns the correct value from the source of truth.

## Related

- [../operations/resilience.md](../operations/resilience.md) — jitter, thundering-herd, treating the external cache as a fallible dependency.
- [../operations/operability.md](../operations/operability.md) — dashboards, SLOs, and the staleness budget.
- [../foundations/data-modeling.md](../foundations/data-modeling.md) — bounded maps and returning copies, not internal state.
- [../decisions/framework-selection.md](../decisions/framework-selection.md) — picking the LRU/TTL library and the external cache.
