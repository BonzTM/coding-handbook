using Microsoft.EntityFrameworkCore;

using Npgsql;

using Orders.Core.Orders;

namespace Orders.Infrastructure.Data.Repositories;

/// <summary>
/// EF Core implementation of the Core persistence port. Every query filters on
/// TenantId; unique violations and concurrency conflicts surface as the typed
/// Core exceptions, never as provider exceptions leaking across the boundary
/// (csharp/services/database.md).
/// </summary>
public sealed class PostgresOrderRepository(OrdersDbContext db) : IOrderRepository
{
    public async Task AddAsync(Order order, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(order);
        db.Orders.Add(order);
        try
        {
            await db.SaveChangesAsync(cancellationToken).ConfigureAwait(false);
        }
        catch (DbUpdateException ex) when (ex.InnerException is PostgresException
        {
            SqlState: PostgresErrorCodes.UniqueViolation,
            ConstraintName: Configurations.OrderConfiguration.TenantExternalReferenceIndex,
        })
        {
            // Insert-and-translate, never a racy pre-check SELECT. Match on
            // SqlState AND ConstraintName so a second unique constraint would
            // map to its own typed error, not this one.
            db.Entry(order).State = EntityState.Detached;
            throw new DuplicateOrderException(order.ExternalReference);
        }
    }

    public Task<Order?> GetAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken)
        => db.Orders
            .FirstOrDefaultAsync(o => o.TenantId == tenantId && o.Id == id, cancellationToken);

    public async Task<OrderPage> ListAsync(
        TenantId tenantId, OrderListQuery query, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(query);

        var rows = db.Orders.Where(o => o.TenantId == tenantId);
        if (query.Cursor is { } cursor)
        {
            var afterId = new OrderId(cursor.Id);
            // PostgreSQL row-value comparison: (created_at, id) > (@t, @id) -
            // one index-friendly predicate over the same composite index the
            // ORDER BY uses (csharp/services/http-services.md, pagination).
            rows = rows.Where(o => EF.Functions.GreaterThan(
                ValueTuple.Create(o.CreatedAt, o.Id),
                ValueTuple.Create(cursor.CreatedAt, afterId)));
        }

        // Fetch one extra row to learn whether a next page exists without a COUNT.
        var items = await rows
            .OrderBy(o => o.CreatedAt)
            .ThenBy(o => o.Id)
            .Take(query.PageSize + 1)
            .ToListAsync(cancellationToken)
            .ConfigureAwait(false);

        if (items.Count <= query.PageSize)
        {
            return new OrderPage(items, NextCursor: null);
        }

        items.RemoveAt(items.Count - 1);
        var last = items[^1];
        return new OrderPage(items, new OrderCursor(last.CreatedAt, last.Id.Value).Encode());
    }

    public async Task<Order> UpdateAsync(Order order, uint expectedVersion, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(order);

        var entry = db.Orders.Attach(order);
        entry.Property(o => o.Version).OriginalValue = expectedVersion;
        entry.State = EntityState.Modified;
        try
        {
            await db.SaveChangesAsync(cancellationToken).ConfigureAwait(false);
            return order;
        }
        catch (DbUpdateConcurrencyException)
        {
            // The xmin token no longer matches (concurrent update or delete).
            // The caller reloads and re-decides; blind retry is not handling.
            entry.State = EntityState.Detached;
            throw new OrderVersionConflictException(order.Id);
        }
    }

    public async Task<bool> DeleteAsync(TenantId tenantId, OrderId id, CancellationToken cancellationToken)
    {
        int deleted = await db.Orders
            .Where(o => o.TenantId == tenantId && o.Id == id)
            .ExecuteDeleteAsync(cancellationToken)
            .ConfigureAwait(false);
        return deleted > 0;
    }
}
