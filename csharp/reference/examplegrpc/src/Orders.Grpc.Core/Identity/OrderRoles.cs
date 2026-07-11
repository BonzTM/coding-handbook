namespace Orders.Grpc.Core.Identity;

/// <summary>
/// The closed set of roles the domain understands. Reads require
/// <see cref="Reader"/>, mutations require <see cref="Writer"/> - checked in
/// <c>OrderService</c>, so authorization holds no matter which transport
/// fronts the domain (mirrors the keystone module's orders.reader/orders.writer
/// role model, csharp/operations/security.md).
/// </summary>
public static class OrderRoles
{
    public const string Reader = "orders.reader";
    public const string Writer = "orders.writer";
}
