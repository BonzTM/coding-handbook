namespace Orders.Core.Orders;

/// <summary>
/// Typed identifier for an order. A raw <see cref="Guid"/> key would let an
/// <c>OrderId</c> silently accept a customer id; the wrapper makes the compiler
/// enforce the difference (csharp/services/database.md, Typed IDs).
/// </summary>
public readonly record struct OrderId(Guid Value)
{
    /// <summary>Version-7 GUIDs keep PostgreSQL b-tree inserts append-friendly.</summary>
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
