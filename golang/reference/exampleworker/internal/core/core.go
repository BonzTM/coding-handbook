// Package core holds the message-processing domain logic and the contracts it
// consumes. It depends only on the standard library and narrowly scoped seams;
// it imports no transport, broker, or storage-implementation packages, per
// golang/foundations/package-design.md.
//
// The domain here is "apply a widget event". The Processor interface is defined
// HERE, in the consumer (the messaging layer), and names what the worker needs;
// implementations (the in-memory WidgetProjector, or a DB-backed one) satisfy
// it structurally.
package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Sentinel errors the messaging layer branches on with errors.Is to classify a
// processing failure as retryable (transient) or not (permanent). Per
// golang/services/eventing-and-messaging.md ### Consumer Rules: validation and
// schema failures are non-retryable and move toward DLQ quickly; transient
// dependency failures retry with bounded backoff.
var (
	// ErrInvalidEvent marks a non-retryable validation/schema failure. The
	// consumer dead-letters it immediately rather than retrying forever.
	ErrInvalidEvent = errors.New("invalid event")
	// ErrTransient marks a retryable dependency failure. The consumer retries it
	// with bounded backoff before dead-lettering.
	ErrTransient = errors.New("transient processing failure")
)

// Clock is the source of time for the worker so behavior is deterministic in
// tests. Production wires a real clock in main; tests pass a fake. The interface
// is kept to the single method the worker actually needs.
type Clock interface {
	// Now returns the current time. Implementations return UTC.
	Now() time.Time
}

// EventType is a stable, low-cardinality event name used for routing,
// classification, and metric labels. It is never a table name or a raw subject.
type EventType string

const (
	// EventWidgetCreated is emitted when a widget is created upstream.
	EventWidgetCreated EventType = "widget.created"
	// EventWidgetDeleted is emitted when a widget is deleted upstream.
	EventWidgetDeleted EventType = "widget.deleted"
)

// WidgetEvent is the decoded domain payload of a widget message. It is a plain
// value type with no wire concerns: the messaging layer decodes the broker
// envelope into this before calling the Processor.
type WidgetEvent struct {
	// Type is the event name (widget.created, widget.deleted).
	Type EventType
	// WidgetID is the aggregate/entity identifier the event concerns. It is the
	// ordering and dedupe key per the eventing doc.
	WidgetID string
	// TenantID scopes the widget to a tenant.
	TenantID string
	// Name is the widget name carried by a create event.
	Name string
	// OccurredAt is the producer timestamp.
	OccurredAt time.Time
}

// Validate enforces the payload invariants. A failure is a non-retryable
// ErrInvalidEvent: replaying a structurally invalid event will never succeed.
func (e WidgetEvent) Validate() error {
	if e.WidgetID == "" {
		return fmt.Errorf("%w: widget id must not be empty", ErrInvalidEvent)
	}
	if e.TenantID == "" {
		return fmt.Errorf("%w: tenant id must not be empty", ErrInvalidEvent)
	}
	switch e.Type {
	case EventWidgetCreated:
		if e.Name == "" {
			return fmt.Errorf("%w: created event must carry a name", ErrInvalidEvent)
		}
	case EventWidgetDeleted:
		// no extra fields required
	default:
		return fmt.Errorf("%w: unknown event type %q", ErrInvalidEvent, e.Type)
	}
	return nil
}

// Processor is the domain-behavior contract the consumer invokes for each
// decoded message. It is intentionally narrow (interface-at-consumer); the
// in-memory WidgetProjector and any DB-backed implementation satisfy it.
//
// Process MUST be idempotent at the domain level where feasible, but the
// consumer ALSO guards it with an inbox/dedupe store keyed by message id so a
// duplicate delivery never invokes Process twice. A returned error wrapping
// ErrInvalidEvent is treated as permanent (dead-letter); anything else (e.g.
// ErrTransient or a raw dependency error) is treated as retryable.
type Processor interface {
	Process(ctx context.Context, e WidgetEvent) error
}

// WidgetProjector is the reference in-memory Processor: it maintains a
// per-(tenant,widget) projection of the latest widget state from the event
// stream. It is safe for concurrent use. Production swaps in a DB-backed
// projector behind the same Processor seam.
type WidgetProjector struct {
	clock Clock

	mu      sync.Mutex
	widgets map[widgetKey]Widget
}

// Widget is the projected widget state.
type Widget struct {
	ID        string
	TenantID  string
	Name      string
	UpdatedAt time.Time
	Deleted   bool
}

type widgetKey struct {
	tenantID string
	id       string
}

// NewWidgetProjector constructs an empty projector. clock is required; passing
// nil is a programming error and will panic on first use rather than silently
// stamping a zero time.
func NewWidgetProjector(clock Clock) *WidgetProjector {
	return &WidgetProjector{clock: clock, widgets: make(map[widgetKey]Widget)}
}

// Process applies a validated widget event to the projection. It validates
// first (a non-retryable failure short-circuits to ErrInvalidEvent) and stamps
// the projection's UpdatedAt from the injected clock, never the wall clock.
func (p *WidgetProjector) Process(ctx context.Context, e WidgetEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := e.Validate(); err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	k := widgetKey{tenantID: e.TenantID, id: e.WidgetID}
	switch e.Type {
	case EventWidgetCreated:
		p.widgets[k] = Widget{
			ID:        e.WidgetID,
			TenantID:  e.TenantID,
			Name:      e.Name,
			UpdatedAt: p.clock.Now().UTC(),
		}
	case EventWidgetDeleted:
		w := p.widgets[k]
		w.ID = e.WidgetID
		w.TenantID = e.TenantID
		w.Deleted = true
		w.UpdatedAt = p.clock.Now().UTC()
		p.widgets[k] = w
	default:
		// Validate already rejected unknown types; defensive belt-and-braces.
		return fmt.Errorf("%w: unknown event type %q", ErrInvalidEvent, e.Type)
	}
	return nil
}

// Get returns the projected widget and whether it is present. It exists so
// tests can assert the projection reflects exactly-once application.
func (p *WidgetProjector) Get(tenantID, id string) (Widget, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	w, ok := p.widgets[widgetKey{tenantID: tenantID, id: id}]
	return w, ok
}
