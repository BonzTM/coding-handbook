namespace Orders.Grpc.Core.Orders;

/// <summary>
/// Typed identifier for an order. A raw <see cref="Guid"/> would let an
/// <c>OrderId</c> silently accept a customer id; the wrapper makes the
/// compiler enforce the difference (csharp/foundations/data-modeling.md).
/// </summary>
public readonly record struct OrderId(Guid Value)
{
    /// <summary>Version-7 GUIDs sort by creation time, keeping keyset pages stable.</summary>
    public static OrderId New() => new(Guid.CreateVersion7());

    public static bool TryParse(string? candidate, out OrderId id)
    {
        if (Guid.TryParse(candidate, out var value))
        {
            id = new OrderId(value);
            return true;
        }

        id = default;
        return false;
    }

    public override string ToString() => Value.ToString("D");
}
