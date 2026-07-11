# Database

Persistence defaults for repos that want visible query cost, explicit transactions, and production-safe migrations.

## Default Approach

- Start with EF Core on the Npgsql/PostgreSQL provider. One `DbContext` per service, owned by `Orders.Infrastructure/Data/`.
- Repository ports (interfaces) live in `Orders.Core`; EF Core implementations live in Infrastructure. Core never references EF Core.
- Reads are no-tracking by default; tracking is an explicit opt-in for write flows.
- Dapper is the escape hatch for measured hot paths, via ADR — not a parallel data layer.

### Suggested Layout

```text
src/Orders.Core/
  Orders/                      # domain types + repository ports (IOrderRepository)
src/Orders.Infrastructure/Data/
  OrdersDbContext.cs
  Configurations/              # IEntityTypeConfiguration<T> per aggregate
  Migrations/                  # EF Core migrations (append-only)
  Repositories/
tests/Orders.IntegrationTests/
  PostgresFixture.cs           # Testcontainers, migrated schema
```

### DbContext Discipline

A `DbContext` is a unit of work, not a connection and not a cache to keep warm.

- **Scoped lifetime only.** Register with `AddDbContextPool<OrdersDbContext>` and let DI hand one instance per request/message scope. Never a singleton, never captured in a field that outlives the scope, never shared across threads — `DbContext` is not thread-safe.
- **No long-lived contexts.** Background workers create a scope per unit of work (`IServiceScopeFactory.CreateAsyncScope()`), do the work, dispose. A context that lives for hours accumulates tracked entities and stale state.
- **No-tracking reads by default.** Set it once on the context options and opt back in for writes:

```csharp
builder.Services.AddDbContextPool<OrdersDbContext>(options =>
    options
        .UseNpgsql(connectionString, npgsql => npgsql.EnableRetryOnFailure())
        .UseQueryTrackingBehavior(QueryTrackingBehavior.NoTracking));
```

  Write paths load with `.AsTracking()` (or attach), mutate, `SaveChangesAsync`. A read path that tracks thousands of entities is a memory leak with extra steps.
- Repositories return domain types or projections, never `IQueryable` across the Core boundary — a leaked `IQueryable` makes query cost invisible and composition untestable.
- Pass the caller's `CancellationToken` into every async EF call ([../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md)).

### Connection Pool Sizing

Two pools exist and they are different things. `AddDbContextPool` recycles `DbContext` *instances* (an allocation optimization). Npgsql's ADO.NET pool owns the *physical connections*, sized by the connection string. The second one causes incidents; set all four limits explicitly from config and never ship defaults unexamined.

```text
Host=db;Database=orders;Username=orders_app;
  Maximum Pool Size=20;Minimum Pool Size=5;
  Connection Lifetime=300;Connection Idle Lifetime=60;Timeout=15
```

- `Maximum Pool Size` — cap total connections per instance. Size to **DB server capacity divided by the number of application instances**, leaving headroom for migrations, admin tools, and replicas. The default (100) looks bounded but is not safe: five instances × 100 exceeds PostgreSQL's default `max_connections` of 100, and one scaled-out deployment exhausts the server for every client.
- `Minimum Pool Size` — keep a sensible floor so the pool does not open and close connections per request and thrash under steady load. The default (0) reintroduces cold-start churn after every quiet period.
- `Connection Lifetime` — bound how long a connection lives. Required when traffic passes through a load balancer, PgBouncer, or a failover-capable endpoint: without it (default 0 = unbounded), connections pin to a stale backend after failover and never rebalance. Set it below the shortest of the backend's connection timeout, the proxy's timeout, and any DNS TTL you must honor.
- `Connection Idle Lifetime` — reap idle connections so the pool shrinks when traffic drops instead of holding stale handles.

Failure modes to recognize:

