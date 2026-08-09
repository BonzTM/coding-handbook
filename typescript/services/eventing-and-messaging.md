# Eventing And Messaging

Producer, consumer, delivery, replay, and schema rules for reliable asynchronous integration.

## Default Approach

Assume at-least-once delivery, validate every message, and make effects idempotent.

### Event Contracts

An event records a completed fact in past tense. Commands request an action and use a distinct contract. Each envelope includes stable event ID, type, schema version, occurred-at UTC instant, producer identity, trace context, and payload.

Parse envelope and payload with Zod before domain use. Define maximum message size and unknown-version behavior. Never trust broker metadata or generic TypeScript parameters as runtime validation.

Evolve schemas additively and test old/new producer-consumer combinations. An incompatible meaning gets a new event type or version and a migration/replay plan.

### Producer Reliability

Publishing after a database commit can lose an event; publishing before it can announce a rolled-back fact. Use a transactional outbox when an event must correspond to PostgreSQL state.

Write domain change and outbox record in one transaction. A bounded publisher claims pending rows, publishes with a stable message ID, and marks success. Duplicate publish remains possible, so consumers stay idempotent.

Direct publish is acceptable only when loss and inconsistency are explicitly tolerable. Do not hide reliability semantics behind a generic broker abstraction.

### Consumer Ownership

Limit concurrent handlers and prefetched messages. Pass a process-lifetime `AbortSignal`; on shutdown stop intake, drain owned handlers within the deadline, then close the broker client.

Acknowledge only after the required durable effect completes. Classify validation, authorization, permanent business, transient dependency, and cancellation failures. Retry only transient failures with bounded exponential backoff and jitter.

### Idempotency And Inbox

Use event ID plus consumer identity as the deduplication key. When the consumer writes PostgreSQL, record inbox receipt and business effect in one transaction. A duplicate returns the previously completed outcome without repeating the effect.

Define retention long enough for broker redelivery and operational replay. Key expiry is a delivery contract, not arbitrary cache cleanup.

### Ordering, DLQ, And Replay

Do not assume global ordering. If per-key ordering matters, define the partition key and behavior when a predecessor is delayed or dead-lettered.

After bounded retries, route poison messages to a DLQ with failure classification and safe metadata. DLQ is not disposal: define alert, owner, inspection, repair, replay, and age objectives.

Replay uses the original stable event ID, preserves or records original occurrence time, is rate-limited, and cannot bypass current schema validation or authorization policy. Prove reprocessing is idempotent before a bulk replay.

### Observability

Propagate trace context without treating it as trusted authorization state. Record low-cardinality event type, consumer, result, attempt bucket, and latency. Message IDs belong in logs and traces, not metric attributes.

Measure queue lag, oldest age, retry/DLQ rate, handler latency, and outbox backlog. Alerts are symptom-based and route to an actionable runbook.

## Common Mistakes And Forbidden Patterns

- Claiming exactly-once delivery from broker settings alone.
- Acknowledging before the durable effect completes.
- Unbounded consumer concurrency, prefetch, retry, or replay.
- Event payloads used as mutable database snapshots without compatibility rules.
- Publishing separately from the transaction when loss is unacceptable.
- Retrying validation and permanent business failures forever.
- DLQ without owner, alert, repair, and replay procedure.
- Message IDs as metric attributes.

## Verification And Proof

- Contract fixtures prove current and supported older message versions.
- Duplicate and concurrent duplicate tests produce one durable effect.
- Outbox tests cover crash before publish, duplicate publish, claim expiry, and backlog recovery.
- Consumer tests cover success, transient retry, retry exhaustion, poison message, abort, and drain.
- Broker integration proves acknowledgement, redelivery, partitioning, and DLQ behavior.
- Replay of a representative batch is bounded, observable, and idempotent.

Related: [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md), [database.md](database.md), and [../operations/resilience.md](../operations/resilience.md).
