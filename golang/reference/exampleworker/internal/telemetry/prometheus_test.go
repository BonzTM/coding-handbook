package telemetry_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/exampleworker/internal/telemetry"
)

func TestPromMetricsExposesCounters(t *testing.T) {
	t.Parallel()

	m := telemetry.NewPromMetrics("exampleworker")
	m.IncConsumed("widget.created", "ack")
	m.IncConsumed("widget.created", "ack")
	m.IncConsumed("widget.created", "dead_lettered")
	m.IncPublished("widget.created")
	m.ObserveProcess("widget.created", "ack", 0.01)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	out := string(body)
	for _, want := range []string{
		`exampleworker_messages_consumed_total{event_type="widget.created",outcome="ack"} 2`,
		`exampleworker_messages_consumed_total{event_type="widget.created",outcome="dead_lettered"} 1`,
		`exampleworker_messages_published_total{event_type="widget.created"} 1`,
		"exampleworker_message_process_duration_seconds_bucket",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("scrape output missing %q\n---\n%s", want, out)
		}
	}
}
