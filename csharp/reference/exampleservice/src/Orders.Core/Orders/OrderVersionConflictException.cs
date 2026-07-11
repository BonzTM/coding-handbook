namespace Orders.Core.Orders;

/// <summary>
/// Optimistic-concurrency conflict: the order changed since the caller read it.
/// Raised by the repository when the concurrency token (xmin) no longer
/// matches; the caller must reload, re-decide, and resubmit - retrying blindly
/// is not handling (csharp/services/database.md). Transport maps it to 409.
/// </summary>
public sealed class OrderVersionConflictException(OrderId orderId)
    : Exception($"Order {orderId} was modified concurrently; reload and retry with the current version.")
{
    public OrderId OrderId { get; } = orderId;
}
