using Orders.Core.Orders;

namespace Orders.Api.Contracts;

/// <summary>
/// Explicit domain-to-wire mapping - the one place the two shapes meet. Plain
/// code, no reflection mappers (csharp/foundations/serialization.md).
/// </summary>
internal static class OrderMappings
{
    public static OrderResponse ToResponse(this Order order) => new()
    {
        OrderId = order.Id.ToString(),
        ExternalReference = order.ExternalReference,
        CustomerId = order.CustomerId,
        Quantity = order.Quantity,
        Status = order.Status,
        Version = order.Version,
        CreatedAt = order.CreatedAt,
        UpdatedAt = order.UpdatedAt,
    };

    public static OrderListResponse ToResponse(this OrderPage page) => new()
    {
        Items = [.. page.Items.Select(ToResponse)],
        NextCursor = page.NextCursor,
    };
}
