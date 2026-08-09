# Database

Persistence defaults for visible SQL, explicit async transactions, and production-safe schema evolution.

## Default Approach

Use SQLAlchemy 2.0 with `asyncpg`, an application-scoped async engine, one `AsyncSession` per request or unit of work, and Alembic migrations. Keep mappings and repositories under `src/<app>/db/`; core owns the repository `Protocol`.

### Suggested Layout

```text
src/<app>/db/
  engine.py             # engine and async_sessionmaker construction
  models.py             # typed ORM mappings or Core tables
  repositories.py       # core Protocol implementations
  unit_of_work.py       # explicit transaction owner when needed
alembic/
  env.py
  versions/
```

The engine is created once in FastAPI lifespan and disposed after all users stop. No global connection, per-request engine, or import-time database I/O.

### Engine And Session Ownership

Construct `create_async_engine()` and one `async_sessionmaker` in composition. A request or worker message opens an `AsyncSession` with `async with`, begins the explicit transaction at the coordinating boundary, commits only on success, and rolls back/closes on failure.

An `AsyncSession` is mutable transaction state and cannot be shared across concurrent tasks. SQLAlchemy states that each concurrent task needs a separate session in its [asyncio documentation](https://docs.sqlalchemy.org/en/20/orm/extensions/asyncio.html#using-asyncsession-with-concurrent-tasks). Do not use ambient scoped sessions to hide ownership.

### Models And Domain Mapping

Use typed `Mapped[T]`/`mapped_column()` ORM declarations or SQLAlchemy Core `Table` objects. Pick per repository; do not mix styles inside one aggregate without a clear mapping boundary. The [SQLAlchemy declarative guide](https://docs.sqlalchemy.org/en/20/orm/declarative_mapping.html) defines typed mappings.

ORM rows are persistence shapes. Repositories map them into plain core values before returning. Core never imports `AsyncSession`, `Mapped`, rows, or database exceptions. Keep DTO aliases, authorization, and HTTP pagination outside database models.

### Query And Loading Discipline

Build statements with SQLAlchemy expressions or `text()` plus bound parameters. Dynamic values never enter SQL through f-strings, `%`, `.format()`, or string concatenation. Dynamic identifiers use an allowlisted mapping to known SQL objects; bind parameters are for values, not identifiers.

Avoid implicit I/O on attribute access. Load required relationships explicitly with `selectinload()` for collections or `joinedload()` where row multiplication is understood, and consider `lazy="raise"` for relationships that must never fetch unexpectedly. SQLAlchemy documents how lazy loading can create N+1 queries and fail under asyncio in its [loader guide](https://docs.sqlalchemy.org/en/20/tutorial/orm_related_objects.html#loader-strategies).

Select only the rows/columns needed, paginate every growing collection with a stable total order, and inspect emitted SQL/query plans for hot paths. Do not return a live ORM object whose later attribute access depends on an open session.

### Transaction Rules

- The use case or unit of work owns a transaction spanning one logical database action.
- Repositories flush when generated values are needed; they do not hide commits.
- No network call, broker publish, user wait, or unbounded computation occurs inside a transaction.
- Lock rows in a deterministic order and use the narrowest isolation/locking semantics that prove the invariant.
- Translate expected integrity/serialization failures once at the repository boundary; preserve the original exception through chaining.

When domain state must imply an event atomically, write an outbox record in the same transaction. Never publish to the broker inside the open transaction and call the dual write atomic.

### Pool Sizing And Timeouts

Set `pool_size`, `max_overflow`, and `pool_timeout` explicitly. Size the per-process maximum so all replicas plus migrations/admin headroom stay below PostgreSQL capacity. SQLAlchemy defines `pool_size`, overflow, and checkout wait in its [engine configuration](https://docs.sqlalchemy.org/en/20/core/engines.html#sqlalchemy.create_engine.params.pool_timeout).

Bound three independent waits:

- connection establishment through asyncpg's connection `timeout`
- pool checkout through SQLAlchemy `pool_timeout`
- statements through PostgreSQL `statement_timeout` and/or asyncpg `command_timeout`

Pass driver settings through reviewed `connect_args`; asyncpg documents `timeout`, `command_timeout`, and `server_settings` in its [API reference](https://magicstack.github.io/asyncpg/current/api/index.html#connection). The request's shorter remaining deadline still wins. Observe checked-out connections, checkout wait, overflow, query duration, errors, and transaction age with low-cardinality labels.

If PgBouncer is present, document its mode and validate prepared-statement/pool behavior; do not stack pools by accident. Configuration changes require a load test showing bounded checkout wait and database headroom.

### Migrations

Alembic revisions are append-only deployment artifacts. Generate a candidate with `alembic revision --autogenerate`, then review every operation, type/default, lock, data rewrite, index strategy, downgrade posture, and generated import. Alembic explicitly says autogeneration is not perfect and requires manual correction in its [autogenerate guide](https://alembic.sqlalchemy.org/en/latest/autogenerate.html).

Apply migrations through an explicit deploy/init step before traffic, never during normal application startup. Only one controlled migrator runs. The service fails readiness when its expected schema is absent or incompatible; it does not repair schema silently.

Expand/contract destructive changes across releases: add compatible shape, deploy code that handles both, backfill in bounded resumable batches, switch reads/writes, verify old use is zero, then remove later. Never edit, reorder, or renumber a revision already applied outside disposable development.

Downgrades are required when they are safe and honest. Destructive data loss is recovered through a documented forward fix or backup restore, not a fictional downgrade. Migration policy follows [contracts and compatibility](../foundations/contracts-and-compatibility.md).

### Repository Ports

Core defines narrow consumer-owned `Protocol`s in domain terms. A repository implements one aggregate/use-case need; it is not a generic CRUD base class or a second ORM. Pass tenant/authorization scope explicitly where persistence must enforce it.

Unit tests use a focused fake to prove core behavior. Integration tests prove the repository, transaction, constraints, and mappings against real PostgreSQL. A fake does not claim SQL proof.

## Common Mistakes And Forbidden Patterns

- Engine or session hidden in a global, context variable, singleton, or router dependency with unclear lifetime.
- One `AsyncSession` shared by sibling tasks or retained beyond its request/message.
- Commits inside low-level helpers, network calls inside transactions, or swallowed rollback failures.
- SQL values interpolated into strings or dynamic identifiers accepted without an allowlist.
- Implicit lazy loads, N+1 queries, or ORM objects crossing into core/HTTP responses.
- Default/unbounded pool behavior, no checkout/connect/statement timeout, or capacity sized per pod without replica math.
- Alembic autogenerate accepted blindly, migration history edited, or migrations run on normal startup.
- Destructive schema and incompatible code shipped in one rolling release.
- SQLite/mocks presented as proof of PostgreSQL queries or migrations.

## Verification And Proof

```bash
uv run alembic upgrade head
uv run pytest -m integration tests/db
make verify
```

On a disposable real PostgreSQL instance, apply all revisions from empty, run repository/constraint/rollback tests, exercise the supported downgrade or forward-recovery posture, and reapply. For expand/contract, prove old and new application versions against the intermediate schema. Load-test expected concurrency and show pool limits, checkout wait, statement deadlines, transaction duration, and database capacity remain bounded.

Related: [data modeling](../foundations/data-modeling.md), [eventing and messaging](eventing-and-messaging.md), [testing](../quality/testing.md), and [add migration](../recipes/add-migration.md).
