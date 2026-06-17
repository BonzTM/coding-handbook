package db_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
)

func widget(id string) core.Widget {
	return core.Widget{ID: id, Name: "name-" + id, CreatedAt: time.Unix(1700000000, 0).UTC()}
}

func TestMemoryCreateGetList(t *testing.T) {
	ctx := context.Background()
	m := db.NewMemory()

	// Create two widgets.
	if err := m.Create(ctx, widget("b")); err != nil {
		t.Fatalf("Create(b): %v", err)
	}
	if err := m.Create(ctx, widget("a")); err != nil {
		t.Fatalf("Create(a): %v", err)
	}

	// Get returns the stored value.
	got, err := m.Get(ctx, "a")
	if err != nil {
		t.Fatalf("Get(a): %v", err)
	}
	if got.ID != "a" || got.Name != "name-a" {
		t.Errorf("Get(a) = %+v, want id=a name=name-a", got)
	}

	// List returns all widgets ordered by ID (stable order is part of the
	// contract, so "a" precedes "b" even though "b" was inserted first).
	list, err := m.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("List len = %d, want 2", len(list))
	}
	if list[0].ID != "a" || list[1].ID != "b" {
		t.Errorf("List order = [%s %s], want [a b]", list[0].ID, list[1].ID)
	}
}

func TestMemoryGetNotFound(t *testing.T) {
	ctx := context.Background()
	m := db.NewMemory()

	_, err := m.Get(ctx, "missing")
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

	list, err := m.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("empty List len = %d, want 0", len(list))
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
	if _, err := m.Get(ctx, "x"); !errors.Is(err, context.Canceled) {
		t.Errorf("Get with cancelled ctx = %v, want context.Canceled", err)
	}
	if _, err := m.List(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("List with cancelled ctx = %v, want context.Canceled", err)
	}
}
