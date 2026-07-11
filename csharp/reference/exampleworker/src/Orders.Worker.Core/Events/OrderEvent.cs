namespace Orders.Worker.Core.Events;

/// <summary>The closed set of event types this worker consumes and relays.
/// Names follow the versioned "orders.order-placed.v1" convention from
/// csharp/services/eventing-and-messaging.md - a breaking payload change means
/// a ".v2" type, never a silent field reinterpretation.</summary>
public static class OrderEventTypes
{
    public const string Placed = "orders.order-placed.v1";
    public const string Cancelled = "orders.order-cancelled.v1";
}

/// <summary>
/// The decoded domain payload of an order message - a plain value type with no
/// wire or broker concerns. The messaging layer deserializes the envelope's
/// <see cref="EventEnvelope.Data"/> into this before calling the processor.
/// Unknown JSON fields are tolerated on read (additive schema evolution);
/// missing required fields fail deserialization, which the consumer classifies
/// as non-retryable.
/// </summary>
public sealed record OrderEvent
{
    /// <summary>Event name; one of <see cref="OrderEventTypes"/>.</summary>
    public required string Type { get; init; }

    /// <summary>The order (aggregate) identifier - the ordering and dedupe key.</summary>
    public required string OrderId { get; init; }

    /// <summary>Tenant scope for the order.</summary>
    public required string TenantId { get; init; }

    /// <summary>Customer-facing reference carried by a placed event.</summary>
    public string? ExternalReference { get; init; }

    /// <summary>Producer timestamp.</summary>
    public DateTimeOffset OccurredAt { get; init; }

    /// <summary>
    /// Enforces the payload invariants. A violation is a non-retryable
    /// <see cref="InvalidEventException"/>: replaying a structurally invalid
    /// event will never succeed, so the consumer dead-letters it immediately
    /// instead of burning the retry budget.
    /// </summary>
    /// <exception cref="InvalidEventException">The payload violates an invariant.</exception>
    public void Validate()
    {
        if (string.IsNullOrWhiteSpace(OrderId))
        {
            throw new InvalidEventException("order id must not be empty");
        }

        if (string.IsNullOrWhiteSpace(TenantId))
        {
            throw new InvalidEventException("tenant id must not be empty");
        }

        switch (Type)
        {
            case OrderEventTypes.Placed:
                if (string.IsNullOrWhiteSpace(ExternalReference))
                {
                    throw new InvalidEventException("placed event must carry an external reference");
                }

                break;
            case OrderEventTypes.Cancelled:
                // No extra fields required.
                break;
            default:
                throw new InvalidEventException($"unknown event type '{Type}'");
        }
    }
}
