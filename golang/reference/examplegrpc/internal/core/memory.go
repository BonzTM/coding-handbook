package core

import (
	"context"
	"sort"
	"sync"
)

// Memory is an in-memory core.Store implementation used by default and in
// tests. It is safe for concurrent use. The zero value is not usable; call
// NewMemory.
//
// Rows are keyed by the composite (tenant_id, id) so the store is tenant-scoped:
// every method filters on tenantID and a widget under one tenant is invisible to
// another, mirroring the tenant_id WHERE clause a SQL store would use.
type Memory struct {
	mu      sync.RWMutex
	widgets map[tenantKey]Widget
}

// tenantKey is the composite primary key (tenant_id, id) the in-memory store
// indexes by, matching a SQL table's per-tenant uniqueness.
type tenantKey struct {
	tenantID string
	id       string
}

// Compile-time proof that *Memory satisfies the consumer-defined Store
// contract.
var _ Store = (*Memory)(nil)

// NewMemory constructs an empty in-memory store.
func NewMemory() *Memory {
	return &Memory{widgets: make(map[tenantKey]Widget)}
}

// Create stores a new widget within w.TenantID, returning ErrAlreadyExists if
// (tenant_id, id) is taken. It honors context cancellation before doing work.
func (m *Memory) Create(ctx context.Context, w Widget) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	k := tenantKey{tenantID: w.TenantID, id: w.ID}
	if _, ok := m.widgets[k]; ok {
		return ErrAlreadyExists
	}
	m.widgets[k] = w
	return nil
}

// Get returns the widget with the given ID within tenantID, or ErrNotFound.
// A widget under a different tenant is reported as ErrNotFound.
func (m *Memory) Get(ctx context.Context, tenantID, id string) (Widget, error) {
	if err := ctx.Err(); err != nil {
		return Widget{}, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	w, ok := m.widgets[tenantKey{tenantID: tenantID, id: id}]
	if !ok {
		return Widget{}, ErrNotFound
	}
	return w, nil
}

// ListPage returns up to limit widgets WITHIN tenantID ordered by the stable
// keyset (CreatedAt, ID), starting strictly after the given cursor. It mirrors
// the keyset query a SQL store would run so offline tests exercise the exact
// same pagination contract: a per-tenant total order on (created_at, id) with id
// as the tiebreaker, and a strict ">" boundary so pages neither overlap nor skip
// rows.
func (m *Memory) ListPage(ctx context.Context, tenantID string, after Cursor, limit int) ([]Widget, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 {
		return nil, nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	all := make([]Widget, 0, len(m.widgets))
	for k, w := range m.widgets {
		if k.tenantID != tenantID {
			continue
		}
		all = append(all, w)
	}
	sort.Slice(all, func(i, j int) bool { return lessKeyset(all[i], all[j]) })

	out := make([]Widget, 0, limit)
	for _, w := range all {
		if !after.IsZero() && !afterCursor(w, after) {
			continue
		}
		out = append(out, w)
		if len(out) == limit {
			break
		}
	}
	return out, nil
}

// Close releases store resources. The in-memory store holds none, so it always
// succeeds; the method satisfies Store so main can close every store uniformly.
func (m *Memory) Close() error { return nil }

// lessKeyset orders widgets by the total, stable keyset (CreatedAt, ID): the
// sort key first, the primary key as the tiebreaker.
func lessKeyset(a, b Widget) bool {
	if a.CreatedAt.Equal(b.CreatedAt) {
		return a.ID < b.ID
	}
	return a.CreatedAt.Before(b.CreatedAt)
}

// afterCursor reports whether w sorts strictly after the cursor under the
// keyset order, i.e. (w.CreatedAt, w.ID) > (cursor.CreatedAt, cursor.ID). This
// is the in-memory equivalent of the SQL row-comparison
// (created_at, id) > ($cursor_created_at, $cursor_id).
func afterCursor(w Widget, c Cursor) bool {
	if w.CreatedAt.Equal(c.CreatedAt) {
		return w.ID > c.ID
	}
	return w.CreatedAt.After(c.CreatedAt)
}
