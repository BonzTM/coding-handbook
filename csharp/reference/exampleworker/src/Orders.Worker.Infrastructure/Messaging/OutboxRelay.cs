using System.Text;

using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

using Orders.Worker.Core.Events;
using Orders.Worker.Core.Messaging;
using Orders.Worker.Infrastructure.Telemetry;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// The transactional-outbox relay's drainable unit: publish one batch of
/// pending records through the <see cref="IEventPublisher"/> port and mark
/// each sent only AFTER a successful publish - reliable publish without a dual
/// write (csharp/services/eventing-and-messaging.md, Outbox And Inbox
/// Patterns). A publish failure stops the batch (preserving order) and leaves
/// the rest pending for the next scan; that deliberate leave-and-retry is why
/// the failure is logged here rather than thrown - this relay IS the caller
/// that can act on it. A crash between publish and mark-sent produces one
/// duplicate, which the consumer's inbox absorbs.
///
/// Separated from the BackgroundService loop so tests drive one deterministic
/// flush at a time.
/// </summary>
internal sealed partial class OutboxRelay(
    IOutboxStore outbox,
    IEventPublisher publisher,
    TimeProvider time,
    IOptions<OutboxOptions> options,
    WorkerMetrics metrics,
    ILogger<OutboxRelay> logger)
{
    /// <summary>Publishes up to one batch of pending records, returning how
    /// many were published and marked sent.</summary>
    public async Task<int> FlushAsync(CancellationToken cancellationToken)
    {
        var pending = await outbox
            .GetPendingAsync(options.Value.BatchSize, cancellationToken).ConfigureAwait(false);
        int published = 0;
        foreach (var record in pending)
        {
            if (!await PublishOneAsync(record, cancellationToken).ConfigureAwait(false))
            {
                break;
            }

            published++;
        }

        return published;
    }

    private async Task<bool> PublishOneAsync(OutboxRecord record, CancellationToken cancellationToken)
    {
        var envelope = new EventEnvelope(
            Id: record.Id,
            Type: record.Type,
            Source: WorkerTelemetry.ServiceName,
            Time: record.OccurredAt,
            Subject: record.Subject,
            DataContentType: "application/json",
            Data: Encoding.UTF8.GetBytes(record.Payload));
        try
        {
            await publisher.PublishAsync(envelope, cancellationToken).ConfigureAwait(false);
            await outbox.MarkSentAsync(record.Id, time.GetUtcNow(), cancellationToken).ConfigureAwait(false);
        }
        catch (OperationCanceledException) when (cancellationToken.IsCancellationRequested)
        {
            throw; // Shutdown bounds the final flush; the record stays pending.
        }
#pragma warning disable CA1031 // The relay is the boundary that converts a publish failure into retry-next-scan.
        catch (Exception ex)
#pragma warning restore CA1031
        {
            Log.PublishFailed(logger, record.Id, record.Type, ex);
            return false;
        }

        metrics.Published(record.Type);
        return true;
    }

    private static partial class Log
    {
        [LoggerMessage(Level = LogLevel.Warning,
            Message = "Outbox publish failed for record {RecordId} ({EventType}); left pending for the next scan")]
        public static partial void PublishFailed(ILogger logger, Guid recordId, string eventType, Exception exception);
    }
}
