# Recipe: Add Event Consumer

Use this when a process handles at-least-once message delivery.

## Files To Touch

- envelope and payload Zod schemas
- core handler and consumer-owned ports
- broker adapter under `src/lib/`
- `src/db/` inbox/dedupe SQL and migration
- lifecycle, retry/DLQ, telemetry, replay, and integration tests

## Steps

1. Document delivery, ordering, duplication, version, maximum size, and acknowledgement semantics.
2. Parse envelope and versioned payload from `unknown` before dispatch.
3. Bound prefetch and concurrent handlers; pass the process `AbortSignal`.
4. Make the durable effect idempotent using event ID plus consumer identity.
5. Commit inbox receipt and PostgreSQL effect in one transaction when possible.
6. Acknowledge only after the required durable effect completes.
7. Classify permanent, validation, authorization, transient, and cancellation failures.
8. Retry only transient failures with bounded backoff; route poison messages to an owned DLQ.
9. Define alert, inspection, repair, rate-limited replay, and retention.

```bash
npm run test:integration -- --runInBand
npm test -- --runInBand src/lib/<broker>/<name>-consumer.test.ts
npm run verify
```

## Invariants To Preserve

- Exactly-once delivery is never claimed from broker settings.
- Duplicate and concurrent duplicate delivery produce one effect.
- Intake stops before in-flight handlers drain during shutdown.
- Retry and replay preserve the original stable event ID.
- DLQ is not disposal; it has owner, alert, repair, and age objective.
- Trace context is propagated but never trusted for authorization.

## Proof

- Tests cover success, duplicate, concurrent duplicate, transient retry, exhaustion, and poison input.
- Broker integration proves acknowledgement, redelivery, ordering scope, and DLQ behavior.
- Abort and shutdown tests prove no early acknowledgement or abandoned owned promise.
- Replay of a representative batch is bounded, observable, and idempotent.
- `npm run verify` is green.
