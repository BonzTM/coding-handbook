package telemetry_test

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/example/exampleservice/internal/config"
	"github.com/example/exampleservice/internal/telemetry"
)

func TestNewTracerProviderOfflineNeverExports(t *testing.T) {
	// With no OTLP endpoint, construction must succeed offline and install a
	// never-sampling provider plus the global W3C propagator. Shutdown is safe.
	ctx := context.Background()
	tp, err := telemetry.NewTracerProvider(ctx, config.TelemetryConfig{}, "svc", "v0")
	if err != nil {
		t.Fatalf("NewTracerProvider (offline): %v", err)
	}

	// A span started from the provider is created but not sampled (never export).
	tracer := otel.Tracer("test")
	_, span := tracer.Start(ctx, "op")
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

	// The global propagator must carry W3C traceparent (TraceContext) and
	// baggage so trace context crosses service boundaries.
	prop := otel.GetTextMapPropagator()
	fields := prop.Fields()
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

func TestNewTracerProviderBatchedWithEndpoint(t *testing.T) {
	// A configured endpoint builds the batched/exporting pipeline. The exporter
	// connects lazily (BatchSpanProcessor), so construction succeeds without a
	// live collector; this proves the config-gated branch is wired.
	ctx := context.Background()
	cfg := config.TelemetryConfig{
		OTLPEndpoint:     "127.0.0.1:4318",
		OTLPInsecure:     true,
		TraceSampleRatio: 1.0,
	}
	tp, err := telemetry.NewTracerProvider(ctx, cfg, "svc", "v1")
	if err != nil {
		t.Fatalf("NewTracerProvider (endpoint): %v", err)
	}

	// With ratio 1.0 and a real exporter pipeline, a root span is sampled.
	_, span := otel.Tracer("test").Start(ctx, "op")
	if !span.SpanContext().IsSampled() {
		t.Error("endpoint provider with ratio 1.0 did not sample a root span")
	}
	span.End()

	shutdownCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if err := tp.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}
