# Recipe: Add Migration

Use this when a feature changes the database schema and the change must roll out safely across mixed-version application instances.

Default tool is **goose**: SQL-first, embeddable via `embed.FS`, runnable from code or CLI, and pairs cleanly with `database/sql` + `sqlc`. Choosing `golang-migrate` or `atlas` instead is an ADR-level decision; see [../decisions/framework-selection.md](../decisions/framework-selection.md) and [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md).

## Files To Touch

- `internal/db/migrations/NNNN_<desc>.sql` — one file with `-- +goose Up` and `-- +goose Down` sections
- `internal/db/migrate.go` — the `embed.FS` holding `migrations/*.sql` plus the apply seam (migrate-on-startup or CI-apply)
- `internal/db/queries.sql` (or `internal/db/queries/*.sql` once split) and the module-root `sqlc.yaml` — if the schema shape changed and the repo uses `sqlc`; layout per [../services/database.md](../services/database.md)
- repository methods and callers under `internal/db` and any `internal/core` seam that consumes the new shape
- integration tests under `internal/db` that run against a real database

## Steps

1. Allocate the next immutable sequence number `NNNN`, following the governing [../services/database.md](../services/database.md) for schema-change discipline. Numbers are append-only: never renumber, edit, or reorder a migration that has shipped.
2. Write the forward (`Up`) SQL **and** the reverse (`Down`) SQL in the same file. The `Down` must restore the prior schema, not just `DROP` the new objects blindly.
3. Classify the change:
   - **Additive / safe** — `ADD COLUMN` (nullable or with a default), `CREATE TABLE`, `CREATE INDEX CONCURRENTLY`, additive enum values. Old code ignores the addition; ship in one release.
   - **Destructive** — `DROP COLUMN`, `DROP TABLE`, `RENAME`, narrowing a type, adding `NOT NULL` to existing data, or any change that breaks code still running during rollout. These are **never** a single-release migration.
4. For every destructive change, do EXPAND/CONTRACT across releases so mixed-version instances stay correct:
   - **Release N (expand):** add the new column/table/index. Make application code write **both** old and new shapes and read the old shape.
   - **Backfill:** populate the new shape from the old (a migration step or a one-off job), idempotent and re-runnable.
   - **Release N+1 (switch):** flip reads to the new shape; keep writing both until every instance is on N+1.
   - **Release N+2 (contract):** stop writing the old shape, then drop it. This `DROP` is the only destructive migration, and it ships only after no running version references the old shape.
5. If the schema shape changed, update `internal/db/queries.sql` (or the `queries/` directory once split) and run `sqlc generate`; reconcile repository methods and callers with the regenerated types.
6. Wire the apply strategy through the `embed.FS` so dev, CI, and prod run identical migrations. Pick one and document it:
   - **CI-apply (default for prod):** a deploy step runs `goose up` against the target before the new app version receives traffic.
   - **Migrate-on-startup:** acceptable for single-writer/dev; guard against concurrent appliers and never let startup apply a destructive migration silently.
7. Document the rollback story in the PR: which `goose down` is safe, which is not (a contracted `DROP` is forward-only — recovery is restore-from-backup, not down-migration), and the operator workflow for each.

## Invariants To Preserve

- Every migration is deploy-safe: either purely additive, or reversible, or staged as expand/contract.
- No destructive change ships in the same release as code that still depends on the old shape.
- Migration order is immutable once shipped — append a new `NNNN`, never edit history.
- Forward and reverse SQL live together; `Down` actually restores the prior schema.
- SQL stays in `internal/db`; regenerated `sqlc` output matches the migrated schema.
- The change is tested against a real database, not a mock or SQLite stand-in for a Postgres feature.

## Proof

- Apply then roll back against a real database: `goose -dir internal/db/migrations <driver> "$TEST_DSN" up` followed by `goose ... down`, then `up` again — proving forward, reverse, and re-apply all succeed. Use the integration approach from [../quality/testing.md](../quality/testing.md) (real DB, integration build tag), not a mocked driver.
- `go test ./internal/db/... -tags=integration -run Repository` — repository tests pass against the freshly migrated schema.
- `sqlc generate` produces no diff (or the committed diff is reviewed) when the schema shape changed.
- For an expand/contract change, a test or documented review proving release N and N+1 application code both run correctly against the release-N schema.
- `make verify` is green.
