# Observability

Logs, traces, metrics, health, and correlation rules for operable TypeScript systems.

## Default Approach

Use Pino for structured logs and OpenTelemetry JS for traces and metrics exported through OTLP.

### Signal Ownership

OpenTelemetry JS traces and metrics are stable. Its logs SDK remains in development, so Pino stays the production logging path. Do not route application logs through an experimental logs SDK merely to claim one telemetry API.

Composition constructs the logger, tracer/meter providers, resource identity, exporters, sampling, and shutdown hooks. Core logic accepts narrow telemetry seams only when domain-significant instrumentation is required.

### Structured Logs

Emit JSON with stable event names and fields: service, version, environment, component, operation, result, request/trace identifiers, and typed error identity. Configure Pino redaction at construction and log once at the acting boundary.

Request IDs, user IDs, message IDs, and raw URLs belong in logs and traces when policy permits, never in metric attributes. Avoid payload logging; record safe size, schema version, or resource references instead.

### Traces

Create server, consumer, scheduled-job, and outbound dependency spans at owned boundaries. Propagate W3C trace context across HTTP and messages but never use trace headers as authentication.

Set span status from operation outcome, record safe typed exceptions, and end spans in every success, failure, timeout, and cancellation path. Sampling is an operational policy; errors and high-value transactions need an intentional retention strategy.

### Metrics

Metrics answer bounded operational questions. Use counters for events, histograms for latency/size, and observable gauges for current backlog or resources where collection is safe.

Attributes use controlled vocabularies such as route template, method, status class, operation, dependency, event type, and result. Never use raw path, SQL, cache key, principal, tenant, request ID, or error message.

Define units and histogram boundaries. A metric without an owner, dashboard use, or alert question is noise.

### Health Endpoints

`/livez` answers whether the process is alive and should be restarted only for a stuck/fatal process. `/readyz` answers whether it can accept new work. Startup readiness remains false until required initialization completes.

Checks are fast, bounded, and safe. Do not expose config, dependency URLs, versions with exploitable detail, or credentials. Optional dependency degradation should not cause a restart loop.

### Correlation And Privacy

Carry trace and request context with the async operation rather than mutable globals. Validate inbound correlation values for length and format, or replace them. Apply data classification and retention to telemetry like any other data store.

### Export Failure And Shutdown

Telemetry must not block core service behavior indefinitely. Export has timeout, queue, retry, and memory bounds. On shutdown, stop intake, drain application work, then flush providers within the remaining deadline.

## Common Mistakes And Forbidden Patterns

- Using the OpenTelemetry logs SDK as the default while its status remains development.
- Logging at every layer or serializing raw requests, config, rows, and errors.
- High-cardinality metric attributes such as IDs, URLs, SQL, or messages.
- Readiness that performs slow unbounded dependency fan-out.
- Liveness tied to an optional downstream outage, causing restart storms.
- Traces or correlation headers trusted as authorization state.
- Exporters allowed unbounded queues or shutdown time.

## Verification And Proof

- Telemetry smoke tests export one trace and metric through the configured OTLP path.
- Captured Pino events prove stable fields, one log, correct level, and redaction.
- Metric review demonstrates bounded attribute values and declared units.
- `/livez` and `/readyz` tests cover startup, ready, draining, and dependency degradation.
- Trace tests cover inbound/outbound propagation, error, timeout, and cancellation.
- Shutdown flushes telemetry within its deadline without delaying termination indefinitely.

Related: [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md), [operability.md](operability.md), and [data-handling.md](data-handling.md). Status anchor: [OpenTelemetry JavaScript](https://opentelemetry.io/docs/languages/js/).
