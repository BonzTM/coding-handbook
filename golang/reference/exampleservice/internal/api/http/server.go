package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/example/exampleservice/internal/config"
	"github.com/example/exampleservice/internal/core"
	"github.com/example/exampleservice/internal/telemetry"
)

// service is the core behavior the transport depends on. It is defined here, at
// the consumer, naming exactly what the handlers call. *core.Service satisfies
// it; tests can substitute a fake.
type service interface {
	CreateWidget(ctx context.Context, id, name string) (core.Widget, error)
	GetWidget(ctx context.Context, id string) (core.Widget, error)
	ListWidgets(ctx context.Context) ([]core.Widget, error)
}

// Server owns the HTTP listener, mux, and middleware wiring. It holds the
// dependencies the handlers need and the readiness flag the shutdown sequence
// flips. It never stores a request context.
type Server struct {
	httpServer   *http.Server
	svc          service
	logger       *slog.Logger
	metrics      telemetry.Metrics
	readiness    *telemetry.Readiness
	maxBodyBytes int64
}

// New constructs a Server with hardened timeouts and the standard middleware
// chain. The readiness flag is shared with the shutdown sequence so it can flip
// the server to unready before draining.
func New(cfg config.HTTPConfig, svc service, logger *slog.Logger, metrics telemetry.Metrics, readiness *telemetry.Readiness) *Server {
	s := &Server{
		svc:          svc,
		logger:       logger,
		metrics:      metrics,
		readiness:    readiness,
		maxBodyBytes: cfg.MaxBodyBytes,
	}

	s.httpServer = &http.Server{
		Addr:    cfg.Addr,
		Handler: s.routes(),
		// Server-hardening defaults, per golang/services/http-services.md.
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}
	return s
}

// routes builds the request handler. Health probes and the application API are
// split onto two muxes so the claim that probes are NOT access-logged is true:
//
//   - The API mux carries the full chain (recovery -> request ID -> logging+
//     metrics -> handler), so every widgets request gets one access-log line
//     and increments request metrics.
//   - The probe mux carries ONLY recovery (and request ID for correlation). It
//     deliberately omits loggingMiddleware so frequent kubelet /livez & /readyz
//     polls neither flood the access log nor inflate request metric counts.
//
// A root mux dispatches by path: the specific probe patterns win over the "/"
// catch-all that fronts the logged API chain. Recovery wraps the root, so a
// panic in either branch is still converted to a 500.
func (s *Server) routes() http.Handler {
	// API mux: widgets routes get the full logging + metrics chain.
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("POST /widgets", s.handleCreateWidget)
	apiMux.HandleFunc("GET /widgets", s.handleListWidgets)
	apiMux.HandleFunc("GET /widgets/{id}", s.handleGetWidget)
	apiHandler := loggingMiddleware(s.logger, s.metrics)(apiMux)

	// Probe mux: liveness vs readiness are distinct endpoints, NOT access-logged
	// and NOT counted in request metrics.
	probeMux := http.NewServeMux()
	probeMux.HandleFunc("GET /livez", s.handleLivez)
	probeMux.HandleFunc("GET /readyz", s.handleReadyz)

	// Root mux: the specific probe patterns take precedence over the catch-all
	// that routes everything else through the logged API chain.
	root := http.NewServeMux()
	root.Handle("GET /livez", probeMux)
	root.Handle("GET /readyz", probeMux)
	root.Handle("/", apiHandler)

	// Request ID and recovery wrap both branches; recovery is outermost so a
	// panic anywhere becomes a 500. Logging lives inside the API branch only.
	var h http.Handler = root
	h = requestIDMiddleware(h)
	h = recoverMiddleware(s.logger)(h)
	return h
}

// ListenAndServe starts serving and blocks until the server is shut down. It
// returns http.ErrServerClosed on a clean Shutdown, which the caller treats as
// the expected outcome.
func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

// Addr returns the configured listen address.
func (s *Server) Addr() string { return s.httpServer.Addr }

// SetReady flips the readiness flag. The shutdown sequence calls SetReady(false)
// before draining so the load balancer stops routing new traffic.
func (s *Server) SetReady(ready bool) { s.readiness.Set(ready) }

// Shutdown gracefully stops the server, draining in-flight requests bounded by
// ctx's deadline. The caller MUST pass a fresh context.WithTimeout, not the
// cancelled root context.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Handler exposes the composed handler for in-process tests (httptest).
func (s *Server) Handler() http.Handler { return s.httpServer.Handler }
