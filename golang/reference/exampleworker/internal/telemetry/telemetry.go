// Package telemetry constructs the process logger, holds the readiness flag,
// and provides a small metrics seam. The metrics seam is an interface so the
// reference build runs and tests without an external metrics dependency (the
// no-op default), while production wires the Prometheus adapter per
// golang/operations/observability.md.
//
// No global logger lives here; the constructed *slog.Logger is returned and
// threaded explicitly into reusable packages.
package telemetry

import (
	"io"
	"log/slog"
	"sync/atomic"

	"github.com/example/exampleworker/internal/config"
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
// shutdown sequence flips to false before draining. For the worker, readiness
// reflects broker connectivity: it flips true once the consumer is subscribed
// and back to false when the broker connection is lost or shutdown begins.
// Liveness is separate and stays green during drain so the platform does not
// kill the pod mid-shutdown.
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

// Ready reports whether the worker is ready (broker connected, consuming).
func (r *Readiness) Ready() bool { return r.ready.Load() }

// Metrics is the low-cardinality metrics seam consumed by the messaging layer.
// It is intentionally small. Production swaps in the Prometheus adapter behind
// this same interface; the reference build uses NopMetrics so it needs no
// external dependency.
//
// Implementations MUST keep label values low-cardinality (event type, outcome
// class), never message IDs, correlation IDs, tenant IDs, or raw subjects, per
// golang/services/eventing-and-messaging.md ### Observability.
type Metrics interface {
	// IncConsumed records one consumed message by event type and outcome
	// ("ack", "retry", "dropped_duplicate", "dead_lettered").
	IncConsumed(eventType, outcome string)
	// IncPublished records one message published by the outbox relay.
	IncPublished(eventType string)
}

// NopMetrics is the default no-op metrics implementation.
type NopMetrics struct{}

// IncConsumed does nothing.
func (NopMetrics) IncConsumed(string, string) {}

// IncPublished does nothing.
func (NopMetrics) IncPublished(string) {}
