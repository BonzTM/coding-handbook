namespace Orders.Core.Orders;

/// <summary>
/// Domain service for the orders resource. Owns business rules and delegates
/// persistence to the <see cref="IOrderRepository"/> port. Takes time as an
/// injected dependency (csharp/foundations/time.md) and never logs - the hosts
/// decide what is log-worthy (csharp/foundations/errors-and-logging.md).
/// </summary>
public sealed class OrderService(IOrderRepository repository, TimeProvider time)
{
    public async Task<Order> CreateAsync(
        TenantId tenantId,
        string externalReference,
        string customerId,
        int quantity,
        CancellationToken cancellationToken)
    {
        var order = Order.Create(tenantId, externalReference, customerId, quantity, time.GetUtcNow());
        await repository.AddAsync(order, cancellationToken).ConfigureAwait(false);
        return order;
    }

    /// <exception cref="OrderNotFoundException">No such order for this tenant.</exception>
    public async Task<Order> GetAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken)
    {
        var order = await repository.GetAsync(tenantId, id, cancellationToken).ConfigureAwait(false);
        return order ?? throw new OrderNotFoundException(id);
    }

    public Task<OrderPage> ListAsync(TenantId tenantId, OrderListQuery query, CancellationToken cancellationToken)
        => repository.ListAsync(tenantId, query, cancellationToken);

    /// <exception cref="OrderNotFoundException">No such order for this tenant.</exception>
    /// <exception cref="OrderVersionConflictException">The order changed since it was read.</exception>
    public async Task<Order> AmendAsync(
        TenantId tenantId,
        OrderId id,
        uint expectedVersion,
        int quantity,
        OrderStatus status,
        CancellationToken cancellationToken)
    {
        var order = await GetAsync(tenantId, id, cancellationToken).ConfigureAwait(false);
        if (order.Version != expectedVersion)
        {
            // Cheap first check; the database's concurrency token remains the
            // authoritative gate inside UpdateAsync.
            throw new OrderVersionConflictException(id);
        }

        order.Amend(quantity, status, time.GetUtcNow());
        return await repository.UpdateAsync(order, expectedVersion, cancellationToken).ConfigureAwait(false);
    }

    /// <exception cref="OrderNotFoundException">No such order for this tenant.</exception>
    public async Task DeleteAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken)
    {
        bool deleted = await repository.DeleteAsync(tenantId, id, cancellationToken).ConfigureAwait(false);
        if (!deleted)
        {
            throw new OrderNotFoundException(id);
        }
    }
}
