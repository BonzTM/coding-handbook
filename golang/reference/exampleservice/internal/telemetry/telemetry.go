// Package telemetry constructs the process logger, holds the readiness flag,
// and provides a small metrics seam. It is stdlib-only on purpose: production
// wires a Prometheus client per golang/operations/observability.md, but the
// reference service must build and test without an external metrics dependency,
// so the seam is an interface with a no-op default and an expvar-backed impl.
//
// No global logger lives here; the constructed *slog.Logger is returned and
// threaded explicitly into reusable packages.
package telemetry

import (
	"expvar"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/example/exampleservice/internal/config"
)

// NewLogger builds the single structured logger for the process. JSON for
// machine-collected environments, text for local development. The level comes
// from config so operators can change verbosity without a code redeploy.
func NewLogger(w io.Writer, cfg config.TelemetryConfig) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	var h slog.Handler
	if cfg.LogJSON {
		h = slog.NewJSONHandler(w, opts)
	} else {
		h = slog.NewTextHandler(w, opts)
	}
	return slog.New(h)
}

// Readiness is a concurrency-safe flag the HTTP /readyz probe reads and the
// shutdown sequence flips to false before draining. Liveness is separate and
// stays green during drain so the platform does not kill the pod mid-shutdown.
//
// The zero value is not ready; call NewReadiness for an explicit initial state.
type Readiness struct {
	ready atomic.Bool
}

// NewReadiness returns a Readiness initialized to the given state.
func NewReadiness(ready bool) *Readiness {
	r := &Readiness{}
	r.ready.Store(ready)
	return r
}

// Set updates the readiness state.
func (r *Readiness) Set(ready bool) { r.ready.Store(ready) }

// Ready reports whether the service is ready to receive traffic.
func (r *Readiness) Ready() bool { return r.ready.Load() }

// Metrics is the low-cardinality metrics seam consumed by the rest of the
// service. It is intentionally tiny (an interface of 1-3 methods, per
// golang/foundations/package-design.md). The reference wires the Prometheus
// adapter (NewPromMetrics) in main behind this same interface; NopMetrics and
// ExpvarMetrics remain drop-in implementations for tests and builds that must
// avoid the external dependency — swap the adapter in main, never the call
// sites.
//
// Implementations MUST keep label values low-cardinality (route patterns and
// status classes, never request IDs, user IDs, or raw paths) per
// golang/operations/observability.md.
type Metrics interface {
	// IncRequest records one handled HTTP request by route pattern and status.
	IncRequest(routePattern, statusClass string)
	// IncWidgetCreated records one successfully created widget.
	IncWidgetCreated()
}

// NopMetrics is the default no-op metrics implementation.
type NopMetrics struct{}

// IncRequest does nothing.
func (NopMetrics) IncRequest(string, string) {}

// IncWidgetCreated does nothing.
func (NopMetrics) IncWidgetCreated() {}

// ExpvarMetrics is a stdlib-only Metrics implementation backed by expvar. It
// demonstrates the seam with real counters without pulling in a metrics
// library; expvar publishes under /debug/vars in the default mux when wired.
type ExpvarMetrics struct {
	mu             sync.Mutex
	requests       *expvar.Map
	widgetsCreated *expvar.Int
}

// NewExpvarMetrics constructs an ExpvarMetrics. The prefix namespaces the
// published variables so multiple instances do not collide.
func NewExpvarMetrics(prefix string) *ExpvarMetrics {
	return &ExpvarMetrics{
		requests:       expvar.NewMap(prefix + "_http_requests_total"),
		widgetsCreated: expvar.NewInt(prefix + "_widgets_created_total"),
	}
}

// IncRequest increments the request counter for a route pattern and status
// class. Both arguments are expected to be low-cardinality.
func (m *ExpvarMetrics) IncRequest(routePattern, statusClass string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests.Add(routePattern+" "+statusClass, 1)
}

// IncWidgetCreated increments the widgets-created counter.
func (m *ExpvarMetrics) IncWidgetCreated() {
	m.widgetsCreated.Add(1)
}
