using System.Text.Json;

using Orders.Worker.Core.Events;

namespace Orders.Worker.UnitTests.Fakes;

/// <summary>Builders for well-formed and malformed deliveries so tests state
/// only what they care about.</summary>
internal static class TestEvents
{
    public static readonly DateTimeOffset Start = new(2026, 7, 1, 12, 0, 0, TimeSpan.Zero);

    public static OrderEvent Placed(
        string orderId = "order-1", string tenantId = "tenant-a", string reference = "ref-1")
        => new()
        {
            Type = OrderEventTypes.Placed,
            OrderId = orderId,
            TenantId = tenantId,
            ExternalReference = reference,
            OccurredAt = Start,
        };

    public static OrderEvent Cancelled(string orderId = "order-1", string tenantId = "tenant-a")
        => new()
        {
            Type = OrderEventTypes.Cancelled,
            OrderId = orderId,
            TenantId = tenantId,
            OccurredAt = Start,
        };

    public static EventEnvelope Envelope(OrderEvent orderEvent, Guid? id = null)
        => new(
            Id: id ?? Guid.NewGuid(),
            Type: orderEvent.Type,
            Source: "tests",
            Time: orderEvent.OccurredAt,
            Subject: orderEvent.OrderId,
            DataContentType: "application/json",
            Data: JsonSerializer.SerializeToUtf8Bytes(orderEvent, OrderEventJsonContext.Default.OrderEvent));

    /// <summary>A syntactically broken payload: decoding it can never succeed,
    /// so the pipeline must dead-letter without retrying.</summary>
    public static EventEnvelope Malformed(Guid? id = null)
        => new(
            Id: id ?? Guid.NewGuid(),
            Type: OrderEventTypes.Placed,
            Source: "tests",
            Time: Start,
            Subject: null,
            DataContentType: "application/json",
            Data: "{not json"u8.ToArray());
}
