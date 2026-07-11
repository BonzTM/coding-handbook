namespace Orders.Core.Orders;

/// <summary>
/// Persistence port for orders, defined where it is consumed (Core), implemented
/// in Orders.Infrastructure. Every operation is tenant-scoped: one tenant can
/// never observe another's rows (csharp/services/database.md).
/// </summary>
public interface IOrderRepository
{
    /// <exception cref="DuplicateOrderException">
    /// The tenant already has an order with the same external reference
    /// (unique constraint, SQLSTATE 23505, translated in the repository).
    /// </exception>
    Task AddAsync(Order order, CancellationToken cancellationToken);

    Task<Order?> GetAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken);

    /// <summary>Keyset-paginated list ordered by (CreatedAt, Id) ascending.</summary>
    Task<OrderPage> ListAsync(TenantId tenantId, OrderListQuery query, CancellationToken cancellationToken);

    /// <exception cref="OrderVersionConflictException">
    /// The row changed (or disappeared) since it was read at
    /// <paramref name="expectedVersion"/> - the caller must reload and re-decide.
    /// </exception>
    Task<Order> UpdateAsync(Order order, uint expectedVersion, CancellationToken cancellationToken);

    /// <returns>false when no matching order exists for the tenant.</returns>
    Task<bool> DeleteAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken);
}
