package messaging

import (
	"context"
	"sync"
	"time"
)

// DeadLetter is a parked message that exhausted its retry budget or failed
// non-retryable validation. It retains enough context for operator-controlled
// replay, per golang/services/eventing-and-messaging.md ### Retries And
// Dead-Letter Behavior ("retain original destination, attempt count, failure
// class, and correlation data").
type DeadLetter struct {
	// Message is the original delivery (ID, type, body, metadata).
	Message Message
	// Topic is the original destination the message arrived on.
	Topic string
	// Attempts is how many delivery attempts were made before parking.
	Attempts int
	// FailureClass is "invalid" (non-retryable validation/schema) or
	// "exhausted" (transient retries exhausted).
	FailureClass string
	// Reason is the last error message, for operator visibility.
	Reason string
	// DeadLetteredAt is when the message was parked (from the injected clock).
	DeadLetteredAt time.Time
}

// Failure classes.
const (
	FailureInvalid   = "invalid"
	FailureExhausted = "exhausted"
)

// DeadLetterStore is the parking contract the consumer writes to after the
// retry budget is exhausted or a non-retryable failure occurs. Defined at the
// consumer; the in-memory implementation satisfies it and a DLQ topic / table
// plugs in unchanged.
type DeadLetterStore interface {
	// Add parks a dead-lettered message.
	Add(ctx context.Context, dl DeadLetter) error
}

// MemoryDLQ is an in-memory DeadLetterStore for offline tests and local dev. It
// is safe for concurrent use.
type MemoryDLQ struct {
	mu      sync.Mutex
	entries []DeadLetter
}

// NewMemoryDLQ constructs an empty dead-letter store.
func NewMemoryDLQ() *MemoryDLQ {
	return &MemoryDLQ{}
}

// Add parks a dead-lettered message.
func (q *MemoryDLQ) Add(ctx context.Context, dl DeadLetter) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	q.entries = append(q.entries, dl)
	return nil
}

// Entries returns a copy of the parked messages for assertions and operator
// tooling.
func (q *MemoryDLQ) Entries() []DeadLetter {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]DeadLetter, len(q.entries))
	copy(out, q.entries)
	return out
}

// Len reports the number of parked messages.
func (q *MemoryDLQ) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.entries)
}

var _ DeadLetterStore = (*MemoryDLQ)(nil)
