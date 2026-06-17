package http

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/example/exampleservice/internal/config"
	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/db"
	"github.com/example/exampleservice/internal/telemetry"
)

// countingMetrics records how many times IncRequest fires so a test can prove
// which routes flow through the logging+metrics middleware.
type countingMetrics struct {
	mu       sync.Mutex
	requests int
	widgets  int
}

func (m *countingMetrics) IncRequest(string, string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests++
}

func (m *countingMetrics) IncWidgetCreated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.widgets++
}

func (m *countingMetrics) requestCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requests
}

// TestProbesNotAccessLoggedOrMetered proves item 6's contract: /livez and
// /readyz are mounted ahead of the logging+metrics middleware, so polling them
// does NOT increment the request metric, while an API call DOES. This is the
// observable difference that makes the routes() comment true.
func TestProbesNotAccessLoggedOrMetered(t *testing.T) {
	metrics := &countingMetrics{}
	svc := core.NewService(db.NewMemory(), fixedClock{t: time.Unix(1700000000, 0).UTC()})
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.HTTPConfig{Addr: ":0", ReadHeaderTimeout: time.Second, MaxBodyBytes: 1 << 20}
	srv := New(cfg, svc, logger, metrics, telemetry.NewReadiness(true))
	h := srv.Handler()

	do := func(method, path string) int {
		req := httptest.NewRequest(method, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	// Hammer the probes: none of these should be metered.
	for range 5 {
		if code := do(http.MethodGet, "/livez"); code != http.StatusOK {
			t.Fatalf("/livez status = %d, want 200", code)
		}
		if code := do(http.MethodGet, "/readyz"); code != http.StatusOK {
			t.Fatalf("/readyz status = %d, want 200", code)
		}
	}
	if got := metrics.requestCount(); got != 0 {
		t.Fatalf("probes were metered: IncRequest called %d times, want 0", got)
	}

	// One API call MUST be metered, proving the API branch keeps the chain.
	if code := do(http.MethodGet, "/widgets"); code != http.StatusOK {
		t.Fatalf("GET /widgets status = %d, want 200", code)
	}
	if got := metrics.requestCount(); got != 1 {
		t.Fatalf("API request metering: IncRequest called %d times, want 1", got)
	}
}
