# Observability

Telemetry defaults for Python services and workers that must be diagnosable in production.

## Default Approach

| Signal | Default | Notes |
|---|---|---|
| logs | stdlib `logging` configured with `dictConfig` | structured JSON in production; one configuration owner |
| metrics | `prometheus-client` | explicit registration and low-cardinality labels |
| traces | OpenTelemetry | automatic adapter spans first; manual spans for owned units of work |
| health | `/livez`, `/readyz`, and `/metrics` | liveness is local; readiness is dependency-aware |

### Structured Logging

Follow [errors and logging](../foundations/errors-and-logging.md): composition configures handlers once, services emit JSON, and libraries add at most `NullHandler`. The stable schema includes `timestamp`, `level`, `service`, `environment`, `version`, `logger`, `operation`, `message`, and, when present, `request_id`, `trace_id`, `span_id`, `error_code`, and safe dependency or message context.
Use lowercase `snake_case` field names and stable machine values. Access logs record method, route template, status class, duration, and response size; never raw URL query strings. Workers record lifecycle, receive, retry, exhaustion, duplicate, settlement, and dead-letter transitions. Log a failure once at the boundary that maps, retries, compensates, or terminates it. Exception text, payloads, credentials, and PII are not fields.

### Correlation Context

Validate an inbound request or correlation ID against a bounded character/length policy, replace invalid input, and store the accepted value in a `contextvars.ContextVar`. Middleware sets the token and resets it in `finally`; background work copies only deliberate safe context. Propagate request/correlation IDs in approved HTTP and message headers and inject `trace_id`/`span_id` from the active OpenTelemetry context into records.
`contextvars` follows asyncio task context, but it is correlation plumbing, not hidden business state or resource ownership. Clear context before pooled worker reuse and test concurrent requests for isolation. IDs belong in logs and traces, never metric labels.

### Metrics

Name application metrics with a stable service namespace and base units: `_seconds`, `_bytes`, and `_total` where the Prometheus data model requires it. Use counters for events, histograms for latency/size, and gauges only for point-in-time state such as in-flight work or queue depth. The client exposes the standard metric types in its [instrumenting guide](https://prometheus.github.io/client_python/instrumenting/).
Labels describe small bounded dimensions such as route template, method, status class, dependency, operation, and outcome. Raw paths, exception messages, URLs, timestamps, request/message/user/tenant IDs, and any unrestricted external value are forbidden. Initialize expected label sets where zero-valued series matter; inject a `CollectorRegistry` into tests and reusable telemetry constructors instead of relying on the global registry.

### Tracing

Configure the OpenTelemetry SDK, resource attributes, sampler, propagators, processor, and exporter once in composition. Instrument FastAPI, the lifespan-owned HTTPX client, and SQLAlchemy at their adapters; the official instrumentations support [FastAPI](https://opentelemetry-python-contrib.readthedocs.io/en/latest/instrumentation/fastapi/fastapi.html), [HTTPX](https://opentelemetry-python-contrib.readthedocs.io/en/latest/instrumentation/httpx/httpx.html), and [SQLAlchemy](https://opentelemetry-python-contrib.readthedocs.io/en/latest/instrumentation/sqlalchemy/sqlalchemy.html). For an async engine, instrument its `sync_engine` as documented.
Automatic spans are the default for transport and library I/O. Add manual spans only around an owned use case, worker-message attempt, scheduled job, or expensive phase that automatic instrumentation cannot explain. Do not span trivial helpers or duplicate the client span. Attributes follow semantic conventions and the same low-cardinality/redaction rules as metrics.
Propagate W3C trace context through HTTP and approved message metadata. Producer send, consumer receive/process, retry, and settlement remain distinguishable. A broken or untrusted incoming trace never bypasses authorization or controls sampling policy.

### Health Endpoints

- `/livez` is cheap and local: process/event-loop viability only. It never calls a dependency.
- `/readyz` is bounded and reports whether this instance can accept work. It is false during startup, drain, or loss of a required dependency.
- `/metrics` exposes Prometheus text separately and is restricted by network/platform policy; it contains no topology secrets.
Readiness checks fan out within a short shared deadline, cache expensive results briefly, and return only safe component states. An optional dependency degrades its feature; it does not fail global readiness without a documented serving invariant.

### Sampling

Use parent-based probabilistic head sampling as the baseline, configured by environment through standard OpenTelemetry settings. Honor an upstream sampling decision inside the trust boundary. Keep errors and rare critical operations observable through metrics and logs; head sampling cannot retroactively retain every failed trace. Tail sampling requires an owned collector pipeline, capacity proof, privacy review, and ADR.
Never make application correctness depend on a span being sampled. Review sampling after traffic or cost changes and record the effective policy in the runbook.

## Common Mistakes And Forbidden Patterns

- Multiple logging configurations, library handlers, duplicate exception logs, or free-form production output.
- Secrets, PII, payloads, query strings, SQL values, or raw exception text in any telemetry signal.
- High-cardinality labels such as request, user, tenant, message, host, or raw-path values.
- Context variables used as an ambient dependency container, or request context leaked between tasks.
- Manual spans duplicating auto-instrumented HTTP/database spans or tracing every helper.
- Capturing all HTTP headers or database statements without explicit redaction and data review.
- Readiness that remains green during drain or critical dependency loss; liveness that queries dependencies.
- Sampling policy changed ad hoc per code path or treated as a reliability mechanism.

## Verification And Proof

```bash
uv run pytest -k "observability or telemetry or request_id or probe"
curl --fail http://localhost:8080/livez
curl --fail http://localhost:8080/readyz
curl --fail http://localhost:8080/metrics
make verify
```
Inspect representative JSON logs for the stable schema and redaction. Run concurrent requests to prove `contextvars` isolation and HTTP/message propagation. Trace one request and one worker message across core, HTTPX, SQLAlchemy, and settlement without duplicate spans. Break a required dependency and start drain: readiness turns false while liveness stays true. Review every metric label against a finite-value inventory and confirm no identifier or PII series exists.
Related: [errors and logging](../foundations/errors-and-logging.md), [HTTP services](../services/http-services.md), [data handling](data-handling.md), and [add metric](../recipes/add-metric.md).
