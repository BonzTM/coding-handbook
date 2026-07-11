using Orders.Core.Orders;
using Orders.Infrastructure.Data;
using Orders.Infrastructure.Data.Repositories;

using Xunit;

namespace Orders.IntegrationTests;

/// <summary>
/// The repository against real PostgreSQL: tenant scoping, SQLSTATE 23505 ->
/// DuplicateOrderException, xmin optimistic concurrency, and row-value keyset
/// pagination - the provider semantics the InMemory provider would lie about
/// (csharp/services/database.md).
/// </summary>
public sealed class PostgresOrderRepositoryTests
{
    private readonly PostgresFixture _postgres;

    public PostgresOrderRepositoryTests(PostgresFixture postgres)
    {
        _postgres = postgres;
    }

    private static TenantId FreshTenant() => new($"t-{Guid.NewGuid():N}"[..20]);

    private static Order NewOrder(TenantId tenant, string reference, DateTimeOffset createdAt)
        => Order.Create(tenant, reference, "cust-1", 1, createdAt);

    [Fact]
    public async Task AddThenGet_RoundTripsWithinTenant()
    {
        TenantId tenant = FreshTenant();
        Order order = NewOrder(tenant, "ref-roundtrip", DateTimeOffset.UtcNow);

        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            await new PostgresOrderRepository(db).AddAsync(order, TestContext.Current.CancellationToken);
        }

        await using OrdersDbContext readDb = _postgres.CreateContext();
        Order? loaded = await new PostgresOrderRepository(readDb).GetAsync(
            tenant, order.Id, TestContext.Current.CancellationToken);

