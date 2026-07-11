using Orders.Worker.Core.Events;

namespace Orders.Worker.Core.Messaging;

/// <summary>Why a message was parked (csharp/services/eventing-and-messaging.md,
/// Retries And Dead-Letter Behavior).</summary>
public enum DeadLetterFailureClass
{
    /// <summary>Non-retryable validation/schema failure - replay can never succeed.</summary>
    Invalid,

    /// <summary>Transient failures exhausted the bounded retry budget.</summary>
    Exhausted,
}

/// <summary>
/// A parked message that exhausted its retry budget or failed non-retryable
/// validation. It retains the original envelope, attempt count, failure class,
/// and reason - enough to diagnose and replay without the original process.
/// Replays are operator-controlled, never an automatic feedback loop.
/// </summary>
/// <param name="Envelope">The original delivery (id, type, payload, metadata).</param>
/// <param name="Attempts">Delivery attempts made before parking.</param>
/// <param name="FailureClass">Terminal classification.</param>
/// <param name="Reason">Last error message, for operator visibility - kept out
/// of metric labels (high cardinality) and put here instead.</param>
/// <param name="DeadLetteredAt">When the message was parked, from the injected
/// <see cref="TimeProvider"/>.</param>
public sealed record DeadLetter(
    EventEnvelope Envelope,
    int Attempts,
    DeadLetterFailureClass FailureClass,
    string Reason,
    DateTimeOffset DeadLetteredAt);

/// <summary>
/// Parking port the consumer writes to after giving up on a message. The
/// in-memory implementation satisfies it here; a DLQ topic or table plugs in
/// unchanged. The per-stream DLQ policy (what lands here, who is paged, how
/// replay works) belongs in the runbook.
/// </summary>
public interface IDeadLetterStore
{
    /// <summary>Parks a dead-lettered message.</summary>
    Task AddAsync(DeadLetter deadLetter, CancellationToken cancellationToken);
}
