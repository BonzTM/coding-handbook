using System.Reflection;

using Microsoft.Extensions.Diagnostics.HealthChecks;

using OpenTelemetry.Logs;
using OpenTelemetry.Metrics;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;

namespace Orders.Grpc.Api.Telemetry;

/// <summary>
/// ALL telemetry wiring, called once from Program.cs - services and
/// interceptors never configure exporters themselves
/// (csharp/operations/observability.md).
///
/// Every signal exports over OTLP; endpoint/protocol/headers come from the
/// standard OTEL_EXPORTER_OTLP_* environment variables (default
/// localhost:4317). The ASP.NET Core instrumentation covers gRPC server spans
/// and RPC metrics - tracing is NOT an interceptor
/// (csharp/services/grpc-services.md). There is no Prometheus /metrics
/// endpoint by design: OTLP push is the handbook default; see this module's
/// README for the swap when the org scrapes.
/// </summary>
internal static class ServiceTelemetryExtensions
{
    public static WebApplicationBuilder AddServiceTelemetry(this WebApplicationBuilder builder)
    {
        string version = typeof(ServiceTelemetryExtensions).Assembly
            .GetCustomAttribute<AssemblyInformationalVersionAttribute>()?.InformationalVersion ?? "dev";

        builder.Services.AddOpenTelemetry()
            .ConfigureResource(resource => resource.AddService(
                serviceName: OrdersGrpcTelemetry.ServiceName, serviceVersion: version))
            .WithTracing(tracing => tracing
                .AddSource(OrdersGrpcTelemetry.ActivitySourceName)
                .AddAspNetCoreInstrumentation()
                .AddOtlpExporter())
            .WithMetrics(metrics => metrics
                .AddMeter(OrdersGrpcMetrics.MeterName)
                .AddAspNetCoreInstrumentation()
                .AddRuntimeInstrumentation()
                .AddOtlpExporter())
            .WithLogging(logging => logging.AddOtlpExporter());

        // Base registration; this module has no external dependency, so
        // readiness has no "ready"-tagged checks yet - a real service adds
        // them here (the keystone adds its database check). The SAME
        // registrations back /livez, /readyz, AND grpc.health.v1. The "self"
        // check exists because the gRPC health service aggregates REGISTERED
        // checks: with zero registrations it answers UNKNOWN, not SERVING.
        builder.Services.AddHealthChecks()
            .AddCheck("self", () => HealthCheckResult.Healthy());

        builder.Services.AddSingleton<OrdersGrpcMetrics>();
        return builder;
    }
}

/// <summary>Telemetry identity: one place for the source names, or they silently export nothing.</summary>
internal static class OrdersGrpcTelemetry
{
    public const string ServiceName = "orders-grpc";
    public const string ActivitySourceName = "Orders.Grpc.Api";
}