        Assert.NotNull(loaded);
        Assert.Equal(order.Id, loaded.Id);
        Assert.Equal("ref-roundtrip", loaded.ExternalReference);
    }

    [Fact]
    public async Task Get_CrossTenant_ReturnsNull()
    {
        TenantId tenant = FreshTenant();
        Order order = NewOrder(tenant, "ref-cross", DateTimeOffset.UtcNow);
        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            await new PostgresOrderRepository(db).AddAsync(order, TestContext.Current.CancellationToken);
        }

        await using OrdersDbContext readDb = _postgres.CreateContext();
        Order? loaded = await new PostgresOrderRepository(readDb).GetAsync(
            FreshTenant(), order.Id, TestContext.Current.CancellationToken);

        Assert.Null(loaded);
    }

    [Fact]
    public async Task Add_DuplicateReferenceSameTenant_MapsUniqueViolationToTypedException()
    {
        TenantId tenant = FreshTenant();
        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            await new PostgresOrderRepository(db).AddAsync(
                NewOrder(tenant, "ref-dup", DateTimeOffset.UtcNow), TestContext.Current.CancellationToken);
        }

        await using OrdersDbContext secondDb = _postgres.CreateContext();
        DuplicateOrderException thrown = await Assert.ThrowsAsync<DuplicateOrderException>(() =>
            new PostgresOrderRepository(secondDb).AddAsync(
                NewOrder(tenant, "ref-dup", DateTimeOffset.UtcNow), TestContext.Current.CancellationToken));

        Assert.Equal("ref-dup", thrown.ExternalReference);
    }

    [Fact]
    public async Task Add_SameReferenceDifferentTenant_Succeeds()
    {
        await using OrdersDbContext db = _postgres.CreateContext();
        var repository = new PostgresOrderRepository(db);

        await repository.AddAsync(
            NewOrder(FreshTenant(), "ref-shared", DateTimeOffset.UtcNow),
            TestContext.Current.CancellationToken);
        await repository.AddAsync(
            NewOrder(FreshTenant(), "ref-shared", DateTimeOffset.UtcNow),
            TestContext.Current.CancellationToken);
    }

    [Fact]
    public async Task List_PagesWithRowValueKeysetAndTenantScope()
    {
        TenantId tenant = FreshTenant();
        DateTimeOffset baseTime = DateTimeOffset.UtcNow;
        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            var repository = new PostgresOrderRepository(db);
            for (int i = 0; i < 5; i++)
            {
                await repository.AddAsync(
                    NewOrder(tenant, $"ref-{i}", baseTime.AddSeconds(i)),
                    TestContext.Current.CancellationToken);
            }

            // Another tenant's row must never appear in the pages below.
            await repository.AddAsync(
                NewOrder(FreshTenant(), "ref-other", baseTime), TestContext.Current.CancellationToken);
        }

        await using OrdersDbContext readDb = _postgres.CreateContext();
        var reader = new PostgresOrderRepository(readDb);

        OrderPage first = await reader.ListAsync(
            tenant, new OrderListQuery(2, null), TestContext.Current.CancellationToken);
        Assert.Equal(["ref-0", "ref-1"], first.Items.Select(o => o.ExternalReference));
        Assert.NotNull(first.NextCursor);

        Assert.True(OrderCursor.TryDecode(first.NextCursor, out OrderCursor cursor));
        OrderPage second = await reader.ListAsync(
            tenant, new OrderListQuery(2, cursor), TestContext.Current.CancellationToken);
        Assert.Equal(["ref-2", "ref-3"], second.Items.Select(o => o.ExternalReference));
        Assert.NotNull(second.NextCursor);

        Assert.True(OrderCursor.TryDecode(second.NextCursor, out cursor));
        OrderPage third = await reader.ListAsync(
            tenant, new OrderListQuery(2, cursor), TestContext.Current.CancellationToken);
        Assert.Equal(["ref-4"], third.Items.Select(o => o.ExternalReference));
        Assert.Null(third.NextCursor);
    }

    [Fact]
    public async Task Update_WithCurrentVersion_BumpsXminToken()
    {
        TenantId tenant = FreshTenant();
        Order order = NewOrder(tenant, "ref-update", DateTimeOffset.UtcNow);
        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            await new PostgresOrderRepository(db).AddAsync(order, TestContext.Current.CancellationToken);
        }

        uint versionAfterCreate = order.Version;
        order.Amend(5, OrderStatus.Confirmed, DateTimeOffset.UtcNow);

        await using OrdersDbContext writeDb = _postgres.CreateContext();
        Order updated = await new PostgresOrderRepository(writeDb).UpdateAsync(
            order, versionAfterCreate, TestContext.Current.CancellationToken);

        Assert.Equal(5, updated.Quantity);
        Assert.NotEqual(versionAfterCreate, updated.Version); // the DATABASE bumped it
    }

    [Fact]
    public async Task Update_WithStaleVersion_ThrowsVersionConflict()
    {
        TenantId tenant = FreshTenant();
        Order order = NewOrder(tenant, "ref-stale", DateTimeOffset.UtcNow);
        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            await new PostgresOrderRepository(db).AddAsync(order, TestContext.Current.CancellationToken);
        }

        uint staleVersion = order.Version;

        // A concurrent writer updates the row first.
        order.Amend(2, OrderStatus.Confirmed, DateTimeOffset.UtcNow);
        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            await new PostgresOrderRepository(db).UpdateAsync(
                order, staleVersion, TestContext.Current.CancellationToken);
        }

        // Retrying with the version read BEFORE that write must conflict.
        order.Amend(3, OrderStatus.Shipped, DateTimeOffset.UtcNow);
        await using OrdersDbContext lateDb = _postgres.CreateContext();
        await Assert.ThrowsAsync<OrderVersionConflictException>(() =>
            new PostgresOrderRepository(lateDb).UpdateAsync(
                order, staleVersion, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task Delete_ScopedToTenant_ReportsWhetherARowWentAway()
    {
        TenantId tenant = FreshTenant();
        Order order = NewOrder(tenant, "ref-delete", DateTimeOffset.UtcNow);
        await using (OrdersDbContext db = _postgres.CreateContext())
        {
            await new PostgresOrderRepository(db).AddAsync(order, TestContext.Current.CancellationToken);
        }

        await using OrdersDbContext workDb = _postgres.CreateContext();
        var repository = new PostgresOrderRepository(workDb);

        // Wrong tenant deletes nothing.
        Assert.False(await repository.DeleteAsync(
            FreshTenant(), order.Id, TestContext.Current.CancellationToken));
        // Owner delete succeeds exactly once.
        Assert.True(await repository.DeleteAsync(
            tenant, order.Id, TestContext.Current.CancellationToken));
        Assert.False(await repository.DeleteAsync(
            tenant, order.Id, TestContext.Current.CancellationToken));
    }
}
