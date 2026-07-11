using System.Text.Json;

using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

using Orders.Worker.Core.Events;
using Orders.Worker.Core.Messaging;
using Orders.Worker.Core.Orders;
using Orders.Worker.Infrastructure.Telemetry;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// Processes ONE delivery to a terminal outcome, implementing the consumer
/// rules from csharp/services/eventing-and-messaging.md: decode -> dedupe
/// (inbox) -> process -> settle. Ack happens only after the terminal decision
/// (durable side effect, duplicate drop, or dead-letter), never after parsing.
///
/// Failure classification:
///   - decode/validation failure  -> non-retryable, dead-letter immediately;
///   - cancellation (drain)       -> nack for redelivery, never dead-letter;
///   - anything else              -> transient, retry in-place with bounded
///     exponential backoff + full jitter, then dead-letter when the budget is
///     exhausted.
///
/// Terminal settlement (ack, inbox compensation, DLQ write) runs on
/// CancellationToken.None on purpose: a decision already made must not be
/// abandoned mid-settlement by the drain token - the caller bounds the whole
/// drain with HostOptions.ShutdownTimeout.
/// </summary>
internal sealed partial class OrderEventHandler(
    IOrderEventProcessor processor,
    IInboxStore inbox,
    IDeadLetterStore deadLetters,
    BackoffPolicy backoff,
    IRetryDelayer delayer,
    TimeProvider time,
    IOptions<ConsumerOptions> options,
    WorkerMetrics metrics,
    ILogger<OrderEventHandler> logger)
{
    public async Task HandleAsync(InboundMessage message, CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(message);
        using var activity = WorkerTelemetry.ActivitySource.StartActivity("orders.event.process");
        activity?.SetTag("messaging.message.id", message.Envelope.Id);
        activity?.SetTag("messaging.message.type", message.Envelope.Type);

        // Decode first. Unparseable bytes will never parse on redelivery:
        // dead-letter immediately, do not touch the retry budget.
        OrderEvent orderEvent;
        try
        {
            orderEvent = Decode(message.Envelope);
        }
        catch (InvalidEventException ex)
        {
            await DeadLetterAsync(message, attempts: 1, DeadLetterFailureClass.Invalid, ex).ConfigureAwait(false);
            return;
        }

        await ProcessWithRetryAsync(message, orderEvent, cancellationToken).ConfigureAwait(false);
    }

    /// <summary>Deserializes the envelope payload through the source-generated
    /// context. Malformed JSON or missing required members surface as the
    /// non-retryable <see cref="InvalidEventException"/>.</summary>
    private static OrderEvent Decode(EventEnvelope envelope)
    {
        try
        {
            return JsonSerializer.Deserialize(envelope.Data.Span, OrderEventJsonContext.Default.OrderEvent)
                ?? throw new InvalidEventException("payload decoded to null");
        }
        catch (JsonException ex)
        {
            throw new InvalidEventException($"payload is not a valid order event: {ex.Message}");
        }
    }

    /// <summary>The bounded retry loop. Attempts are 1-based; MaxAttempts
    /// includes the first delivery, so the loop is statically bounded by
    /// configuration validated at startup.</summary>
    private async Task ProcessWithRetryAsync(
        InboundMessage message, OrderEvent orderEvent, CancellationToken cancellationToken)
    {
        int maxAttempts = options.Value.MaxAttempts;
        Exception? lastError = null;
        for (int attempt = 1; attempt <= maxAttempts; attempt++)
        {
            try
            {
                await ProcessOnceAsync(message, orderEvent, cancellationToken).ConfigureAwait(false);
                return;
            }
            catch (OperationCanceledException) when (cancellationToken.IsCancellationRequested)
            {
                // Drain, not a message failure: nack so the message is
                // redelivered later - never lost, never dead-lettered.
                await NackAsync(message).ConfigureAwait(false);
                return;
            }
            catch (InvalidEventException ex)
            {
                await DeadLetterAsync(message, attempt, DeadLetterFailureClass.Invalid, ex).ConfigureAwait(false);
                return;
            }
#pragma warning disable CA1031 // The retry loop is the boundary that classifies unknown failures as transient.
            catch (Exception ex)
#pragma warning restore CA1031
            {
                lastError = ex;
                if (attempt < maxAttempts && !await WaitBeforeRetryAsync(message, attempt, ex, cancellationToken).ConfigureAwait(false))
                {
                    return;
                }
            }
        }

        // Budget exhausted on transient failures.
        await DeadLetterAsync(message, maxAttempts, DeadLetterFailureClass.Exhausted, lastError).ConfigureAwait(false);
    }

    /// <summary>
    /// One attempt: inbox dedupe guard, then the domain processor. A duplicate
    /// (id already recorded) acks and drops WITHOUT re-invoking the processor -
    /// the exactly-once-processing guarantee under at-least-once delivery. The
    /// dedupe record is written before the side effect; if the side effect
    /// fails the record is compensated away so a retry is not mistaken for a
    /// duplicate. A SQL inbox gets both behaviors from one transaction that
    /// commits or rolls back atomically (csharp/services/eventing-and-messaging.md).
    /// </summary>
    private async Task ProcessOnceAsync(
        InboundMessage message, OrderEvent orderEvent, CancellationToken cancellationToken)
    {
        bool alreadyProcessed = await inbox
            .MarkProcessedAsync(message.Envelope.Id, cancellationToken).ConfigureAwait(false);
        if (alreadyProcessed)
        {
            metrics.Consumed(message.Envelope.Type, WorkerMetrics.Outcomes.DroppedDuplicate);
            Log.DuplicateDropped(logger, message.Envelope.Id, message.Envelope.Type);
            await AckAsync(message).ConfigureAwait(false);
            return;
        }

        try
        {
            await processor.ProcessAsync(orderEvent, cancellationToken).ConfigureAwait(false);
        }
        catch
        {
            await CompensateInboxAsync(message.Envelope.Id).ConfigureAwait(false);
            throw;
        }

        metrics.Consumed(message.Envelope.Type, WorkerMetrics.Outcomes.Ack);
        await AckAsync(message).ConfigureAwait(false);
    }

    /// <summary>Waits the jittered backoff before the next attempt. Returns
    /// false when the drain cancelled the wait - the message is nacked for
    /// redelivery and the caller stops retrying.</summary>
    private async Task<bool> WaitBeforeRetryAsync(
        InboundMessage message, int attempt, Exception cause, CancellationToken cancellationToken)
    {
        var delay = backoff.NextDelay(attempt);
        metrics.Consumed(message.Envelope.Type, WorkerMetrics.Outcomes.Retry);
        Log.Retrying(logger, message.Envelope.Id, message.Envelope.Type, attempt, delay, cause.Message);
        try
        {
            await delayer.DelayAsync(delay, cancellationToken).ConfigureAwait(false);
            return true;
        }
        catch (OperationCanceledException) when (cancellationToken.IsCancellationRequested)
        {
            await NackAsync(message).ConfigureAwait(false);
            return false;
        }
    }

    /// <summary>Parks the message with its attempt count, failure class, and
    /// reason, then ACKS it: the consumer has terminally given up, so the
    /// broker must not redeliver it (that would loop forever). A DLQ write
    /// failure is logged and the message still acked for the same reason.</summary>
    private async Task DeadLetterAsync(
        InboundMessage message, int attempts, DeadLetterFailureClass failureClass, Exception? cause)
    {
        var deadLetter = new DeadLetter(
            message.Envelope, attempts, failureClass, cause?.Message ?? "", time.GetUtcNow());
        try
        {
            await deadLetters.AddAsync(deadLetter, CancellationToken.None).ConfigureAwait(false);
        }
#pragma warning disable CA1031 // Terminal settlement must not throw past the consume loop.
        catch (Exception ex)
#pragma warning restore CA1031
        {
            Log.DeadLetterWriteFailed(logger, message.Envelope.Id, ex);
        }

        metrics.Consumed(message.Envelope.Type, WorkerMetrics.Outcomes.DeadLettered);
        Log.DeadLettered(
            logger, message.Envelope.Id, message.Envelope.Type, attempts, failureClass, deadLetter.Reason);
        await AckAsync(message).ConfigureAwait(false);
    }

    private async Task CompensateInboxAsync(Guid eventId)
    {
        if (inbox is IInboxCompensation compensation)
        {
            await compensation.RemoveAsync(eventId, CancellationToken.None).ConfigureAwait(false);
        }
    }

    private async Task AckAsync(InboundMessage message)
    {
        try
        {
            await message.AckAsync().ConfigureAwait(false);
        }
#pragma warning disable CA1031 // Settlement failures are logged, never thrown past the loop.
        catch (Exception ex)
#pragma warning restore CA1031
        {
            Log.SettleFailed(logger, "ack", message.Envelope.Id, ex);
        }
    }

    private async Task NackAsync(InboundMessage message)
    {
        try
        {
            await message.NackAsync().ConfigureAwait(false);
        }
#pragma warning disable CA1031 // Settlement failures are logged, never thrown past the loop.
        catch (Exception ex)
#pragma warning restore CA1031
        {
            Log.SettleFailed(logger, "nack", message.Envelope.Id, ex);
        }
    }

    /// <summary>Source-generated log methods with stable fields for every
    /// delivery-state transition (CA1848;
    /// csharp/foundations/errors-and-logging.md).</summary>
    private static partial class Log
    {
        [LoggerMessage(Level = LogLevel.Information,
            Message = "Duplicate delivery dropped: {MessageId} ({EventType})")]
        public static partial void DuplicateDropped(ILogger logger, Guid messageId, string eventType);

        [LoggerMessage(Level = LogLevel.Warning,
            Message = "Retrying message {MessageId} ({EventType}), attempt {Attempt}, backoff {Backoff}: {Reason}")]
        public static partial void Retrying(
            ILogger logger, Guid messageId, string eventType, int attempt, TimeSpan backoff, string reason);

        [LoggerMessage(Level = LogLevel.Warning,
            Message = "Message dead-lettered: {MessageId} ({EventType}), attempts {Attempts}, class {FailureClass}: {Reason}")]
        public static partial void DeadLettered(
            ILogger logger, Guid messageId, string eventType, int attempts,
            DeadLetterFailureClass failureClass, string reason);

        [LoggerMessage(Level = LogLevel.Error,
            Message = "Dead-letter write failed for message {MessageId}")]
        public static partial void DeadLetterWriteFailed(ILogger logger, Guid messageId, Exception exception);

        [LoggerMessage(Level = LogLevel.Error,
            Message = "Message settlement ({Settlement}) failed for {MessageId}")]
        public static partial void SettleFailed(ILogger logger, string settlement, Guid messageId, Exception exception);
    }
}
