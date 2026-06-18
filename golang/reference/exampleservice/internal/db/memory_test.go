package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
)

// testTenant is the tenant every widget in these tests belongs to. Cross-tenant
// isolation is exercised separately in TestMemoryTenantIsolation.
const testTenant = "t1"

func widget(id string) core.Widget {
	return core.Widget{ID: id, TenantID: testTenant, Name: "name-" + id, CreatedAt: time.Unix(1700000000, 0).UTC()}
}

// widgetAt builds a widget with a distinct CreatedAt so keyset ordering is
// driven by the primary sort key rather than only the id tiebreaker.
func widgetAt(id string, offset time.Duration) core.Widget {
	return core.Widget{ID: id, TenantID: testTenant, Name: "name-" + id, CreatedAt: time.Unix(1700000000, 0).UTC().Add(offset)}
}

func TestMemoryCreateGetList(t *testing.T) {
	ctx := context.Background()
	m := db.NewMemory()

	// Create two widgets that share a CreatedAt, so the id tiebreaker orders them.
	if err := m.Create(ctx, widget("b")); err != nil {
		t.Fatalf("Create(b): %v", err)
	}
	if err := m.Create(ctx, widget("a")); err != nil {
		t.Fatalf("Create(a): %v", err)
	}

	// Get returns the stored value.
	got, err := m.Get(ctx, testTenant, "a")
	if err != nil {
		t.Fatalf("Get(a): %v", err)
	}
	if got.ID != "a" || got.Name != "name-a" {
		t.Errorf("Get(a) = %+v, want id=a name=name-a", got)
	}

	// ListPage returns widgets ordered by the keyset (CreatedAt, ID); the
	// CreatedAts tie so id wins, putting "a" before "b" even though "b" was
	// inserted first.
	list, err := m.ListPage(ctx, testTenant, core.Cursor{}, 10)
	if err != nil {
		t.Fatalf("ListPage: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("ListPage len = %d, want 2", len(list))
	}
	if list[0].ID != "a" || list[1].ID != "b" {
		t.Errorf("ListPage order = [%s %s], want [a b]", list[0].ID, list[1].ID)
	}
}

// TestMemoryListPageKeysetWalk proves the keyset contract end to end against
// the in-memory store: pages of size 2 partition the set in (CreatedAt, ID)
// order, the cursor resumes strictly after the previous page, and the walk
// covers every row exactly once without overlap or gaps.
func TestMemoryListPageKeysetWalk(t *testing.T) {
	ctx := context.Background()
	m := db.NewMemory()
	// Distinct CreatedAt per widget; insertion order is shuffled relative to time.
	seed := []struct {
		id     string
		offset time.Duration
	}{
		{"w3", 3 * time.Second},
		{"w1", 1 * time.Second},
		{"w5", 5 * time.Second},
		{"w2", 2 * time.Second},
		{"w4", 4 * time.Second},
	}
	for _, s := range seed {
		if err := m.Create(ctx, widgetAt(s.id, s.offset)); err != nil {
			t.Fatalf("seed %s: %v", s.id, err)
		}
	}

	var got []string
	after := core.Cursor{}
	for page := 0; page < len(seed)+1; page++ {
		batch, err := m.ListPage(ctx, testTenant, after, 2)
		if err != nil {
			t.Fatalf("ListPage: %v", err)
		}
		if len(batch) == 0 {
			break
		}
		if len(batch) > 2 {
			t.Fatalf("page size = %d, want <= 2", len(batch))
		}
		for _, w := range batch {
			got = append(got, w.ID)
		}
		last := batch[len(batch)-1]
		after = core.Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}

	want := []string{"w1", "w2", "w3", "w4", "w5"}
	if len(got) != len(want) {
		t.Fatalf("walked %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestMemoryListPageZeroLimit proves a non-positive limit returns no rows
// rather than the whole table.
func TestMemoryListPageZeroLimit(t *testing.T) {
	ctx := context.Background()
	m := db.NewMemory()
	if err := m.Create(ctx, widget("a")); err != nil {
		t.Fatalf("Create: %v", err)
	}
	list, err := m.ListPage(ctx, testTenant, core.Cursor{}, 0)
	if err != nil {
		t.Fatalf("ListPage: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("ListPage(limit=0) len = %d, want 0", len(list))
	}
}

func TestMemoryGetNotFound(t *testing.T) {
	ctx := context.Background()
	m := db.NewMemory()

	_, err := m.Get(ctx, testTenant, "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("Get(missing) error = %v, want core.ErrNotFound", err)
	}
}

func TestMemoryCreateDuplicate(t *testing.T) {
	ctx := context.Background()
	m := db.NewMemory()

	if err := m.Create(ctx, widget("dup")); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	err := m.Create(ctx, widget("dup"))
	if !errors.Is(err, core.ErrAlreadyExists) {
		t.Fatalf("duplicate Create error = %v, want core.ErrAlreadyExists", err)
	}
}

func TestMemoryEmptyList(t *testing.T) {
	ctx := context.Background()
	m := db.NewMemory()

	list, err := m.ListPage(ctx, testTenant, core.Cursor{}, 10)
	if err != nil {
		t.Fatalf("ListPage: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("empty ListPage len = %d, want 0", len(list))
	}
}

// TestMemoryTenantIsolation proves the store scopes every read by tenant: a
// widget created under one tenant is invisible (ErrNotFound / absent from the
// list) to another, and the same id may be reused across tenants.
func TestMemoryTenantIsolation(t *testing.T) {
	ctx := context.Background()
	m := db.NewMemory()

	ts := time.Unix(1700000000, 0).UTC()
	if err := m.Create(ctx, core.Widget{ID: "shared", TenantID: "t1", Name: "t1-widget", CreatedAt: ts}); err != nil {
		t.Fatalf("create t1: %v", err)
	}
	// Same id under a different tenant: allowed (composite key), distinct row.
	if err := m.Create(ctx, core.Widget{ID: "shared", TenantID: "t2", Name: "t2-widget", CreatedAt: ts}); err != nil {
		t.Fatalf("create t2: %v", err)
	}

	// A cross-tenant read returns the OTHER tenant's row never; each sees its own.
	got1, err := m.Get(ctx, "t1", "shared")
	if err != nil || got1.Name != "t1-widget" {
		t.Fatalf("Get(t1) = %+v, %v; want t1-widget", got1, err)
	}
	got2, err := m.Get(ctx, "t2", "shared")
	if err != nil || got2.Name != "t2-widget" {
		t.Fatalf("Get(t2) = %+v, %v; want t2-widget", got2, err)
	}

	// A tenant with no rows sees zero rows / not-found, never another's data.
	if _, missErr := m.Get(ctx, "t3", "shared"); !errors.Is(missErr, core.ErrNotFound) {
		t.Errorf("Get(t3) error = %v, want ErrNotFound (no cross-tenant leak)", missErr)
	}
	list, err := m.ListPage(ctx, "t3", core.Cursor{}, 10)
	if err != nil {
		t.Fatalf("ListPage(t3): %v", err)
	}
	if len(list) != 0 {
		t.Errorf("ListPage(t3) len = %d, want 0 (cross-tenant isolation)", len(list))
	}

	// t1's list contains only t1's row.
	t1list, err := m.ListPage(ctx, "t1", core.Cursor{}, 10)
	if err != nil {
		t.Fatalf("ListPage(t1): %v", err)
	}
	if len(t1list) != 1 || t1list[0].Name != "t1-widget" {
		t.Errorf("ListPage(t1) = %+v, want [t1-widget]", t1list)
	}
}

func TestMemoryHonorsContextCancellation(t *testing.T) {
	// Every method checks ctx.Err() before doing work; a cancelled context must
	// short-circuit rather than mutate or read state.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := db.NewMemory()

	if err := m.Create(ctx, widget("x")); !errors.Is(err, context.Canceled) {
		t.Errorf("Create with cancelled ctx = %v, want context.Canceled", err)
	}
	if _, err := m.Get(ctx, testTenant, "x"); !errors.Is(err, context.Canceled) {
		t.Errorf("Get with cancelled ctx = %v, want context.Canceled", err)
	}
	if _, err := m.ListPage(ctx, testTenant, core.Cursor{}, 10); !errors.Is(err, context.Canceled) {
		t.Errorf("ListPage with cancelled ctx = %v, want context.Canceled", err)
	}
}
