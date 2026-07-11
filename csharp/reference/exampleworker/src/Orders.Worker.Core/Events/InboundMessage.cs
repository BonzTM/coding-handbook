namespace Orders.Worker.Core.Events;

/// <summary>
/// One delivery from the broker: the envelope plus the settlement handles the
/// broker implementation bound to this delivery. The consumer calls exactly one
/// of <see cref="AckAsync"/> or <see cref="NackAsync"/> per message - Ack after
/// the terminal decision (durable side effect, duplicate drop, or dead-letter),
/// Nack to return the message for redelivery. The handles are delegates
/// supplied by the broker adapter so the consumer never depends on a concrete
/// broker type (csharp/services/eventing-and-messaging.md).
/// </summary>
public sealed class InboundMessage
{
    private readonly Func<ValueTask>? _ack;
    private readonly Func<ValueTask>? _nack;

    public InboundMessage(EventEnvelope envelope, Func<ValueTask>? ack = null, Func<ValueTask>? nack = null)
    {
        ArgumentNullException.ThrowIfNull(envelope);
        Envelope = envelope;
        _ack = ack;
        _nack = nack;
    }

    public EventEnvelope Envelope { get; }

    /// <summary>Settles the delivery as handled. Safe to call when the broker
    /// bound no handle (a test message): it is then a no-op.</summary>
    public ValueTask AckAsync() => _ack?.Invoke() ?? ValueTask.CompletedTask;

    /// <summary>Returns the delivery to the broker for redelivery
    /// (at-least-once semantics).</summary>
    public ValueTask NackAsync() => _nack?.Invoke() ?? ValueTask.CompletedTask;
}
