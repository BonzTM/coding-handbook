# Recipe: Add Event Consumer

Use this when the repo needs to consume queue or stream messages and turn them into durable behavior.

## Files To Touch

- the eventing or broker adapter package
- the owning core service invoked by the consumer
- inbox or dedupe storage if duplicate delivery matters
- telemetry and readiness wiring if the consumer is operationally significant
- consumer tests and integration tests

## Steps

1. Define the accepted contract and compatibility policy before writing handler logic.
2. Decode, validate, and classify failures into retryable versus terminal.
3. Make the handler idempotent before enabling retries.
4. Choose the ack or settlement point only after durable side effects complete.
5. Document ordering guarantees, retry limits, and dead-letter behavior.

## Invariants To Preserve

- duplicate delivery is safe
- schema or validation failures do not retry forever
- per-key ordering is preserved when the contract requires it
- DLQ or parked-message flows preserve enough metadata for replay and diagnosis

## Proof

- duplicate-delivery and replay tests
- integration test against a real broker path or realistic emulator when semantics matter
- failure-path test proving retry exhaustion and DLQ or operator-visible handling
- telemetry review for receive, retry, success, and dead-letter transitions
