# Recipe: Add Database Feature

Use this when a feature changes schema, queries, or transaction behavior.

## Files To Touch

- `internal/db/migrations/*`
- SQL query files or repository methods under `internal/db`
- `internal/core/...` if domain behavior changes
- integration tests for storage behavior

## Steps

1. Add the migration and decide whether it is deploy-safe with mixed-version application instances.
2. Add or update SQL queries and regenerate code if the repo uses `sqlc`.
3. Update repository code and keep storage-specific mapping out of core.
4. Update the core service to call the repository seam, not raw SQL.
5. Add integration tests that exercise the real database path.

## Invariants To Preserve

- SQL stays in storage packages
- transactions are owned by the coordinating layer
- queries remain parameterized
- migrations and application code can coexist safely during rollout

## Proof

- migration apply test on a clean database
- integration test for the changed behavior
- generation check if `sqlc` or similar tooling is used
- rollback or compatibility review for destructive schema changes
