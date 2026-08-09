# Eventing and Messaging

Message contracts, at-least-once delivery, settlement, replay, and worker ownership are explicit system behavior.

## Default Approach

Keep application code broker-neutral through core-owned publisher/handler `Protocol`s and a thin broker adapter under `src/<app>/workers/`. Assume at-least-once delivery, make handlers idempotent before adding retries, and use an outbox when database state must imply publication.

The broker library remains deferred to [framework selection](../decisions/framework-selection.md). No framework may hide acknowledgement, offset, retry, prefetch, or shutdown behavior from the owning worker.

### Contract Source And Envelope

Published payloads are versioned Pydantic v2 models under an owned `api/events/` or adapter contract package. Parse bytes/dictionaries into the boundary model, reject malformed data, then map once into plain domain values. Independently evolving consumers ignore additive unknown fields while validating every field they use.

Every envelope carries:

- stable event/message ID for dedupe and traceability
- stable type and explicit schema version
- producing source
- aware UTC occurred-at time
- correlation and trace context
- optional subject/ordering key
- content type and schema reference when a registry exists

Names describe domain facts or commands, not table names. Schemas, examples, and compatibility policy are reviewed artifacts under [contracts and compatibility](../foundations/contracts-and-compatibility.md).

### Publisher Rules

Publish only after the use case decides the domain outcome. Map a domain event into the versioned payload at the publisher adapter. Include stable metadata and choose a key matching the documented per-entity ordering contract.

Direct best-effort publication after a database commit is not atomic. If failure would leave durable state without its required event, write an outbox record in the same transaction and relay later. A handler does not keep a database transaction open while waiting for a broker.

### Consumer Worker Shape

One composition-owned root task supervises the broker connection, receive loop, in-flight message tasks, and settlement. The loop:

1. receives within a bounded prefetch/queue limit
2. decodes and validates the envelope/payload
3. checks durable inbox/idempotency state when required
4. invokes the core handler under a per-message timeout
5. durably records the outcome
6. acknowledges, retries, rejects, or dead-letters explicitly

Use `asyncio.TaskGroup` for the owned set and a semaphore or broker prefetch to cap in-flight work. Never create one unbounded task per message. The queue limit, handler concurrency, per-message deadline, and shutdown drain are typed configuration with safe caps.

### At-Least-Once And Idempotency

Assume the same message can arrive concurrently, after a successful side effect, and long after the first delivery. A handler is idempotent by domain design or guarded by a durable inbox keyed by consumer identity plus event ID. The inbox reservation and local side effect share one transaction where possible.

Mark completion only after durable effects succeed. A duplicate of a completed message acknowledges without reapplying the effect. An in-progress/stale reservation has an explicit recovery rule. Never infer exactly-once behavior from a broker feature or a process-local set.

### Settlement And Failure Classification

- Acknowledge/commit only after durable success or recognized completed duplicate.
- Reject without retry for invalid schema, unsupported version, authorization/policy rejection, or permanent domain failure; preserve operator evidence according to DLQ policy.
- Nack/retry transient dependency, concurrency, or availability failures within one documented attempt/age budget.
- Do not settle when cancellation interrupts work unless the broker's lease/visibility behavior makes redelivery explicit and safe.

Broker adapters translate these decisions into ack/nack/offset/visibility operations. Core returns domain outcomes and never imports broker delivery types.

### Retries And Dead-Letter Policy

Retries are bounded exponential backoff with full jitter, consume the message's overall age/deadline budget, and apply only to retryable outcomes. Preserve attempt count in broker metadata or durable state; process restarts do not reset an infinite retry loop.

After exhaustion, dead-letter or park the original envelope with original destination, receive time, attempt count, safe failure class/code, correlation context, and schema version. Do not copy secrets or raw exception tracebacks. Define DLQ retention, access control, alerting, ownership, replay approval, and purge behavior before enabling it.

### Outbox And Inbox

The outbox row contains the serialized/versioned event or enough immutable data to build it deterministically, destination/key, event ID, creation time, and relay state. The relay claims a bounded batch, publishes idempotently where supported, and marks sent. A crash between publish and mark-sent causes duplicate publication by design; consumers remain idempotent.

The inbox records consumer name, event ID, processing state, and completion/expiry metadata. Put uniqueness in the database. Bound retention against the maximum broker replay/redelivery window; deleting dedupe state earlier re-enables effects.

### Ordering And Concurrency

Promise ordering only within a documented key/partition. Global ordering is forbidden. Preserve one-key sequencing when work fans out; more concurrency comes from independent keys/partitions. A failed message may block or delay later messages for the same key according to an explicit policy.

Handlers tolerate replay and document whether out-of-order events are rejected, ignored by version/sequence, buffered within a bound, or applied commutatively. Broker arrival order alone is not domain version control.

### Replay

Replay is a named operator workflow with source range/filter, target consumer version, rate/concurrency limits, dry-run where possible, idempotency retention check, telemetry, pause/abort, and audit record. Do not route DLQ records back into the live topic automatically.

Before replay, prove current code can parse old schema versions and that required side effects are idempotent. Rebuild/projection consumers distinguish replay telemetry from live lag without changing business meaning.

### Observability And Health

Emit low-cardinality metrics for receives, successful handling, duplicates, retries, permanent rejects, settlement failures, handler duration, in-flight count, backlog/oldest age, outbox lag, and DLQ count. Message IDs and ordering keys belong in safe logs/traces, never metric labels.

Readiness becomes false when the worker cannot receive or safely settle required messages, during startup, and before drain. Liveness remains local. Shutdown stops intake, drains in-flight messages within a timeout, settles terminal results, cancels/awaits the TaskGroup, and closes broker/storage clients.

## Common Mistakes And Forbidden Patterns

- Broker payload/delivery types or Pydantic models in core.
- Exactly-once assumed, or idempotency implemented with process memory.
- Unbounded prefetch, queue, task creation, retry count, replay rate, or drain.
- Ack before durable effect, or cancellation swallowed and message acknowledged.
- Validation/permanent failures retried forever; transient failures dead-lettered immediately.
- Database commit plus direct broker publish treated as atomic without an outbox.
- Inbox/outbox state with no uniqueness, retention, claim recovery, or real-storage test.
- Global ordering promised, or per-key sequencing destroyed by fan-out.
- Automatic DLQ replay loop, raw exception/payload copied to DLQ, or no operator owner.
- Message ID, tenant, subject, or ordering key used as a metric label.

## Verification And Proof

```bash
uv run pytest -k "event or message or outbox or inbox"
uv run pytest -m integration
make verify
```

Prove exact payload/envelope shapes, unknown-field/version behavior, duplicate and concurrent delivery, crash-after-effect-before-ack, bounded prefetch, handler timeout/cancellation, retry classification and exhaustion, settlement failure, outbox crash windows, inbox retention, per-key ordering, replay/out-of-order behavior, DLQ evidence/redaction, drain, and telemetry. Run real broker/storage integration tests for every broker-specific claim.

Related: [database](database.md), [serialization](../foundations/serialization.md), [testing](../quality/testing.md), and [resilience](../operations/resilience.md).
