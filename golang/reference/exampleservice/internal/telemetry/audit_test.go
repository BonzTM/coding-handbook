package telemetry_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/telemetry"
)

// fixedClock is a deterministic telemetry.AuditClock for the audit-record tests.
type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

// TestAuditLoggerEmitsFullSchema proves an audit record carries the full
// who/what/when/where schema with a UTC timestamp and the dedicated log_type
// marker, and that the clock-supplied time is forced to UTC.
func TestAuditLoggerEmitsFullSchema(t *testing.T) {
	var buf bytes.Buffer
	// A non-UTC clock proves Emit normalizes the stamp to UTC.
	loc := time.FixedZone("UTC-5", -5*3600)
	clk := fixedClock{t: time.Date(2026, 6, 17, 12, 0, 0, 0, loc)}
	a := telemetry.NewAuditLogger(&buf, clk)

	a.Emit(context.Background(), telemetry.AuditEvent{
		Actor:     "user-1",
		Tenant:    "tenant-1",
		Action:    "widget.create",
		Resource:  "w1",
		Result:    telemetry.AuditSuccess,
		RequestID: "req-abc",
	})

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("decode audit record: %v", err)
	}

	for k, want := range map[string]string{
		"actor":      "user-1",
		"tenant":     "tenant-1",
		"action":     "widget.create",
		"resource":   "w1",
		"result":     string(telemetry.AuditSuccess),
		"request_id": "req-abc",
		"log_type":   "audit",
	} {
		got, ok := rec[k].(string)
		if !ok || got != want {
			t.Errorf("field %q = %v, want %q", k, rec[k], want)
		}
	}

	ts, ok := rec["time"].(string)
	if !ok {
		t.Fatalf("audit record missing time field: %v", rec)
	}
	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t.Fatalf("parse audit time %q: %v", ts, err)
	}
	if parsed.Location() != time.UTC {
		t.Errorf("audit time location = %v, want UTC", parsed.Location())
	}
	if !parsed.Equal(clk.t) {
		t.Errorf("audit time = %v, want same instant as %v", parsed, clk.t)
	}
}

// TestNopAuditLoggerDiscards proves the no-op audit logger is a safe sink that
// never panics, so callers need no nil check before Emit.
func TestNopAuditLoggerDiscards(t *testing.T) {
	a := telemetry.NopAuditLogger()
	a.Emit(context.Background(), telemetry.AuditEvent{Action: "authenticate", Result: telemetry.AuditFailure})
}
