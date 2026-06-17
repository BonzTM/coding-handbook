package telemetry_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"

	"github.com/example/exampleservice/internal/telemetry"
)

func TestPromMetricsCountsRequests(t *testing.T) {
	m := telemetry.NewPromMetrics("exampleservice")

	m.IncRequest("GET /widgets", "2xx")
	m.IncRequest("GET /widgets", "2xx")
	m.IncRequest("POST /widgets", "4xx")
	m.IncWidgetCreated()
	m.IncWidgetCreated()

	const wantRequests = `
# HELP exampleservice_http_requests_total Total handled HTTP requests by route pattern and status class.
# TYPE exampleservice_http_requests_total counter
exampleservice_http_requests_total{route="GET /widgets",status_class="2xx"} 2
exampleservice_http_requests_total{route="POST /widgets",status_class="4xx"} 1
`
	if err := testutil.GatherAndCompare(m.Registry(), strings.NewReader(wantRequests), "exampleservice_http_requests_total"); err != nil {
		t.Errorf("requests_total mismatch: %v", err)
	}

	const wantCreated = `
# HELP exampleservice_widgets_created_total Total successfully created widgets.
# TYPE exampleservice_widgets_created_total counter
exampleservice_widgets_created_total 2
`
	if err := testutil.GatherAndCompare(m.Registry(), strings.NewReader(wantCreated), "exampleservice_widgets_created_total"); err != nil {
		t.Errorf("widgets_created_total mismatch: %v", err)
	}
}

func TestPromMetricsObserveRequestRecordsLatency(t *testing.T) {
	m := telemetry.NewPromMetrics("exampleservice")
	m.ObserveRequest("GET /widgets", "2xx", 0.05)
	m.ObserveRequest("GET /widgets", "2xx", 0.15)

	// The histogram must have observed two samples for the labeled series.
	count, err := histogramSampleCount(m, "GET /widgets", "2xx")
	if err != nil {
		t.Fatalf("read histogram: %v", err)
	}
	if count != 2 {
		t.Errorf("histogram sample count = %d, want 2", count)
	}
}

func TestPromMetricsHandlerExposesRegistry(t *testing.T) {
	m := telemetry.NewPromMetrics("exampleservice")
	m.IncRequest("GET /widgets", "2xx")

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("/metrics status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "exampleservice_http_requests_total") {
		t.Errorf("/metrics body missing request counter:\n%s", body)
	}
	// Runtime collectors are registered, so a Go runtime gauge is also exposed.
	if !strings.Contains(body, "go_goroutines") {
		t.Errorf("/metrics body missing go_goroutines (runtime collector not registered)")
	}
}

// histogramSampleCount gathers the registry and returns the observation count
// of the request-duration histogram for the given labels.
func histogramSampleCount(m *telemetry.PromMetrics, route, class string) (uint64, error) {
	families, err := m.Registry().Gather()
	if err != nil {
		return 0, err
	}
	for _, mf := range families {
		if mf.GetName() != "exampleservice_http_request_duration_seconds" {
			continue
		}
		for _, metric := range mf.GetMetric() {
			if labelsMatch(metric.GetLabel(), route, class) {
				return metric.GetHistogram().GetSampleCount(), nil
			}
		}
	}
	return 0, nil
}

func labelsMatch(labels []*dto.LabelPair, route, class string) bool {
	var gotRoute, gotClass string
	for _, l := range labels {
		switch l.GetName() {
		case "route":
			gotRoute = l.GetValue()
		case "status_class":
			gotClass = l.GetValue()
		}
	}
	return gotRoute == route && gotClass == class
}
