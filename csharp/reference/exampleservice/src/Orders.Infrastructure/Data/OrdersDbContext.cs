using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Storage.ValueConversion;

using Orders.Core.Orders;

namespace Orders.Infrastructure.Data;

/// <summary>
/// The one DbContext this service owns - a unit of work, not a connection and
/// not a cache (csharp/services/database.md). Registered with
/// AddDbContextPool, scoped per request; reads are no-tracking by default (set
/// on the context options), write flows opt back in explicitly.
/// </summary>
public sealed class OrdersDbContext(DbContextOptions<OrdersDbContext> options) : DbContext(options)
{
    public DbSet<Order> Orders => Set<Order>();

    public DbSet<IdempotencyRecord> IdempotencyRecords => Set<IdempotencyRecord>();

    protected override void ConfigureConventions(ModelConfigurationBuilder configurationBuilder)
    {
        // Typed IDs convert once, by convention - never per-property
        // (csharp/services/database.md, Typed IDs).
        configurationBuilder.Properties<OrderId>().HaveConversion<OrderIdConverter>();
        configurationBuilder.Properties<TenantId>().HaveConversion<TenantIdConverter>();
    }

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(OrdersDbContext).Assembly);
    }

    internal sealed class OrderIdConverter() : ValueConverter<OrderId, Guid>(
        id => id.Value, value => new OrderId(value));

    internal sealed class TenantIdConverter() : ValueConverter<TenantId, string>(
        id => id.Value, value => new TenantId(value));
}
