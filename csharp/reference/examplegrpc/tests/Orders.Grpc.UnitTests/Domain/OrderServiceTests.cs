using Microsoft.Extensions.Time.Testing;

using Orders.Grpc.Core.Identity;
using Orders.Grpc.Core.Orders;

using Xunit;

namespace Orders.Grpc.UnitTests.Domain;

/// <summary>
/// Domain rules against the real in-memory store with a FakeTimeProvider
/// (csharp/foundations/time.md): role checks deny before any store work,
/// tenant scoping holds, duplicates surface typed, pagination and streaming
/// follow the keyset order.
/// </summary>
public sealed class OrderServiceTests
{
    private static readonly CallerPrincipal _writer = new(
        "writer", new TenantId("tenant-a"), [OrderRoles.Reader, OrderRoles.Writer]);

    private static readonly CallerPrincipal _readerOnly = new(
        "reader", new TenantId("tenant-a"), [OrderRoles.Reader]);

    private static readonly CallerPrincipal _writerOnly = new(
        "writer-only", new TenantId("tenant-a"), [OrderRoles.Writer]);

    private static readonly CallerPrincipal _otherTenant = new(
        "intruder", new TenantId("tenant-b"), [OrderRoles.Reader, OrderRoles.Writer]);

    private readonly FakeTimeProvider _time = new(new DateTimeOffset(2026, 7, 1, 12, 0, 0, TimeSpan.Zero));
    private readonly InMemoryOrderStore _store = new();
    private readonly OrderService _service;

    public OrderServiceTests()
    {
        _service = new OrderService(_store, _time);
    }

    [Fact]
    public async Task CreateAsync_WithoutWriterRole_IsDenied()
    {
        var denied = await Assert.ThrowsAsync<PermissionDeniedException>(() =>
            _service.CreateAsync(_readerOnly, "ord-1", "cust-1", 1, TestContext.Current.CancellationToken));

        Assert.Equal(OrderRoles.Writer, denied.RequiredRole);
    }

    [Fact]
    public async Task GetAsync_WithoutReaderRole_IsDenied()
    {
        var denied = await Assert.ThrowsAsync<PermissionDeniedException>(() =>
            _service.GetAsync(_writerOnly, OrderId.New(), TestContext.Current.CancellationToken));

        Assert.Equal(OrderRoles.Reader, denied.RequiredRole);
    }

    [Fact]
    public async Task CreateAsync_StampsCreationTimeFromInjectedClock()
    {
        var order = await _service.CreateAsync(
            _writer, "ord-1", "cust-1", 2, TestContext.Current.CancellationToken);

        Assert.Equal(_time.GetUtcNow(), order.CreatedAt);
        Assert.Equal(_writer.Tenant, order.TenantId);
    }

    [Fact]
    public async Task CreateAsync_DuplicateReferenceInTenant_Throws()
    {
        _ = await _service.CreateAsync(_writer, "ord-1", "cust-1", 1, TestContext.Current.CancellationToken);

        var duplicate = await Assert.ThrowsAsync<DuplicateOrderException>(() =>
            _service.CreateAsync(_writer, "ord-1", "cust-2", 5, TestContext.Current.CancellationToken));

        Assert.Equal("ord-1", duplicate.ExternalReference);
    }

    [Fact]
    public async Task CreateAsync_SameReferenceOtherTenant_Succeeds()
    {
        _ = await _service.CreateAsync(_writer, "ord-1", "cust-1", 1, TestContext.Current.CancellationToken);

        var other = await _service.CreateAsync(
            _otherTenant, "ord-1", "cust-9", 1, TestContext.Current.CancellationToken);

        Assert.Equal(_otherTenant.Tenant, other.TenantId);
    }

    [Fact]
    public async Task GetAsync_CrossTenant_IsIndistinguishableFromMissing()
    {
        var order = await _service.CreateAsync(
            _writer, "ord-1", "cust-1", 1, TestContext.Current.CancellationToken);

        _ = await Assert.ThrowsAsync<OrderNotFoundException>(() =>
            _service.GetAsync(_otherTenant, order.Id, TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task ListAsync_WalksAllPagesInKeysetOrder()
    {
        var created = await SeedAsync(count: 5);

        var seen = new List<OrderId>();
        string? cursorToken = null;
        // Bounded walk: at most one page per seeded order.
        for (int page = 0; page < created.Count; page++)
        {
            OrderCursor? cursor = null;
            if (cursorToken is not null)
            {
                Assert.True(OrderCursor.TryDecode(cursorToken, out var decoded));
                cursor = decoded;
            }

            var result = await _service.ListAsync(
                _writer, new OrderListQuery(2, cursor), TestContext.Current.CancellationToken);
            seen.AddRange(result.Items.Select(o => o.Id));
            cursorToken = result.NextCursor;
            if (cursorToken is null)
            {
                break;
            }
        }

        Assert.Equal(created.Select(o => o.Id), seen);
    }

    [Fact]
    public async Task StreamAsync_YieldsEveryOrderInKeysetOrder()
    {
        var created = await SeedAsync(count: 4);

        var streamed = new List<OrderId>();
        await foreach (var order in _service.StreamAsync(_writer, TestContext.Current.CancellationToken))
        {
            streamed.Add(order.Id);
        }

        Assert.Equal(created.Select(o => o.Id), streamed);
    }

    [Fact]
    public async Task StreamAsync_HonorsCancellationMidStream()
    {
        _ = await SeedAsync(count: 3);
        using var cts = new CancellationTokenSource();

        await Assert.ThrowsAnyAsync<OperationCanceledException>(async () =>
        {
            await foreach (var order in _service.StreamAsync(_writer, cts.Token))
            {
                // Cancel after the first yielded order; the next iteration
                // must observe it and stop.
                await cts.CancelAsync();
            }
        });
    }

    private async Task<List<Order>> SeedAsync(int count)
    {
        var created = new List<Order>(count);
        for (int i = 0; i < count; i++)
        {
            created.Add(await _service.CreateAsync(
                _writer, $"ord-{i:D3}", "cust-1", i + 1, TestContext.Current.CancellationToken));
            // Distinct timestamps keep the keyset order unambiguous.
            _time.Advance(TimeSpan.FromSeconds(1));
        }

        return created;
    }
}
