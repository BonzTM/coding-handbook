# Database

Persistence defaults for repos that want visible SQL, explicit transactions, and production-safe migrations.

## Default Approach

- Start with `database/sql`.
- Use hand-written SQL first and adopt `sqlc` when query count or mapping noise justifies generation.
- Keep SQL and storage mapping in `internal/db`, not in handlers or core packages.

### Suggested Layout

```text
internal/db/
  migrations/
  queries/
  repository.go
  tx.go
  sqlc.yaml          # if the repo uses sqlc
```

### Connection Pool Sizing

The `*sql.DB` is a pool, not a connection. Its defaults are wrong for production and the wrong values silently cause some of the most common incidents. Set all four limits explicitly, drive them from `internal/config` keys, and never ship the defaults.

- `SetMaxOpenConns` — cap total open connections. Size to **DB server capacity divided by the number of application instances**, leaving headroom for migrations, admin tools, and replicas. Unbounded (the default of `0`) lets one instance fan out under load and exhaust the database's connection limit, taking down every client.
- `SetMaxIdleConns` — keep a sensible floor (commonly equal to or a fraction of `MaxOpenConns`) so the pool does not open and close connections per request and thrash. Must be `<= MaxOpenConns`; a floor that is too low under steady load reintroduces churn.
- `SetConnMaxLifetime` — bound how long a connection lives. Required when traffic passes through a load balancer, proxy, or failover-capable endpoint: without it, connections pin to a stale backend after failover and never rebalance. Set it below the shortest of the backend's idle/connection timeout, the proxy's timeout, and any DNS TTL you must honor.
- `SetConnMaxIdleTime` — reap idle connections so the pool shrinks when traffic drops instead of holding stale handles.

Failure modes to recognize:

- **Pool exhaustion** — too-low `MaxOpenConns` (or a slow query holding connections) makes callers block in `db.Conn`/`QueryContext` until their context deadline; symptom is request queueing and timeouts, not DB errors.
- **Stale connections behind a proxy/failover** — missing or too-long `ConnMaxLifetime` keeps routing to a drained or failed-over backend.
- **Churn** — `ConnMaxLifetime` or `ConnMaxIdleTime` set too aggressively, or `MaxIdleConns` too low, forces constant reconnects and TLS handshakes.

Pool waits must be observable. Export `db.Stats()` (`WaitCount`, `WaitDuration`, `InUse`, `Idle`) as metrics so saturation is visible before it becomes a timeout. See [../operations/observability.md](../operations/observability.md) for emitting them with low-cardinality labels.

External HTTP clients are the same problem in a different pool: never use `http.DefaultClient`, and tune `http.Transport` `MaxIdleConnsPerHost` and `IdleConnTimeout` to the downstream's capacity and keep-alive window. See [../recipes/add-external-client.md](../recipes/add-external-client.md) and [../operations/resilience.md](../operations/resilience.md).

### Transaction Rules

- Transactions are owned by the layer coordinating multiple writes, not by low-level helper methods by default.
- Use `BeginTx` with the caller's context.
- Keep transaction scope small and avoid network I/O inside transactions.

### Migrations

Default tool is **goose**: SQL-first, embedded via `embed.FS`, run from code or CLI, and pairs with `database/sql` + `sqlc`. `golang-migrate` and `atlas` are alternatives that need an ADR ([../decisions/framework-selection.md](../decisions/framework-selection.md)). Full procedure: [../recipes/add-migration.md](../recipes/add-migration.md).

- Every schema change is versioned with an append-only sequence number and ships forward (`Up`) **and** reverse (`Down`) SQL in the same file.
- **Treat schema shape and rollout behavior as a contract with running application versions.** During a rollout, old and new application code run against the same schema. A migration is deploy-safe only if every concurrently-running version still works.
- Additive changes (`ADD COLUMN` nullable/defaulted, `CREATE TABLE`, `CREATE INDEX CONCURRENTLY`) are safe to ship in one release because old code ignores them.
- **Destructive changes (drop, rename, narrow, add `NOT NULL`) use expand/contract across releases:** add the new shape and dual-write -> backfill -> switch reads -> drop the old shape only in a later release once no running version references it. Never drop or rename in the same release as code that still reads the old shape.
- Embed migrations via `embed.FS` so dev, CI, and prod apply identical files. Pick one apply strategy and document it: CI-apply before traffic (default for prod) or migrate-on-startup (dev/single-writer only, guarded against concurrent appliers). Startup must never silently apply a destructive migration.
- **Migration order is immutable once shipped.** Append a new file; never edit, renumber, or reorder history. A contracted `DROP` is forward-only — recovery is restore-from-backup, not a down-migration.
- Idempotency only when the tool expects it; otherwise fail loudly.

### Outbox And Inbox

- If business logic must commit DB state and emit an event together, default to a transactional outbox.
- Write domain state and an outbox row in one transaction, then relay that outbox record asynchronously.
- Use an inbox or durable dedupe table on the consumer side when duplicate processing would be harmful.
- Keep relay jobs and dedupe storage explicit; do not hide them behind a magical repository abstraction that erases failure semantics.

## Common Mistakes And Forbidden Patterns

- ORM-first design that hides SQL and query cost.
- Manual SQL string concatenation.
- Database handles hidden behind globals.
- Repository interfaces so broad that they become a second domain model.
- Tests that mock SQL behavior but never prove migrations or real queries.
- Editing, renumbering, or reordering a migration that has already shipped instead of appending a new one.
- A destructive migration (drop, rename, narrow, add `NOT NULL`) shipped in the same release as code that still depends on the old shape, instead of staged expand/contract.
- Migrations that only have an `Up` and no reverse, or a `Down` that drops new objects without restoring the prior schema.
- Startup silently applying destructive migrations without an explicit operator workflow.
- Dual-write designs that update the database and publish to a broker in separate success paths without an outbox or equivalent failure model.
- Shipping the default pool: unbounded `MaxOpenConns`, or any of the four limits left unset.
- No `ConnMaxLifetime` on a connection that reaches a load balancer, proxy, or failover-capable endpoint.
- Pool limits aggressive enough to churn connections instead of reusing them.
- Pool saturation (`WaitCount`/`WaitDuration`) that is invisible because `db.Stats()` is never exported.
- Using `http.DefaultClient` or an untuned `http.Transport` for outbound calls.

## Verification And Proof

- `sqlc generate` or equivalent generation check when the repo uses `sqlc`
- migration apply-then-rollback-then-reapply test against a real database (`goose up` / `goose down` / `goose up`), proving forward, reverse, and re-apply all succeed
- repository tests pass against the freshly migrated schema, and regenerated `sqlc` output matches it
- for expand/contract changes, proof that release N and N+1 application code both run correctly against the release-N schema
- real integration tests for query behavior, transactions, and rollback paths
- proof that callers pass contexts through `QueryContext`, `ExecContext`, or `BeginTx`
- outbox or inbox integration tests when event delivery depends on storage-backed durability
- proof that all four pool limits are set from config and that `MaxIdleConns <= MaxOpenConns`
- a load test that drives expected peak concurrency and shows the pool config holds without runaway `WaitDuration` or connection churn
- `db.Stats()` exported as metrics, with pool-wait saturation visible on a dashboard
