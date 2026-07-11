using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

using Orders.Worker.Core.Events;
using Orders.Worker.Infrastructure.Health;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// The consume loop, composed as a BackgroundService per
/// csharp/services/eventing-and-messaging.md (Consumer Rules): it streams
/// deliveries from the broker-neutral <see cref="IMessageSource"/> port and
/// hands each to <see cref="OrderEventHandler"/> in its own DI scope - the
/// scope lives exactly as long as the unit of work (a database-backed
/// processor gets its scoped DbContext from this).
///
/// Graceful drain contract: cancelling <c>stoppingToken</c> stops pulling NEW
/// messages (the source's stream ends), but a message already dequeued is
/// finished and settled before ExecuteAsync returns. The host bounds the whole
/// drain with HostOptions.ShutdownTimeout - the drain test proves both halves.
/// The <see cref="OperationCanceledException"/> from the stream is orderly
/// shutdown, not an error, and is the one place it is deliberately caught.
/// </summary>
internal sealed partial class OrderEventsConsumer(
    IMessageSource source,
    IServiceScopeFactory scopeFactory,
    WorkerReadiness readiness,
    ILogger<OrderEventsConsumer> logger) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        // Yield so host startup is not blocked waiting on the first delivery.
        await Task.Yield();
        readiness.SetReady(true);
        Log.Started(logger);
        try
        {
            await foreach (var message in source.ReadAllAsync(stoppingToken).ConfigureAwait(false))
            {
                await using var scope = scopeFactory.CreateAsyncScope();
                var handler = scope.ServiceProvider.GetRequiredService<OrderEventHandler>();
                await handler.HandleAsync(message, stoppingToken).ConfigureAwait(false);
            }
        }
        catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
        {
            // Orderly shutdown: the drain token ended the subscription stream.
        }
        finally
        {
            readiness.SetReady(false);
            Log.Stopped(logger);
        }
    }

    private static partial class Log
    {
        [LoggerMessage(Level = LogLevel.Information, Message = "Order events consumer started")]
        public static partial void Started(ILogger logger);

        [LoggerMessage(Level = LogLevel.Information, Message = "Order events consumer stopped")]
        public static partial void Stopped(ILogger logger);
    }
}
