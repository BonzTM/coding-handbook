using Microsoft.Extensions.Time.Testing;

using Orders.Core.Orders;
using Orders.UnitTests.Fakes;

using Xunit;

namespace Orders.UnitTests.Domain;

/// <summary>
/// Domain-service rules against the in-memory fake repository and a
/// FakeTimeProvider - no clock, no database, fully deterministic
/// (csharp/foundations/time.md, csharp/quality/testing.md).
/// </summary>
public sealed class OrderServiceTests
{
    private static readonly TenantId _tenantA = new("tenant-a");
    private static readonly TenantId _tenantB = new("tenant-b");
    private static readonly DateTimeOffset _start = new(2026, 7, 1, 12, 0, 0, TimeSpan.Zero);

    private readonly FakeTimeProvider _time = new(_start);
    private readonly InMemoryOrderRepository _repository = new();
    private readonly OrderService _service;

    public OrderServiceTests()
    {
        _service = new OrderService(_repository, _time);
    }

    [Fact]
    public async Task CreateAsync_StampsTimestampsFromInjectedClock()
    {
        _time.Advance(TimeSpan.FromMinutes(7));

        Order order = await _service.CreateAsync(
            _tenantA, "ref-1", "cust-1", 2, TestContext.Current.CancellationToken);

        Assert.Equal(_start + TimeSpan.FromMinutes(7), order.CreatedAt);
        Assert.Equal(order.CreatedAt, order.UpdatedAt);
    }

    [Fact]
    public async Task CreateAsync_DuplicateExternalReference_Throws()
    {
        await _service.CreateAsync(_tenantA, "ref-1", "cust-1", 1, TestContext.Current.CancellationToken);

        await Assert.ThrowsAsync<DuplicateOrderException>(() =>
            _service.CreateAsync(_tenantA, "ref-1", "cust-2", 2, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task CreateAsync_SameReferenceDifferentTenant_Succeeds()
    {
        await _service.CreateAsync(_tenantA, "ref-1", "cust-1", 1, TestContext.Current.CancellationToken);

        Order other = await _service.CreateAsync(
            _tenantB, "ref-1", "cust-1", 1, TestContext.Current.CancellationToken);

        Assert.Equal(_tenantB, other.TenantId);
    }

    [Fact]
    public async Task GetAsync_UnknownOrder_ThrowsNotFound()
    {
        await Assert.ThrowsAsync<OrderNotFoundException>(() =>
            _service.GetAsync(_tenantA, OrderId.New(), TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task GetAsync_CrossTenant_IsIndistinguishableFromNotFound()
    {
        Order order = await _service.CreateAsync(
            _tenantA, "ref-1", "cust-1", 1, TestContext.Current.CancellationToken);

        await Assert.ThrowsAsync<OrderNotFoundException>(() =>
            _service.GetAsync(_tenantB, order.Id, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task ListAsync_PagesThroughAllOrdersWithoutOverlap()
    {
        for (int i = 0; i < 5; i++)
        {
            _time.Advance(TimeSpan.FromSeconds(1)); // distinct CreatedAt per order
            await _service.CreateAsync(
                _tenantA, $"ref-{i}", "cust-1", 1, TestContext.Current.CancellationToken);
        }

        OrderPage first = await _service.ListAsync(
            _tenantA, new OrderListQuery(2, null), TestContext.Current.CancellationToken);
        Assert.Equal(2, first.Items.Count);
        Assert.NotNull(first.NextCursor);

        Assert.True(OrderCursor.TryDecode(first.NextCursor, out OrderCursor cursor));
        OrderPage second = await _service.ListAsync(
            _tenantA, new OrderListQuery(2, cursor), TestContext.Current.CancellationToken);
        Assert.Equal(2, second.Items.Count);
        Assert.NotNull(second.NextCursor);

        Assert.True(OrderCursor.TryDecode(second.NextCursor, out cursor));
        OrderPage third = await _service.ListAsync(
            _tenantA, new OrderListQuery(2, cursor), TestContext.Current.CancellationToken);
        Assert.Single(third.Items);
        Assert.Null(third.NextCursor); // last page

        string[] seen =
        [
            .. first.Items.Select(o => o.ExternalReference),
            .. second.Items.Select(o => o.ExternalReference),
            .. third.Items.Select(o => o.ExternalReference),
        ];
        Assert.Equal(["ref-0", "ref-1", "ref-2", "ref-3", "ref-4"], seen);
    }

    [Fact]
    public async Task ListAsync_IsTenantScoped()
    {
        await _service.CreateAsync(_tenantA, "ref-a", "cust-1", 1, TestContext.Current.CancellationToken);
        await _service.CreateAsync(_tenantB, "ref-b", "cust-1", 1, TestContext.Current.CancellationToken);

        OrderPage page = await _service.ListAsync(
            _tenantA, new OrderListQuery(null, null), TestContext.Current.CancellationToken);

        Order item = Assert.Single(page.Items);
        Assert.Equal("ref-a", item.ExternalReference);
    }

    [Fact]
    public async Task AmendAsync_UpdatesFieldsAndTimestamp()
    {
        Order created = await _service.CreateAsync(
            _tenantA, "ref-1", "cust-1", 1, TestContext.Current.CancellationToken);
        _time.Advance(TimeSpan.FromMinutes(1));

        Order amended = await _service.AmendAsync(
            _tenantA, created.Id, created.Version, 5, OrderStatus.Confirmed,
            TestContext.Current.CancellationToken);

        Assert.Equal(5, amended.Quantity);
        Assert.Equal(OrderStatus.Confirmed, amended.Status);
        Assert.Equal(_start + TimeSpan.FromMinutes(1), amended.UpdatedAt);
    }

    [Fact]
    public async Task AmendAsync_StaleVersion_ThrowsVersionConflict()
    {
        Order created = await _service.CreateAsync(
            _tenantA, "ref-1", "cust-1", 1, TestContext.Current.CancellationToken);

        await Assert.ThrowsAsync<OrderVersionConflictException>(() =>
            _service.AmendAsync(
                _tenantA, created.Id, created.Version + 1, 5, OrderStatus.Confirmed,
                TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task DeleteAsync_RemovesOrder()
    {
        Order created = await _service.CreateAsync(
            _tenantA, "ref-1", "cust-1", 1, TestContext.Current.CancellationToken);

        await _service.DeleteAsync(_tenantA, created.Id, TestContext.Current.CancellationToken);

        await Assert.ThrowsAsync<OrderNotFoundException>(() =>
            _service.GetAsync(_tenantA, created.Id, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task DeleteAsync_UnknownOrder_ThrowsNotFound()
    {
        await Assert.ThrowsAsync<OrderNotFoundException>(() =>
            _service.DeleteAsync(_tenantA, OrderId.New(), TestContext.Current.CancellationToken));
    }
}
