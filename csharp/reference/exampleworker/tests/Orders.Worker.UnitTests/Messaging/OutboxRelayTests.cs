using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.Extensions.Options;
using Microsoft.Extensions.Time.Testing;

using Orders.Worker.Core.Messaging;
using Orders.Worker.Infrastructure.Messaging;
using Orders.Worker.Infrastructure.Telemetry;
using Orders.Worker.UnitTests.Fakes;

using Xunit;

namespace Orders.Worker.UnitTests.Messaging;

/// <summary>
/// The transactional-outbox relay: pending records publish in order and are
/// marked sent only AFTER a successful publish; a publish failure leaves them
/// pending for the next scan (reliable publish, no dual-write loss); and the
/// relay service performs one final flush inside the shutdown drain.
/// </summary>
public sealed class OutboxRelayTests
{
    private const string Topic = "orders.events";

    [Fact]
    public async Task Flush_PublishesPendingInOrderAndMarksSent()
    {
        var harness = new RelayHarness();
        var ids = new[] { Guid.NewGuid(), Guid.NewGuid(), Guid.NewGuid() };
        foreach (var id in ids)
        {
            await harness.Outbox.AddAsync(NewRecord(id), TestContext.Current.CancellationToken);
        }

        int published = await harness.Relay.FlushAsync(TestContext.Current.CancellationToken);

        Assert.Equal(3, published);
        Assert.Equal(0, harness.Outbox.PendingCount);
        var delivered = await ReadAsync(harness.Broker, count: 3);
        Assert.Equal(ids, delivered.Select(message => message.Envelope.Id));
        // The relayed envelope carries the record's contract metadata.
        Assert.All(delivered, message =>
        {
            Assert.Equal("orders.order-placed.v1", message.Envelope.Type);
            Assert.Equal("order-1", message.Envelope.Subject);
            Assert.Equal(WorkerTelemetry.ServiceName, message.Envelope.Source);
        });

        // A second flush is a no-op: everything is marked sent.
        Assert.Equal(0, await harness.Relay.FlushAsync(TestContext.Current.CancellationToken));
    }

    [Fact]
    public async Task Flush_PublishFailureLeavesRecordPendingForNextScan()
    {
        var harness = new RelayHarness();
        harness.Broker.Close(); // every publish now fails
        await harness.Outbox.AddAsync(NewRecord(Guid.NewGuid()), TestContext.Current.CancellationToken);

        int published = await harness.Relay.FlushAsync(TestContext.Current.CancellationToken);

        Assert.Equal(0, published);
        Assert.Equal(1, harness.Outbox.PendingCount);
    }

    [Fact]
    public async Task RelayService_FinalFlushPublishesRecordsEnqueuedBeforeShutdown()
    {
        // Poll interval of one hour: the ONLY flush that can publish the
        // record is the final one on shutdown.
        var harness = new RelayHarness(pollInterval: TimeSpan.FromHours(1));
        await harness.Outbox.AddAsync(NewRecord(Guid.NewGuid()), TestContext.Current.CancellationToken);
        using var service = new OutboxRelayService(
            harness.Relay,
            harness.Time,
            Options.Create(harness.Options),
            NullLogger<OutboxRelayService>.Instance);

        await service.StartAsync(TestContext.Current.CancellationToken);
        await service.StopAsync(TestContext.Current.CancellationToken);

        Assert.Equal(0, harness.Outbox.PendingCount);
        var delivered = Assert.Single(await ReadAsync(harness.Broker, count: 1));
        Assert.Equal("orders.order-placed.v1", delivered.Envelope.Type);
    }

    private static OutboxRecord NewRecord(Guid id) => new()
    {
        Id = id,
        Type = "orders.order-placed.v1",
        Payload = """{"type":"orders.order-placed.v1","orderId":"order-1","tenantId":"tenant-a","externalReference":"ref-1"}""",
        Subject = "order-1",
        OccurredAt = TestEvents.Start,
    };

    private static async Task<List<Orders.Worker.Core.Events.InboundMessage>> ReadAsync(
        InMemoryBroker broker, int count)
    {
        var received = new List<Orders.Worker.Core.Events.InboundMessage>();
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

    /// <summary>Relay + in-memory broker/outbox on a FakeTimeProvider, wired
    /// exactly as the composition root does.</summary>
    private sealed class RelayHarness
    {
        public RelayHarness(TimeSpan? pollInterval = null)
        {
            Options = new OutboxOptions { PollInterval = pollInterval ?? TimeSpan.FromSeconds(1), BatchSize = 100 };
            var publisher = new BrokerEventPublisher(
                Broker,
                Microsoft.Extensions.Options.Options.Create(new MessagingOptions { Topic = Topic }));
            Relay = new OutboxRelay(
                Outbox,
                publisher,
                Time,
                Microsoft.Extensions.Options.Options.Create(Options),
                new WorkerMetrics(MeterFactory),
                NullLogger<OutboxRelay>.Instance);
        }

        public FakeTimeProvider Time { get; } = new(TestEvents.Start);

        public InMemoryBroker Broker { get; } = new();

        public InMemoryOutboxStore Outbox { get; } = new();

        public TestMeterFactory MeterFactory { get; } = new();

        public OutboxOptions Options { get; }

        public OutboxRelay Relay { get; }
    }
}
