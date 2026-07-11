using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.Extensions.Time.Testing;

using Orders.Worker.Core.Events;
using Orders.Worker.Core.Messaging;
using Orders.Worker.Core.Orders;
using Orders.Worker.Infrastructure.Messaging;
using Orders.Worker.Infrastructure.Telemetry;

namespace Orders.Worker.UnitTests.Fakes;

/// <summary>
/// Wires an <see cref="OrderEventHandler"/> with FakeTimeProvider, the
/// recording delayer, in-memory stores, and a pinned jitter source (1.0 =
/// every delay equals its exponential ceiling), so pipeline tests are fully
/// deterministic with no real sleeps.
/// </summary>
internal sealed class HandlerHarness
{
    public HandlerHarness(ConsumerOptions? options = null, double jitter = 1.0)
    {
        Options = options ?? new ConsumerOptions
        {
            MaxAttempts = 4,
            BaseBackoff = TimeSpan.FromMilliseconds(100),
            MaxBackoff = TimeSpan.FromSeconds(1),
        };
        Time = new FakeTimeProvider(TestEvents.Start);
        Delayer = new RecordingRetryDelayer(Time);
        Backoff = new BackoffPolicy(Options.BaseBackoff, Options.MaxBackoff, () => jitter);
        MeterFactory = new TestMeterFactory();
        Metrics = new WorkerMetrics(MeterFactory);
    }

    public ConsumerOptions Options { get; }

    public FakeTimeProvider Time { get; }

    public RecordingRetryDelayer Delayer { get; }

    public BackoffPolicy Backoff { get; }

    public TestMeterFactory MeterFactory { get; }

    public WorkerMetrics Metrics { get; }

    public InMemoryInboxStore Inbox { get; } = new();

    public InMemoryDeadLetterStore DeadLetters { get; } = new();

    public OrderEventHandler CreateHandler(IOrderEventProcessor processor)
        => new(
            processor,
            Inbox,
            DeadLetters,
            Backoff,
            Delayer,
            Time,
            Microsoft.Extensions.Options.Options.Create(Options),
            Metrics,
            NullLogger<OrderEventHandler>.Instance);
}

/// <summary>An <see cref="InboundMessage"/> plus counters for its settlement
/// handles, so tests assert exactly one terminal settle per delivery.</summary>
internal sealed class TestDelivery
{
    private int _acks;
    private int _nacks;

    public TestDelivery(EventEnvelope envelope)
    {
        Message = new InboundMessage(
            envelope,
            ack: () =>
            {
                Interlocked.Increment(ref _acks);
                return ValueTask.CompletedTask;
            },
            nack: () =>
            {
                Interlocked.Increment(ref _nacks);
                return ValueTask.CompletedTask;
            });
    }

    public InboundMessage Message { get; }

    public int Acks => Volatile.Read(ref _acks);

    public int Nacks => Volatile.Read(ref _nacks);
}
