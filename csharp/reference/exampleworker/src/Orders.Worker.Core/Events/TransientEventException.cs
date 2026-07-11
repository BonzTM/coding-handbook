namespace Orders.Worker.Core.Events;

/// <summary>
/// A RETRYABLE dependency failure during event processing (store unavailable,
/// timeout, ...). The consumer retries it with bounded exponential backoff and
/// full jitter, then dead-letters with failure class
/// <see cref="Messaging.DeadLetterFailureClass.Exhausted"/> once the attempt
/// budget is spent. Processors may throw it to mark a failure explicitly
/// transient; any exception that is not <see cref="InvalidEventException"/> or
/// a cancellation gets the same retry treatment.
/// </summary>
public sealed class TransientEventException(string reason) : Exception(reason);
