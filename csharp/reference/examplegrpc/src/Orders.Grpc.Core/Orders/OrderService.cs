using System.Runtime.CompilerServices;

using Orders.Grpc.Core.Identity;

namespace Orders.Grpc.Core.Orders;

/// <summary>
/// Domain service for the orders resource. Owns business rules - including
/// role checks, so authorization holds no matter which transport fronts the
/// domain (the Go reference makes the same choice in core). Takes the
/// authenticated <see cref="CallerPrincipal"/> and time as explicit inputs
/// (csharp/foundations/time.md) and never logs - the host decides what is
/// log-worthy (csharp/foundations/errors-and-logging.md).
/// </summary>
public sealed class OrderService(IOrderStore store, TimeProvider time)
{
    /// <exception cref="PermissionDeniedException">The principal lacks the writer role.</exception>
    /// <exception cref="OrderValidationException">One or more inputs are invalid.</exception>
    /// <exception cref="DuplicateOrderException">The tenant already uses this external reference.</exception>
    public async Task<Order> CreateAsync(
        CallerPrincipal principal,
        string externalReference,
        string customerId,
        int quantity,
        CancellationToken cancellationToken)
    {
        Require(principal, OrderRoles.Writer);
        var order = Order.Create(
            principal.Tenant, externalReference, customerId, quantity, time.GetUtcNow());
        await store.AddAsync(order, cancellationToken).ConfigureAwait(false);
        return order;
    }

    /// <exception cref="PermissionDeniedException">The principal lacks the reader role.</exception>
    /// <exception cref="OrderNotFoundException">No such order for this tenant.</exception>
    public async Task<Order> GetAsync(CallerPrincipal principal, OrderId id, CancellationToken cancellationToken)
    {
        Require(principal, OrderRoles.Reader);
        var order = await store.GetAsync(principal.Tenant, id, cancellationToken).ConfigureAwait(false);
        return order ?? throw new OrderNotFoundException(id);
    }

    /// <exception cref="PermissionDeniedException">The principal lacks the reader role.</exception>
    public Task<OrderPage> ListAsync(CallerPrincipal principal, OrderListQuery query, CancellationToken cancellationToken)
    {
        Require(principal, OrderRoles.Reader);
        return store.ListAsync(principal.Tenant, query, cancellationToken);
    }

    /// <summary>
    /// Streams every order in the principal's tenant in (CreatedAt, Id) order
    /// by paging the store - no unbounded snapshot is materialized. The loop
    /// is bounded: each iteration advances the keyset cursor, so it runs at
    /// most ceil(rows / MaxPageSize) times over the data that exists.
    /// Cancellation (client cancel or deadline) stops it between pages.
    /// </summary>
    /// <exception cref="PermissionDeniedException">The principal lacks the reader role.</exception>
    public async IAsyncEnumerable<Order> StreamAsync(
        CallerPrincipal principal,
        [EnumeratorCancellation] CancellationToken cancellationToken)
    {
        Require(principal, OrderRoles.Reader);

        string? cursorToken = null;
        do
        {
            OrderCursor? cursor = null;
            if (cursorToken is not null)
            {
                if (!OrderCursor.TryDecode(cursorToken, out var decoded))
                {
                    // The token came from this process one page ago; failing to
                    // decode it is a programming error, not caller input.
                    throw new InvalidOperationException("The store returned an undecodable next cursor.");
                }

                cursor = decoded;
            }

            var page = await store
                .ListAsync(principal.Tenant, new OrderListQuery(OrderListQuery.MaxPageSize, cursor), cancellationToken)
                .ConfigureAwait(false);
            foreach (var order in page.Items)
            {
                cancellationToken.ThrowIfCancellationRequested();
                yield return order;
            }

            cursorToken = page.NextCursor;
        }
        while (cursorToken is not null);
    }

    /// <exception cref="PermissionDeniedException">The principal lacks the required role.</exception>
    private static void Require(CallerPrincipal principal, string role)
    {
        ArgumentNullException.ThrowIfNull(principal);
        if (!principal.HasRole(role))
        {
            throw new PermissionDeniedException(principal.Subject, role);
        }
    }
}
