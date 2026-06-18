package db

import (
	"context"
	"sync"
	"time"

	"github.com/example/exampleservice/internal/core"
)

// MemoryIdempotency is an in-memory core.IdempotencyStore for offline tests and
// the default in-memory build. It is safe for concurrent use. The zero value is
// not usable; call NewMemoryIdempotency.
//
// Entries are TTL-bounded: an entry whose age exceeds the configured TTL is
// treated as absent so an abandoned in-flight key (a process that crashed mid
// request) does not wedge the slot forever and a long-expired completed key can
// be reused. The SQL implementation behind the integration tag enforces the
// same TTL with a stored expires_at column; see idempotency_postgres.go.
type MemoryIdempotency struct {
	ttl time.Duration

	mu      sync.Mutex
	entries map[core.IdempotencyKey]*idempotencyEntry
}

// idempotencyEntry is one tracked key. A nil record means the request is still
// in flight (leased); a non-nil record is the completed, replayable response.
type idempotencyEntry struct {
	requestHash string
	createdAt   time.Time
	record      *core.IdempotencyRecord
}

// Compile-time proof that *MemoryIdempotency satisfies the consumer-defined
// core.IdempotencyStore contract.
var _ core.IdempotencyStore = (*MemoryIdempotency)(nil)

// NewMemoryIdempotency constructs an empty store with the given TTL. A
// non-positive ttl panics: an unbounded idempotency store is a memory leak and
// a configuration error, not a runtime fallback.
func NewMemoryIdempotency(ttl time.Duration) *MemoryIdempotency {
	if ttl <= 0 {
		panic("db: idempotency TTL must be positive")
	}
	return &MemoryIdempotency{
		ttl:     ttl,
		entries: make(map[core.IdempotencyKey]*idempotencyEntry),
	}
}

// expired reports whether e is older than the TTL relative to now.
func (m *MemoryIdempotency) expired(e *idempotencyEntry, now time.Time) bool {
	return now.Sub(e.createdAt) >= m.ttl
}

// Begin implements core.IdempotencyStore. It reserves an unused or expired key,
// replays a completed same-hash key, and reports the in-flight / mismatch
// conflicts the recipe requires.
func (m *MemoryIdempotency) Begin(ctx context.Context, key core.IdempotencyKey, requestHash string, now time.Time) (*core.IdempotencyRecord, bool, error) {
	if err := ctx.Err(); err != nil {
		return nil, false, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.entries[key]
	if ok && !m.expired(e, now) {
		switch {
		case e.record == nil:
			// Reserved but not completed: a concurrent duplicate is still running.
			return nil, false, core.ErrIdempotencyInFlight
		case e.requestHash != requestHash:
			// Completed under a different request body: the key is being misused.
			return nil, false, core.ErrIdempotencyKeyMismatch
		default:
			// Completed with the same body: replay the stored response. Copy the
			// body so a caller mutating the returned slice cannot corrupt the
			// stored record (a later replay must be byte-identical).
			replay := *e.record
			replay.ResponseBody = append([]byte(nil), e.record.ResponseBody...)
			return &replay, false, nil
		}
	}

	// Unused or expired: reserve the slot for this request.
	m.entries[key] = &idempotencyEntry{requestHash: requestHash, createdAt: now}
	return nil, true, nil
}

// Complete implements core.IdempotencyStore: it records the final response for a
// leased key so future duplicates replay it.
func (m *MemoryIdempotency) Complete(ctx context.Context, key core.IdempotencyKey, rec core.IdempotencyRecord, now time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	stored := rec
	// Copy the body so a later mutation of the caller's slice cannot corrupt the
	// replayable record.
	stored.ResponseBody = append([]byte(nil), rec.ResponseBody...)
	m.entries[key] = &idempotencyEntry{
		requestHash: rec.RequestHash,
		createdAt:   now,
		record:      &stored,
	}
	return nil
}

// Release implements core.IdempotencyStore: it abandons a leased key so a failed
// request can be retried rather than staying wedged in flight. Releasing a key
// that has already completed is a no-op (the completed record wins).
func (m *MemoryIdempotency) Release(ctx context.Context, key core.IdempotencyKey) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.entries[key]; ok && e.record == nil {
		delete(m.entries, key)
	}
	return nil
}
