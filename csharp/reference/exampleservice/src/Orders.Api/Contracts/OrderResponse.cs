using Orders.Core.Orders;

namespace Orders.Api.Contracts;

/// <summary>
/// Wire shape of one order. A dedicated DTO, not the Core aggregate - a Core
/// refactor must never be a silent wire change
/// (csharp/foundations/serialization.md).
/// </summary>
internal sealed record OrderResponse
{
    public required string OrderId { get; init; }

    public required string ExternalReference { get; init; }

    public required string CustomerId { get; init; }

    public required int Quantity { get; init; }

    public required OrderStatus Status { get; init; }

    /// <summary>Concurrency token; echo it back in PUT.</summary>
    public required uint Version { get; init; }

    public required DateTimeOffset CreatedAt { get; init; }

    public required DateTimeOffset UpdatedAt { get; init; }
}
