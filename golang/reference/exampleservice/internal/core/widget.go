// Package core holds the widgets domain logic and the contracts it consumes.
// It depends only on the standard library and narrowly scoped seams; it imports
// no transport or storage-implementation packages, per
// golang/foundations/package-design.md.
//
// The Store interface is defined HERE, in the consumer, and names what the
// service needs ("Store"), not what an implementer is. The in-memory and
// database/sql implementations live in internal/db and satisfy it
// structurally.
package core

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Sentinel errors callers (notably the HTTP layer) branch on with errors.Is.
// They are part of the package contract.
var (
	// ErrNotFound is returned when a widget does not exist.
	ErrNotFound = errors.New("widget not found")
	// ErrAlreadyExists is returned when creating a widget whose ID is taken.
	ErrAlreadyExists = errors.New("widget already exists")
	// ErrInvalidWidget is returned when a widget fails validation.
	ErrInvalidWidget = errors.New("invalid widget")
)

// Widget is the widgets domain entity. It is a plain value type with no wire
// concerns: serialization lives on dedicated DTOs in the transport package.
type Widget struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

// Clock is the source of time for the service so behavior is deterministic in
// tests. Production wires a real clock in main; tests pass a fake. The
// interface is kept to the single method the service actually needs.
type Clock interface {
	// Now returns the current time. Implementations return UTC.
	Now() time.Time
}

// Store is the persistence contract the service consumes. It is intentionally
// narrow (interface-at-consumer); implementations live in internal/db.
type Store interface {
	// Create persists a new widget. It returns ErrAlreadyExists if a widget
	// with the same ID is already stored.
	Create(ctx context.Context, w Widget) error
	// Get returns the widget with the given ID, or ErrNotFound.
	Get(ctx context.Context, id string) (Widget, error)
	// List returns all widgets in a stable order.
	List(ctx context.Context) ([]Widget, error)
}

// Service is the widgets domain service. It is constructed with its
// dependencies and threads context explicitly; it never stores a context and
// never reads the wall clock directly.
type Service struct {
	store Store
	clock Clock
}

// NewService constructs a Service. store and clock are required; passing nil is
// a programming error and will panic on first use rather than silently
// misbehave.
func NewService(store Store, clock Clock) *Service {
	return &Service{store: store, clock: clock}
}

const maxNameLen = 128

// CreateWidget validates the input, stamps a UTC creation time from the
// injected clock, and persists the widget. Validation failures return
// ErrInvalidWidget; a duplicate ID surfaces the store's ErrAlreadyExists.
func (s *Service) CreateWidget(ctx context.Context, id, name string) (Widget, error) {
	if id == "" {
		return Widget{}, errInvalidf("id must not be empty")
	}
	if name == "" {
		return Widget{}, errInvalidf("name must not be empty")
	}
	if len(name) > maxNameLen {
		return Widget{}, errInvalidf("name must be at most %d bytes", maxNameLen)
	}

	w := Widget{
		ID:        id,
		Name:      name,
		CreatedAt: s.clock.Now().UTC(),
	}
	if err := s.store.Create(ctx, w); err != nil {
		// Wrap with %w so callers can still match ErrAlreadyExists / context.
		return Widget{}, err
	}
	return w, nil
}

// GetWidget returns the widget with the given ID. A missing widget surfaces
// ErrNotFound for the caller to map.
func (s *Service) GetWidget(ctx context.Context, id string) (Widget, error) {
	if id == "" {
		return Widget{}, errInvalidf("id must not be empty")
	}
	return s.store.Get(ctx, id)
}

// ListWidgets returns all widgets in a stable order.
func (s *Service) ListWidgets(ctx context.Context) ([]Widget, error) {
	return s.store.List(ctx)
}

// errInvalidf builds an ErrInvalidWidget-wrapping error with a specific reason
// so the message is actionable while errors.Is(err, ErrInvalidWidget) holds.
func errInvalidf(format string, args ...any) error {
	return &invalidError{reason: fmt.Sprintf(format, args...)}
}

type invalidError struct{ reason string }

func (e *invalidError) Error() string { return "invalid widget: " + e.reason }

// Unwrap lets errors.Is(err, ErrInvalidWidget) succeed.
func (e *invalidError) Unwrap() error { return ErrInvalidWidget }
