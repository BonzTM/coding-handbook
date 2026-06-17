package telemetry_test

import (
	"bytes"
	"expvar"
	"log/slog"
	"strings"
	"testing"

	"github.com/example/exampleservice/internal/config"
	"github.com/example/exampleservice/internal/telemetry"
)

func TestReadinessFlip(t *testing.T) {
	// The zero-arg constructor sets the explicit initial state, and Set flips it.
	// This is the flag the /readyz probe reads and shutdown drives to false.
	r := telemetry.NewReadiness(false)
	if r.Ready() {
		t.Fatal("NewReadiness(false) should not be ready")
	}

	r.Set(true)
	if !r.Ready() {
		t.Fatal("after Set(true) should be ready")
	}

	r.Set(false)
	if r.Ready() {
		t.Fatal("after Set(false) should not be ready (shutdown drain)")
	}
}

func TestReadinessInitialTrue(t *testing.T) {
	r := telemetry.NewReadiness(true)
	if !r.Ready() {
		t.Fatal("NewReadiness(true) should be ready")
	}
}

func TestNopMetricsNoPanic(t *testing.T) {
	// The default seam must be safe to call and do nothing.
	var m telemetry.Metrics = telemetry.NopMetrics{}
	m.IncRequest("GET /widgets", "2xx")
	m.IncWidgetCreated()
}

func TestExpvarMetricsCounts(t *testing.T) {
	// The expvar-backed seam records real counters. Use a unique prefix so the
	// global expvar registry does not collide across test runs/cases.
	const prefix = "exampleservice_test_expvarmetrics"
	m := telemetry.NewExpvarMetrics(prefix)

	m.IncRequest("GET /widgets", "2xx")
	m.IncRequest("GET /widgets", "2xx")
	m.IncRequest("POST /widgets", "4xx")
	m.IncWidgetCreated()
	m.IncWidgetCreated()
	m.IncWidgetCreated()

	requests, ok := expvar.Get(prefix + "_http_requests_total").(*expvar.Map)
	if !ok {
		t.Fatal("requests var is not an *expvar.Map")
	}
	if got := requests.Get("GET /widgets 2xx").String(); got != "2" {
		t.Errorf("GET /widgets 2xx count = %s, want 2", got)
	}
	if got := requests.Get("POST /widgets 4xx").String(); got != "1" {
		t.Errorf("POST /widgets 4xx count = %s, want 1", got)
	}

	created, ok := expvar.Get(prefix + "_widgets_created_total").(*expvar.Int)
	if !ok {
		t.Fatal("widgets-created var is not an *expvar.Int")
	}
	if got := created.Value(); got != 3 {
		t.Errorf("widgets_created_total = %d, want 3", got)
	}
}

func TestNewLoggerJSONAndText(t *testing.T) {
	// JSON handler emits a structured line at the configured level; text does
	// not. This exercises both branches of the constructor.
	var jsonBuf bytes.Buffer
	jl := telemetry.NewLogger(&jsonBuf, config.TelemetryConfig{LogLevel: slog.LevelInfo, LogJSON: true})
	jl.Info("hello", "k", "v")
	if !strings.Contains(jsonBuf.String(), `"msg":"hello"`) {
		t.Errorf("JSON logger output missing structured msg: %q", jsonBuf.String())
	}

	var textBuf bytes.Buffer
	tl := telemetry.NewLogger(&textBuf, config.TelemetryConfig{LogLevel: slog.LevelInfo, LogJSON: false})
	tl.Info("hello", "k", "v")
	if !strings.Contains(textBuf.String(), "msg=hello") {
		t.Errorf("text logger output missing msg=hello: %q", textBuf.String())
	}

	// A sub-info level message must be filtered out at LevelWarn.
	var warnBuf bytes.Buffer
	wl := telemetry.NewLogger(&warnBuf, config.TelemetryConfig{LogLevel: slog.LevelWarn, LogJSON: true})
	wl.Info("filtered")
	if warnBuf.Len() != 0 {
		t.Errorf("LevelWarn logger emitted an Info line: %q", warnBuf.String())
	}
}
