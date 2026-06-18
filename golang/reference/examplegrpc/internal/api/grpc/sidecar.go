package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/example/examplegrpc/internal/config"
	"github.com/example/examplegrpc/internal/telemetry"
)

// Sidecar is the small HTTP server that exposes what gRPC cannot: the Prometheus
// /metrics scrape endpoint and the Kubernetes /livez and /readyz probes.
// Liveness is unconditional (the process is up); readiness reflects the shared
// telemetry.Readiness flag the shutdown sequence flips to false before draining.
type Sidecar struct {
	srv       *http.Server
	readiness *telemetry.Readiness
}

// NewSidecar builds the metrics/probes HTTP server. metricsHandler is the
// promhttp handler from telemetry.PromMetrics.
func NewSidecar(cfg config.HTTPConfig, metricsHandler http.Handler, readiness *telemetry.Readiness) *Sidecar {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metricsHandler)
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	s := &Sidecar{readiness: readiness}
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !s.readiness.Ready() {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	s.srv = &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}
	return s
}

// ListenAndServe starts the sidecar listener. It returns http.ErrServerClosed
// after a clean Shutdown, which the caller treats as success.
func (s *Sidecar) ListenAndServe() error {
	if err := s.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("sidecar serve: %w", err)
	}
	return nil
}

// Shutdown drains the sidecar under the given (bounded) context.
func (s *Sidecar) Shutdown(ctx context.Context) error {
	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("sidecar shutdown: %w", err)
	}
	return nil
}

// Addr returns the configured listen address for startup logging.
func (s *Sidecar) Addr() string { return s.srv.Addr }

// Listen opens a TCP listener on addr. It is a small helper so main can bind the
// gRPC listener explicitly (and surface a bind error early) rather than letting
// grpc.Serve own address resolution. A background context is used because the
// listener's lifetime is the process, not a single request.
func Listen(addr string) (net.Listener, error) {
	var lc net.ListenConfig
	lis, err := lc.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", addr, err)
	}
	return lis, nil
}
