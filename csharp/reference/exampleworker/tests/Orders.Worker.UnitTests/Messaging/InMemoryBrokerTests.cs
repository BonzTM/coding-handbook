using Orders.Worker.Core.Events;
using Orders.Worker.Infrastructure.Messaging;
using Orders.Worker.UnitTests.Fakes;

using Xunit;

namespace Orders.Worker.UnitTests.Messaging;

/// <summary>
/// The in-memory broker's delivery semantics: FIFO per topic, at-least-once
/// redelivery on nack, closed-broker behavior, and the depth/health surface
/// the readiness probe and lag gauge read.
/// </summary>
public sealed class InMemoryBrokerTests
{
    private const string Topic = "orders.events";

    [Fact]
    public async Task PublishThenSubscribe_DeliversInOrder()
    {
        var broker = new InMemoryBroker();
        var first = TestEvents.Envelope(TestEvents.Placed(orderId: "order-1"));
        var second = TestEvents.Envelope(TestEvents.Placed(orderId: "order-2"));
        await broker.PublishAsync(Topic, first, TestContext.Current.CancellationToken);
        await broker.PublishAsync(Topic, second, TestContext.Current.CancellationToken);

        var received = await ReadAsync(broker, count: 2);

        Assert.Equal([first.Id, second.Id], received.Select(message => message.Envelope.Id));
    }

    [Fact]
    public async Task Nack_RedeliversTheSameEnvelope()
    {
        var broker = new InMemoryBroker();
        var envelope = TestEvents.Envelope(TestEvents.Placed());
        await broker.PublishAsync(Topic, envelope, TestContext.Current.CancellationToken);

        var firstDelivery = Assert.Single(await ReadAsync(broker, count: 1));
        await firstDelivery.NackAsync();
        var redelivery = Assert.Single(await ReadAsync(broker, count: 1));

        Assert.Equal(envelope.Id, redelivery.Envelope.Id);
    }

    [Fact]
    public async Task Close_CompletesOpenSubscriptions()
    {
        var broker = new InMemoryBroker();
        // Start the subscription first: MoveNextAsync runs the iterator up to
        // its (pending) wait-to-read, so Close deterministically ends an
        // ESTABLISHED subscription rather than racing its creation.
        var subscription = broker.SubscribeAsync(Topic, TestContext.Current.CancellationToken)
            .GetAsyncEnumerator(TestContext.Current.CancellationToken);
        await using (subscription.ConfigureAwait(false))
        {
            var firstRead = subscription.MoveNextAsync();

            broker.Close();

            Assert.False(await firstRead);
        }
    }

    [Fact]
    public async Task Publish_AfterClose_Throws()
    {
        var broker = new InMemoryBroker();
        broker.Close();

        await Assert.ThrowsAsync<InvalidOperationException>(() => broker.PublishAsync(
            Topic, TestEvents.Envelope(TestEvents.Placed()), TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task Depth_ReportsQueuedMessages()
    {
        var broker = new InMemoryBroker();
        Assert.Equal(0, broker.Depth(Topic));

        await broker.PublishAsync(Topic, TestEvents.Envelope(TestEvents.Placed(orderId: "order-1")), TestContext.Current.CancellationToken);
        await broker.PublishAsync(Topic, TestEvents.Envelope(TestEvents.Placed(orderId: "order-2")), TestContext.Current.CancellationToken);
        Assert.Equal(2, broker.Depth(Topic));

        _ = await ReadAsync(broker, count: 2);
        Assert.Equal(0, broker.Depth(Topic));
    }

    [Fact]
    public void Health_TogglesAndClosePinsUnhealthy()
    {
        var broker = new InMemoryBroker();
        Assert.True(broker.IsHealthy);

        broker.SetHealthy(false);
        Assert.False(broker.IsHealthy);

        broker.SetHealthy(true);
        broker.Close();
        Assert.False(broker.IsHealthy);
    }

    /// <summary>Reads exactly <paramref name="count"/> deliveries, bounded by
    /// the test's cancellation token so a broken broker fails the test instead
    /// of hanging it.</summary>
    private static async Task<List<InboundMessage>> ReadAsync(InMemoryBroker broker, int count)
    {
        var received = new List<InboundMessage>();
        await foreach (var message in broker.SubscribeAsync(Topic, TestContext.Current.CancellationToken))
        {
            received.Add(message);
            if (received.Count == count)
            {
                break;
            }
        }

        return received;
    }
}
