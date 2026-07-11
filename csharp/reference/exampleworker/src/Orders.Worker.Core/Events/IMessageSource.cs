namespace Orders.Worker.Core.Events;

/// <summary>
/// Consumer port - the broker-neutral seam from
/// csharp/services/eventing-and-messaging.md. The consume loop iterates this
/// stream and never touches a broker client type. At-least-once delivery is
/// assumed: the same envelope id may be yielded more than once, which is why
/// the consumer dedupes through <see cref="Messaging.IInboxStore"/>.
/// </summary>
public interface IMessageSource
{
    /// <summary>Streams deliveries until <paramref name="cancellationToken"/>
    /// is cancelled (orderly shutdown: stop pulling new work) or the broker
    /// closes the subscription (the stream completes).</summary>
    IAsyncEnumerable<InboundMessage> ReadAllAsync(CancellationToken cancellationToken);
}
