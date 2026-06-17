package messaging

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/example/exampleworker/internal/core"
	"github.com/example/exampleworker/internal/telemetry"
)

// OutboxRecord is one pending publish written transactionally alongside the
// domain state change, per golang/services/eventing-and-messaging.md ### Outbox
// And Inbox Patterns ("write business state and an outbox row in one DB
// transaction, then relay that row to the broker asynchronously"). The relay
// reads pending records, publishes them, and marks them sent — so a process
// that commits state but crashes before publishing still publishes on the next
// relay scan (no dual-write loss).
type OutboxRecord struct {
	// ID is the stable record/event id; it becomes the published Message.ID so
	// the consumer's inbox can dedupe a redelivered relay.
	ID string
	// Topic is the destination topic.
	Topic string
	// Type is the event type.
	Type string
	// Body is the encoded payload.
	Body []byte
	// CreatedAt is when the record was enqueued (producer time).
	CreatedAt time.Time
	// SentAt is set when the relay confirms publication. Zero means pending.
	SentAt time.Time
}

// OutboxStore is the pending-message store contract the relay drains. Defined
// at the consumer; the in-memory implementation satisfies it and a SQL outbox
// table (SELECT ... WHERE sent_at IS NULL ... FOR UPDATE SKIP LOCKED) plugs in
// unchanged.
type OutboxStore interface {
	// Add enqueues a pending record. In a DB build this is the INSERT that runs
	// in the same transaction as the domain write.
	Add(ctx context.Context, rec OutboxRecord) error
	// Pending returns up to limit unsent records in enqueue order.
	Pending(ctx context.Context, limit int) ([]OutboxRecord, error)
	// MarkSent records that id was published at sentAt so it is not relayed
	// again.
	MarkSent(ctx context.Context, id string, sentAt time.Time) error
}

// MemoryOutbox is an in-memory OutboxStore for offline tests and local dev. It
// preserves enqueue order and is safe for concurrent use.
type MemoryOutbox struct {
	mu      sync.Mutex
	order   []string
	records map[string]OutboxRecord
}

// NewMemoryOutbox constructs an empty outbox.
func NewMemoryOutbox() *MemoryOutbox {
	return &MemoryOutbox{records: make(map[string]OutboxRecord)}
}

// Add enqueues a pending record, preserving first-seen order.
func (o *MemoryOutbox) Add(ctx context.Context, rec OutboxRecord) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if _, ok := o.records[rec.ID]; !ok {
		o.order = append(o.order, rec.ID)
	}
	o.records[rec.ID] = rec
	return nil
}

// Pending returns up to limit unsent records in enqueue order.
func (o *MemoryOutbox) Pending(ctx context.Context, limit int) ([]OutboxRecord, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	out := make([]OutboxRecord, 0, limit)
	for _, id := range o.order {
		rec := o.records[id]
		if !rec.SentAt.IsZero() {
			continue
		}
		out = append(out, rec)
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

// MarkSent stamps a record as published.
func (o *MemoryOutbox) MarkSent(ctx context.Context, id string, sentAt time.Time) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	rec, ok := o.records[id]
	if !ok {
		return fmt.Errorf("outbox: mark sent: unknown id %q", id)
	}
	rec.SentAt = sentAt
	o.records[id] = rec
	return nil
}

// PendingCount reports how many records are still unsent (for assertions).
func (o *MemoryOutbox) PendingCount() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	n := 0
	for _, id := range o.order {
		if o.records[id].SentAt.IsZero() {
			n++
		}
	}
	return n
}

var _ OutboxStore = (*MemoryOutbox)(nil)

// Relay is the transactional-outbox relay: it periodically drains pending
// records from the OutboxStore, publishes each to the broker, and marks it sent
// only AFTER a successful publish. A publish failure leaves the record pending
// so the next scan retries it — reliable publish without a dual write. The
// relay is bound to a context and is the errgroup member for the relay loop.
type Relay struct {
	store     OutboxStore
	broker    Broker
	clock     core.Clock
	metrics   telemetry.Metrics
	logger    *slog.Logger
	interval  time.Duration
	batchSize int
}

// RelayDeps bundles the relay's collaborators.
type RelayDeps struct {
	Store   OutboxStore
	Broker  Broker
	Clock   core.Clock
	Metrics telemetry.Metrics
	Logger  *slog.Logger
}

// NewRelay constructs a Relay. interval is the scan period; batchSize caps rows
// claimed per scan. Required deps must be non-nil.
func NewRelay(interval time.Duration, batchSize int, deps RelayDeps) *Relay {
	var metrics telemetry.Metrics = telemetry.NopMetrics{}
	if deps.Metrics != nil {
		metrics = deps.Metrics
	}
	return &Relay{
		store:     deps.Store,
		broker:    deps.Broker,
		clock:     deps.Clock,
		metrics:   metrics,
		logger:    deps.Logger,
		interval:  interval,
		batchSize: batchSize,
	}
}

// Run scans on a ticker until ctx is cancelled, then performs one final flush so
// records enqueued just before shutdown are published within the drain budget.
func (r *Relay) Run(ctx context.Context) error {
	r.logger.InfoContext(ctx, "outbox relay started", "interval", r.interval.String())
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Final flush on a detached, bounded context so a record enqueued just
			// before shutdown still publishes. main orders this within the drain.
			flushCtx, cancel := context.WithTimeout(context.Background(), r.interval)
			_, err := r.Flush(flushCtx)
			cancel()
			if err != nil {
				r.logger.WarnContext(ctx, "final outbox flush incomplete", "error", err.Error())
			}
			r.logger.InfoContext(ctx, "outbox relay stopped")
			return nil
		case <-ticker.C:
			if _, err := r.Flush(ctx); err != nil {
				// A scan error is logged and retried next tick; it is not fatal to
				// the worker.
				r.logger.WarnContext(ctx, "outbox scan failed", "error", err.Error())
			}
		}
	}
}

// Flush publishes one batch of pending records and marks each sent. It returns
// the number successfully published. It is exported so tests can drive a single
// deterministic relay cycle and main can flush on shutdown. A publish failure
// stops the batch (preserving order) and leaves the rest pending for the next
// scan.
func (r *Relay) Flush(ctx context.Context) (int, error) {
	pending, err := r.store.Pending(ctx, r.batchSize)
	if err != nil {
		return 0, fmt.Errorf("read pending: %w", err)
	}
	published := 0
	for _, rec := range pending {
		msg := Message{
			ID:     rec.ID,
			Type:   rec.Type,
			Source: "exampleworker",
			Time:   rec.CreatedAt,
			Body:   rec.Body,
		}
		if perr := r.broker.Publish(ctx, rec.Topic, msg); perr != nil {
			// Leave this and the rest pending; retry next scan. Returning the
			// count already published lets a test assert partial progress.
			return published, fmt.Errorf("publish %q: %w", rec.ID, perr)
		}
		if merr := r.store.MarkSent(ctx, rec.ID, r.clock.Now().UTC()); merr != nil {
			return published, fmt.Errorf("mark sent %q: %w", rec.ID, merr)
		}
		r.metrics.IncPublished(rec.Type)
		published++
	}
	return published, nil
}
