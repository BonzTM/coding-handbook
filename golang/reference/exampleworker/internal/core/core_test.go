package core_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/exampleworker/internal/core"
	"github.com/example/exampleworker/internal/testutil"
)

func TestWidgetEventValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		event   core.WidgetEvent
		wantErr bool
	}{
		{
			name:  "valid created",
			event: core.WidgetEvent{Type: core.EventWidgetCreated, WidgetID: "w1", TenantID: "t1", Name: "n"},
		},
		{
			name:  "valid deleted",
			event: core.WidgetEvent{Type: core.EventWidgetDeleted, WidgetID: "w1", TenantID: "t1"},
		},
		{
			name:    "missing widget id",
			event:   core.WidgetEvent{Type: core.EventWidgetCreated, TenantID: "t1", Name: "n"},
			wantErr: true,
		},
		{
			name:    "missing tenant",
			event:   core.WidgetEvent{Type: core.EventWidgetCreated, WidgetID: "w1", Name: "n"},
			wantErr: true,
		},
		{
			name:    "created without name",
			event:   core.WidgetEvent{Type: core.EventWidgetCreated, WidgetID: "w1", TenantID: "t1"},
			wantErr: true,
		},
		{
			name:    "unknown type",
			event:   core.WidgetEvent{Type: "widget.exploded", WidgetID: "w1", TenantID: "t1"},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.event.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr && !errors.Is(err, core.ErrInvalidEvent) {
				t.Errorf("err %v must wrap ErrInvalidEvent", err)
			}
		})
	}
}

func TestWidgetProjectorProcess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	clk := testutil.NewFakeClock(now)
	p := core.NewWidgetProjector(clk)

	created := core.WidgetEvent{Type: core.EventWidgetCreated, WidgetID: "w1", TenantID: "t1", Name: "gadget"}
	if err := p.Process(context.Background(), created); err != nil {
		t.Fatalf("Process created: %v", err)
	}
	w, ok := p.Get("t1", "w1")
	if !ok || w.Name != "gadget" || !w.UpdatedAt.Equal(now) {
		t.Fatalf("projection after create = %+v, ok=%v", w, ok)
	}

	clk.Advance(time.Minute)
	del := core.WidgetEvent{Type: core.EventWidgetDeleted, WidgetID: "w1", TenantID: "t1"}
	if err := p.Process(context.Background(), del); err != nil {
		t.Fatalf("Process deleted: %v", err)
	}
	w, _ = p.Get("t1", "w1")
	if !w.Deleted || !w.UpdatedAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("projection after delete = %+v", w)
	}
}

func TestWidgetProjectorProcessInvalid(t *testing.T) {
	t.Parallel()

	p := core.NewWidgetProjector(testutil.NewFakeClock(time.Unix(0, 0).UTC()))
	err := p.Process(context.Background(), core.WidgetEvent{Type: core.EventWidgetCreated})
	if !errors.Is(err, core.ErrInvalidEvent) {
		t.Fatalf("err = %v, want ErrInvalidEvent", err)
	}
}

func TestWidgetProjectorContextCancelled(t *testing.T) {
	t.Parallel()

	p := core.NewWidgetProjector(testutil.NewFakeClock(time.Unix(0, 0).UTC()))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := p.Process(ctx, core.WidgetEvent{Type: core.EventWidgetCreated, WidgetID: "w1", TenantID: "t1", Name: "n"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}