- **Pool exhaustion** — a too-low `Maximum Pool Size` (or a slow query holding connections) makes callers queue for up to `Timeout` seconds (default 15), then throw "the connection pool has been exhausted"; the symptom is request queueing and timeouts before it is DB errors.
- **Stale connections behind a proxy/failover** — missing or too-long `Connection Lifetime` keeps routing to a drained or failed-over backend.
- **Churn** — lifetimes set too aggressively, or no minimum pool size, forces constant reconnects and TLS handshakes.

Pool waits must be observable. Npgsql publishes connection-pool counters (busy/idle connections, pending requests) through `System.Diagnostics.Metrics` under the `Npgsql` meter — add it to the OpenTelemetry `MeterProvider` (`AddMeter("Npgsql")`) so saturation is visible before it becomes a timeout. See [../operations/observability.md](../operations/observability.md).

Outbound HTTP clients are the same problem in a different pool: `SocketsHttpHandler` connection limits and idle timeouts need the same deliberate sizing. See [../recipes/add-external-client.md](../recipes/add-external-client.md) and [../operations/resilience.md](../operations/resilience.md).

### Query Hygiene

- **No N+1.** Lazy-loading proxies are forbidden. Load related data explicitly with `Include`, or better, project: a `Select` into exactly the fields the caller needs is one round trip, no tracking, no over-fetch. Prefer projection for read endpoints; reserve `Include` for write flows that mutate the aggregate. Use `AsSplitQuery()` when a multi-`Include` query would explode into a cartesian product.
- **Compiled queries for hot paths.** EF compiles LINQ per query shape; on measured hot paths, hoist that cost once:

```csharp
private static readonly Func<OrdersDbContext, OrderId, CancellationToken, Task<Order?>> GetById =
    EF.CompileAsyncQuery((OrdersDbContext db, OrderId id, CancellationToken ct) =>
        db.Orders.FirstOrDefault(o => o.Id == id));
```

- Raw SQL only through `FromSql`/`SqlQuery` interpolation (parameterized by the compiler) — never string concatenation.
- Every list query is paginated with a bounded page size. Unbounded `ToListAsync()` over a growing table is a latency time bomb.

### Transaction Rules

- `SaveChangesAsync` is already atomic for one unit of work. Explicit transactions exist for coordinating multiple saves or raw SQL — owned by the layer coordinating the writes, not hidden in low-level helpers.
- `EnableRetryOnFailure` turns on the retrying execution strategy for transient faults. With it enabled, user-initiated transactions must run *inside* the strategy so the whole unit retries, not a torn half:

```csharp
var strategy = db.Database.CreateExecutionStrategy();
await strategy.ExecuteAsync(async () =>
{
    await using var tx = await db.Database.BeginTransactionAsync(ct);
    await db.SaveChangesAsync(ct);
    await outbox.EnqueueAsync(evt, ct);
    await tx.CommitAsync(ct);
});
```

- Keep transaction scope small and avoid network I/O inside transactions — a held connection doing HTTP is pool exhaustion in progress.

### Migrations

Default tool is **EF Core migrations**, living in `Orders.Infrastructure/Data/Migrations/`. Full procedure: [../recipes/add-migration.md](../recipes/add-migration.md). Standalone migration tools (Flyway, Liquibase, DbUp) need an ADR ([../decisions/framework-selection.md](../decisions/framework-selection.md)).

```bash
dotnet ef migrations add AddOrderExternalReference \
  --project src/Orders.Infrastructure --startup-project src/Orders.Api
```

- Every migration ships a real `Up` **and** a real `Down` in the same file. A scaffolded `Down` that EF could not infer is reviewed and completed by hand, not shipped empty.
- **Apply is an explicit operator step, never a side effect of normal startup.** Pick one and document it:
  - a `--migrate` flag: the host applies migrations and exits, run as a Kubernetes Job or init step before traffic (default for prod);
  - an EF migrations bundle (`dotnet ef migrations bundle`) executed by the deploy pipeline;
  - an idempotent SQL script (`dotnet ef migrations script --idempotent`) for DBA-gated environments.

