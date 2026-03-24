# Observability

Telemetry defaults for Go services and workers that need to be debuggable in production.

## Default Approach

| Signal | Default | Notes |
|---|---|---|
| logs | `log/slog` | structured, machine-readable in production |
| metrics | Prometheus client | low-cardinality, explicit registration |
| traces | OpenTelemetry when the system is distributed or latency-sensitive | propagate context end to end |
| health | `/livez`, `/readyz`, and `/metrics` for services | readiness is dependency-aware, not just process-up |

### Logging

- Standard fields should include stable operation context such as `service`, `operation`, `request_id`, and `trace_id` when available.
- Access logs should record method, route pattern, status, duration, and major error class.
- Background workers should log start, stop, retries, and terminal failures.
- Message producers and consumers should log publish, receive, retry, exhaustion, duplicate drop, and dead-letter transitions with stable event metadata.

### Metrics

- Counters for attempts and failures.
- Histograms for latency.
- Gauges only when a point-in-time measurement is the right model.
- Inject registries into tests and reusable packages instead of relying on one global default registry everywhere.
- For messaging flows, track publish failures, consume failures, handler latency, retries, duplicate drops, backlog age, and DLQ count.

### Tracing

- Trace boundaries usually start at HTTP handlers, gRPC methods, worker loop iterations, and external client calls.
- Avoid creating spans for trivial helpers; trace units of work that help explain latency or failures.
- For asynchronous work, propagate correlation and trace context through message metadata and model producer send, consumer receive, processing, and settlement as the main spans.

## Common Mistakes And Forbidden Patterns

- High-cardinality labels such as request IDs, user IDs, raw paths, or timestamps.
- Logs that duplicate the same failure at every layer.
- Instrumentation hidden behind a bespoke abstraction so deep that standard tooling no longer fits.
- Readiness endpoints that stay green when the database or critical downstream is unavailable.
- Using message IDs, subjects, or tenant IDs as metric labels.

## Verification And Proof

- `curl http://localhost:8080/metrics` returns scrapeable output
- log review shows stable structured fields and no secret leakage
- one traced request or job run proves context propagation across core, storage, and external-client boundaries
- readiness toggles correctly when a critical dependency is unavailable
- messaging flows expose observable publish, retry, duplicate-drop, and DLQ behavior before release
