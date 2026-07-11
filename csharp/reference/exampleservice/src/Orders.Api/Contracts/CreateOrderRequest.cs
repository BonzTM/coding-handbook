using System.ComponentModel.DataAnnotations;

using Orders.Core.Orders;

namespace Orders.Api.Contracts;

/// <summary>
/// Wire DTO for POST /orders. Validation attributes live here, never on Core
/// domain types; shape/range checks run before the handler via the built-in
/// minimal-API validation (csharp/services/http-services.md).
/// </summary>
public sealed record CreateOrderRequest
{
    [Required(AllowEmptyStrings = false)]
    [StringLength(Order.MaxReferenceLength, MinimumLength = 1)]
    public required string ExternalReference { get; init; }

    [Required(AllowEmptyStrings = false)]
    [StringLength(Order.MaxCustomerIdLength, MinimumLength = 1)]
    public required string CustomerId { get; init; }

    [Range(1, Order.MaxQuantity)]
    public required int Quantity { get; init; }
}
