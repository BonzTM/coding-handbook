# Recipe: Add Migration

Use this when the PostgreSQL schema must change across a mixed-version rollout.

## Files To Touch

- `alembic/versions/<revision>_<slug>.py`
- SQLAlchemy metadata/mappings and repository callers
- migration and repository integration tests
- deployment migration step and rollback notes when behavior changes

## Steps

1. Update metadata, then generate a candidate: `uv run alembic revision --autogenerate -m "<description>"`.
2. Review every emitted operation. Delete unrelated churn; supply explicit names; inspect server defaults, nullability, indexes, constraints, and type changes.
3. Hand-edit `upgrade()` and `downgrade()` so the revision expresses the intended SQL. Never trust autogenerate as review.
4. Classify the change as additive or destructive. Stage rename/drop/`NOT NULL`/type narrowing through expand, backfill, switch, then contract releases.
5. Apply only through the explicit migration job; normal application startup never calls Alembic.

## Invariants To Preserve

- Shipped revisions are immutable and form one reviewed history.
- Old and new application versions work throughout rollout.
- Backfills are bounded, restartable, observable, and separate when table size makes a transaction unsafe.
- Downgrade is honest; irreversible data loss is documented as forward-only with restore procedure.

## Proof

```bash
uv run alembic upgrade head
uv run alembic downgrade -1
uv run alembic upgrade head
uv run pytest -m integration -k migration
make verify
```

Run against a clean real PostgreSQL database and an upgraded representative schema. Governing doc: [database](../services/database.md).
