using Microsoft.Extensions.Options;

using Orders.Worker.Core.Events;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// <see cref="IEventPublisher"/> adapter over the in-memory broker. This file
/// is the ONLY seam a real broker replaces: a NatsEventPublisher or
/// RabbitMqEventPublisher implements the same one-method port and is swapped
/// in the composition root - Core, the consumer, and the relay never change
/// (csharp/services/eventing-and-messaging.md).
/// </summary>
internal sealed class BrokerEventPublisher(InMemoryBroker broker, IOptions<MessagingOptions> options)
    : IEventPublisher
{
    public Task PublishAsync(EventEnvelope envelope, CancellationToken cancellationToken)
        => broker.PublishAsync(options.Value.Topic, envelope, cancellationToken);
}

/// <summary>
/// <see cref="IMessageSource"/> adapter over the in-memory broker: the
/// consumer's subscription to the configured topic.
/// </summary>
internal sealed class BrokerMessageSource(InMemoryBroker broker, IOptions<MessagingOptions> options)
    : IMessageSource
{
    public IAsyncEnumerable<InboundMessage> ReadAllAsync(CancellationToken cancellationToken)
        => broker.SubscribeAsync(options.Value.Topic, cancellationToken);
}
