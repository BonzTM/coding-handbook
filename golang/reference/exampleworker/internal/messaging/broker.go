// Package messaging holds the broker abstraction and the delivery logic that
// sits between a message broker and the core domain: a consume loop with
// idempotent processing, bounded retry, and a dead-letter path, plus the
// transactional-outbox relay for reliable publish.
//
// A specific broker (Kafka, NATS, RabbitMQ, SQS, ...) is an ADR/framework
// selection decision per golang/services/eventing-and-messaging.md, so this
// package demonstrates the SHAPE behind a Broker interface. The in-memory
// implementation (memory.go) makes the whole flow offline-testable; a real
// broker client plugs into the same interface with no change to the consumer or
// relay.
package messaging

import (
	"context"
	"time"
)

// Message is one delivery from the broker. It carries a stable ID used for
// dedupe and traceability and a CloudEvents-style metadata envelope, per
// golang/services/eventing-and-messaging.md ### Contract Source And Envelope.
// The Body is the raw encoded payload; the consumer decodes it into a domain
// type before processing.
//
// Ack and Nack settle the delivery. The consumer calls exactly one of them per
// message: Ack after the durable side effect succeeds, Nack to return the
// message for redelivery. They are supplied by the Broker implementation (a
// closure capturing the broker-specific delivery handle) so the consumer never
// depends on a concrete broker.
type Message struct {
	// ID is the stable message/event identifier. The inbox dedupe store is keyed
	// by it so duplicate delivery is processed exactly once.
	ID string
	// Type is the event type (low-cardinality), used for routing and metrics.
	Type string
	// Source is the producing service/subsystem.
	Source string
	// Subject is the optional entity/aggregate identifier (the ordering key).
	Subject string
	// Time is the producer timestamp.
	Time time.Time
	// Body is the raw encoded payload (e.g. JSON).
	Body []byte

	// ack settles the message as successfully handled.
	ack func() error
	// nack returns the message for redelivery.
	nack func() error
}

// Ack settles the message as successfully handled. It is safe to call once; the
// Broker implementation decides redelivery semantics for a double-settle.
func (m Message) Ack() error {
	if m.ack == nil {
		return nil
	}
	return m.ack()
}

// Nack returns the message to the broker for redelivery.
func (m Message) Nack() error {
	if m.nack == nil {
		return nil
	}
	return m.nack()
}

// Broker is the broker-neutral contract the consumer and relay depend on. It is
// intentionally small: subscribe to a topic to receive a channel of messages,
// and publish a message to a topic. Implementations own the wire protocol,
// delivery guarantees (at-least-once is assumed), and settlement handles.
type Broker interface {
	// Subscribe returns a receive-only channel of messages for the topic. The
	// channel is closed when ctx is cancelled or the broker is closed, which is
	// how the consume loop learns to stop pulling new work. At-least-once
	// delivery is assumed, so the same message ID may arrive more than once.
	Subscribe(ctx context.Context, topic string) (<-chan Message, error)
	// Publish sends a message to the topic. It returns an error if the broker is
	// unreachable; the outbox relay retries on the next scan rather than losing
	// the row.
	Publish(ctx context.Context, topic string, msg Message) error
}

// Healther is the optional seam a Broker can satisfy to report connectivity for
// the worker's /readyz probe. The in-memory broker implements it; a real client
// reports its connection state. A Broker that does not implement it is treated
// as always healthy.
type Healther interface {
	// Healthy reports whether the broker connection is usable.
	Healthy() bool
}
