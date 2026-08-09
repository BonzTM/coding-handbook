# Recipe: Add Migration

Use this when PostgreSQL schema changes.

## Files To Touch

- `migrations/<timestamp>-<description>.js`
- affected SQL, Zod row schemas, and mappers under `src/db/`
- Testcontainers migration tests
- deployment job, changelog, and rollback notes
- backfill code and telemetry when expand/migrate/contract is required

## Steps

1. Create a migration with the repository script:

   ```bash
   npm run migrate -- create <description>
   ```

2. Classify it as additive, backfill, constraint, destructive, or long-locking.
3. Use expand/migrate/contract for rename, removal, narrowing, or new required data.
4. Keep old and new application versions compatible throughout rollout and rollback.
5. Make backfills bounded, resumable, idempotent, observable, and independently deployable.
6. Measure lock and rewrite behavior with production-representative data for risky DDL.
7. Update SQL, row parsing, domain mapping, and contract documentation.
8. Define forward repair or restore; do not claim a down migration recovers lost data.
9. Run the same `node-pg-migrate` command used by deployment in tests.

## Invariants To Preserve

- Migrations are explicit deployment jobs, never per-replica startup work.
- Shipped migration history is immutable; corrections are new migrations.
- No destructive contract step lands while running code uses the old shape.
- Schema values and identifiers remain code-owned and reviewed.
- Failure and retry cannot leave an unobservable half-completed backfill.
- PostgreSQL image and supported prior schema are pinned in proof.

## Proof

- Apply all migrations to an empty database.
- Apply the new migration from the prior supported schema.
- Run mixed-version read/write tests for expand and contract changes.
- Prove backfill interruption and resume where applicable.
- `npm run test:integration` and `npm run verify` are green.
