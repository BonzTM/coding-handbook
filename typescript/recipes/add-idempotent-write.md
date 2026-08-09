# Recipe: Add Idempotent Write

Use this when clients may retry a state-changing HTTP request.

## Files To Touch

- request and response schemas under `src/api/`
- core write use case and idempotency contract
- `src/db/` dedupe store, SQL, row schema, and migration
- transport and PostgreSQL integration tests
- metrics, retention config, and API documentation

## Steps

1. Define key syntax, scope, retention, request fingerprint, replayed response, and conflict behavior.
2. Parse and bound the `Idempotency-Key`; never log or metric-label the raw key.
3. Hash a canonical request representation when key reuse with different input must conflict.
4. Atomically reserve or complete the key with the durable business effect.
5. Store the stable response status and body needed for contractually identical replay.
6. Define concurrent in-progress behavior: bounded wait, conflict, or retry guidance.
7. Recover expired claims safely and keep retention longer than supported client retry windows.
8. Emit bounded first/replay/conflict/in-progress outcome metrics.
9. Document which operations are idempotent and how clients retry.

```bash
npm run test:integration -- --runInBand
npm test -- --runInBand src/api/<feature>-route.test.ts
npm run verify
```

## Invariants To Preserve

- Duplicate and concurrent duplicate requests produce one durable effect.
- The dedupe record and business effect share one consistency boundary.
- Key reuse with a different canonical request cannot replay unrelated success.
- Authorization is re-evaluated according to the documented replay policy.
- Retention cleanup is bounded and cannot race active claims incorrectly.
- Raw keys and request bodies remain absent from telemetry.

## Proof

- Real PostgreSQL tests cover first request, sequential replay, and concurrent duplicate.
- Tests cover key conflict, expired claim, failed effect, cancellation, and cleanup.
- Replay matches the documented status, headers, and body contract.
- Metrics distinguish outcomes without high-cardinality values.
- `npm run verify` is green.
