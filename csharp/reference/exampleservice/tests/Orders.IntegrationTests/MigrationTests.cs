using Microsoft.EntityFrameworkCore;

using Orders.Core.Orders;
using Orders.Infrastructure.Data;

using Xunit;

namespace Orders.IntegrationTests;

/// <summary>
/// Proof the committed migrations apply cleanly to a real PostgreSQL and
/// produce the schema the code relies on (csharp/services/database.md).
/// </summary>
public sealed class MigrationTests
{
    private readonly PostgresFixture _postgres;

    public MigrationTests(PostgresFixture postgres)
    {
        _postgres = postgres;
    }

    [Fact]
    public async Task InitialCreate_IsApplied()
    {
        await using OrdersDbContext db = _postgres.CreateContext();

        IEnumerable<string> applied = await db.Database.GetAppliedMigrationsAsync(
            TestContext.Current.CancellationToken);

        Assert.Contains(applied, name => name.EndsWith("_InitialCreate", StringComparison.Ordinal));
    }

    [Fact]
    public async Task Schema_SupportsRoundTripAndXminConcurrencyToken()
    {
        var order = Order.Create(
            new TenantId($"mig-{Guid.NewGuid():N}"[..20]), "ref-schema", "cust-1", 1,
            DateTimeOffset.UtcNow);

        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            db.Orders.Add(order);
            await db.SaveChangesAsync(TestContext.Current.CancellationToken);
        }

        // xmin came back from the database: the concurrency token is real.
        Assert.NotEqual(0u, order.Version);

        await using OrdersDbContext reader = _postgres.CreateContext();
        Order? loaded = await reader.Orders.SingleOrDefaultAsync(
            o => o.Id == order.Id, TestContext.Current.CancellationToken);
        Assert.NotNull(loaded);
        Assert.Equal(order.ExternalReference, loaded.ExternalReference);
        Assert.Equal(order.Version, loaded.Version);
    }
}
