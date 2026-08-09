# Recipe: Add Idempotent Write

Use this when a non-GET HTTP operation must survive client retries without repeating its effect.

## Files To Touch

- HTTP idempotency dependency/adapter and core coordinating port
- Alembic revision plus SQLAlchemy idempotency repository
- write path and real-PostgreSQL integration tests

## Steps

1. Require and bound `Idempotency-Key`; scope storage by principal/tenant, route, and key.
2. Hash a canonical fingerprint of method, route, relevant headers, and request body.
3. In one transaction, claim the unique key, perform the domain write, and store status, content type, and exact response bytes.
4. Replay a matching completed key byte-identically. Reject a mismatched fingerprint with `422`; reject an in-flight duplicate with the documented `409` or `425`.
5. Define a client-visible retention window and a bounded reaper.

## Invariants To Preserve

- Domain write and completed idempotency record commit atomically.
- Exactly one concurrent claimant executes the effect; conflicts use the database constraint.
- Responses never cross tenant/principal scope and are replayed from stored bytes, not regenerated state.
- Expiry exceeds the supported client retry window and cleanup is observable.

## Proof

```bash
uv run pytest -k idempotency
uv run pytest -m integration -k idempotency
make verify
```

Prove sequential replay, concurrent first use, mismatched body, missing key, expiry, rollback, and crash-window semantics against real PostgreSQL. Governing docs: [HTTP services](../services/http-services.md) and [database](../services/database.md).
