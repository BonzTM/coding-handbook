package messaging

import (
	"context"
	"sync"
)

// InboxStore is the durable dedupe contract the consumer uses to process each
// message exactly once under at-least-once delivery, per
// golang/services/eventing-and-messaging.md ### Outbox And Inbox Patterns ("use
// an inbox or durable dedupe table keyed by event ID when duplicate delivery
// would be harmful"). It is defined here, at the consumer; the in-memory
// implementation satisfies it and a SQL-backed inbox (insert-on-conflict-do-
// nothing keyed by message id) plugs in unchanged.
type InboxStore interface {
	// MarkProcessed records that the message id has been durably processed. It
	// returns alreadyProcessed=true if the id was already recorded (a duplicate
	// delivery), in which case the consumer drops the message after ack without
	// re-invoking the domain. Recording the id and the side effect must be one
	// logical action; the SQL inbox does both in the same transaction.
	MarkProcessed(ctx context.Context, id string) (alreadyProcessed bool, err error)
	// Seen reports whether an id has been recorded, without recording it. It is
	// a read-only helper for tests and operator tooling.
	Seen(ctx context.Context, id string) (bool, error)
}

// MemoryInbox is an in-memory InboxStore for offline tests and local dev. It is
// safe for concurrent use. The zero value is not usable; call NewMemoryInbox.
type MemoryInbox struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

// NewMemoryInbox constructs an empty inbox.
func NewMemoryInbox() *MemoryInbox {
	return &MemoryInbox{seen: make(map[string]struct{})}
}

// MarkProcessed records id and reports whether it was already present. The
// check-and-record is atomic under the mutex, mirroring the SQL inbox's
// insert-on-conflict-do-nothing so two concurrent deliveries of the same id
// cannot both be treated as first-seen.
func (m *MemoryInbox) MarkProcessed(ctx context.Context, id string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.seen[id]; ok {
		return true, nil
	}
	m.seen[id] = struct{}{}
	return false, nil
}

// remove deletes a dedupe record. The consumer calls it (via an interface
// assertion) to roll back a record after the domain side effect failed, so a
// retry is not mistaken for a duplicate. A SQL inbox rolls back the enclosing
// transaction instead and would not need this hook.
func (m *MemoryInbox) remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.seen, id)
}

// Seen reports whether id has been recorded.
func (m *MemoryInbox) Seen(ctx context.Context, id string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.seen[id]
	return ok, nil
}

var _ InboxStore = (*MemoryInbox)(nil)
