package telemetry

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PromMetrics is the production-grade Metrics implementation backed by a
// dedicated Prometheus registry. It satisfies the same telemetry.Metrics seam
// the messaging layer consumes, so swapping it for NopMetrics is a wiring
// change in main, never a call-site change. It additionally records handler
// latency via the optional ObserveProcess method.
//
// Every label is deliberately low-cardinality (event type, outcome class), per
// golang/services/eventing-and-messaging.md ### Observability: message IDs,
// correlation IDs, tenant IDs, and raw subjects are NEVER used as labels
// because they would blow up the time-series cardinality.
type PromMetrics struct {
	registry       *prometheus.Registry
	consumed       *prometheus.CounterVec
	published      *prometheus.CounterVec
	processSeconds *prometheus.HistogramVec
}

// NewPromMetrics constructs a PromMetrics on a fresh, private registry (not the
// global default registry) so tests and multiple instances do not collide. The
// namespace prefixes every metric name, e.g. "exampleworker_messages_consumed_total".
func NewPromMetrics(namespace string) *PromMetrics {
	reg := prometheus.NewRegistry()

	consumed := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "messages_consumed_total",
		Help:      "Total consumed messages by event type and outcome.",
	}, []string{"event_type", "outcome"})

	published := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "messages_published_total",
		Help:      "Total messages published by the outbox relay by event type.",
	}, []string{"event_type"})

	processSeconds := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "message_process_duration_seconds",
		Help:      "Message handler latency in seconds by event type and outcome.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"event_type", "outcome"})

	// Register on the private registry alongside the standard process and Go
	// runtime collectors so /metrics also exposes runtime gauges.
	reg.MustRegister(
		consumed,
		published,
		processSeconds,
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	return &PromMetrics{
		registry:       reg,
		consumed:       consumed,
		published:      published,
		processSeconds: processSeconds,
	}
}

// IncConsumed records one consumed message. Both labels must be low-cardinality
// (an event type, not a message ID; an outcome class, not a free-form reason).
func (m *PromMetrics) IncConsumed(eventType, outcome string) {
	m.consumed.WithLabelValues(eventType, outcome).Inc()
}

// IncPublished increments the published counter for an event type.
func (m *PromMetrics) IncPublished(eventType string) {
	m.published.WithLabelValues(eventType).Inc()
}

// ObserveProcess records handler latency under the same low-cardinality labels.
func (m *PromMetrics) ObserveProcess(eventType, outcome string, seconds float64) {
	m.processSeconds.WithLabelValues(eventType, outcome).Observe(seconds)
}

// Handler returns the promhttp handler serving this instance's registry. It is
// mounted at GET /metrics on the probe sidecar.
func (m *PromMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{Registry: m.registry})
}

// Registry exposes the underlying registry for tests that gather metrics
// directly.
func (m *PromMetrics) Registry() *prometheus.Registry { return m.registry }
