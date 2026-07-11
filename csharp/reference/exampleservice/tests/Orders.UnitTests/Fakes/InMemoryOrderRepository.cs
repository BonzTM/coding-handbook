using Orders.Core.Orders;

namespace Orders.UnitTests.Fakes;

/// <summary>
/// Hand-rolled in-memory implementation of the Core persistence port
/// (csharp/quality/testing.md: fakes before mocks). It honors the SAME
/// contract as the PostgreSQL repository - tenant scoping on every operation,
/// duplicate external references rejected, keyset pagination over
/// (CreatedAt, Id) - so offline tests prove the same behavior the integration
/// suite proves against the real database.
///
/// Honesty note: the real database bumps the xmin-backed Version on every
/// write; this fake cannot (the setter is private, by design), so stored
/// versions stay at their created value. Version-bump fidelity is proven by
/// the Testcontainers suite.
/// </summary>
public sealed class InMemoryOrderRepository : IOrderRepository
{
    private readonly Lock _gate = new();
    private readonly Dictionary<(string Tenant, Guid Id), Order> _orders = [];

    public Task AddAsync(Order order, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(order);
        lock (_gate)
        {
            bool duplicateReference = _orders.Values.Any(existing =>
                existing.TenantId == order.TenantId
                && string.Equals(existing.ExternalReference, order.ExternalReference, StringComparison.Ordinal));
            if (duplicateReference)
            {
                throw new DuplicateOrderException(order.ExternalReference);
            }

            _orders.Add((order.TenantId.Value, order.Id.Value), order);
        }

        return Task.CompletedTask;
    }

    public Task<Order?> GetAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken)
    {
        lock (_gate)
        {
            _orders.TryGetValue((tenantId.Value, id.Value), out Order? order);
            return Task.FromResult(order);
        }
    }

    public Task<OrderPage> ListAsync(TenantId tenantId, OrderListQuery query, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(query);
        lock (_gate)
        {
            IEnumerable<Order> rows = _orders.Values
                .Where(order => order.TenantId == tenantId)
                .OrderBy(order => order.CreatedAt)
                .ThenBy(order => order.Id.Value);

            if (query.Cursor is { } cursor)
            {
                rows = rows.Where(order =>
                    order.CreatedAt > cursor.CreatedAt
                    || (order.CreatedAt == cursor.CreatedAt && CompareIds(order.Id.Value, cursor.Id) > 0));
            }

            // Same fetch-one-extra trick as the SQL repository: learn whether a
            // next page exists without a COUNT.
            var items = rows.Take(query.PageSize + 1).ToList();
            if (items.Count <= query.PageSize)
            {
                return Task.FromResult(new OrderPage(items, NextCursor: null));
            }

            items.RemoveAt(items.Count - 1);
            Order last = items[^1];
            return Task.FromResult(
                new OrderPage(items, new OrderCursor(last.CreatedAt, last.Id.Value).Encode()));
        }
    }

    public Task<Order> UpdateAsync(Order order, uint expectedVersion, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(order);
        lock (_gate)
        {
            if (!_orders.TryGetValue((order.TenantId.Value, order.Id.Value), out Order? stored)
                || stored.Version != expectedVersion)
            {
                throw new OrderVersionConflictException(order.Id);
            }

            _orders[(order.TenantId.Value, order.Id.Value)] = order;
            return Task.FromResult(order);
        }
    }

    public Task<bool> DeleteAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken)
    {
        lock (_gate)
        {
            return Task.FromResult(_orders.Remove((tenantId.Value, id.Value)));
        }
    }

    /// <summary>PostgreSQL uuid ordering equals .NET's Guid comparison for the v7 ids this service generates.</summary>
    private static int CompareIds(Guid left, Guid right) => left.CompareTo(right);
}
