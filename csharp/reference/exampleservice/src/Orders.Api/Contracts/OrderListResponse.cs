namespace Orders.Api.Contracts;

/// <summary>
/// The one list envelope every collection endpoint uses: items plus pagination
/// metadata. <c>nextCursor</c> is null on the last page; clients pass it back
/// verbatim (csharp/services/http-services.md, List Endpoints And Pagination).
/// </summary>
internal sealed record OrderListResponse
{
    public required IReadOnlyList<OrderResponse> Items { get; init; }

    public required string? NextCursor { get; init; }
}
