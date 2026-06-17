package health_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/exampleworker/internal/config"
	"github.com/example/exampleworker/internal/health"
	"github.com/example/exampleworker/internal/telemetry"
)

type stubBroker struct{ healthy bool }

func (s *stubBroker) Healthy() bool { return s.healthy }

func newSidecar(t *testing.T, ready, brokerHealthy bool, metrics telemetry.Metrics) *health.Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return health.New(config.HTTPConfig{Addr: ":0", ReadHeaderTimeout: time.Second}, logger, health.Deps{
		Readiness: telemetry.NewReadiness(ready),
		Broker:    &stubBroker{healthy: brokerHealthy},
		Metrics:   metrics,
	})
}

func get(t *testing.T, s *health.Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}

func TestLivezAlwaysOK(t *testing.T) {
	t.Parallel()
	// Liveness is independent of readiness/broker so the platform never kills a
	// draining pod.
	s := newSidecar(t, false, false, telemetry.NopMetrics{})
	rec := get(t, s, "/livez")
	if rec.Code != http.StatusOK {
		t.Errorf("/livez = %d, want 200", rec.Code)
	}
}

func TestReadyzReflectsBrokerAndReadiness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ready         bool
		brokerHealthy bool
		want          int
	}{
		{name: "ready and broker healthy", ready: true, brokerHealthy: true, want: http.StatusOK},
		{name: "not ready", ready: false, brokerHealthy: true, want: http.StatusServiceUnavailable},
		{name: "broker unhealthy", ready: true, brokerHealthy: false, want: http.StatusServiceUnavailable},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := newSidecar(t, tc.ready, tc.brokerHealthy, telemetry.NopMetrics{})
			rec := get(t, s, "/readyz")
			if rec.Code != tc.want {
				t.Errorf("/readyz = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}

func TestSetReadyFlipsProbe(t *testing.T) {
	t.Parallel()
	s := newSidecar(t, true, true, telemetry.NopMetrics{})
	if get(t, s, "/readyz").Code != http.StatusOK {
		t.Fatal("expected ready before SetReady(false)")
	}
	s.SetReady(false)
	if get(t, s, "/readyz").Code != http.StatusServiceUnavailable {
		t.Error("expected not-ready after SetReady(false)")
	}
}

func TestMetricsMountedWhenExposed(t *testing.T) {
	t.Parallel()
	// The Prometheus adapter exposes /metrics; NopMetrics does not.
	s := newSidecar(t, true, true, telemetry.NewPromMetrics("exampleworker"))
	if get(t, s, "/metrics").Code != http.StatusOK {
		t.Error("/metrics should be served when a Prometheus registry is wired")
	}

	nop := newSidecar(t, true, true, telemetry.NopMetrics{})
	if get(t, nop, "/metrics").Code != http.StatusNotFound {
		t.Error("/metrics should be absent without a registry")
	}
}

func TestSidecarShutdown(t *testing.T) {
	t.Parallel()
	s := newSidecar(t, true, true, telemetry.NopMetrics{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown: %v", err)
	}
}
