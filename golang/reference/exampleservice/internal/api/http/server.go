package http

import (
	"context"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/example/exampleservice/internal/auth"
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
	ListWidgetsPage(ctx context.Context, after core.Cursor, pageSize int) (core.Page, error)
}

// Server owns the HTTP listener, mux, and middleware wiring. It holds the
// dependencies the handlers need and the readiness flag the shutdown sequence
// flips. It never stores a request context.
type Server struct {
	httpServer     *http.Server
	svc            service
	logger         *slog.Logger
	metrics        telemetry.Metrics
	metricsHandler http.Handler
	readiness      *telemetry.Readiness
	maxBodyBytes   int64
	// verifier validates Bearer tokens. nil selects local/dev mode (auth
	// disabled): the auth middleware injects a synthetic principal so the
	// tenant-scoped service still functions offline.
	verifier auth.Verifier
	// idempotency persists idempotent-write responses for replay. Required.
	idempotency core.IdempotencyStore
	// clock stamps idempotency TTLs; injectable for deterministic tests.
	clock clock
}

// Deps bundles the identity and idempotency dependencies the server wires on top
// of the core service. Grouping them keeps New's signature stable as the
// enterprise layer grows and documents each seam in one place.
type Deps struct {
	// Verifier validates Bearer tokens; nil runs the service in local/dev mode
	// (AUTH_ENABLED=false) with a synthetic principal.
	Verifier auth.Verifier
	// Idempotency is the Idempotency-Key store for unsafe writes. Required.
	Idempotency core.IdempotencyStore
	// Clock stamps idempotency TTLs. Required (production wires the real clock).
	Clock clock
}

// metricsExposer is the optional seam a Metrics implementation can satisfy to
// publish a scrape endpoint. The Prometheus adapter implements it (returning a
// promhttp handler); the no-op and expvar seams do not, so /metrics is mounted
// only when a real registry is wired.
type metricsExposer interface {
	Handler() http.Handler
}

// New constructs a Server with hardened timeouts and the standard middleware
// chain. The readiness flag is shared with the shutdown sequence so it can flip
// the server to unready before draining.
func New(cfg config.HTTPConfig, svc service, logger *slog.Logger, metrics telemetry.Metrics, readiness *telemetry.Readiness, deps Deps) *Server {
	s := &Server{
		svc:          svc,
		logger:       logger,
		metrics:      metrics,
		readiness:    readiness,
		maxBodyBytes: cfg.MaxBodyBytes,
		verifier:     deps.Verifier,
		idempotency:  deps.Idempotency,
		clock:        deps.Clock,
	}
	// If the metrics impl publishes a scrape endpoint (the Prometheus adapter
	// does), mount it on the probe-side mux so it is neither access-logged nor
	// counted in request metrics.
	if exposer, ok := metrics.(metricsExposer); ok {
		s.metricsHandler = exposer.Handler()
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

// routes builds the request handler. Health probes, the scrape endpoint, and
// the application API are split onto separate muxes so the claim that probes and
// /metrics are NOT access-logged or over-instrumented is true:
//
//   - The API mux carries the full chain (recovery -> request ID -> otelhttp
//     span -> logging+metrics -> handler), so every widgets request is a span,
//     gets one access-log line, and increments request metrics.
//   - The probe mux carries ONLY recovery (and request ID for correlation). It
//     deliberately omits otelhttp and loggingMiddleware so frequent kubelet
//     /livez & /readyz polls and Prometheus /metrics scrapes neither flood the
//     access log, create per-scrape spans, nor inflate request metric counts.
//
// A root mux dispatches by path: the specific probe and /metrics patterns win
// over the "/" catch-all that fronts the logged API chain. Recovery wraps the
// root, so a panic in either branch is still converted to a 500.
func (s *Server) routes() http.Handler {
	// API mux: widgets routes get the full logging + metrics chain, wrapped in an
	// otelhttp span. otelhttp is innermost of the cross-cutting layers so the span
	// is open while logging records trace_id/span_id; the route pattern names the
	// span (low cardinality), never the raw path.
	apiMux := http.NewServeMux()
	// Per-route AUTHZ (requireRole) wraps each handler so the role requirement is
	// declared next to the route. The create route additionally runs the
	// Idempotency-Key middleware (innermost of the per-route layers, just before
	// the handler) so only an authorized write consumes a key. The per-resource
	// tenant ownership check is enforced inside the tenant-scoped store, so a
	// cross-tenant read is a 404 — authorization is at the boundary, not in a
	// helper.
	createHandler := idempotencyMiddleware(s.idempotency, s.clock, s.logger)(http.HandlerFunc(s.handleCreateWidget))
	apiMux.HandleFunc("POST /widgets", requireRole(core.RoleWriter, s.logger, createHandler.ServeHTTP))
	apiMux.HandleFunc("GET /widgets", requireRole(core.RoleReader, s.logger, s.handleListWidgets))
	apiMux.HandleFunc("GET /widgets/{id}", requireRole(core.RoleReader, s.logger, s.handleGetWidget))
	var apiHandler http.Handler = apiMux
	// Cross-cutting chain, innermost first: logging+metrics wrap the routed
	// handlers (so the matched route pattern and authz/idempotency outcome are in
	// the access log), AUTHN wraps logging (a rejected token is still logged and
	// traced), and otelhttp is outermost so the span is open for every layer.
	// Order of execution: otelhttp -> auth -> logging/metrics -> authz ->
	// idempotency -> handler.
	apiHandler = loggingMiddleware(s.logger, s.metrics)(apiHandler)
	apiHandler = authMiddleware(s.verifier, s.logger)(apiHandler)
	apiHandler = otelhttp.NewHandler(apiHandler, "http.server")

	// Probe mux: liveness vs readiness are distinct endpoints, NOT access-logged,
	// NOT traced, and NOT counted in request metrics. The Prometheus scrape
	// endpoint joins them when a registry is wired.
	probeMux := http.NewServeMux()
	probeMux.HandleFunc("GET /livez", s.handleLivez)
	probeMux.HandleFunc("GET /readyz", s.handleReadyz)
	if s.metricsHandler != nil {
		probeMux.Handle("GET /metrics", s.metricsHandler)
	}

	// Root mux: the specific probe and /metrics patterns take precedence over the
	// catch-all that routes everything else through the logged API chain.
	root := http.NewServeMux()
	root.Handle("GET /livez", probeMux)
	root.Handle("GET /readyz", probeMux)
	if s.metricsHandler != nil {
		root.Handle("GET /metrics", probeMux)
	}
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
