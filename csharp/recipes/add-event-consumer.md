# Recipe: Add Event Consumer

Use this when the repo needs to consume queue or stream messages and turn them into durable behavior.

## Files To Touch

- the consumer loop as a `BackgroundService` plus the broker adapter under `src/Orders.Infrastructure/Messaging/`
- the owning Core handler the consumer invokes (behind the repo-owned consumer seam interface in `src/Orders.Core`)
- inbox/dedupe storage in `src/Orders.Infrastructure/Data` (processed-message table via [add-migration.md](add-migration.md)) if duplicate delivery matters — with at-least-once delivery it always does
- telemetry and readiness wiring if the consumer is operationally significant
- consumer tests under `tests/Orders.UnitTests` and broker tests under `tests/Orders.IntegrationTests`

## Steps

1. Define the accepted contract and compatibility policy before writing handler logic: which event name/version this consumer accepts, and what it does with unknown versions. Decode through a `JsonSerializerContext` with the explicit unknown-field policy from [../foundations/serialization.md](../foundations/serialization.md).
2. Decode, validate, and classify failures into retryable (broker down, DB timeout) versus terminal (schema violation, business-rule rejection). Terminal failures go straight to the DLQ — they must never retry forever.
3. Make the handler idempotent before enabling retries: record the event ID in the inbox table (unique constraint) in the same transaction as the handler's side effects, and skip cleanly when the insert conflicts. Duplicate delivery then converges instead of double-applying.
4. Choose the ack/settlement point only after durable side effects complete: process → commit (side effects + inbox row) → ack. Ack-then-process loses messages on crash; the inbox makes the resulting redelivery safe.
5. Honor `stoppingToken` throughout the loop (see [add-background-worker.md](add-background-worker.md)) so shutdown never abandons an in-flight message mid-settlement.
6. Document ordering guarantees, retry limits, backoff, and dead-letter behavior. DLQ'd messages must carry enough metadata (event ID, error, attempt count, original payload) for diagnosis and replay.

## Invariants To Preserve

- duplicate delivery is safe: the inbox dedupe and the handler's side effects commit atomically
- schema or validation failures do not retry forever; retryable and terminal failures take different paths
- per-key ordering is preserved when the contract requires it
- DLQ or parked-message flows preserve enough metadata for replay and diagnosis, and replay from the DLQ re-enters the idempotent path
- Core references no broker SDK; only the Infrastructure adapter does

## Proof

- duplicate-delivery and replay tests: deliver the same event twice (and replay a DLQ'd message), assert side effects applied exactly once
- integration test against a real broker path (Testcontainers) or realistic emulator when semantics matter, behind the `-Integration` switch
- failure-path test proving retry exhaustion lands the message in the DLQ with its diagnostic metadata intact
- telemetry review for receive, retry, success, and dead-letter transitions
- run `pwsh ./verify.ps1`

Governing doc: [eventing-and-messaging.md](../services/eventing-and-messaging.md). The producing side is [add-event-publisher.md](add-event-publisher.md); the inbox table ships via [add-migration.md](add-migration.md).
