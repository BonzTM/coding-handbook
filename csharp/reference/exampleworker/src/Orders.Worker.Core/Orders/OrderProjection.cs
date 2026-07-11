namespace Orders.Worker.Core.Orders;

/// <summary>
/// The projected state of one order, maintained from the event stream by
/// <see cref="OrderProjector"/>. A cancelled order keeps its row as a
/// tombstone (<see cref="Cancelled"/> true) so a late or replayed event on the
/// same key still has something to land on.
/// </summary>
public sealed record OrderProjection
{
    public required string OrderId { get; init; }

    public required string TenantId { get; init; }

    public string? ExternalReference { get; init; }

    public bool Cancelled { get; init; }

    /// <summary>Stamped from the injected <see cref="TimeProvider"/>, never
    /// the wall clock (csharp/foundations/time.md).</summary>
    public DateTimeOffset UpdatedAt { get; init; }
}
