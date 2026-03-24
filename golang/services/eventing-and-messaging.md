# Eventing and Messaging

Event-driven systems need the same level of rigor as HTTP or database boundaries. Messages are contracts, delivery is a runtime behavior, and replay is part of the design, not an edge case.

## Default Approach

- Stay broker-neutral in application code when practical.
- Assume at-least-once delivery from day one.
- Make consumers idempotent before adding retries.
- Treat ordering as per-entity or per-key, never global.
- When a service both commits DB state and publishes messages, default to an outbox pattern.

## Contract Source And Envelope

- Every published event needs one obvious source of truth.
- For broadly consumed event streams, prefer a documented schema source such as AsyncAPI plus schema files under `api/`.
- Carry stable metadata in a CloudEvents-style envelope even if the repo does not adopt CloudEvents literally.

Recommended metadata fields:

- `id`: stable event identifier used for dedupe and traceability
- `type`: event name, not a table name or vague verb
- `source`: producing service or subsystem
- `time`: producer timestamp
- `subject`: optional entity identifier or logical subject
- `datacontenttype`: payload encoding such as `application/json`
- `dataschema`: schema URI or compatibility anchor when the repo uses one

## Schema Evolution And Compatibility

- Prefer additive-only changes.
- Consumers should tolerate unknown fields.
- Producers must not silently change field meaning in place.
- Breaking changes deserve a new contract version or a new event type, not a surprise field reinterpretation.
- If compatibility governance matters across teams, prefer Protobuf or Avro plus automated validation over undocumented JSON drift.

## Publisher Rules

- Publish after domain behavior is decided, not before.
- Include correlation and trace context in message metadata.
- Choose keys that match the ordering or dedupe contract, usually an aggregate or entity ID.
- Do not emit best-effort side effects from request handlers and hope a retry somewhere else fixes it later.

## Consumer Rules

- A consumer should decode, validate, dedupe, call core logic, and only then acknowledge or commit offset/settlement.
- Ack after durable side effects, not after parsing.
- Validation or schema failures are usually non-retryable and should move toward operator visibility or DLQ policy quickly.
- Transient dependency failures can retry with bounded backoff and jitter.

## Ordering And Concurrency

- Promise ordering only within one key or partition.
- Scaling consumers usually means more keys or partitions, not more concurrency on one ordered stream.
- If work fans out asynchronously, preserve per-key sequencing in the owning worker.
- One redelivery on an ordered key may require replay or delay of later events on that same key.

## Retries And Dead-Letter Behavior

- Retry only transient failures.
- Use bounded exponential backoff with jitter.
- Stop retrying after a documented budget; then dead-letter, park, or surface the event for operator action.
- A dead-letter record should retain original destination, attempt count, failure class, and correlation data.
- Replays should be operator-controlled, not an infinite feedback loop.

## Outbox And Inbox Patterns

- If a service writes domain state and emits an event as one logical action, use a transactional outbox by default.
- Write business state and an outbox row in one DB transaction, then relay that row to the broker asynchronously.
- On the consumer side, use an inbox or durable dedupe table keyed by event ID when duplicate delivery would be harmful.
- Avoid dual writes to the database and broker in separate success paths without an explicit failure model.

## Suggested Layout

```text
api/events/                 # optional published schemas or AsyncAPI
internal/messaging/
  producer/
  consumer/
  relay/                    # outbox relay if needed
internal/db/
  migrations/
  outbox.sql
```

Keep the exact package names repo-specific. The important rule is ownership: contracts in one place, delivery logic in one place, business rules in core packages.

## Observability

- Log publish, receive, retry, exhaustion, and dead-letter transitions with stable fields.
- Trace producer send, consumer receive, handler execution, and settlement where tracing is enabled.
- Track at least: publish failures, consume failures, retry count, duplicate drops, handler latency, backlog age, and DLQ count.
- Message IDs and correlation IDs belong in logs and traces, not as high-cardinality metric labels.

## Testing And Proof

- Contract tests prove payload shape and compatibility expectations.
- Duplicate-delivery tests prove consumer idempotency.
- Replay and ordering tests prove per-key behavior.
- Retry and DLQ tests prove failure classification and bounded exhaustion behavior.
- Integration tests should run against real infrastructure or realistic emulators when the broker semantics matter.

## Common Mistakes And Forbidden Patterns

- Publishing events directly from handlers while skipping a durable outbox where atomicity matters.
- Treating exactly-once delivery as a default assumption.
- Using global ordering as a product promise.
- Putting event schemas only in generated code or only in tribal knowledge.
- Retrying validation failures forever.
- Using message IDs, raw subjects, or tenant IDs as metric labels.

## Verification And Proof

- event schemas or examples validate in CI when the repo publishes them
- producers and consumers have targeted contract and replay tests
- outbox or inbox flows are covered by real storage integration tests when present
- publish/consume telemetry is visible in logs, metrics, or traces before the feature is called done
