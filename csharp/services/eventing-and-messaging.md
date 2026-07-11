# Eventing and Messaging

Event-driven systems need the same level of rigor as HTTP or database boundaries. Messages are contracts, delivery is a runtime behavior, and replay is part of the design, not an edge case.

## Default Approach

- Stay broker-neutral in application code: the seam is a pair of small repo-owned interfaces in `Orders.Core` (publisher port, message source), with the broker adapter in `Orders.Infrastructure/Messaging/`. Core and handlers never reference a broker client library.
- **No MassTransit, no Wolverine, no broker framework by default.** A framework that owns your topology, serialization, and retry semantics is an architectural decision — it requires an ADR ([../decisions/framework-selection.md](../decisions/framework-selection.md)). The seam keeps the option open without the coupling.
- When the spec is silent on broker choice, default to NATS JetStream or RabbitMQ behind the seam — both are self-hostable, both model at-least-once delivery honestly.
- Assume at-least-once delivery from day one.
- Make consumers idempotent before adding retries.
- Treat ordering as per-entity or per-key, never global.
- When a service both commits DB state and publishes messages, default to an outbox pattern.

The seam:

```csharp
// Orders.Core — ports; no broker types anywhere in the signatures
public interface IEventPublisher
{
    Task PublishAsync(EventEnvelope envelope, CancellationToken cancellationToken);
}

public interface IMessageSource
{
    IAsyncEnumerable<InboundMessage> ReadAllAsync(CancellationToken cancellationToken);
}
```

## Contract Source And Envelope

- Every published event needs one obvious source of truth: schema files versioned under `api/events/<stream>/v1/` (JSON Schema for JSON payloads, `.proto` when the org already runs protobuf events), owned by the repo like any other contract ([../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md)).
- For broadly consumed streams, prefer a documented schema source such as AsyncAPI plus the schema files under `api/`.
- Carry stable metadata in a CloudEvents-style envelope even if the repo does not adopt CloudEvents literally:

```csharp
public sealed record EventEnvelope(
    Guid Id,                    // stable event identifier: dedupe + traceability
    string Type,                // event name ("orders.order-placed.v1"), not a table name
    string Source,              // producing service or subsystem
    DateTimeOffset Time,        // producer timestamp, from TimeProvider
    string? Subject,            // optional entity identifier or logical subject
    string DataContentType,     // payload encoding, e.g. "application/json"
    ReadOnlyMemory<byte> Data); // serialized payload
```

- Payload serialization at this boundary follows [../foundations/serialization.md](../foundations/serialization.md): `System.Text.Json` with a source-generated `JsonSerializerContext`, unknown-fields policy explicit.

## Schema Evolution And Compatibility

- Prefer additive-only changes.
- Consumers must tolerate unknown fields.
- Producers must not silently change field meaning in place.
- Breaking changes deserve a new contract version or a new event type (`orders.order-placed.v2`), not a surprise field reinterpretation.
- If compatibility governance matters across teams, prefer Protobuf or Avro plus automated validation over undocumented JSON drift.

## Publisher Rules

- Publish after domain behavior is decided, not before.
- Include correlation and trace context in message metadata — propagate the current `Activity` as `traceparent` in message headers so consumer spans link to the producing request ([../operations/observability.md](../operations/observability.md)).
- Choose keys that match the ordering or dedupe contract, usually an aggregate or entity ID (`OrderId` as the subject/partition key).
- Do not emit best-effort side effects from request handlers and hope a retry somewhere else fixes it later — that is what the outbox is for.

## Consumer Rules

- A consumer should decode, validate, dedupe, call core logic, and only then acknowledge or commit offset/settlement.
- Ack after durable side effects, not after parsing.
- Consumers run as a `BackgroundService` and honor `stoppingToken` end-to-end — shutdown means finish the in-flight message within `HostOptions.ShutdownTimeout` and stop, never abandon mid-settlement:

```csharp
public sealed class OrderEventsConsumer(
    IMessageSource source,
    IServiceScopeFactory scopeFactory,
    ILogger<OrderEventsConsumer> logger) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        await foreach (var message in source.ReadAllAsync(stoppingToken))
        {
            await using var scope = scopeFactory.CreateAsyncScope();
            var handler = scope.ServiceProvider.GetRequiredService<OrderEventHandler>();
            await handler.HandleAsync(message, stoppingToken);
        }
    }
}
```

  One DI scope per message: the scoped `DbContext` and the handler live exactly as long as the unit of work ([database.md](database.md)).
- Validation or schema failures are usually non-retryable and should move toward operator visibility or DLQ policy quickly.
- Transient dependency failures can retry with bounded backoff and jitter.
- Do not swallow `OperationCanceledException` triggered by `stoppingToken` — that is orderly shutdown, not an error.

## Ordering And Concurrency

- Promise ordering only within one key or partition.
- Scaling consumers usually means more keys or partitions, not more concurrency on one ordered stream.
- If work fans out asynchronously (`Task.WhenAll`, channels), preserve per-key sequencing in the owning worker.
- One redelivery on an ordered key may require replay or delay of later events on that same key.

## Retries And Dead-Letter Behavior

- Retry only transient failures — classify first, retry second.
- Use bounded exponential backoff with full jitter. Delays go through the injected `TimeProvider` so tests advance a `FakeTimeProvider` instead of sleeping ([../foundations/time.md](../foundations/time.md)):

```csharp
var delay = Backoff.Next(attempt);                    // bounded exponential + full jitter
await Task.Delay(delay, timeProvider, stoppingToken); // TimeProvider-aware overload
```

