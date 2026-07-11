using System.Diagnostics;

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

using Orders.Worker.Core.Orders;
using Orders.Worker.Infrastructure;
using Orders.Worker.Infrastructure.Messaging;
using Orders.Worker.UnitTests.Fakes;

using Xunit;

namespace Orders.Worker.UnitTests.Host;

/// <summary>
/// The graceful-drain contract on a REAL generic host wired through the same
/// composition extension as production: stopping the host stops pulling new
/// messages but finishes and settles the in-flight one - and the whole drain
/// is bounded by HostOptions.ShutdownTimeout, so a wedged handler cannot block
/// shutdown past the budget (csharp/services/eventing-and-messaging.md,
/// Verification And Proof).
/// </summary>
public sealed class DrainTests
{
    private const string Topic = "orders.events";

    [Fact]
    public async Task Stop_FinishesAndSettlesTheInFlightMessage()
    {
        var processor = new BlockingProcessor();
        using var host = BuildHost(processor, shutdownTimeout: TimeSpan.FromSeconds(10));
        await host.StartAsync(TestContext.Current.CancellationToken);
        var broker = host.Services.GetRequiredService<InMemoryBroker>();
        var envelope = TestEvents.Envelope(TestEvents.Placed());
        await broker.PublishAsync(Topic, envelope, TestContext.Current.CancellationToken);
        // Hold the message in-flight, THEN trigger the drain.
        await processor.Started.WaitAsync(TimeSpan.FromSeconds(5), TestContext.Current.CancellationToken);

        var stopping = host.StopAsync(CancellationToken.None);

        // The host must not finish stopping while the in-flight message is
        // unfinished (the consumer is draining, not abandoning).
        await Task.Delay(TimeSpan.FromMilliseconds(100), TestContext.Current.CancellationToken);
        Assert.False(stopping.IsCompleted, "host stopped before the in-flight message settled");

        processor.Release();
        await stopping;

        Assert.Equal(1, processor.Processed);
        Assert.True(await host.Services.GetRequiredService<InMemoryInboxStore>()
            .SeenAsync(envelope.Id, TestContext.Current.CancellationToken));
        Assert.Equal(0, host.Services.GetRequiredService<InMemoryDeadLetterStore>().Count);
    }

    [Fact]
    public async Task Stop_IsBoundedByShutdownTimeoutWhenAHandlerWedges()
    {
        var processor = new BlockingProcessor();
        using var host = BuildHost(processor, shutdownTimeout: TimeSpan.FromSeconds(1));
        await host.StartAsync(TestContext.Current.CancellationToken);
        var broker = host.Services.GetRequiredService<InMemoryBroker>();
        await broker.PublishAsync(
            Topic, TestEvents.Envelope(TestEvents.Placed()), TestContext.Current.CancellationToken);
        await processor.Started.WaitAsync(TimeSpan.FromSeconds(5), TestContext.Current.CancellationToken);

        // Never release: the handler is wedged. StopAsync must still return
        // within the ShutdownTimeout budget (plus scheduling slack), because
        // the host abandons the drain at the deadline.
        var stopwatch = Stopwatch.StartNew();
        try
        {
            await host.StopAsync(CancellationToken.None);
        }
        catch (OperationCanceledException)
        {
            // Some host versions surface the abandoned drain as a cancellation;
            // either way the bound below is what matters.
        }

        stopwatch.Stop();
        Assert.True(
            stopwatch.Elapsed < TimeSpan.FromSeconds(8),
            $"drain was not bounded by ShutdownTimeout (took {stopwatch.Elapsed})");

        processor.Release(); // let the wedged task finish so the host disposes cleanly
    }

    /// <summary>A production-shaped host: the same composition extension
    /// Program.cs uses, with the processor swapped behind its Core port and a
    /// test-sized drain budget.</summary>
    private static IHost BuildHost(IOrderEventProcessor processor, TimeSpan shutdownTimeout)
    {
        var builder = Microsoft.Extensions.Hosting.Host.CreateApplicationBuilder();
        builder.Configuration.AddInMemoryCollection(new Dictionary<string, string?>
        {
            ["Messaging:Topic"] = Topic,
            ["Consumer:MaxAttempts"] = "3",
            ["Consumer:BaseBackoff"] = "00:00:00.010",
            ["Consumer:MaxBackoff"] = "00:00:00.050",
            ["Outbox:PollInterval"] = "00:00:00.200",
            ["Outbox:BatchSize"] = "10",
        });
        builder.Logging.ClearProviders();
        builder.Services.AddSingleton(TimeProvider.System);
        builder.Services.AddOrdersWorkerMessaging(builder.Configuration);
        // Last registration wins for single resolution: the blocking processor
        // replaces the projector behind the same port.
        builder.Services.AddSingleton(processor);
        builder.Services.Configure<HostOptions>(options => options.ShutdownTimeout = shutdownTimeout);
        return builder.Build();
    }
}
