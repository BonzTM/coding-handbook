package telemetry_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/example/examplegrpc/internal/config"
	"github.com/example/examplegrpc/internal/telemetry"
)

func TestNewLoggerJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := telemetry.NewLogger(&buf, slog.LevelInfo, true)
	logger.Info("hello", "k", "v")
	if !strings.Contains(buf.String(), `"msg":"hello"`) {
		t.Errorf("expected JSON log, got: %s", buf.String())
	}
}

func TestReadiness(t *testing.T) {
	r := telemetry.NewReadiness(false)
	if r.Ready() {
		t.Error("should start not ready")
	}
	r.Set(true)
	if !r.Ready() {
		t.Error("should be ready after Set(true)")
	}
}

func TestPromMetricsExposesRPCCounter(t *testing.T) {
	m := telemetry.NewPromMetrics("examplegrpc")
	m.ObserveRPC("/widget.v1.WidgetService/GetWidget", "OK", 0.01)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	m.Handler().ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "examplegrpc_grpc_requests_total") {
		t.Errorf("metrics output missing grpc_requests_total:\n%s", body)
	}
}

func TestNewTracerProviderOfflineNeverExports(t *testing.T) {
	ctx := context.Background()
	tp, err := telemetry.NewTracerProvider(ctx, config.TelemetryConfig{}, "svc", "v0")
	if err != nil {
		t.Fatalf("NewTracerProvider (offline): %v", err)
	}
	_, span := otel.Tracer("test").Start(ctx, "op")
	if span.SpanContext().IsSampled() {
		t.Error("offline provider sampled a span; want NeverSample")
	}
	span.End()

	shutdownCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if err := tp.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestNewTracerProviderSetsW3CPropagator(t *testing.T) {
	ctx := context.Background()
	tp, err := telemetry.NewTracerProvider(ctx, config.TelemetryConfig{}, "svc", "v0")
	if err != nil {
		t.Fatalf("NewTracerProvider: %v", err)
	}
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			t.Errorf("tracer provider shutdown: %v", err)
		}
	})

	fields := otel.GetTextMapPropagator().Fields()
	want := map[string]bool{"traceparent": false, "baggage": false}
	for _, f := range fields {
		if _, ok := want[f]; ok {
			want[f] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("global propagator missing field %q (fields=%v)", k, fields)
		}
	}
}
