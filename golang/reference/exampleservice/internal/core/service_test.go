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

func TestCreateWidgetHappyPath(t *testing.T) {
	svc := newService()
	ctx := context.Background()

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
	ctx := context.Background()

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
	ctx := context.Background()

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
	_, err := svc.GetWidget(context.Background(), "missing")
	if !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestListWidgetsStableOrder(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	for _, id := range []string{"c", "a", "b"} {
		if _, err := svc.CreateWidget(ctx, id, "n"); err != nil {
			t.Fatalf("create %s: %v", id, err)
		}
	}
	got, err := svc.ListWidgets(ctx)
	if err != nil {
		t.Fatalf("ListWidgets: %v", err)
	}
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, w := range got {
		if w.ID != want[i] {
			t.Errorf("position %d = %q, want %q", i, w.ID, want[i])
		}
	}
}
