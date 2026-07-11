# Recipe: Add Event Publisher

Use this when a feature needs to emit messages or events to another system or internal stream.

## Files To Touch

- the producing Core service or use case, plus the `IEventPublisher` port (seam interface) in `src/Orders.Core`
- the broker adapter and outbox relay under `src/Orders.Infrastructure/Messaging/`
- `api/events/` — the payload contract (envelope shape, stable event name, versioned schema) when the event leaves the service
- outbox storage in `src/Orders.Infrastructure/Data` (outbox table via [add-migration.md](add-migration.md)) if DB state and publish success must stay coordinated
- producer tests under `tests/Orders.UnitTests` and outbox/relay tests under `tests/Orders.IntegrationTests`

## Steps

1. Define the event contract in `api/events/` before writing producer code: a stable, versioned event name (`orders.order-created.v1`), the envelope (event ID, type, `occurredAt` as `DateTimeOffset`, correlation ID), and the payload schema. Serialization at this boundary goes through a `JsonSerializerContext` per [../foundations/serialization.md](../foundations/serialization.md).
2. Decide whether publishing is best-effort, request-coupled, or requires a transactional outbox. Default to the outbox whenever the business requires that a committed DB state change and its event agree.
3. For the outbox path, write the outbox row in the same transaction as the domain change: the use case adds an `OutboxMessage` entity through the same `DbContext` unit of work, so one `SaveChangesAsync` commits both or neither. Never publish to the broker inside the request and hope it aligns with the commit.
4. Build the event payload from domain state in Core, not from EF entities or transport DTO leakage. The producing code depends only on the `IEventPublisher` port; the broker adapter lives in Infrastructure.
5. Add correlation metadata (from `Activity.Current`) and a durable, time-ordered event ID (`Guid.CreateVersion7()`).
6. Implement the relay as a `BackgroundService` ([add-background-worker.md](add-background-worker.md)): drain pending outbox rows → publish through the seam → mark sent. A crash between commit and publish recovers on the next scan — delivery is at-least-once, so downstream dedupe relies on the stable event ID. Decide retry policy, ordering key, and terminal failure handling before shipping.

## Invariants To Preserve

- published payloads have one source of truth — the contract in `api/events/`, versioned, additive-only within a version
- event IDs are stable enough for downstream dedupe and tracing
- producer retries do not silently duplicate non-idempotent side effects
- the outbox is used when DB state and publish success must be atomic from the business perspective; the outbox write and the domain write share one transaction
- Core references no broker SDK; only the Infrastructure adapter does

## Proof

- contract tests for payload shape and metadata (serialize through the `JsonSerializerContext`, assert the wire shape)
- integration test for the publish path or outbox relay, including crash-between-commit-and-publish recovery, behind the `-Integration` switch
- negative test for broker failure or relay lag behavior
- observability review showing publish success, failure, and retry signals
- run `pwsh ./verify.ps1`

Governing doc: [eventing-and-messaging.md](../services/eventing-and-messaging.md). The consuming side is [add-event-consumer.md](add-event-consumer.md); the outbox table ships via [add-migration.md](add-migration.md).