```csharp
if (args.Contains("--migrate", StringComparer.Ordinal))
{
    await using var scope = app.Services.CreateAsyncScope();
    var db = scope.ServiceProvider.GetRequiredService<OrdersDbContext>();
    await db.Database.MigrateAsync();
    return 0;
}
```

  `Database.Migrate()` in the normal startup path is forbidden: N replicas race to apply, startup silently runs destructive DDL, and rollback becomes impossible to reason about.
- **Treat schema shape and rollout behavior as a contract with running application versions.** During a rollout, old and new code run against the same schema. A migration is deploy-safe only if every concurrently-running version still works.
- Additive changes (`ADD COLUMN` nullable/defaulted, `CREATE TABLE`, `CREATE INDEX CONCURRENTLY`) are safe in one release because old code ignores them.
- **Destructive changes (drop, rename, narrow, add `NOT NULL`) use expand/contract across releases:** add the new shape and dual-write → backfill → switch reads → drop the old shape only in a later release once no running version references it. Never drop or rename in the same release as code that still reads the old shape.
- **Migration order is immutable once shipped.** Append a new migration; never edit, renumber, or reorder history — `__EFMigrationsHistory` is append-only. A contracted `DROP` is forward-only; recovery is restore-from-backup, not a down-migration.
- Keep model and migrations in lockstep: `dotnet ef migrations has-pending-model-changes` fails when the model drifted without a migration — run it in the verify gate.

### Typed IDs

Raw `Guid`/`long` keys let an `OrderId` silently accept a `CustomerId`. Wrap keys in domain-typed IDs in Core and convert once, by convention:

```csharp
public readonly record struct OrderId(Guid Value)
{
    public static OrderId New() => new(Guid.CreateVersion7());
    public static OrderId Parse(string s) => new(Guid.Parse(s));
}
```

```csharp
// OrdersDbContext
protected override void ConfigureConventions(ModelConfigurationBuilder builder)
    => builder.Properties<OrderId>().HaveConversion<OrderIdConverter>();

internal sealed class OrderIdConverter() : ValueConverter<OrderId, Guid>(
    id => id.Value, value => new OrderId(value));
```

Version-7 GUIDs keep PostgreSQL b-tree inserts append-friendly; random v4 keys fragment the index. See [../foundations/data-modeling.md](../foundations/data-modeling.md).

### Concurrency Tokens

Lost updates are silent without a concurrency token. Default: map PostgreSQL's `xmin` system column — zero schema cost, updated by every write:

```csharp
// entity: public uint Version { get; set; }
builder.Property(o => o.Version).IsRowVersion();   // maps to xmin on Npgsql
```

(On SQL Server the same pattern maps a `rowversion` column.) A conflicting write throws `DbUpdateConcurrencyException`; map it to the typed conflict error Core defines, and let the transport render it (HTTP 409 / gRPC `Aborted`). Retrying blindly is not handling — reload, re-decide, or surface to the caller.

### Unique Violations To Typed Conflicts

Uniqueness is enforced by the database, so the database is where "already exists" surfaces. Do not pre-check with a `SELECT` (racy); insert and translate the violation once, in the repository:

```csharp
try
{
    await db.SaveChangesAsync(ct);
}
catch (DbUpdateException ex) when (ex.InnerException is PostgresException
    {
        SqlState: PostgresErrorCodes.UniqueViolation,   // 23505
        ConstraintName: "ix_orders_external_reference",
    })
{
    throw new DuplicateOrderException(order.ExternalReference);
}
```

Match on `SqlState` *and* `ConstraintName` — a table with two unique constraints must map to two distinct typed errors. The typed exception is Core's ([../foundations/errors-and-logging.md](../foundations/errors-and-logging.md)); the `PostgresException` inspection stays in Infrastructure.

### When Dapper

EF Core is the default; Dapper enters only via ADR for a *measured* hot path ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).

- Same repository port in Core; the Dapper implementation is an Infrastructure detail behind it.
- Parameterized SQL only; SQL strings live next to the repository, not scattered.
- Dapper queries run against the schema EF migrations own — Dapper never gets its own migration story.

