# Recipe: Add Event Publisher

Use this when a feature needs to emit messages or events to another system or internal stream.

## Files To Touch

- the producing core service or use case
- the eventing or broker adapter package
- `api/` or contract docs if the event is published outside one package
- outbox storage and relay code if DB state and publish success must stay coordinated
- producer tests and integration tests

## Steps

1. Define the event contract, envelope, and stable event name before writing producer code.
2. Decide whether publishing is best-effort, request-coupled, or requires a transactional outbox.
3. Build the event payload from domain state, not from transport or database row leakage.
4. Add correlation metadata and a durable event ID.
5. Decide retry policy, ordering key, and terminal failure handling.

## Invariants To Preserve

- published payloads have one source of truth
- event IDs are stable enough for downstream dedupe and tracing
- producer retries do not silently duplicate non-idempotent side effects
- outbox is used when DB state and publish success must be atomic from the business perspective

## Proof

- contract tests for payload shape and metadata
- integration test for the publish path or outbox relay
- negative test for broker failure or relay lag behavior
- observability review showing publish success, failure, and retry signals
