package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/example/exampleservice/internal/config"
	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
	"github.com/example/exampleservice/internal/telemetry"
)

// TestMetricsEndpointExposedAndNotMetered proves the Prometheus scrape endpoint
// is mounted (because the metrics impl exposes a Handler) and that scraping it
// is NOT itself counted as a request, while an API call is.
func TestMetricsEndpointExposedAndNotMetered(t *testing.T) {
	metrics := telemetry.NewPromMetrics("exampleservice")
	svc := core.NewService(db.NewMemory(), fixedClock{t: time.Unix(1700000000, 0).UTC()})
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	cfg := config.HTTPConfig{Addr: ":0", ReadHeaderTimeout: time.Second, MaxBodyBytes: 1 << 20}
	srv := New(cfg, svc, logger, metrics, telemetry.NewReadiness(true), testDeps())
	h := srv.Handler()

	do := func(method, path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}

	// Drive one API request so the counter has a sample.
	if rec := do(http.MethodGet, "/widgets"); rec.Code != http.StatusOK {
		t.Fatalf("GET /widgets = %d, want 200", rec.Code)
	}

	// Scrape twice; scrapes must not be counted.
	for range 2 {
		if rec := do(http.MethodGet, "/metrics"); rec.Code != http.StatusOK {
			t.Fatalf("GET /metrics = %d, want 200", rec.Code)
		}
	}

	rec := do(http.MethodGet, "/metrics")
	body := rec.Body.String()
	if !strings.Contains(body, `exampleservice_http_requests_total{route="GET /widgets",status_class="2xx"} 1`) {
		t.Errorf("expected exactly one counted GET /widgets request; body:\n%s", body)
	}
	if strings.Contains(body, `route="GET /metrics"`) {
		t.Errorf("/metrics scrape was metered; body:\n%s", body)
	}
}

// TestMetricsEndpointAbsentWithoutExposer proves the endpoint is only mounted
// when the metrics impl publishes a Handler. The no-op seam does not, so
// /metrics falls through to the API 404.
func TestMetricsEndpointAbsentWithoutExposer(t *testing.T) {
	srv := newTestServer(t, true) // wired with telemetry.NopMetrics{}
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("/metrics with no exposer = %d, want 404", rec.Code)
	}
}

// TestAccessLogIncludesTraceID proves the logging middleware enriches the access
// log with trace_id/span_id when a span is active (otelhttp wraps the chain).
func TestAccessLogIncludesTraceID(t *testing.T) {
	// A real always-sample provider so the request span has a valid context.
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			t.Errorf("tracer provider shutdown: %v", err)
		}
	})
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, nil))
	svc := core.NewService(db.NewMemory(), fixedClock{t: time.Unix(1700000000, 0).UTC()})
	cfg := config.HTTPConfig{Addr: ":0", ReadHeaderTimeout: time.Second, MaxBodyBytes: 1 << 20}
	srv := New(cfg, svc, logger, telemetry.NopMetrics{}, telemetry.NewReadiness(true), testDeps())

	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/widgets", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /widgets = %d, want 200", rec.Code)
	}

	var line struct {
		Msg     string `json:"msg"`
		TraceID string `json:"trace_id"`
		SpanID  string `json:"span_id"`
		Route   string `json:"route"`
	}
	// The access log line has msg="request"; decode the matching JSON object.
	for _, raw := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var probe struct {
			Msg string `json:"msg"`
		}
		if err := json.Unmarshal([]byte(raw), &probe); err != nil {
			continue
		}
		if probe.Msg == "request" {
			if err := json.Unmarshal([]byte(raw), &line); err != nil {
				t.Fatalf("decode access log line: %v", err)
			}
			break
		}
	}
	if line.Msg != "request" {
		t.Fatalf("no access log line found in:\n%s", buf.String())
	}
	if line.Route != "GET /widgets" {
		t.Errorf("route = %q, want GET /widgets", line.Route)
	}
	if len(line.TraceID) != 32 {
		t.Errorf("trace_id = %q, want 32 hex chars", line.TraceID)
	}
	if len(line.SpanID) != 16 {
		t.Errorf("span_id = %q, want 16 hex chars", line.SpanID)
	}
}
