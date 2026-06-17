package core_test

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/example/examplegrpc/internal/core"
	"github.com/example/examplegrpc/internal/testutil"
)

const (
	testTenant  = "tenant-a"
	otherTenant = "tenant-b"
)

func writerCtx() context.Context {
	return core.WithPrincipal(context.Background(), core.Principal{
		Subject:  "user-1",
		TenantID: testTenant,
		Roles:    []core.Role{core.RoleReader, core.RoleWriter},
	})
}

func readerCtx(tenant string) context.Context {
	return core.WithPrincipal(context.Background(), core.Principal{
		Subject:  "user-ro",
		TenantID: tenant,
		Roles:    []core.Role{core.RoleReader},
	})
}

func newService(t *testing.T) (*core.Service, *testutil.FakeClock) {
	t.Helper()
	clk := testutil.NewFakeClock(time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC))
	return core.NewService(core.NewMemory(), clk), clk
}

func TestCreateAndGetWidget(t *testing.T) {
	svc, _ := newService(t)
	ctx := writerCtx()

	w, err := svc.CreateWidget(ctx, "w1", "Widget One")
	if err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}
	if w.TenantID != testTenant {
		t.Errorf("TenantID = %q, want %q (must come from principal)", w.TenantID, testTenant)
	}
	if w.CreatedAt.IsZero() || w.CreatedAt.Location() != time.UTC {
		t.Errorf("CreatedAt = %v, want a UTC stamp from the clock", w.CreatedAt)
	}

	got, err := svc.GetWidget(ctx, "w1")
	if err != nil {
		t.Fatalf("GetWidget: %v", err)
	}
	if got != w {
		t.Errorf("GetWidget = %+v, want %+v", got, w)
	}
}

func TestGetWidgetNotFound(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.GetWidget(readerCtx(testTenant), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestCreateWidgetDuplicate(t *testing.T) {
	svc, _ := newService(t)
	ctx := writerCtx()
	if _, err := svc.CreateWidget(ctx, "dup", "first"); err != nil {
		t.Fatalf("first CreateWidget: %v", err)
	}
	_, err := svc.CreateWidget(ctx, "dup", "second")
	if !errors.Is(err, core.ErrAlreadyExists) {
		t.Fatalf("err = %v, want ErrAlreadyExists", err)
	}
}

func TestCreateWidgetValidation(t *testing.T) {
	svc, _ := newService(t)
	ctx := writerCtx()
	cases := map[string]struct{ id, name string }{
		"empty id":   {"", "n"},
		"empty name": {"id", ""},
		"long name":  {"id", string(make([]byte, 129))},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := svc.CreateWidget(ctx, tc.id, tc.name)
			if !errors.Is(err, core.ErrInvalidWidget) {
				t.Fatalf("err = %v, want ErrInvalidWidget", err)
			}
		})
	}
}

func TestAuthzUnauthenticatedAndForbidden(t *testing.T) {
	svc, _ := newService(t)

	if _, err := svc.CreateWidget(context.Background(), "x", "y"); !errors.Is(err, core.ErrUnauthenticated) {
		t.Errorf("no principal: err = %v, want ErrUnauthenticated", err)
	}
	if _, err := svc.CreateWidget(readerCtx(testTenant), "x", "y"); !errors.Is(err, core.ErrForbidden) {
		t.Errorf("reader writing: err = %v, want ErrForbidden", err)
	}
}

func TestTenantIsolation(t *testing.T) {
	svc, _ := newService(t)
	if _, err := svc.CreateWidget(writerCtx(), "shared", "a"); err != nil {
		t.Fatalf("create: %v", err)
	}
	// Another tenant cannot see it.
	if _, err := svc.GetWidget(readerCtx(otherTenant), "shared"); !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("cross-tenant get: err = %v, want ErrNotFound", err)
	}
}

func TestListWidgetsKeysetPagination(t *testing.T) {
	svc, clk := newService(t)
	ctx := writerCtx()

	// Seed 5 widgets at distinct, increasing timestamps so the keyset order is
	// deterministic.
	for i := range 5 {
		clk.Advance(time.Second)
		if _, err := svc.CreateWidget(ctx, "w"+strconv.Itoa(i), "name"); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}

	// Page size 2 -> pages of 2, 2, 1.
	var seen []string
	cursor := core.Cursor{}
	pages := 0
	for {
		page, err := svc.ListWidgetsPage(ctx, cursor, 2)
		if err != nil {
			t.Fatalf("ListWidgetsPage: %v", err)
		}
		pages++
		for _, w := range page.Widgets {
			seen = append(seen, w.ID)
		}
		if page.NextCursor.IsZero() {
			break
		}
		cursor = page.NextCursor
		if pages > 10 {
			t.Fatal("pagination did not terminate")
		}
	}

	want := []string{"w0", "w1", "w2", "w3", "w4"}
	if len(seen) != len(want) {
		t.Fatalf("saw %v widgets, want %v", seen, want)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("order mismatch at %d: got %v want %v", i, seen, want)
		}
	}
	if pages != 3 {
		t.Errorf("pages = %d, want 3", pages)
	}
}

func TestClampPageSize(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, core.DefaultPageSize},
		{-1, core.DefaultPageSize},
		{10, 10},
		{core.MaxPageSize + 50, core.MaxPageSize},
	}
	for _, tc := range cases {
		if got := core.ClampPageSize(tc.in); got != tc.want {
			t.Errorf("ClampPageSize(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestCursorRoundTrip(t *testing.T) {
	c := core.Cursor{CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 6, time.UTC), ID: "abc"}
	tok := core.EncodeCursor(c)
	if tok == "" {
		t.Fatal("non-zero cursor encoded to empty token")
	}
	got, err := core.DecodeCursor(tok)
	if err != nil {
		t.Fatalf("DecodeCursor: %v", err)
	}
	if !got.CreatedAt.Equal(c.CreatedAt) || got.ID != c.ID {
		t.Errorf("round-trip = %+v, want %+v", got, c)
	}

	if core.EncodeCursor(core.Cursor{}) != "" {
		t.Error("zero cursor must encode to empty token")
	}
	if _, err := core.DecodeCursor("!!!not-base64!!!"); !errors.Is(err, core.ErrInvalidCursor) {
		t.Errorf("bad token err = %v, want ErrInvalidCursor", err)
	}
}

func TestListContextCancellation(t *testing.T) {
	svc, _ := newService(t)
	ctx, cancel := context.WithCancel(writerCtx())
	cancel()
	if _, err := svc.ListWidgetsPage(ctx, core.Cursor{}, 10); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}
