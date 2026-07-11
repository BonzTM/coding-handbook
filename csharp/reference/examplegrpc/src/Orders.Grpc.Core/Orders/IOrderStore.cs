using Orders.Grpc.Core.Identity;

namespace Orders.Grpc.Core.Orders;

/// <summary>
/// Persistence port for orders, defined where it is consumed (Core) and named
/// for what the service needs, not what an implementer is - the C# twin of the
/// Go reference's consumer-side Store interface. Every operation is
/// tenant-scoped: one tenant can never observe another's rows. This module
/// satisfies it with <see cref="InMemoryOrderStore"/>; a database-backed
/// service implements the same port in an Infrastructure project (see the
/// keystone module's IOrderRepository).
/// </summary>
public interface IOrderStore
{
    /// <exception cref="DuplicateOrderException">
    /// The tenant already has an order with the same external reference
    /// (uniqueness enforced by the store, never by a racy pre-check read).
    /// </exception>
    Task AddAsync(Order order, CancellationToken cancellationToken);

    Task<Order?> GetAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken);

    /// <summary>Keyset-paginated list ordered by (CreatedAt, Id) ascending.</summary>
    Task<OrderPage> ListAsync(TenantId tenantId, OrderListQuery query, CancellationToken cancellationToken);
}
