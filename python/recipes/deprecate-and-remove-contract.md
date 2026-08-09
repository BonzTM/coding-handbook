# Recipe: Deprecate And Remove A Contract

Use this to retire an HTTP, gRPC, event, public Python, or database contract over multiple releases.

## Files To Touch

- contract source and adapter serving it
- low-cardinality deprecation usage telemetry
- changelog/release notes and mixed-version compatibility tests
- Alembic revision only in the later contract phase for schema removal

## Steps

1. Ship the additive replacement before or with deprecation; keep old behavior compatible.
2. For HTTP, add standards-compliant deprecation documentation and a `Sunset` HTTP-date header; pin any `Deprecation` header form only after verifying the current standard. For protobuf, mark deprecated; for Python public APIs, emit a documented warning; for DB, stop new writes before drop.
3. Measure usage at the exact serving/reading boundary with finite labels and announce replacement, window, and earliest removal release.
4. Remove only after observed zero use for the full window across supported consumers.
5. Reserve removed protobuf names/numbers; stage database drops through expand/contract; announce the breaking removal and bump version appropriately.

## Invariants To Preserve

- A calendar date alone never authorizes removal; telemetry and supported-consumer evidence do.
- Old/new clients and mixed application versions work throughout the window.
- Removal preserves a migration path, rollback story, and immutable schema/protobuf history.
- Deprecated secrets or PII never become telemetry labels.

## Proof

```bash
uv run pytest -k 'deprecat or compat'
uv run pytest -m integration -k migration
make verify
```

For HTTP, assert the `Sunset` value and later `404`/documented `410`; cite the zero-usage query in the removal change. Governing doc: [contracts and compatibility](../foundations/contracts-and-compatibility.md).
