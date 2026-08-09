# Recipe: Add Event Publisher

Use this when a completed fact must be published reliably.

## Files To Touch

- event envelope and payload schemas under the owning feature
- core publication port
- `src/db/` outbox SQL, row parsing, and migration
- bounded relay adapter and lifecycle wiring
- contract, integration, replay, telemetry, and runbook artifacts

## Steps

1. Name a past-tense event and define stable ID, type, version, occurred-at, producer, and payload.
2. Parse and serialize through reviewed Zod schemas; set a maximum message size.
3. Decide whether direct-publish loss is acceptable; otherwise use a transactional outbox.
4. Write the domain change and outbox row in one PostgreSQL transaction.
5. Claim rows in bounded batches with lease/ownership semantics.
6. Publish using the stable event ID, then mark success; assume duplicate publish is possible.
7. Bound relay concurrency, timeout, retry, backoff, and shutdown drain.
8. Define additive evolution and old/new consumer compatibility.
9. Observe outbox backlog, oldest age, attempts, result, and publish latency.

```bash
npm run test:integration -- --runInBand
npm test -- --runInBand src/lib/<broker>/<name>-publisher.test.ts
npm run verify
```

## Invariants To Preserve

- Domain code imports no broker client or transport envelope type.
- State and publish intent are atomic when loss is unacceptable.
- Stable event IDs survive retries and replay.
- Messages, batches, concurrency, leases, and retries are bounded.
- Payloads and IDs do not become metric attributes.
- Direct publish documents the tolerated inconsistency explicitly.

## Proof

- Contract fixtures prove current and supported reader versions.
- Integration tests cover commit, rollback, crash before mark, duplicate publish, and lease expiry.
- Broker proof confirms headers, acknowledgement, and redelivery behavior.
- Backlog recovery and shutdown drain stay within configured bounds.
- `npm run verify` is green.
