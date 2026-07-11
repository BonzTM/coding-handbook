using System.Reflection;

using OpenTelemetry.Logs;
using OpenTelemetry.Metrics;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;

namespace Orders.Api.Telemetry;

/// <summary>
/// ALL telemetry wiring, called once from Program.cs - endpoints, workers, and
/// clients never configure exporters themselves
/// (csharp/operations/observability.md).
///
/// Every signal exports over OTLP; endpoint/protocol/headers come from the
/// standard OTEL_EXPORTER_OTLP_* environment variables (default
/// localhost:4317). There is no Prometheus /metrics endpoint by design: OTLP
/// push is the handbook default, and the OTel Prometheus exporter is added
/// only when the org scrapes - that decision is documented in this module's
/// README.
/// </summary>
internal static class ServiceTelemetryExtensions
{
    public static WebApplicationBuilder AddServiceTelemetry(this WebApplicationBuilder builder)
    {
        string version = typeof(ServiceTelemetryExtensions).Assembly
            .GetCustomAttribute<AssemblyInformationalVersionAttribute>()?.InformationalVersion ?? "dev";

        builder.Services.AddOpenTelemetry()
            .ConfigureResource(resource => resource.AddService(
                serviceName: OrdersTelemetry.ServiceName, serviceVersion: version))
            .WithTracing(tracing => tracing
                .AddSource(OrdersTelemetry.ActivitySourceName)
                .AddAspNetCoreInstrumentation()
                .AddHttpClientInstrumentation()
                .AddOtlpExporter())
            .WithMetrics(metrics => metrics
                .AddMeter(OrdersMetrics.MeterName)
                // Npgsql publishes connection-pool counters (busy/idle, pending
                // requests) - pool saturation must be visible before it becomes
                // a timeout (csharp/services/database.md).
                .AddMeter("Npgsql")
                .AddAspNetCoreInstrumentation()
                .AddHttpClientInstrumentation()
                .AddRuntimeInstrumentation()
                .AddOtlpExporter())
            .WithLogging(logging => logging.AddOtlpExporter());

        // Base registration; the "ready"-tagged database check is added by
        // AddOrdersInfrastructure, which owns the DbContext.
        builder.Services.AddHealthChecks();

        builder.Services.AddSingleton<OrdersMetrics>();
        builder.Services.AddSingleton<AuditLogger>();
        return builder;
    }
}

/// <summary>Telemetry identity: one place for the source names, or they silently export nothing.</summary>
internal static class OrdersTelemetry
{
    public const string ServiceName = "orders";
    public const string ActivitySourceName = "Orders.Api";
}
