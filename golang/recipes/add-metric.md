# Recipe: Add Metric

Use this when you need a new counter, histogram, or gauge to make a behavior observable in production.

## Files To Touch

- `internal/telemetry` — extend the metrics seam and register the instrument on the injected registry/meter (see `golang/reference/exampleservice/internal/telemetry/telemetry.go`)
- the emitting package — call the seam method at the point where the event happens
- `internal/telemetry` tests and the emitting package's tests
- the metric's documentation (the seam method's doc comment is the canonical spec for its name and labels)

## Steps

1. Choose the instrument: counter for monotonic events (attempts, failures, items processed), histogram for distributions (latency, payload size), gauge for a point-in-time level (queue depth, ready replicas). Default to counter + histogram; reach for a gauge only when a snapshot is the correct model.
2. Choose a stable, namespaced name following the producing system's convention (e.g. `<service>_widgets_created_total`, `<service>_http_request_duration_seconds`). The name is a contract; downstream dashboards and alerts depend on it.
3. Choose FINITE, low-cardinality labels. Every label value must come from a bounded set known at code-review time (route pattern, status class, outcome enum). If you cannot enumerate the values, it is not a label.
4. Add a method to the `Metrics` interface in `internal/telemetry`, document its name and labels on the method, and implement it in `NopMetrics` (no-op) and the real backend. Register the instrument on the INJECTED registry/meter passed into the constructor — never `prometheus.DefaultRegisterer` or any package-global.
5. Pass the metrics seam through the existing telemetry wiring into the emitting package; do not let the package reach for a global.
6. Emit at the right place: increment counters on the actual event (success and failure separately), observe histograms around the operation, set gauges from the owner of the measured resource.
7. For histograms, choose buckets deliberately for the expected range (latency buckets differ from byte-size buckets); do not accept defaults without checking they cover your distribution.

## Invariants To Preserve

- NO request IDs, user IDs, tenant IDs, raw paths, timestamps, or any unbounded value as a label
- the registry/meter is injected through the telemetry seam, never a global default
- the metric name and its label set are a stable contract; renaming or adding a label is a breaking change for dashboards and alerts
- failures and successes are distinguishable (separate counters or an outcome label with a finite value set)
- histogram buckets are chosen for the measured distribution, not left implicit
- reusable packages depend on the small `Metrics` interface, not a concrete client

## Proof

- a test asserting the instrument moves: `go test ./internal/telemetry/... -run Metrics` (mirror `TestExpvarMetricsCounts` — emit, then read the series value back with `expvar.Get` and assert the count/observation)
- `make verify`
- a scrape showing the new series and its label set against a production build: `curl -s http://localhost:8080/metrics | grep <metric_name>` (the Prometheus `/metrics` handler is wired in production per `golang/operations/observability.md`; the reference build's expvar seam is read in-process by the test above and is not mounted on an HTTP endpoint)
- a cardinality sanity check: enumerate every possible value of each label and confirm the set is finite and small (against a production build, `curl -s http://localhost:8080/metrics | grep <metric_name> | wc -l` stays bounded under load); cross-check against the forbidden-label list in `golang/operations/observability.md`

See `golang/operations/observability.md` for the metrics defaults (counters for attempts/failures, histograms for latency, gauges only for point-in-time, injected registries) and the forbidden high-cardinality patterns, and `golang/reference/exampleservice/internal/telemetry/telemetry.go` for the injected-seam pattern with low-cardinality labels.
