# Recipe: Add Metric

Use this when production behavior needs a new counter, histogram, or gauge.

## Files To Touch

- `src/<app>/telemetry/metrics.py`
- the emitting adapter/use-case boundary
- metric tests and dashboards/alerts when they consume the series

## Steps

1. Use a counter for monotonic events, histogram for distributions, and gauge only for owned point-in-time state.
2. Choose one stable namespaced name, base unit, help text, and finite label set before implementation.
3. Register the collector once in the injected registry; do not register at import or against global state in reusable code.
4. Emit at the point where the outcome is known. Choose histogram buckets from expected/service-level ranges.
5. Review the aggregate series count across labels, replicas, and workers.

## Invariants To Preserve

- No request, user, tenant, event, raw URL/path, timestamp, or exception text as a label.
- Metric name and label set are compatibility contracts.
- Success, error, retry, and skipped outcomes remain distinguishable without unbounded values.
- Core does not import prometheus-client; use a narrow telemetry seam when core must report behavior.

## Proof

```bash
uv run pytest tests/telemetry -k metric
curl --fail --silent http://localhost:8080/metrics | grep '<metric_name>'
make verify
```

Assert collector values and enumerate every possible label value in review. Governing doc: [observability](../operations/observability.md).
