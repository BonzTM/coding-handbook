namespace Orders.Grpc.Core.Orders;

/// <summary>
/// One page of a keyset-paginated list. <see cref="NextCursor"/> is null on
/// the last page; otherwise the client passes it back verbatim (as
/// page_token) to fetch the next page.
/// </summary>
public sealed record OrderPage(IReadOnlyList<Order> Items, string? NextCursor);
