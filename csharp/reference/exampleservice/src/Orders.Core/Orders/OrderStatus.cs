namespace Orders.Core.Orders;

/// <summary>
/// Order lifecycle states. Values are pinned and <c>Unknown = 0</c> is reserved
/// so a missing or defaulted value is distinguishable from a real state
/// (csharp/foundations/data-modeling.md). On the wire the names serialize as
/// strings, never numbers (csharp/foundations/serialization.md).
/// </summary>
public enum OrderStatus
{
    Unknown = 0,
    Pending = 1,
    Confirmed = 2,
    Shipped = 3,
    Cancelled = 4,
}
