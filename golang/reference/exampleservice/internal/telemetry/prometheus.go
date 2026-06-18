package telemetry

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PromMetrics is the production-grade Metrics implementation backed by a
// dedicated Prometheus registry. It satisfies the same telemetry.Metrics seam
// the rest of the service consumes, so swapping it for NopMetrics/ExpvarMetrics
// is a wiring change in main, never a call-site change. It additionally records
// request latency via the optional ObserveRequest method.
//
// Every label is deliberately low-cardinality (matched route pattern, status
// class), per golang/operations/observability.md: request IDs, user IDs, tenant
// IDs, and raw paths are NEVER used as labels because they would blow up the
// time-series cardinality.
type PromMetrics struct {
	registry       *prometheus.Registry
	requests       *prometheus.CounterVec
	requestSeconds *prometheus.HistogramVec
	widgetsCreated prometheus.Counter
}

// NewPromMetrics constructs a PromMetrics on a fresh, private registry (not the
// global default registry) so tests and multiple instances do not collide. The
// namespace prefixes every metric name, e.g. "exampleservice_http_requests_total".
func NewPromMetrics(namespace string) *PromMetrics {
	reg := prometheus.NewRegistry()

	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "http_requests_total",
		Help:      "Total handled HTTP requests by route pattern and status class.",
	}, []string{"route", "status_class"})

	requestSeconds := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "http_request_duration_seconds",
		Help:      "HTTP request latency in seconds by route pattern and status class.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"route", "status_class"})

	widgetsCreated := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "widgets_created_total",
		Help:      "Total successfully created widgets.",
	})

	// Register on the private registry alongside the standard process and Go
	// runtime collectors so /metrics also exposes runtime gauges.
	reg.MustRegister(
		requests,
		requestSeconds,
		widgetsCreated,
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	return &PromMetrics{
		registry:       reg,
		requests:       requests,
		requestSeconds: requestSeconds,
		widgetsCreated: widgetsCreated,
	}
}

// IncRequest records one handled request. Both labels must be low-cardinality
// (a route pattern, not a raw path; a status class, not a code).
func (m *PromMetrics) IncRequest(routePattern, statusClass string) {
	m.requests.WithLabelValues(routePattern, statusClass).Inc()
}

// IncWidgetCreated increments the widgets-created counter.
func (m *PromMetrics) IncWidgetCreated() { m.widgetsCreated.Inc() }

// ObserveRequest records request latency under the same low-cardinality labels.
// The HTTP middleware calls IncRequest plus ObserveRequest once per served
// request; the split keeps the narrow Metrics interface unchanged while letting
// the Prometheus adapter also publish a latency histogram.
func (m *PromMetrics) ObserveRequest(routePattern, statusClass string, seconds float64) {
	m.requestSeconds.WithLabelValues(routePattern, statusClass).Observe(seconds)
}

// Handler returns the promhttp handler serving this instance's registry. It is
// mounted at GET /metrics ahead of the heavy middleware so the scrape endpoint
// is neither access-logged nor counted in request metrics.
func (m *PromMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{Registry: m.registry})
}

// Registry exposes the underlying registry for tests that gather metrics
// directly.
func (m *PromMetrics) Registry() *prometheus.Registry { return m.registry }
