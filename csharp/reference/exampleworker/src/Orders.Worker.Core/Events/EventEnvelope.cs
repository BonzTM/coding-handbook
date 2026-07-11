namespace Orders.Worker.Core.Events;

/// <summary>
/// CloudEvents-style metadata envelope carried by every message, exactly as
/// specified in csharp/services/eventing-and-messaging.md (Contract Source And
/// Envelope). The payload is opaque bytes here; the consumer decodes it into a
/// domain type (<see cref="OrderEvent"/>) before any business logic runs.
/// </summary>
/// <param name="Id">Stable event identifier: the inbox dedupe store is keyed by
/// it, so an at-least-once duplicate delivery is processed exactly once.</param>
/// <param name="Type">Event name ("orders.order-placed.v1"), low-cardinality -
/// used for routing, metrics, and DLQ classification. Never a table name.</param>
/// <param name="Source">Producing service or subsystem.</param>
/// <param name="Time">Producer timestamp, stamped from <see cref="TimeProvider"/>.</param>
/// <param name="Subject">Optional entity identifier - the ordering/partition
/// key (the order id for order events).</param>
/// <param name="DataContentType">Payload encoding, e.g. "application/json".</param>
/// <param name="Data">Serialized payload.</param>
public sealed record EventEnvelope(
    Guid Id,
    string Type,
    string Source,
    DateTimeOffset Time,
    string? Subject,
    string DataContentType,
    ReadOnlyMemory<byte> Data);
