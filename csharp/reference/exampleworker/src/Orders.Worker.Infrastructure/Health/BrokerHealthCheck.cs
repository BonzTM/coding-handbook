using Microsoft.Extensions.Diagnostics.HealthChecks;

using Orders.Worker.Infrastructure.Messaging;

namespace Orders.Worker.Infrastructure.Health;

/// <summary>
/// The "ready"-tagged health check behind /readyz: healthy only when the
/// consumer is subscribed AND the broker reports connectivity. /livez never
/// runs it - an unreachable broker means shed work, not restart the pod
/// (csharp/operations/observability.md).
/// </summary>
internal sealed class BrokerHealthCheck(InMemoryBroker broker, WorkerReadiness readiness) : IHealthCheck
{
    public Task<HealthCheckResult> CheckHealthAsync(
        HealthCheckContext context, CancellationToken cancellationToken = default)
    {
        if (!readiness.IsReady)
        {
            return Task.FromResult(HealthCheckResult.Unhealthy("Consumer is not subscribed."));
        }

        if (!broker.IsHealthy)
        {
            return Task.FromResult(HealthCheckResult.Unhealthy("Broker connection is unavailable."));
        }

        return Task.FromResult(HealthCheckResult.Healthy("Consuming."));
    }
}
