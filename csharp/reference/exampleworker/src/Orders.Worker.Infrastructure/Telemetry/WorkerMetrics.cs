using System.Diagnostics;
using System.Diagnostics.Metrics;

namespace Orders.Worker.Infrastructure.Telemetry;

/// <summary>Telemetry identity: one place for the source names, or they
/// silently export nothing.</summary>
public static class WorkerTelemetry
{
    public const string ServiceName = "orders-worker";
    public const string ActivitySourceName = "Orders.Worker";

    /// <summary>Consumer-side spans (receive -> handle -> settle), per
    /// csharp/services/eventing-and-messaging.md (Observability). Static and
    /// never disposed on purpose: it lives exactly as long as the process.</summary>
    public static readonly ActivitySource ActivitySource = new(ActivitySourceName);
}

/// <summary>
/// The worker's domain instruments. Created through IMeterFactory (never a
/// static Meter) so tests get isolated meters (csharp/operations/observability.md).
/// Every tag is deliberately low-cardinality (event type, outcome class):
/// message ids, correlation ids, and tenants belong in logs and traces, never
/// in metric labels (csharp/services/eventing-and-messaging.md).
/// </summary>
public sealed class WorkerMetrics
{
    public const string MeterName = "Orders.Worker";

    /// <summary>Consume outcomes, matching the pipeline's terminal decisions.</summary>
    public static class Outcomes
    {
        public const string Ack = "ack";
        public const string Retry = "retry";
        public const string DroppedDuplicate = "dropped_duplicate";
        public const string DeadLettered = "dead_lettered";
    }

    private readonly Meter _meter;
    private readonly Counter<long> _consumed;
    private readonly Counter<long> _published;

    public WorkerMetrics(IMeterFactory meterFactory)
    {
        ArgumentNullException.ThrowIfNull(meterFactory);
        // The factory owns the Meter it hands out: it caches instances and
        // disposes them with the container, so disposing here would break
        // other holders of the same cached Meter.
#pragma warning disable CA2000 // Dispose objects before losing scope
        _meter = meterFactory.Create(MeterName);
#pragma warning restore CA2000
        _consumed = _meter.CreateCounter<long>(
            "orders.worker.messages.consumed", unit: "{message}",
            description: "Consumed messages by event type and outcome.");
        _published = _meter.CreateCounter<long>(
            "orders.worker.messages.published", unit: "{message}",
            description: "Messages published by the outbox relay, by event type.");
    }

    public void Consumed(string eventType, string outcome)
        => _consumed.Add(1,
            new KeyValuePair<string, object?>("event.type", eventType),
            new KeyValuePair<string, object?>("outcome", outcome));

    public void Published(string eventType)
        => _published.Add(1, new KeyValuePair<string, object?>("event.type", eventType));

    /// <summary>
    /// Registers the queue-depth gauges. Called once from the composition root
    /// with callbacks into the in-memory broker and stores; a production
    /// deployment derives consumer lag from the broker's own metrics and DLQ /
    /// outbox depth from monitored queries - the gauge NAMES are the contract,
    /// the callbacks are this module's in-memory stand-in. Consumer lag is the
    /// earliest signal a consumer is losing; alert on it, not on the eventual
    /// timeout storm (csharp/services/eventing-and-messaging.md).
    /// </summary>
    public void RegisterQueueGauges(Func<long> consumerLag, Func<long> dlqDepth, Func<long> outboxPending)
    {
        ArgumentNullException.ThrowIfNull(consumerLag);
        ArgumentNullException.ThrowIfNull(dlqDepth);
        ArgumentNullException.ThrowIfNull(outboxPending);
        _meter.CreateObservableGauge(
            "orders.worker.consumer.lag", consumerLag, unit: "{message}",
            description: "Messages waiting on the subscribed topic (backlog).");
        _meter.CreateObservableGauge(
            "orders.worker.dlq.depth", dlqDepth, unit: "{message}",
            description: "Messages parked in the dead-letter store.");
        _meter.CreateObservableGauge(
            "orders.worker.outbox.pending", outboxPending, unit: "{message}",
            description: "Outbox records not yet relayed to the broker.");
    }
}
