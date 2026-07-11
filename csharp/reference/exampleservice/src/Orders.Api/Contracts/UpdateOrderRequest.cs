using System.ComponentModel.DataAnnotations;

using Orders.Core.Orders;

namespace Orders.Api.Contracts;

/// <summary>
/// Wire DTO for PUT /orders/{id}. Carries the concurrency token the caller
/// read; a stale token is a 409 conflict, never a silent lost update
/// (csharp/services/database.md, Concurrency Tokens).
/// </summary>
public sealed record UpdateOrderRequest
{
    [Range(1, Order.MaxQuantity)]
    public required int Quantity { get; init; }

    [DeniedValues(OrderStatus.Unknown)]
    public required OrderStatus Status { get; init; }

    /// <summary>The version from the last read of this order.</summary>
    public required uint Version { get; init; }
}
