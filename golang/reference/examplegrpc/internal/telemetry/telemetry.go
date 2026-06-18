// Package telemetry constructs the process logger, holds the readiness flag,
// and provides the Prometheus metrics seam plus the config-gated OTel tracing
// pipeline. No global logger lives here; the constructed *slog.Logger is
// returned and threaded explicitly into reusable packages.
package telemetry

import (
	"io"
	"log/slog"
	"sync/atomic"
)

// NewLogger builds the single structured logger for the process. JSON for
// machine-collected environments, text for local development. The level comes
// from config so operators can change verbosity without a code redeploy.
func NewLogger(w io.Writer, level slog.Level, json bool) *slog.Logger {
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler
	if json {
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
