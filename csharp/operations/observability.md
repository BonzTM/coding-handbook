# Observability

Telemetry defaults for .NET services and workers that need to be debuggable in production.

## Default Approach

All three signals flow through OpenTelemetry with OTLP export as the default; the Prometheus-scrape exporter is added only when the org scrapes.

| Signal | Default | Notes |
|---|---|---|
| logs | `ILogger<T>` + `[LoggerMessage]`, bridged through the OpenTelemetry logging provider | structured, machine-readable in production |
| metrics | `Meter` (`System.Diagnostics.Metrics`) via OpenTelemetry | low-cardinality, explicit instrument registration |
| traces | `ActivitySource` via OpenTelemetry when the system is distributed or latency-sensitive | propagate context end to end |
| export | OTLP exporter for all three signals | OTel Prometheus exporter only when the org scrapes |
| health | `/livez` and `/readyz` for services | readiness is dependency-aware, not just process-up |

### Wiring

All telemetry wiring lives in one `AddServiceTelemetry()` extension (in `<App>.Api/Telemetry/` or the shared `ServiceDefaults`-style project; see [../foundations/shared-constructs.md](../foundations/shared-constructs.md)). `Program.cs` calls it once; endpoints, workers, and clients never configure exporters themselves.

```csharp
namespace Orders.Api.Telemetry;

public static class ServiceTelemetryExtensions
{
    public static WebApplicationBuilder AddServiceTelemetry(this WebApplicationBuilder builder)
    {
        builder.Services.AddOpenTelemetry()
            .ConfigureResource(r => r.AddService(serviceName: "orders"))
            .WithTracing(t => t
                .AddSource(OrdersTelemetry.ActivitySourceName)
                .AddAspNetCoreInstrumentation()
                .AddHttpClientInstrumentation()
                .AddOtlpExporter())
            .WithMetrics(m => m
                .AddMeter(OrdersTelemetry.MeterName)
                .AddAspNetCoreInstrumentation()
                .AddHttpClientInstrumentation()
                .AddRuntimeInstrumentation()
                .AddOtlpExporter())
            .WithLogging(l => l.AddOtlpExporter());

        builder.Services.AddHealthChecks();
        return builder;
    }
}
```

`OrdersTelemetry` holds the single `ActivitySource` and the meter name as constants; a source or meter that is not registered here exports nothing, silently.

### Logging

- Standard fields include stable operation context: `service`, `operation`, and `RequestId`; `TraceId`/`SpanId` are attached automatically by the OpenTelemetry bridge from `Activity.Current` — never hand-copy them into message templates.
- Use source-generated `[LoggerMessage]` methods with structured templates and stable event names; see [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md) for the log schema and log-once rule.
- Access logs record method, route pattern, status, duration, and major error class — the built-in ASP.NET Core instrumentation provides this; do not add a second access-log layer.
- Background workers log start, stop, retries, and terminal failures.
- Message producers and consumers log publish, receive, retry, exhaustion, duplicate drop, and dead-letter transitions with stable event metadata (see [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md)).

### Metrics

- Counters for attempts and failures; histograms for latency; gauges (`ObservableGauge`) only when a point-in-time measurement is the right model.
- Create instruments through `IMeterFactory` injected via DI — never a `static Meter` — so tests get isolated meters instead of one process-global default. See [../recipes/add-metric.md](../recipes/add-metric.md).
- RED per endpoint (rate, errors, duration): the built-in `http.server.request.duration` histogram, tagged with route pattern and status code, covers all three. Add domain meters only for business-level events.
- Labels stay low-cardinality: route *pattern*, status class, dependency name, result. Never request IDs, user IDs, raw paths, or timestamps.
- For messaging flows, track publish failures, consume failures, handler latency, retries, duplicate drops, backlog age, and DLQ count.

### Tracing

- Trace boundaries start at HTTP endpoints, gRPC methods, worker loop iterations, and outbound client calls — the ASP.NET Core and HttpClient instrumentation covers the edges; start `Activity` spans from the service's `ActivitySource` only for units of work that explain latency or failures.
- Avoid creating spans for trivial helpers.
- For asynchronous work, propagate `traceparent` through message metadata and model producer send, consumer receive, processing, and settlement as the main spans.

### Health Endpoints

Liveness answers "should the platform restart this process"; readiness answers "should this instance receive traffic". They are different questions with different costs.

```csharp
app.MapHealthChecks("/livez", new HealthCheckOptions { Predicate = _ => false });
app.MapHealthChecks("/readyz", new HealthCheckOptions { Predicate = c => c.Tags.Contains("ready") });
```

- `/livez` runs no dependency checks — cheap, allocation-light, and never coupled to downstreams. A dead database must not cause a restart loop.
- `/readyz` runs the checks tagged `ready` (database, broker, critical downstream) and gates traffic: not-ready removes the instance from rotation while it keeps running.
- Register dependency checks with `AddHealthChecks().AddCheck(..., tags: ["ready"])`; wiring lives in `AddServiceTelemetry()` or beside it, never per-endpoint.

### Sampling

- Traces: parent-based head sampling. Development runs always-on; production runs a parent-based ratio configured via standard OTel configuration (environment), so a decision made at the edge is honored end to end.
- Metrics are aggregates — never sampled.
- Logs are not sampled at the source; control volume with level configuration per category. The audit stream is never sampled or rate-limited (see [security.md](security.md)).

## Common Mistakes And Forbidden Patterns

- High-cardinality labels such as request IDs, user IDs, raw paths, or timestamps.
- Logs that duplicate the same failure at every layer, or `Console.WriteLine`/string interpolation that destroys structure.
- A `static Meter` or per-request `ActivitySource`/`Meter` construction instead of one registered source and `IMeterFactory` injection.
- An `ActivitySource` or `Meter` name missing from `AddSource`/`AddMeter`, so its telemetry is silently dropped.
- Instrumentation hidden behind a bespoke abstraction so deep that standard OTel tooling no longer fits.
- `/livez` that checks dependencies (restart loops on a downstream outage) or `/readyz` that stays green when the database or a critical downstream is unavailable.
- Using message IDs, subjects, or tenant IDs as metric labels.

## Verification And Proof

- OTLP export verified against a local collector (or `curl` the Prometheus endpoint when the scrape exporter is enabled) shows the service resource, RED histograms per route, and domain instruments
- log review shows stable structured fields, `TraceId` present on request-scoped lines, and no secret leakage
- one traced request or job run proves context propagation across endpoint, core, storage, and external-client boundaries
- `/livez` stays green and `/readyz` toggles correctly when a critical dependency is stopped (Testcontainers makes this a cheap integration test)
- messaging flows expose observable publish, retry, duplicate-drop, and DLQ behavior before release
- run `pwsh ./verify.ps1` — telemetry wiring compiles under warnings-as-errors and unit tests cover the meter-backed code paths via `IMeterFactory`
