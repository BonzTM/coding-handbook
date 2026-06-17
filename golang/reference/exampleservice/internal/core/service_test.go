package core_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
)

// fixedClock is a deterministic core.Clock for tests.
type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

func newService() *core.Service {
	return core.NewService(db.NewMemory(), fixedClock{t: time.Unix(1700000000, 0).UTC()})
}

// authedCtx returns a context carrying a principal with both roles in tenant
// "t1", the default identity these service tests act as. Authorization edge
// cases (missing principal, missing role, cross-tenant) have dedicated tests.
func authedCtx() context.Context {
	return core.WithPrincipal(context.Background(), core.Principal{
		Subject:  "tester",
		TenantID: "t1",
		Roles:    []core.Role{core.RoleReader, core.RoleWriter},
	})
}

func TestCreateWidgetHappyPath(t *testing.T) {
	svc := newService()
	ctx := authedCtx()

	w, err := svc.CreateWidget(ctx, "w1", "Widget One")
	if err != nil {
		t.Fatalf("CreateWidget: unexpected error: %v", err)
	}
	if w.ID != "w1" || w.Name != "Widget One" {
		t.Errorf("widget = %+v, want id=w1 name=Widget One", w)
	}
	if w.CreatedAt.IsZero() || w.CreatedAt.Location() != time.UTC {
		t.Errorf("CreatedAt = %v, want a non-zero UTC time", w.CreatedAt)
	}

	got, err := svc.GetWidget(ctx, "w1")
	if err != nil {
		t.Fatalf("GetWidget: unexpected error: %v", err)
	}
	if got != w {
		t.Errorf("GetWidget = %+v, want %+v", got, w)
	}
}

func TestCreateWidgetDuplicate(t *testing.T) {
	svc := newService()
	ctx := authedCtx()

	if _, err := svc.CreateWidget(ctx, "dup", "first"); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := svc.CreateWidget(ctx, "dup", "second")
	if !errors.Is(err, core.ErrAlreadyExists) {
		t.Fatalf("second create: error = %v, want ErrAlreadyExists", err)
	}
}

func TestCreateWidgetValidation(t *testing.T) {
	svc := newService()
	ctx := authedCtx()

	tests := []struct {
		name string
		id   string
		wnam string
	}{
		{"empty id", "", "ok"},
		{"empty name", "id", ""},
		{"name too long", "id", string(make([]byte, 200))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateWidget(ctx, tt.id, tt.wnam)
			if !errors.Is(err, core.ErrInvalidWidget) {
				t.Errorf("error = %v, want ErrInvalidWidget", err)
			}
		})
	}
}

func TestGetWidgetNotFound(t *testing.T) {
	svc := newService()
	_, err := svc.GetWidget(authedCtx(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

// TestListWidgetsStableOrder seeds widgets that all share one CreatedAt (the
// fixed clock), so the keyset order falls back to the id tiebreaker and the
// page is returned in id order regardless of insertion order.
func TestListWidgetsStableOrder(t *testing.T) {
	svc := newService()
	ctx := authedCtx()
	for _, id := range []string{"c", "a", "b"} {
		if _, err := svc.CreateWidget(ctx, id, "n"); err != nil {
			t.Fatalf("create %s: %v", id, err)
		}
	}
	page, err := svc.ListWidgetsPage(ctx, core.Cursor{}, 0)
	if err != nil {
		t.Fatalf("ListWidgetsPage: %v", err)
	}
	want := []string{"a", "b", "c"}
	if len(page.Widgets) != len(want) {
		t.Fatalf("len = %d, want %d", len(page.Widgets), len(want))
	}
	for i, w := range page.Widgets {
		if w.ID != want[i] {
			t.Errorf("position %d = %q, want %q", i, w.ID, want[i])
		}
	}
	// One page holds all three widgets, so this is the last page: empty cursor.
	if !page.NextCursor.IsZero() {
		t.Errorf("NextCursor = %+v, want zero (last page)", page.NextCursor)
	}
}

// TestClampPageSize pins the server-side page-size policy: non-positive falls
// back to the default, oversized clamps to the max, in-range passes through.
func TestClampPageSize(t *testing.T) {
	tests := []struct {
		name      string
		requested int
		want      int
	}{
		{"zero uses default", 0, core.DefaultPageSize},
		{"negative uses default", -5, core.DefaultPageSize},
		{"in range passes through", 10, 10},
		{"at max passes through", core.MaxPageSize, core.MaxPageSize},
		{"over max is clamped", core.MaxPageSize + 1, core.MaxPageSize},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := core.ClampPageSize(tt.requested); got != tt.want {
				t.Errorf("ClampPageSize(%d) = %d, want %d", tt.requested, got, tt.want)
			}
		})
	}
}

// TestListWidgetsPageWalk walks every page via the opaque cursor and proves the
// pages partition the full set with no overlaps or gaps, the page size is
// honored (and clamped), and the last page reports a zero cursor.
func TestListWidgetsPageWalk(t *testing.T) {
	ctx := authedCtx()
	// Distinct CreatedAt per widget so the primary sort key (CreatedAt) drives
	// the order; ids are intentionally NOT in time order.
	base := time.Unix(1700000000, 0).UTC()
	store := db.NewMemory()
	svc := core.NewService(store, fixedClock{t: base})
	ids := []string{"e", "a", "d", "b", "c"}
	for i, id := range ids {
		// Seed under the same tenant the principal acts as ("t1") so the
		// tenant-scoped list returns them.
		w := core.Widget{ID: id, TenantID: "t1", Name: "n", CreatedAt: base.Add(time.Duration(i) * time.Second)}
		if err := store.Create(ctx, w); err != nil {
			t.Fatalf("seed %s: %v", id, err)
		}
	}

	var got []string
	cursor := core.Cursor{}
	for pages := 0; ; pages++ {
		if pages > len(ids) {
			t.Fatal("pagination did not terminate")
		}
		page, err := svc.ListWidgetsPage(ctx, cursor, 2)
		if err != nil {
			t.Fatalf("ListWidgetsPage: %v", err)
		}
		if len(page.Widgets) > 2 {
			t.Fatalf("page size = %d, want <= 2", len(page.Widgets))
		}
		for _, w := range page.Widgets {
			got = append(got, w.ID)
		}
		if page.NextCursor.IsZero() {
			break
		}
		// Round-trip the cursor through the opaque encoding the wire uses.
		next, err := core.DecodeCursor(core.EncodeCursor(page.NextCursor))
		if err != nil {
			t.Fatalf("cursor round-trip: %v", err)
		}
		cursor = next
	}

	// Widgets come back in CreatedAt order, which is insertion order here.
	want := ids
	if len(got) != len(want) {
		t.Fatalf("walked %d widgets, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d = %q, want %q", i, got[i], want[i])
		}
	}
}
