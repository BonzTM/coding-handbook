namespace Orders.Worker.Core.Events;

/// <summary>
/// Publisher port - the broker-neutral seam from
/// csharp/services/eventing-and-messaging.md. Core and the outbox relay depend
/// only on this interface; the broker adapter (in-memory here, NATS/RabbitMQ in
/// production) lives in Orders.Worker.Infrastructure/Messaging and is the only
/// code that knows a wire protocol.
/// </summary>
public interface IEventPublisher
{
    /// <summary>Publishes one envelope. A failure must surface to the caller -
    /// the outbox relay leaves the record pending and retries on the next scan
    /// rather than losing it.</summary>
    Task PublishAsync(EventEnvelope envelope, CancellationToken cancellationToken);
}
