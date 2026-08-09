# Recipe: Add Database Feature

Use this when behavior changes schema, SQL, mapping, or transaction ownership.

## Files To Touch

- `src/<app>/db/` mappings, repository, and session/transaction code
- `src/<app>/core/` port or use case
- `alembic/versions/` when schema changes
- real-PostgreSQL integration tests

## Steps

1. Define the consumer-owned repository `Protocol` in core using domain values.
2. Add the SQLAlchemy 2.0 async implementation under `db`; map ORM/Core rows at the adapter edge.
3. Keep one explicit `AsyncSession` transaction owned by the coordinating repository/unit of work.
4. Parameterize every query and bound result size, statement time, and lock scope.
5. Add the migration first when the schema changes and preserve mixed-version rollout safety.

## Invariants To Preserve

- Core imports no SQLAlchemy type; routers never receive sessions.
- No network call occurs while a database transaction is open.
- Commits and rollbacks are explicit; sessions are closed on every path.
- PostgreSQL behavior is proven against PostgreSQL, not SQLite or a mock.

## Proof

```bash
uv run alembic upgrade head
uv run pytest -m integration -k '<repository_or_feature>'
make verify
```

Prove rollback/error paths, query cardinality, and mixed-version compatibility. Governing doc: [database](../services/database.md).
