using System.Runtime.CompilerServices;
using System.Threading.Channels;

using Orders.Worker.Core.Events;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// In-memory, channel-backed broker for offline tests and local development.
/// It models at-least-once delivery honestly: a nacked message is re-enqueued
/// for redelivery, so the consumer's idempotency (inbox dedupe) and retry
/// paths are exercised without any external infrastructure. A real broker
/// client (NATS JetStream / RabbitMQ per csharp/services/eventing-and-messaging.md)
/// replaces the two thin adapters over this class - the consumer, relay, and
/// Core never change.
///
/// One bounded channel per topic models a single competing-consumer queue.
/// Publishers wait when the channel is full (backpressure, never unbounded
/// growth); a nack re-enqueue drops instead of blocking the consumer when the
/// buffer is full, the same trade the Go reference documents - a real broker
/// applies its own redelivery/visibility policy there.
/// </summary>
public sealed class InMemoryBroker(int capacity = 256)
{
    private readonly Lock _gate = new();
    private readonly Dictionary<string, Channel<EventEnvelope>> _topics = new(StringComparer.Ordinal);
    private bool _closed;
    private bool _healthy = true;

    /// <summary>Broker connectivity for the /readyz probe. Tests flip it with
    /// <see cref="SetHealthy"/>; a real client reports its connection state.</summary>
    public bool IsHealthy
    {
        get
        {
            lock (_gate)
            {
                return _healthy && !_closed;
            }
        }
    }

    /// <summary>Enqueues an envelope on the topic, waiting for space when the
    /// bounded channel is full.</summary>
    /// <exception cref="InvalidOperationException">The broker is closed.</exception>
    public async Task PublishAsync(string topic, EventEnvelope envelope, CancellationToken cancellationToken)
    {
        ArgumentException.ThrowIfNullOrEmpty(topic);
        ArgumentNullException.ThrowIfNull(envelope);
        await TopicChannel(topic).Writer.WriteAsync(envelope, cancellationToken).ConfigureAwait(false);
    }

    /// <summary>
    /// Streams deliveries for the topic until <paramref name="cancellationToken"/>
    /// is cancelled or the broker is closed (the stream then completes). Each
    /// delivery's Ack is a no-op (a dequeued message is settled once handled);
    /// Nack re-enqueues the same envelope, modeling at-least-once redelivery.
    /// </summary>
    public async IAsyncEnumerable<InboundMessage> SubscribeAsync(
        string topic, [EnumeratorCancellation] CancellationToken cancellationToken)
    {
        ArgumentException.ThrowIfNullOrEmpty(topic);
        var reader = TopicChannel(topic).Reader;
        while (await reader.WaitToReadAsync(cancellationToken).ConfigureAwait(false))
        {
            while (reader.TryRead(out var envelope))
            {
                var captured = envelope;
                yield return new InboundMessage(
                    captured,
                    ack: static () => ValueTask.CompletedTask,
                    nack: () => RequeueAsync(topic, captured));
            }
        }
    }

    /// <summary>Queued (not yet delivered) messages on a topic - the consumer
    /// lag the metrics gauge reports. A real deployment reads lag from the
    /// broker's own metrics instead.</summary>
    public int Depth(string topic)
    {
        ArgumentException.ThrowIfNullOrEmpty(topic);
        lock (_gate)
        {
            return _topics.TryGetValue(topic, out var channel) ? channel.Reader.Count : 0;
        }
    }

    /// <summary>Toggles reported connectivity so tests can drive /readyz.</summary>
    public void SetHealthy(bool healthy)
    {
        lock (_gate)
        {
            _healthy = healthy;
        }
    }

    /// <summary>Marks the broker closed and completes every topic channel so
    /// subscribers' streams end. Idempotent. Publishing afterwards throws.</summary>
    public void Close()
    {
        lock (_gate)
        {
            if (_closed)
            {
                return;
            }

            _closed = true;
            foreach (var channel in _topics.Values)
            {
                channel.Writer.TryComplete();
            }
        }
    }

    private Channel<EventEnvelope> TopicChannel(string topic)
    {
        lock (_gate)
        {
            if (_closed)
            {
                throw new InvalidOperationException("The in-memory broker is closed.");
            }

            if (!_topics.TryGetValue(topic, out var channel))
            {
                channel = Channel.CreateBounded<EventEnvelope>(new BoundedChannelOptions(capacity)
                {
                    FullMode = BoundedChannelFullMode.Wait,
                });
                _topics[topic] = channel;
            }

            return channel;
        }
    }

    /// <summary>Re-enqueues a nacked envelope, dropping (never blocking the
    /// consumer) when the broker is closed or the buffer is full.</summary>
    private ValueTask RequeueAsync(string topic, EventEnvelope envelope)
    {
        lock (_gate)
        {
            if (!_closed && _topics.TryGetValue(topic, out var channel))
            {
                channel.Writer.TryWrite(envelope);
            }
        }

        return ValueTask.CompletedTask;
    }
}
