using Microsoft.Extensions.Time.Testing;

using Orders.Worker.Core.Events;
using Orders.Worker.Core.Orders;
using Orders.Worker.UnitTests.Fakes;

using Xunit;

namespace Orders.Worker.UnitTests.Domain;

/// <summary>Projection rules against a FakeTimeProvider - no wall clock, fully
/// deterministic (csharp/foundations/time.md).</summary>
public sealed class OrderProjectorTests
{
    private readonly FakeTimeProvider _time = new(TestEvents.Start);
    private readonly OrderProjector _projector;

    public OrderProjectorTests()
    {
        _projector = new OrderProjector(_time);
    }

    [Fact]
    public async Task Placed_CreatesProjectionStampedFromInjectedClock()
    {
        _time.Advance(TimeSpan.FromMinutes(3));

        await _projector.ProcessAsync(TestEvents.Placed(), TestContext.Current.CancellationToken);

        var projection = _projector.GetProjection("tenant-a", "order-1");
        Assert.NotNull(projection);
        Assert.Equal("ref-1", projection.ExternalReference);
        Assert.False(projection.Cancelled);
        Assert.Equal(TestEvents.Start + TimeSpan.FromMinutes(3), projection.UpdatedAt);
    }

    [Fact]
    public async Task Cancelled_MarksTheExistingProjection()
    {
        await _projector.ProcessAsync(TestEvents.Placed(), TestContext.Current.CancellationToken);

        await _projector.ProcessAsync(TestEvents.Cancelled(), TestContext.Current.CancellationToken);

        var projection = _projector.GetProjection("tenant-a", "order-1");
        Assert.NotNull(projection);
        Assert.True(projection.Cancelled);
        Assert.Equal("ref-1", projection.ExternalReference);
    }

    [Fact]
    public async Task Cancelled_BeforePlaced_LeavesATombstoneThePlaceCannotRevive()
    {
        // Redelivery can interleave events on one key; a cancel must never be
        // silently undone by a late place.
        await _projector.ProcessAsync(TestEvents.Cancelled(), TestContext.Current.CancellationToken);
        await _projector.ProcessAsync(TestEvents.Placed(), TestContext.Current.CancellationToken);

        var projection = _projector.GetProjection("tenant-a", "order-1");
        Assert.NotNull(projection);
        Assert.True(projection.Cancelled);
    }

    [Fact]
    public async Task ReplayingTheSameEvent_ConvergesToTheSameState()
    {
        var placed = TestEvents.Placed();
        await _projector.ProcessAsync(placed, TestContext.Current.CancellationToken);
        var first = _projector.GetProjection("tenant-a", "order-1");

        await _projector.ProcessAsync(placed, TestContext.Current.CancellationToken);

        Assert.Equal(first, _projector.GetProjection("tenant-a", "order-1"));
    }

    [Fact]
    public async Task TenantsAreIsolated()
    {
        await _projector.ProcessAsync(TestEvents.Placed(tenantId: "tenant-a"), TestContext.Current.CancellationToken);

        Assert.Null(_projector.GetProjection("tenant-b", "order-1"));
    }

    [Fact]
    public async Task InvalidEvent_ThrowsWithoutMutatingState()
    {
        var invalid = TestEvents.Placed() with { Type = "orders.unknown.v1" };

        await Assert.ThrowsAsync<InvalidEventException>(() =>
            _projector.ProcessAsync(invalid, TestContext.Current.CancellationToken));

        Assert.Null(_projector.GetProjection("tenant-a", "order-1"));
    }
}
