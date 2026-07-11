namespace Orders.Core.Orders;

/// <summary>
/// One page of a keyset-paginated list. <see cref="NextCursor"/> is null on the
/// last page; otherwise the client passes it back verbatim to fetch the next
/// page. The contract is identical for the PostgreSQL repository and the
/// in-memory test fake, so offline tests prove the same behavior.
/// </summary>
public sealed record OrderPage(IReadOnlyList<Order> Items, string? NextCursor);
