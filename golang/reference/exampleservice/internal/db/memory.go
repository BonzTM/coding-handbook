// Package db holds the persistence implementations of the core.Store
// contract. It depends on internal/core for the domain types and sentinel
// errors; it never imports the transport layer, per the dependency direction
// in golang/foundations/package-design.md.
//
// memory.go is the in-memory test/dev double; postgres.go is the real
// database/sql repository.
package db

import (
	"context"
	"sort"
	"sync"

	"github.com/example/exampleservice/internal/core"
)

// Memory is an in-memory core.Store implementation used by default and in
// tests. It is safe for concurrent use. The zero value is not usable; call
// NewMemory.
type Memory struct {
	mu      sync.RWMutex
	widgets map[string]core.Widget
}

// Compile-time proof that *Memory satisfies the consumer-defined core.Store
// contract. This catches a signature drift here, where the implementation
// lives, rather than only at the call site in main.
var _ core.Store = (*Memory)(nil)

// NewMemory constructs an empty in-memory store.
func NewMemory() *Memory {
	return &Memory{widgets: make(map[string]core.Widget)}
}

// Create stores a new widget, returning core.ErrAlreadyExists if the ID is
// taken. It honors context cancellation before doing work.
func (m *Memory) Create(ctx context.Context, w core.Widget) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.widgets[w.ID]; ok {
		return core.ErrAlreadyExists
	}
	m.widgets[w.ID] = w
	return nil
}

// Get returns the widget with the given ID, or core.ErrNotFound.
func (m *Memory) Get(ctx context.Context, id string) (core.Widget, error) {
	if err := ctx.Err(); err != nil {
		return core.Widget{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.widgets[id]
	if !ok {
		return core.Widget{}, core.ErrNotFound
	}
	return w, nil
}

// List returns all widgets ordered by ID so the result is stable.
func (m *Memory) List(ctx context.Context) ([]core.Widget, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]core.Widget, 0, len(m.widgets))
	for _, w := range m.widgets {
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
