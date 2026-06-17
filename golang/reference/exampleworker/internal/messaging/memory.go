package messaging

import (
	"context"
	"errors"
	"sync"
)

// ErrBrokerClosed is returned by Publish after the in-memory broker is closed.
var ErrBrokerClosed = errors.New("messaging: broker closed")

// MemoryBroker is an in-memory, channel-backed Broker for offline tests and
// local development. It models at-least-once delivery: a Nack'd message is
// re-enqueued for redelivery, so the consumer's idempotency (inbox dedupe) and
// retry paths are exercised without any external infrastructure.
//
// It is safe for concurrent use. The zero value is not usable; call
// NewMemoryBroker.
type MemoryBroker struct {
	mu      sync.Mutex
	topics  map[string]chan Message
	buffer  int
	closed  bool
	healthy bool
}

// MemoryBrokerOption configures a MemoryBroker.
type MemoryBrokerOption func(*MemoryBroker)

// WithBuffer sets the per-topic channel buffer. A larger buffer lets a test
// publish a batch without a consumer draining concurrently.
func WithBuffer(n int) MemoryBrokerOption {
	return func(b *MemoryBroker) { b.buffer = n }
}

// NewMemoryBroker constructs an empty in-memory broker that reports healthy.
func NewMemoryBroker(opts ...MemoryBrokerOption) *MemoryBroker {
	b := &MemoryBroker{
		topics:  make(map[string]chan Message),
		buffer:  64,
		healthy: true,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// topicChan returns the channel for a topic, creating it on first use. The
// caller must hold b.mu.
func (b *MemoryBroker) topicChan(topic string) chan Message {
	ch, ok := b.topics[topic]
	if !ok {
		ch = make(chan Message, b.buffer)
		b.topics[topic] = ch
	}
	return ch
}

// Subscribe returns a receive-only view of the topic channel. Cancelling ctx or
// closing the broker closes the channel, which stops the consume loop. A single
// shared channel per topic models a single consumer group / competing-consumer
// queue.
func (b *MemoryBroker) Subscribe(ctx context.Context, topic string) (<-chan Message, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil, ErrBrokerClosed
	}
	ch := b.topicChan(topic)

	// A relay goroutine closes the delivered channel when ctx is cancelled so the
	// consume loop's range terminates without the broker having to be closed.
	out := make(chan Message, b.buffer)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				select {
				case out <- msg:
				case <-ctx.Done():
					// Cancelled mid-handoff: put the message back on the underlying
					// topic channel (non-blocking) so it is not lost — a redelivery
					// for whoever subscribes next, modeling at-least-once. If the
					// buffer is full it stays only on this side; the consumer's
					// in-flight drain already finished any message it had pulled.
					b.requeue(topic, msg)
					return
				}
			}
		}
	}()
	return out, nil
}

// requeue puts a message back on the underlying topic channel after a cancelled
// handoff, non-blocking so a full buffer never wedges the relay goroutine.
func (b *MemoryBroker) requeue(topic string, msg Message) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	ch := b.topicChan(topic)
	b.mu.Unlock()
	select {
	case ch <- msg:
	default:
	}
}

// Publish enqueues a message on the topic. The returned Message's Ack is a
// no-op (delivery is considered settled once handled) and Nack re-enqueues the
// message for redelivery, modeling at-least-once semantics.
func (b *MemoryBroker) Publish(ctx context.Context, topic string, msg Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return ErrBrokerClosed
	}
	ch := b.topicChan(topic)
	b.mu.Unlock()

	// Bind settlement handles. Nack re-publishes the same message (preserving its
	// ID) so the consumer sees a redelivery; Ack is a no-op for the in-memory
	// queue.
	delivery := msg
	delivery.ack = func() error { return nil }
	delivery.nack = func() error {
		return b.redeliver(topic, msg)
	}

	select {
	case ch <- delivery:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// redeliver re-enqueues a message after a Nack. It drops silently if the broker
// is closed (shutdown drains in-flight; a Nack racing close is not an error).
func (b *MemoryBroker) redeliver(topic string, msg Message) error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	ch := b.topicChan(topic)
	b.mu.Unlock()

	delivery := msg
	delivery.ack = func() error { return nil }
	delivery.nack = func() error { return b.redeliver(topic, msg) }

	select {
	case ch <- delivery:
		return nil
	default:
		// Buffer full: drop to avoid blocking the consumer. A real broker would
		// apply its own redelivery/visibility policy here.
		return nil
	}
}

// SetHealthy toggles the reported health, so tests can drive the /readyz probe.
func (b *MemoryBroker) SetHealthy(v bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.healthy = v
}

// Healthy reports broker connectivity for the worker's readiness probe.
func (b *MemoryBroker) Healthy() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.healthy && !b.closed
}

// Close marks the broker closed and closes every topic channel so subscribers
// terminate. It is idempotent.
func (b *MemoryBroker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil
	}
	b.closed = true
	for _, ch := range b.topics {
		close(ch)
	}
	return nil
}

// Compile-time proof that *MemoryBroker satisfies the consumer-defined seams.
var (
	_ Broker   = (*MemoryBroker)(nil)
	_ Healther = (*MemoryBroker)(nil)
)
