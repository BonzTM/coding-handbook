using Orders.Grpc.Core.Identity;
using Orders.Grpc.Core.Orders;

namespace Orders.Grpc.UnitTests.Fakes;

/// <summary>
/// A store whose every operation throws the configured exception - the
/// deterministic way to drive the exception-mapping interceptor's shielding
/// and cancellation paths through the real transport.
/// </summary>
public sealed class ThrowingOrderStore(Func<Exception> exceptionFactory) : IOrderStore
{
    public Task AddAsync(Order order, CancellationToken cancellationToken)
        => Task.FromException(exceptionFactory());

    public Task<Order?> GetAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken)
        => Task.FromException<Order?>(exceptionFactory());

    public Task<OrderPage> ListAsync(TenantId tenantId, OrderListQuery query, CancellationToken cancellationToken)
        => Task.FromException<OrderPage>(exceptionFactory());
}