### Outbox And Inbox

- If business logic must commit DB state and emit an event together, default to a transactional outbox: write domain state and an `OutboxMessage` row in one transaction, then relay asynchronously.
- Use an inbox or durable dedupe table on the consumer side when duplicate processing would be harmful.
- Keep relay workers and dedupe storage explicit; do not hide them behind a repository abstraction that erases failure semantics. Full pattern: [eventing-and-messaging.md](eventing-and-messaging.md).

### Integration Tests

Mocked `DbSet`s prove nothing about SQL, translation, or constraints. Integration tests run against real PostgreSQL via Testcontainers, on the migrated schema:

```csharp
public sealed class PostgresFixture : IAsyncLifetime
{
    private readonly PostgreSqlContainer _container = new PostgreSqlBuilder().Build();

    public string ConnectionString => _container.GetConnectionString();

    public async ValueTask InitializeAsync()
    {
        await _container.StartAsync();
        await using var db = CreateContext();
        await db.Database.MigrateAsync();
    }

    public async ValueTask DisposeAsync() => await _container.DisposeAsync();
}
```

Integration tests run behind the explicit `-Integration` switch of the verify gate (Docker is not guaranteed on every dev machine). See [../quality/testing.md](../quality/testing.md).

## Common Mistakes And Forbidden Patterns

- `DbContext` as a singleton, shared across threads, or held beyond its unit of work.
- Lazy-loading proxies, or `Include`-everything queries where a projection was the answer — N+1 and over-fetch by default.
- Repositories returning `IQueryable` across the Core boundary, or repository interfaces so broad they become a second domain model.
- Raw SQL string concatenation instead of parameterized `FromSql` interpolation.
- Tests that mock `DbSet`/`DbContext` but never prove migrations or real queries.
- `Database.Migrate()` on normal startup — racing replicas and silent destructive DDL.
- Editing, renumbering, or reordering a migration that has already shipped instead of appending a new one.
- A destructive migration (drop, rename, narrow, add `NOT NULL`) shipped in the same release as code that still depends on the old shape, instead of staged expand/contract.
- Migrations with an empty scaffolded `Down` that cannot restore the prior schema.
- Dual-write designs that update the database and publish to a broker in separate success paths without an outbox.
- Explicit transactions outside the execution strategy while `EnableRetryOnFailure` is on — torn retries.
- Catching `DbUpdateException` without inspecting `SqlState`/`ConstraintName`, or swallowing `DbUpdateConcurrencyException`.
- Pre-checking uniqueness with a `SELECT` instead of translating the constraint violation.
- Shipping the default connection pool: any of the four Npgsql limits left unset, or no `Connection Lifetime` on a path through a proxy/failover endpoint.
- Pool limits aggressive enough to churn connections instead of reusing them.
- Pool saturation invisible because the `Npgsql` meter is never exported.

## Verification And Proof

- run `pwsh ./verify.ps1` (restore (locked), format-check, build (warnings-as-errors), test, audit); integration tests behind the `-Integration` switch
- `dotnet ef migrations has-pending-model-changes` proves model and migrations have not drifted
- migration apply-then-rollback-then-reapply test against real PostgreSQL (`MigrateAsync()` → `dotnet ef database update <previous>` → `MigrateAsync()`), proving forward, reverse, and re-apply all succeed
- repository integration tests pass against the freshly migrated schema via Testcontainers
- for expand/contract changes, proof that release N and N+1 application code both run correctly against the release-N schema
- a test that a duplicate insert surfaces the typed conflict (assert on the domain exception, not on `PostgresException`)
- a concurrency test: two loads, two conflicting saves, second save observes `DbUpdateConcurrencyException` mapped to the typed conflict
- proof that callers pass `CancellationToken` through every async EF call
- proof that all four pool limits are set from config, and a load test at expected peak concurrency showing no runaway pool-wait or connection churn
- the `Npgsql` meter exported through OpenTelemetry, with pool saturation visible on a dashboard
