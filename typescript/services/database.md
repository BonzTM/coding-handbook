# Database

PostgreSQL access, transaction, migration, and row-validation rules for TypeScript services.

## Default Approach

Use `pg`, parameterized SQL, Zod row parsing, and `node-pg-migrate`; keep persistence behind core-owned ports.

### Connection Ownership

Composition creates one bounded pool from validated configuration and injects the adapter. Set connection, statement, lock, and idle transaction timeouts appropriate to the service budget. Close the pool during graceful shutdown.

Do not acquire a client until needed. Release it in `finally` on every path. Pass `AbortSignal` through where the driver and owned adapter support it; a caller deadline still bounds waiting and follow-up work.

Construct one bounded pool from typed configuration:

```ts
import { Pool } from "pg";

export function createPool(databaseUrl: string): Pool {
  return new Pool({
    connectionString: databaseUrl,
    application_name: "example-service",
    max: 10,
    connectionTimeoutMillis: 2_000,
    idleTimeoutMillis: 30_000,
    statement_timeout: 5_000,
    query_timeout: 6_000,
    lock_timeout: 1_000,
    idle_in_transaction_session_timeout: 5_000,
  });
}
```

Pool size and deadlines come from bounded configuration in a real service; the values above show the required fields, not universal tuning.

### SQL And Parameters

Write explicit SQL with parameter placeholders for every value. Identifiers cannot be parameterized; choose them from code-owned allowlists rather than user input.

Select named columns, constrain result counts, and define deterministic ordering for pagination. Avoid `SELECT *`, string interpolation, offset pagination on large mutable sets, and queries without a reviewed bound.

### Row Parsing And Mapping

Treat database results as untrusted. Parse rows with Zod schemas that reflect driver runtime representations, then map into domain values. Do not assert `QueryResult<T>` as proof that PostgreSQL returned `T`.

Keep SQL aliases, row schemas, and mappers together. Handle `bigint`, numeric, timestamps, null, arrays, and JSON explicitly. A schema mismatch is an internal failure with safe telemetry, not silent coercion.

```ts
import { z } from "zod";

const widgetRowSchema = z.strictObject({
  id: z.uuid(),
  name: z.string(),
  created_at: z.date(),
  sequence: z.string().regex(/^(0|[1-9]\d*)$/),
});

export async function findWidget(pool: Pool, id: WidgetId): Promise<Widget | null> {
  const result = await pool.query(
    "SELECT id, name, created_at, sequence FROM widgets WHERE id = $1",
    [id],
  );
  const row: unknown = result.rows[0];
  if (row === undefined) return null;
  const parsed = widgetRowSchema.parse(row);
  return {
    id: parseWidgetId(parsed.id),
    name: parsed.name,
    createdAt: new Date(parsed.created_at.getTime()),
    sequence: BigInt(parsed.sequence),
  };
}
```

### Transactions

A transaction owns one checked-out client and one business consistency boundary. Keep it short, avoid remote calls while locks are held, and pass a transaction-scoped repository object inward.

Always `BEGIN`, then `COMMIT` on success or `ROLLBACK` on failure, and release in `finally`. Preserve the original cause if rollback also fails. Define isolation and retry policy from the anomaly being prevented.

```ts
import type { Pool, PoolClient } from "pg";

export async function inTransaction<T>(
  pool: Pool,
  operation: (client: PoolClient) => Promise<T>,
): Promise<T> {
  const client = await pool.connect();
  try {
    await client.query("BEGIN");
    const value = await operation(client);
    await client.query("COMMIT");
    return value;
  } catch (error: unknown) {
    try {
      await client.query("ROLLBACK");
    } catch (rollbackError: unknown) {
      throw new AggregateError([error, rollbackError], "transaction and rollback failed");
    }
    throw error;
  } finally {
    client.release();
  }
}
```

Retries are bounded and apply only to recognized transient PostgreSQL conditions when the entire transaction is safe to repeat. Idempotency and side effects outside PostgreSQL must be addressed first.

### Migrations

Use `node-pg-migrate` as an explicit deployment step, never application startup. Commit ordered migrations and test them on an empty database and the prior supported schema.

Use expand/migrate/contract for rolling compatibility: add nullable or compatible structures, deploy mixed-version code, backfill in bounded resumable batches, then enforce and remove after old code is gone.

Destructive or long-locking changes require measured production-scale evidence, a rollout window, monitoring, and rollback or restore plan. Down migrations do not create recoverability for lost data.

### Query Builders And ORMs

Handwritten SQL is the default. Kysely is the first escalation when compositional query volume justifies it. Any ORM requires an ADR covering schema ownership, migrations, generated code, query visibility, performance, and exit strategy.

Regardless of tool, parameterization, row parsing, transaction boundaries, real PostgreSQL tests, and operational review remain mandatory.

### Integration Tests

Use Testcontainers with a pinned PostgreSQL image. Apply production migrations, isolate test state, and exercise constraints, transactions, concurrency, row mapping, and representative query behavior.

## Common Mistakes And Forbidden Patterns

- SQL values or identifiers interpolated from external input.
- Generic result typing or ORM models treated as runtime row validation.
- `SELECT *`, unbounded reads, or pagination without stable ordering.
- A pool client not released in `finally`.
- Remote network calls inside a database transaction.
- Migrations run automatically by every service instance.
- Destructive migration merged without mixed-version and restore planning.
- Mocked `pg` tests presented as database proof.

## Verification And Proof

- Testcontainers tests run every query and migration against pinned PostgreSQL.
- Empty and prior-schema migration paths succeed; mixed-version reads/writes remain compatible.
- Injection tests prove values never become SQL syntax and dynamic identifiers are allowlisted.
- Row tests cover nulls, large integers, numerics, timestamps, JSON, and malformed representations.
- Transaction tests prove commit, rollback, retry bounds, conflict, and cleanup.
- Representative query plans and lock impact are reviewed for high-risk changes.

Related: [../foundations/data-modeling.md](../foundations/data-modeling.md), [../quality/testing.md](../quality/testing.md), and [../operations/deployment.md](../operations/deployment.md). Driver anchors: [node-postgres pool](https://node-postgres.com/apis/pool) and [transactions](https://node-postgres.com/features/transactions).
