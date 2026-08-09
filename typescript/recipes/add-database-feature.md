# Recipe: Add Database Feature

Use this when adding PostgreSQL queries, transactions, or row mapping.

## Files To Touch

- the consumer-owned port in `src/core/<feature>.ts`
- SQL, row schema, and mapper in `src/db/<feature>-repository.ts`
- migration under `migrations/` when schema changes
- Testcontainers integration tests
- telemetry and query-plan evidence for material paths

## Steps

1. Define the narrow port from the core consumer's needs.
2. Write explicit parameterized SQL with named columns, stable ordering, and result bounds.
3. Treat `result.rows` as unknown and parse each row with a colocated Zod schema.
4. Map driver representations into domain values explicitly.
5. Acquire a pool client only for a transaction and release it in `finally`.
6. Keep remote calls outside transactions and define isolation from the anomaly prevented.
7. Bound connection, statement, lock, query, and caller deadlines.
8. Add a migration recipe when persistence shape changes.
9. Run production migrations in the integration suite against pinned PostgreSQL.

```bash
npm run test:integration -- --runInBand
npm run typecheck
npm run verify
```

## Invariants To Preserve

- `src/core/` imports neither `pg` nor database row types.
- Values use placeholders; dynamic identifiers come only from code-owned allowlists.
- No `SELECT *`, unbounded query, or unstable pagination ordering.
- Row generic annotations never substitute for runtime parsing.
- Transactions commit, roll back, and release deterministically.
- Integration proof uses real PostgreSQL, not mocked `pg`.

## Proof

- Testcontainers tests cover empty, one, malformed, constraint, and boundary representations.
- Transaction tests prove commit, rollback, conflict, and cleanup.
- Injection-shaped values remain data and cannot alter SQL syntax.
- Representative high-risk query plans and lock impact are reviewed.
- `npm run verify` and required integration CI are green.
