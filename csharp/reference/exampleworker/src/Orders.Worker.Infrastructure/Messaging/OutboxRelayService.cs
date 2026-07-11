using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// The publisher-side BackgroundService: scans the outbox on a TimeProvider-
/// driven PeriodicTimer and drains one batch per tick through
/// <see cref="OutboxRelay"/>. On shutdown it performs one FINAL flush on a
/// detached, PollInterval-bounded token, so a record enqueued just before the
/// drain still publishes within the shutdown budget - the relay-service test
/// proves it (csharp/services/eventing-and-messaging.md).
/// </summary>
internal sealed partial class OutboxRelayService(
    OutboxRelay relay,
    TimeProvider time,
    IOptions<OutboxOptions> options,
    ILogger<OutboxRelayService> logger) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        var interval = options.Value.PollInterval;
        Log.Started(logger, interval);
        try
        {
            using var timer = new PeriodicTimer(interval, time);
            while (await timer.WaitForNextTickAsync(stoppingToken).ConfigureAwait(false))
            {
                await relay.FlushAsync(stoppingToken).ConfigureAwait(false);
            }
        }
        catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
        {
            // Orderly shutdown; fall through to the final flush.
        }

        await FinalFlushAsync(interval).ConfigureAwait(false);
        Log.Stopped(logger);
    }

    /// <summary>One last drain on a fresh, bounded token - the stopping token
    /// is already cancelled and must not abort work we still owe the outbox.</summary>
    private async Task FinalFlushAsync(TimeSpan budget)
    {
        using var cts = new CancellationTokenSource(budget, time);
        try
        {
            await relay.FlushAsync(cts.Token).ConfigureAwait(false);
        }
        catch (OperationCanceledException) when (cts.IsCancellationRequested)
        {
            Log.FinalFlushIncomplete(logger, budget);
        }
    }

    private static partial class Log
    {
        [LoggerMessage(Level = LogLevel.Information,
            Message = "Outbox relay started, poll interval {PollInterval}")]
        public static partial void Started(ILogger logger, TimeSpan pollInterval);

        [LoggerMessage(Level = LogLevel.Warning,
            Message = "Final outbox flush exceeded its budget ({Budget}); remaining records stay pending")]
        public static partial void FinalFlushIncomplete(ILogger logger, TimeSpan budget);

        [LoggerMessage(Level = LogLevel.Information, Message = "Outbox relay stopped")]
        public static partial void Stopped(ILogger logger);
    }
}
