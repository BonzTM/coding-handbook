namespace Orders.Grpc.Core.Orders;

/// <summary>
/// Validated list-query parameters. Page size is clamped server-side; an
/// unbounded list is never honored. A non-positive page_size falls back to the
/// default rather than being rejected (csharp/services/grpc-services.md keeps
/// the same policy as the HTTP module's list endpoint).
/// </summary>
public sealed record OrderListQuery
{
    public const int DefaultPageSize = 20;
    public const int MaxPageSize = 100;

    public OrderListQuery(int? requestedPageSize, OrderCursor? cursor)
    {
        PageSize = Clamp(requestedPageSize);
        Cursor = cursor;
    }

    public int PageSize { get; }

    public OrderCursor? Cursor { get; }

    private static int Clamp(int? requested) => requested switch
    {
        null => DefaultPageSize,
        < 1 => DefaultPageSize,
        > MaxPageSize => MaxPageSize,
        _ => requested.Value,
    };
}
