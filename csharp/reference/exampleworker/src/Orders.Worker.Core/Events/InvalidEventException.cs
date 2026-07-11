namespace Orders.Worker.Core.Events;

/// <summary>
/// A NON-RETRYABLE validation or schema failure: the payload can never be
/// processed successfully, no matter how often it is redelivered. The consumer
/// dead-letters the message immediately with failure class
/// <see cref="Messaging.DeadLetterFailureClass.Invalid"/> instead of retrying
/// (csharp/services/eventing-and-messaging.md, Retries And Dead-Letter
/// Behavior). Any other exception from processing is treated as transient and
/// retried with bounded backoff.
/// </summary>
public sealed class InvalidEventException(string reason) : Exception(reason);
