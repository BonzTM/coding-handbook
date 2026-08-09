# Recipe: Add Event Consumer

Use this when a queue or stream message drives durable behavior.

## Files To Touch

- broker adapter and boundary Pydantic envelope/payload model
- core use case and inbox/dedupe repository plus migration
- telemetry/readiness and consumer integration tests

## Steps

1. Decode and validate the supported contract version before invoking core.
2. Classify schema/authorization/business failures as terminal and transient dependency failures as retryable.
3. In one transaction, claim the stable event ID in the inbox and commit the durable side effect; duplicates become no-ops.
4. Acknowledge only after the durable terminal decision. Bound retries with jitter and preserve original metadata in DLQ/parked records.
5. Bound per-message time, payload size, concurrency, prefetch, and shutdown drain.

## Invariants To Preserve

- At-least-once duplicate delivery is safe; validation failures never poison-retry forever.
- Required per-key ordering is preserved through partition key and concurrency policy.
- Cancellation stops intake, settles owned work deliberately, and is re-raised.
- Replay cannot bypass validation, authorization, dedupe, or telemetry.

## Proof

```bash
uv run pytest -k 'consumer or duplicate or replay'
uv run pytest -m integration -k 'consumer or dead_letter'
make verify
```

Prove duplicate delivery, retry exhaustion, DLQ metadata, replay, ordering, broker reconnect, and bounded shutdown. Governing doc: [eventing and messaging](../services/eventing-and-messaging.md).