- Stop retrying after a documented budget; then dead-letter, park, or surface the event for operator action.
- A dead-letter record retains original destination, attempt count, failure class, and correlation data — enough to diagnose and replay without the original process.
- Replays are operator-controlled, not an infinite feedback loop.
- DLQ policy per stream is written down in the runbook: what lands there, who is paged, how replay works.

## Outbox And Inbox Patterns

- If a service writes domain state and emits an event as one logical action, use a transactional outbox by default: save the domain change and an `OutboxMessage` row in the same EF Core transaction, then relay asynchronously.

```csharp
public sealed class OutboxMessage
{
    public Guid Id { get; init; }
    public required string Type { get; init; }
    // The broker subject / partition key — the aggregate ID per the publisher
    // rules above, so ordering-per-key survives the outbox hop.
    public required string Subject { get; init; }
    public required string Payload { get; init; }
    public DateTimeOffset OccurredAt { get; init; }
    public DateTimeOffset? SentAt { get; set; }
}
```

```csharp
// same unit of work: one SaveChangesAsync, one transaction, no dual write
db.Orders.Add(order);
db.Outbox.Add(OutboxMessage.From(orderPlaced, timeProvider));
await db.SaveChangesAsync(ct);
```

- A relay `BackgroundService` drains pending rows (ordered by `OccurredAt`, `SentAt IS NULL`), publishes through `IEventPublisher`, and marks rows sent. Crash between publish and mark-sent means a duplicate — which is why consumers are idempotent, not why the relay gets cleverer.
- On the consumer side, use an inbox — a durable dedupe table keyed by event `Id` — when duplicate delivery would be harmful. Insert the inbox row in the same transaction as the handler's state change; a unique violation on the event ID *is* the duplicate detection ([database.md](database.md), [../recipes/add-idempotent-write.md](../recipes/add-idempotent-write.md)).
- Outbox and inbox tables are ordinary EF Core migrations in `Orders.Infrastructure/Data/Migrations/` — see [database.md](database.md#migrations).
- Avoid dual writes to the database and broker in separate success paths without an explicit failure model.

## Suggested Layout

```text
api/events/orders/v1/            # payload schemas, versioned like any contract
src/Orders.Core/Events/          # ports (IEventPublisher, IMessageSource), envelope, domain events
src/Orders.Infrastructure/Messaging/
  NatsEventPublisher.cs          # broker adapter (or RabbitMqEventPublisher)
  OutboxRelayService.cs          # BackgroundService draining the outbox
  OrderEventsConsumer.cs         # BackgroundService consuming
src/Orders.Infrastructure/Data/
  Migrations/                    # outbox + inbox tables live with the schema
```

Keep the exact names repo-specific. The important rule is ownership: contracts in one place, delivery logic in one place, business rules in Core.

## Observability

- Log publish, receive, retry, exhaustion, and dead-letter transitions with stable fields via `[LoggerMessage]` templates ([../foundations/errors-and-logging.md](../foundations/errors-and-logging.md)).
- Trace producer send, consumer receive, handler execution, and settlement with `ActivitySource` spans; link consumer spans to the producer via the propagated `traceparent` header.
- Track at least: publish failures, consume failures, retry count, duplicate drops, handler latency, **consumer lag / backlog age**, outbox backlog (oldest unsent row age), and DLQ count. Consumer lag is the earliest signal a consumer is losing — alert on it, not on the eventual timeout storm.
- Message IDs and correlation IDs belong in logs and traces, not as high-cardinality metric labels.

## Testing And Proof

- Contract tests prove payload shape and compatibility expectations against the schemas in `api/events/`.
- Duplicate-delivery tests prove consumer idempotency: deliver the same envelope twice, assert one state change and one duplicate-drop metric.
- Replay and ordering tests prove per-key behavior.
- Retry and DLQ tests prove failure classification and bounded exhaustion — driven by `FakeTimeProvider`, no real sleeps.
- An in-memory `IEventPublisher`/`IMessageSource` fake that models at-least-once delivery (duplicates, redelivery) keeps the whole flow offline-testable; broker-semantics tests run against the real broker via Testcontainers behind the `-Integration` switch ([../quality/testing.md](../quality/testing.md)).

## Common Mistakes And Forbidden Patterns

- Publishing events directly from handlers while skipping a durable outbox where atomicity matters.
- Treating exactly-once delivery as a default assumption.
- Using global ordering as a product promise.
- Pulling in MassTransit/Wolverine (or scattering broker client types through Core) without the ADR the seam exists to make unnecessary.
- Putting event schemas only in generated code or only in tribal knowledge.
- A consumer loop that ignores `stoppingToken`, blocking shutdown until Kubernetes sends SIGKILL mid-settlement.
- Retrying validation failures forever.
- `Task.Delay` with hardcoded clocks in retry paths — untestable sleeps instead of `TimeProvider`.
- Using message IDs, raw subjects, or tenant IDs as metric labels.

## Verification And Proof

- event schemas or examples validate in CI when the repo publishes them
- producers and consumers have targeted contract, duplicate-delivery, and replay tests; run `pwsh ./verify.ps1` (restore (locked), format-check, build (warnings-as-errors), test, audit)
- outbox and inbox flows are covered by real-storage integration tests (Testcontainers, `-Integration` switch)
- a shutdown test: cancelling the host mid-stream lets the in-flight message settle and the consumer stop within the shutdown timeout
- publish/consume telemetry — including consumer lag and outbox backlog — is visible in logs, metrics, or traces before the feature is called done
