# Recipe: Add Event Publisher

Use this when durable behavior emits an event or message.

## Files To Touch

- core-owned publisher `Protocol` and producing use case
- boundary Pydantic envelope/payload model and broker adapter
- Alembic outbox revision/repository/relay when database state and publish meet
- contract, relay, and real-broker tests

## Steps

1. Define stable event name/version, event ID, occurred-at UTC timestamp, correlation metadata, payload, and compatibility policy.
2. Map from domain state at the adapter boundary; do not publish ORM rows or transport objects.
3. When state and event must agree, insert the domain change and outbox record in one PostgreSQL transaction.
4. Relay pending records with bounded batches/concurrency, publish through the adapter, then mark sent. Make the relay restartable.
5. Define ordering key, bounded retry/jitter, terminal handling, and retention before enabling production traffic.

## Invariants To Preserve

- Outbox closes the commit/publish gap but still permits duplicate delivery; stable event IDs enable consumer dedupe.
- Publishing never occurs inside the database transaction.
- Payloads contain no secrets and have one schema source of truth.
- Broker types do not enter core.

## Proof

```bash
uv run pytest -k 'event_contract or outbox'
uv run pytest -m integration -k 'publisher or outbox_relay'
make verify
```

Prove crash recovery between commit/publish/mark-sent, broker failure, duplicate publish, ordering, and relay lag signals. Governing doc: [eventing and messaging](../services/eventing-and-messaging.md).
