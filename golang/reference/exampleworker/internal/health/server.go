// Package health is the worker's small HTTP sidecar: it serves liveness,
// readiness, and the Prometheus scrape endpoint so the platform can probe the
// worker and Prometheus can scrape it. The worker is not an HTTP service, so
// this listener carries no application routes and no heavy middleware — only
// the probes and /metrics, none of which are access-logged.
//
// Readiness reflects broker connectivity: /readyz is healthy only when the
// readiness flag is set AND the broker reports healthy. Liveness is independent
// and stays green during drain so the platform does not kill a draining pod.
package health

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/example/exampleworker/internal/config"
	"github.com/example/exampleworker/internal/telemetry"
)

// brokerHealth reports broker connectivity for the readiness probe. The
// in-memory broker and a real client satisfy it; main passes the broker (or a
// constant-healthy stub when the broker does not report health).
type brokerHealth interface {
	Healthy() bool
}

// metricsExposer is the optional seam a Metrics implementation satisfies to
// publish a scrape endpoint. The Prometheus adapter implements it; the no-op
// seam does not, so /metrics is mounted only when a real registry is wired.
type metricsExposer interface {
	Handler() http.Handler
}

// Server owns the sidecar HTTP listener and the readiness wiring.
type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	readiness  *telemetry.Readiness
	broker     brokerHealth
}

// Deps bundles the sidecar's collaborators.
type Deps struct {
	// Readiness is the shared flag the consumer flips once subscribed and the
	// shutdown sequence flips back before draining. Required.
	Readiness *telemetry.Readiness
	// Broker reports connectivity for /readyz. Required.
	Broker brokerHealth
	// Metrics optionally publishes /metrics (the Prometheus adapter does).
	Metrics telemetry.Metrics
}

// New constructs the sidecar server with hardened timeouts.
func New(cfg config.HTTPConfig, logger *slog.Logger, deps Deps) *Server {
	s := &Server{
		logger:    logger,
		readiness: deps.Readiness,
		broker:    deps.Broker,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", s.handleLivez)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	if exposer, ok := deps.Metrics.(metricsExposer); ok {
		mux.Handle("GET /metrics", exposer.Handler())
	}

	s.httpServer = &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
	return s
}

// ListenAndServe starts serving and blocks until Shutdown. It returns
// http.ErrServerClosed on a clean Shutdown, which the caller treats as the
// expected outcome.
func (s *Server) ListenAndServe() error { return s.httpServer.ListenAndServe() }

// Addr returns the configured listen address.
func (s *Server) Addr() string { return s.httpServer.Addr }

// SetReady flips the readiness flag. The shutdown sequence calls SetReady(false)
// before draining so the platform stops routing readiness-gated traffic.
func (s *Server) SetReady(ready bool) { s.readiness.Set(ready) }

// Shutdown gracefully stops the sidecar, bounded by ctx's deadline. The caller
// MUST pass a fresh context.WithTimeout, not the cancelled root context.
func (s *Server) Shutdown(ctx context.Context) error { return s.httpServer.Shutdown(ctx) }

// Handler exposes the composed handler for in-process tests (httptest).
func (s *Server) Handler() http.Handler { return s.httpServer.Handler }

// handleLivez reports process liveness: if this handler runs, the process is
// alive. It must NOT depend on broker readiness, or the platform would kill a
// draining pod.
func (s *Server) handleLivez(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	s.writePlain(w, r, "ok")
}

// handleReadyz reports readiness to consume: the readiness flag is set AND the
// broker reports healthy. The shutdown sequence flips the flag to false before
// draining.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if !s.readiness.Ready() || !s.broker.Healthy() {
		w.WriteHeader(http.StatusServiceUnavailable)
		s.writePlain(w, r, "not ready")
		return
	}
	w.WriteHeader(http.StatusOK)
	s.writePlain(w, r, "ready")
}

// writePlain writes a short text probe body. The status is already committed,
// so a write error is unactionable for the client; it is logged once so a
// broken probe connection is observable rather than silently dropped.
func (s *Server) writePlain(w http.ResponseWriter, r *http.Request, body string) {
	if _, err := io.WriteString(w, body); err != nil {
		s.logger.WarnContext(r.Context(), "write probe body failed", "error", err.Error())
	}
}
