namespace Orders.Grpc.Core.Orders;

/// <summary>
/// The order does not exist for the caller's tenant. A cross-tenant read maps
/// here too, so another tenant's data is indistinguishable from "not found"
/// (csharp/operations/security.md). Transport maps it to NOT_FOUND.
/// </summary>
public sealed class OrderNotFoundException(OrderId orderId)
    : Exception($"Order {orderId} was not found.")
{
    public OrderId OrderId { get; } = orderId;
}
