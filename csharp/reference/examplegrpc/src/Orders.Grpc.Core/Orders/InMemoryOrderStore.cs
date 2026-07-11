using Orders.Grpc.Core.Identity;

namespace Orders.Grpc.Core.Orders;

/// <summary>
/// The module's production store - deliberately in-memory, mirroring
/// golang/reference/examplegrpc: this reference proves the TRANSPORT patterns
/// (interceptors, error details, deadlines, TLS gating), and a real database
/// would only duplicate what the keystone exampleservice module already
/// proves against PostgreSQL. It lives in Core (not a one-class
/// Infrastructure project) because it depends on nothing app-level.
///
/// Race-safe by construction: all access to the map happens under one lock
/// (the store is a process-wide singleton). Keyset pagination has the same
/// observable contract as the keystone's SQL row-value comparison, so the
/// transport tests prove behavior a database-backed port must also satisfy.
/// </summary>
public sealed class InMemoryOrderStore : IOrderStore
{
    private readonly Lock _mutex = new();
    private readonly Dictionary<OrderId, Order> _orders = [];

    public Task AddAsync(Order order, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(order);
        cancellationToken.ThrowIfCancellationRequested();

        lock (_mutex)
        {
            bool duplicate = _orders.Values.Any(existing =>
                existing.TenantId == order.TenantId
                && string.Equals(existing.ExternalReference, order.ExternalReference, StringComparison.Ordinal));
            if (duplicate)
            {
                throw new DuplicateOrderException(order.ExternalReference);
            }

            _orders.Add(order.Id, order);
        }

        return Task.CompletedTask;
    }

    public Task<Order?> GetAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken)
    {
        cancellationToken.ThrowIfCancellationRequested();

        lock (_mutex)
        {
            // Tenant scoping in the storage layer: an order under another
            // tenant is reported as absent, indistinguishable from missing.
            return Task.FromResult(
                _orders.TryGetValue(id, out var order) && order.TenantId == tenantId ? order : null);
        }
    }

    public Task<OrderPage> ListAsync(TenantId tenantId, OrderListQuery query, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(query);
        cancellationToken.ThrowIfCancellationRequested();

        lock (_mutex)
        {
            // Fetch one extra row: if it exists there is at least one more page,
            // and the trimmed last row becomes the next cursor. Same
            // more-pages detection as the keystone's SQL repository.
            var page = _orders.Values
                .Where(order => order.TenantId == tenantId)
                .Where(order => query.Cursor is not { } cursor
                    || order.CreatedAt > cursor.CreatedAt
                    || (order.CreatedAt == cursor.CreatedAt && order.Id.Value.CompareTo(cursor.Id) > 0))
                .OrderBy(order => order.CreatedAt)
                .ThenBy(order => order.Id.Value)
                .Take(query.PageSize + 1)
                .ToList();

            string? nextCursor = null;
            if (page.Count > query.PageSize)
            {
                page.RemoveAt(page.Count - 1);
                var last = page[^1];
                nextCursor = new OrderCursor(last.CreatedAt, last.Id.Value).Encode();
            }

            return Task.FromResult(new OrderPage(page, nextCursor));
        }
    }
}
