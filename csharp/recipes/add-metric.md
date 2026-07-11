# Recipe: Add Metric

Use this when you need a new counter, histogram, or gauge to make a behavior observable in production.

Governing doc: [`csharp/operations/observability.md`](../operations/observability.md).

## Files To Touch

- `src/Orders.Api/Telemetry/OrdersMetrics.cs` — the shared telemetry construct: create the instrument here (see [../foundations/shared-constructs.md](../foundations/shared-constructs.md))
- the emitting class — call the metrics method at the point where the event happens
- `AddServiceTelemetry()` wiring — only if the instrument lives on a NEW `Meter` name that OpenTelemetry must subscribe to (`WithMetrics(m => m.AddMeter(...))`)
- telemetry tests and the emitting class's tests
- the metric's documentation (the XML doc comment on the emitting method is the canonical spec for its name and tags)

## Steps

1. Choose the instrument: `Counter<long>` for monotonic events (attempts, failures, items processed), `Histogram<double>` for distributions (latency, payload size), an observable gauge (`CreateObservableGauge`) for a point-in-time level (queue depth, ready replicas). Default to counter + histogram; reach for a gauge only when a snapshot is the correct model.
2. Choose a stable, namespaced name following OpenTelemetry conventions: dot-separated lowercase with the unit passed separately (`orders.created` with `unit: "{order}"`, `orders.fulfillment.duration` with `unit: "s"`). The name is a contract; downstream dashboards and alerts depend on it.
3. Choose FINITE, low-cardinality tags. Every tag value must come from a bounded set known at code-review time (route pattern, status class, outcome enum). If you cannot enumerate the values, it is not a tag.
4. Create the instrument in `OrdersMetrics`, whose constructor takes the injected `IMeterFactory` and calls `meterFactory.Create(OrdersMetrics.MeterName)` — never `new Meter(...)` in a static field and never a static instrument. Add a small, named emitting method (`OrderCreated(string outcome)`) that owns the `TagList`; document the metric name and every tag on that method.
5. Inject `OrdersMetrics` (registered once as a singleton) into the emitting class through the constructor; do not let the class reach for a static.
6. Emit at the right place: increment counters on the actual event (success and failure separately, or one counter with a finite `outcome` tag), record histograms around the operation, back gauges with a callback owned by the measured resource.
7. For histograms, choose bucket boundaries deliberately for the expected range via an OpenTelemetry view (`AddView(...)` with `ExplicitBucketHistogramConfiguration`) — latency buckets differ from byte-size buckets; do not accept defaults without checking they cover your distribution.

## Invariants To Preserve

- NO request IDs, user IDs, tenant IDs, raw paths, timestamps, or any unbounded value as a tag
- the `Meter` comes from the injected `IMeterFactory` through the telemetry construct, never a static `Meter` or static instrument
- the metric name and its tag set are a stable contract; renaming or adding a tag is a breaking change for dashboards and alerts
- failures and successes are distinguishable (separate counters or an outcome tag with a finite value set)
- histogram buckets are chosen for the measured distribution, not left implicit
- emitting classes depend on `OrdersMetrics` (the narrow construct), not on `Meter` or the OpenTelemetry SDK

## Proof

- a test asserting the instrument moves, using `MetricCollector<T>` (namespace `Microsoft.Extensions.Diagnostics.Metrics.Testing`, package `Microsoft.Extensions.Diagnostics.Testing`) — build an `IMeterFactory` from a `ServiceCollection` with `AddMetrics()`, point a collector at the meter name and instrument name, emit, then assert the snapshot:

  ```csharp
  using var collector = new MetricCollector<long>(meterFactory, OrdersMetrics.MeterName, "orders.created");
  metrics.OrderCreated("accepted");
  var m = Assert.Single(collector.GetMeasurementSnapshot());
  Assert.Equal(1, m.Value);
  Assert.Equal("accepted", m.Tags["outcome"]);
  ```

- run `pwsh ./verify.ps1`
- an export check showing the new series and its tag set: with the OTLP exporter wired per [../operations/observability.md](../operations/observability.md), confirm the series in the collector backend; where the org scrapes, `curl -s http://localhost:9464/metrics | grep orders_created` against a locally running host
- a cardinality sanity check: enumerate every possible value of each tag and confirm the set is finite and small (the scraped/exported series count for the metric stays bounded under load); cross-check against the forbidden-tag list in [../operations/observability.md](../operations/observability.md)
