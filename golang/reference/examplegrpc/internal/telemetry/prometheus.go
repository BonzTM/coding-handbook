package telemetry

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PromMetrics is the Prometheus-backed metrics sink for the gRPC server. gRPC
// has no metrics endpoint of its own, so the RPC counters and latency histogram
// recorded here by the metrics interceptor are exposed over the HTTP sidecar's
// GET /metrics.
//
// Every label is deliberately low-cardinality (full method name, status code
// string), per golang/operations/observability.md: request IDs, user IDs,
// tenant IDs, and message payloads are NEVER used as labels.
type PromMetrics struct {
	registry   *prometheus.Registry
	rpcs       *prometheus.CounterVec
	rpcSeconds *prometheus.HistogramVec
}

// NewPromMetrics constructs a PromMetrics on a fresh, private registry (not the
// global default registry) so tests and multiple instances do not collide. The
// namespace prefixes every metric name, e.g. "examplegrpc_grpc_requests_total".
func NewPromMetrics(namespace string) *PromMetrics {
	reg := prometheus.NewRegistry()

	rpcs := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "grpc_requests_total",
		Help:      "Total handled gRPC requests by full method and status code.",
	}, []string{"grpc_method", "grpc_code"})

	rpcSeconds := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "grpc_request_duration_seconds",
		Help:      "gRPC request latency in seconds by full method and status code.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"grpc_method", "grpc_code"})

	reg.MustRegister(
		rpcs,
		rpcSeconds,
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	return &PromMetrics{registry: reg, rpcs: rpcs, rpcSeconds: rpcSeconds}
}

// ObserveRPC records one handled RPC: increments the counter and observes the
// latency under the same low-cardinality labels. method is the full gRPC method
// ("/widget.v1.WidgetService/GetWidget") and code is the canonical status code
// string ("OK", "NotFound", ...).
func (m *PromMetrics) ObserveRPC(method, code string, seconds float64) {
	m.rpcs.WithLabelValues(method, code).Inc()
	m.rpcSeconds.WithLabelValues(method, code).Observe(seconds)
}

// Handler returns the promhttp handler serving this instance's registry. It is
// mounted at GET /metrics on the sidecar server.
func (m *PromMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{Registry: m.registry})
}

// Registry exposes the underlying registry for tests that gather metrics
// directly.
func (m *PromMetrics) Registry() *prometheus.Registry { return m.registry }
