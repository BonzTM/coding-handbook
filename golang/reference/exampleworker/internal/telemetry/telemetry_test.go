package telemetry_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/example/exampleworker/internal/config"
	"github.com/example/exampleworker/internal/telemetry"
)

func TestNewLoggerJSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := telemetry.NewLogger(&buf, config.TelemetryConfig{LogLevel: slog.LevelInfo, LogJSON: true})
	logger.Info("hello", "k", "v")

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("log line is not JSON: %v (%q)", err, buf.String())
	}
	if rec["msg"] != "hello" || rec["k"] != "v" {
		t.Errorf("unexpected log record: %v", rec)
	}
}

func TestNewLoggerTextLevelFilter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := telemetry.NewLogger(&buf, config.TelemetryConfig{LogLevel: slog.LevelWarn, LogJSON: false})
	logger.Info("suppressed")
	logger.Warn("shown")

	out := buf.String()
	if strings.Contains(out, "suppressed") {
		t.Error("info line should be filtered at warn level")
	}
	if !strings.Contains(out, "shown") {
		t.Error("warn line should be present")
	}
}

func TestReadiness(t *testing.T) {
	t.Parallel()

	r := telemetry.NewReadiness(false)
	if r.Ready() {
		t.Error("new readiness should start not ready")
	}
	r.Set(true)
	if !r.Ready() {
		t.Error("readiness should be true after Set(true)")
	}
	r.Set(false)
	if r.Ready() {
		t.Error("readiness should be false after Set(false)")
	}
}

func TestNopMetrics(t *testing.T) {
	t.Parallel()
	// NopMetrics must satisfy the seam and never panic.
	var m telemetry.Metrics = telemetry.NopMetrics{}
	m.IncConsumed("widget.created", "ack")
	m.IncPublished("widget.created")
}

func TestNewTracerProviderOffline(t *testing.T) {
	t.Parallel()

	tp, err := telemetry.NewTracerProvider(context.Background(), config.TelemetryConfig{}, "svc", "v0")
	if err != nil {
		t.Fatalf("NewTracerProvider: %v", err)
	}
	if err := tp.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}
