namespace Orders.Core.Orders;

/// <summary>
/// Validated list-query parameters. Page size is clamped server-side; an
/// unbounded list is never honored (csharp/services/http-services.md).
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
