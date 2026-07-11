using System.Diagnostics.Metrics;

namespace Orders.Api.Telemetry;

/// <summary>
/// Domain-level instruments. Created through IMeterFactory (never a static
/// Meter) so tests get isolated meters; RED metrics per endpoint come from the
/// built-in ASP.NET Core instrumentation, not hand-rolled counters
/// (csharp/operations/observability.md).
/// </summary>
internal sealed class OrdersMetrics
{
    public const string MeterName = "Orders.Api";

    private readonly Counter<long> _ordersCreated;
    private readonly Counter<long> _idempotentReplays;

    public OrdersMetrics(IMeterFactory meterFactory)
    {
        ArgumentNullException.ThrowIfNull(meterFactory);
        // The factory owns the Meter it hands out: it caches instances and
        // disposes them with the container, so disposing here would break
        // other holders of the same cached Meter.
#pragma warning disable CA2000 // Dispose objects before losing scope
        var meter = meterFactory.Create(MeterName);
#pragma warning restore CA2000
        _ordersCreated = meter.CreateCounter<long>(
            "orders.created", unit: "{order}", description: "Orders successfully created.");
        _idempotentReplays = meter.CreateCounter<long>(
            "orders.idempotent_replays", unit: "{request}",
            description: "POST /orders requests answered from a stored idempotent response.");
    }

    public void OrderCreated() => _ordersCreated.Add(1);

    public void IdempotentReplay() => _idempotentReplays.Add(1);
}
