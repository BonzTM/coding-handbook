using System.Reflection;

using OpenTelemetry.Logs;
using OpenTelemetry.Metrics;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;

using Orders.Worker.Infrastructure.Telemetry;

namespace Orders.Worker.Telemetry;

/// <summary>
/// ALL telemetry wiring, called once from Program.cs - the consumer, relay,
/// and stores never configure exporters themselves
/// (csharp/operations/observability.md).
///
/// Every signal exports over OTLP; endpoint/protocol/headers come from the
/// standard OTEL_EXPORTER_OTLP_* environment variables (default
/// localhost:4317). There is no Prometheus /metrics endpoint by design: OTLP
/// push is the handbook default, and the OTel Prometheus exporter is added
/// only when the org scrapes - that decision is documented in this module's
/// README (the Go reference's sidecar scrapes; this module pushes).
/// </summary>
internal static class ServiceTelemetryExtensions
{
    public static WebApplicationBuilder AddServiceTelemetry(this WebApplicationBuilder builder)
    {
        string version = typeof(ServiceTelemetryExtensions).Assembly
            .GetCustomAttribute<AssemblyInformationalVersionAttribute>()?.InformationalVersion ?? "dev";

        builder.Services.AddOpenTelemetry()
            .ConfigureResource(resource => resource.AddService(
                serviceName: WorkerTelemetry.ServiceName, serviceVersion: version))
            .WithTracing(tracing => tracing
                .AddSource(WorkerTelemetry.ActivitySourceName)
                .AddAspNetCoreInstrumentation()
                .AddOtlpExporter())
            .WithMetrics(metrics => metrics
                .AddMeter(WorkerMetrics.MeterName)
                .AddAspNetCoreInstrumentation()
                .AddRuntimeInstrumentation()
                .AddOtlpExporter())
            .WithLogging(logging => logging.AddOtlpExporter());

        // Base registration; the "ready"-tagged broker check is added by
        // AddOrdersWorkerMessaging, which owns the broker.
        builder.Services.AddHealthChecks();
        return builder;
    }
}
