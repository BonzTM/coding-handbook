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

### Transaction Rules

- Transactions are owned by the layer coordinating multiple writes, not by low-level helper methods by default.
- Use `BeginTx` with the caller's context.
- Keep transaction scope small and avoid network I/O inside transactions.

### Migration Rules

- Every schema change is versioned and reversible when practical.
- Migrations should be idempotent only when the migration tool expects it; otherwise fail loudly.
- Application startup should never silently apply destructive migrations without an explicit operator workflow.
- Treat schema shape and rollout behavior as a contract with running application versions, not just a local dev concern.

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
- Dual-write designs that update the database and publish to a broker in separate success paths without an outbox or equivalent failure model.

## Verification And Proof

- `sqlc generate` or equivalent generation check when the repo uses `sqlc`
- migration apply test against a clean database
- real integration tests for query behavior, transactions, and rollback paths
- proof that callers pass contexts through `QueryContext`, `ExecContext`, or `BeginTx`
- outbox or inbox integration tests when event delivery depends on storage-backed durability
