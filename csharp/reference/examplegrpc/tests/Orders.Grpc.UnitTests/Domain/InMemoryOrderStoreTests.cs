using Orders.Grpc.Core.Identity;
using Orders.Grpc.Core.Orders;

using Xunit;

namespace Orders.Grpc.UnitTests.Domain;

/// <summary>
/// Store-level contracts a database-backed implementation must also satisfy:
/// tenant scoping in the storage layer and stable (CreatedAt, Id) keyset
/// pagination including the created-at tie-break.
/// </summary>
public sealed class InMemoryOrderStoreTests
{
    private static readonly TenantId _tenantA = new("tenant-a");
    private static readonly TenantId _tenantB = new("tenant-b");
    private static readonly DateTimeOffset _now = new(2026, 7, 1, 12, 0, 0, TimeSpan.Zero);

    private readonly InMemoryOrderStore _store = new();

    [Fact]
    public async Task GetAsync_OtherTenantsOrder_ReturnsNull()
    {
        var order = Order.Create(_tenantA, "ord-1", "cust-1", 1, _now);
        await _store.AddAsync(order, TestContext.Current.CancellationToken);

        Assert.Null(await _store.GetAsync(_tenantB, order.Id, TestContext.Current.CancellationToken));
        Assert.NotNull(await _store.GetAsync(_tenantA, order.Id, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task ListAsync_IsScopedToTheTenant()
    {
        await _store.AddAsync(Order.Create(_tenantA, "ord-1", "cust-1", 1, _now), TestContext.Current.CancellationToken);
        await _store.AddAsync(Order.Create(_tenantB, "ord-2", "cust-2", 1, _now), TestContext.Current.CancellationToken);

        var page = await _store.ListAsync(
            _tenantA, new OrderListQuery(10, null), TestContext.Current.CancellationToken);

        var only = Assert.Single(page.Items);
        Assert.Equal(_tenantA, only.TenantId);
        Assert.Null(page.NextCursor);
    }

    [Fact]
    public async Task ListAsync_EqualTimestamps_TieBreakOnIdNeverSkipsOrDuplicates()
    {
        // Same CreatedAt on purpose: the id tie-break must produce a total
        // order, or keyset pages skip/duplicate rows.
        for (int i = 0; i < 3; i++)
        {
            await _store.AddAsync(
                Order.Create(_tenantA, $"ord-{i}", "cust-1", 1, _now), TestContext.Current.CancellationToken);
        }

        var first = await _store.ListAsync(
            _tenantA, new OrderListQuery(2, null), TestContext.Current.CancellationToken);
        Assert.NotNull(first.NextCursor);
        Assert.True(OrderCursor.TryDecode(first.NextCursor, out var cursor));

        var second = await _store.ListAsync(
            _tenantA, new OrderListQuery(2, cursor), TestContext.Current.CancellationToken);

        var all = first.Items.Concat(second.Items).Select(o => o.Id.Value).ToList();
        Assert.Equal(3, all.Count);
        Assert.Equal(all.Distinct().Count(), all.Count);
        Assert.Null(second.NextCursor);
    }
}
