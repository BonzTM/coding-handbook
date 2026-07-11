using Orders.Worker.Core.Events;

namespace Orders.Worker.Core.Orders;

/// <summary>
/// The reference in-memory <see cref="IOrderEventProcessor"/>: it maintains a
/// per-(tenant, order) projection of the latest order state from the event
/// stream. Applying the same event twice converges to the same state, and the
/// consumer's inbox additionally guarantees the second application never
/// happens. Production swaps in a database-backed projector behind the same
/// port with one registration change in the composition root.
///
/// Thread-safe: the consume loop is sequential, but tests and operator tooling
/// may read concurrently.
/// </summary>
public sealed class OrderProjector(TimeProvider time) : IOrderEventProcessor
{
    private readonly Lock _gate = new();
    private readonly Dictionary<(string TenantId, string OrderId), OrderProjection> _orders = [];

    /// <summary>Validates then applies one event to the projection. Validation
    /// failures throw <see cref="InvalidEventException"/> (non-retryable);
    /// timestamps come from the injected clock.</summary>
    public Task ProcessAsync(OrderEvent orderEvent, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(orderEvent);
        cancellationToken.ThrowIfCancellationRequested();
        orderEvent.Validate();

        lock (_gate)
        {
            Apply(orderEvent);
        }

        return Task.CompletedTask;
    }

    /// <summary>Returns the projected order, or null when the key was never
    /// seen. Exists so tests can assert exactly-once application.</summary>
    public OrderProjection? GetProjection(string tenantId, string orderId)
    {
        lock (_gate)
        {
            return _orders.TryGetValue((tenantId, orderId), out var projection) ? projection : null;
        }
    }

    private void Apply(OrderEvent orderEvent)
    {
        var key = (orderEvent.TenantId, orderEvent.OrderId);
        var now = time.GetUtcNow();
        _orders[key] = orderEvent.Type switch
        {
            OrderEventTypes.Placed => new OrderProjection
            {
                OrderId = orderEvent.OrderId,
                TenantId = orderEvent.TenantId,
                ExternalReference = orderEvent.ExternalReference,
                Cancelled = ExistingCancelled(key),
                UpdatedAt = now,
            },
            OrderEventTypes.Cancelled => (ExistingOrTombstone(orderEvent)) with
            {
                Cancelled = true,
                UpdatedAt = now,
            },
            // Validate() already rejected unknown types; defensive backstop.
            _ => throw new InvalidEventException($"unknown event type '{orderEvent.Type}'"),
        };
    }

    /// <summary>A cancel that raced ahead of its place event must not be
    /// un-cancelled by the late place - per-key ordering is the contract, but
    /// redelivery can still interleave (csharp/services/eventing-and-messaging.md).</summary>
    private bool ExistingCancelled((string TenantId, string OrderId) key)
        => _orders.TryGetValue(key, out var existing) && existing.Cancelled;

    private OrderProjection ExistingOrTombstone(OrderEvent orderEvent)
    {
        if (_orders.TryGetValue((orderEvent.TenantId, orderEvent.OrderId), out var existing))
        {
            return existing;
        }

        return new OrderProjection
        {
            OrderId = orderEvent.OrderId,
            TenantId = orderEvent.TenantId,
            ExternalReference = orderEvent.ExternalReference,
        };
    }
}
