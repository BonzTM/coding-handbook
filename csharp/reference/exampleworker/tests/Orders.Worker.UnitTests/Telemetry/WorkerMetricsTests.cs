using System.Diagnostics.Metrics;

using Orders.Worker.Infrastructure.Telemetry;
using Orders.Worker.UnitTests.Fakes;

using Xunit;

namespace Orders.Worker.UnitTests.Telemetry;

/// <summary>
/// The metrics surface via an in-proc MeterListener: counters carry only
/// low-cardinality tags, and the consumer-lag / DLQ-depth / outbox-pending
/// gauges report whatever their composition-supplied callbacks read
/// (csharp/services/eventing-and-messaging.md, Observability).
/// </summary>
public sealed class WorkerMetricsTests : IDisposable
{
    private readonly TestMeterFactory _meterFactory = new();
    private readonly MeterListener _listener = new();
    private readonly Dictionary<string, List<(long Value, Dictionary<string, object?> Tags)>> _measurements = [];

    public WorkerMetricsTests()
    {
        _listener.InstrumentPublished = (instrument, listener) =>
        {
            if (string.Equals(instrument.Meter.Name, WorkerMetrics.MeterName, StringComparison.Ordinal))
            {
                listener.EnableMeasurementEvents(instrument);
            }
        };
        _listener.SetMeasurementEventCallback<long>((instrument, value, tags, _) =>
        {
            var recorded = new Dictionary<string, object?>(StringComparer.Ordinal);
            foreach (var tag in tags)
            {
                recorded[tag.Key] = tag.Value;
            }

            lock (_measurements)
            {
                if (!_measurements.TryGetValue(instrument.Name, out var list))
                {
                    list = [];
                    _measurements[instrument.Name] = list;
                }

                list.Add((value, recorded));
            }
        });
        _listener.Start();
    }

    [Fact]
    public void Consumed_RecordsEventTypeAndOutcomeTags()
    {
        var metrics = new WorkerMetrics(_meterFactory);

        metrics.Consumed("orders.order-placed.v1", WorkerMetrics.Outcomes.Retry);

        var (value, tags) = Assert.Single(Measurements("orders.worker.messages.consumed"));
        Assert.Equal(1, value);
        Assert.Equal("orders.order-placed.v1", tags["event.type"]);
        Assert.Equal("retry", tags["outcome"]);
    }

    [Fact]
    public void Published_RecordsEventTypeTag()
    {
        var metrics = new WorkerMetrics(_meterFactory);

        metrics.Published("orders.order-placed.v1");

        var (value, tags) = Assert.Single(Measurements("orders.worker.messages.published"));
        Assert.Equal(1, value);
        Assert.Equal("orders.order-placed.v1", tags["event.type"]);
    }

    [Fact]
    public void QueueGauges_ReportTheRegisteredCallbacks()
    {
        var metrics = new WorkerMetrics(_meterFactory);
        metrics.RegisterQueueGauges(consumerLag: () => 7, dlqDepth: () => 2, outboxPending: () => 5);

        _listener.RecordObservableInstruments();

        Assert.Equal(7, Assert.Single(Measurements("orders.worker.consumer.lag")).Value);
        Assert.Equal(2, Assert.Single(Measurements("orders.worker.dlq.depth")).Value);
        Assert.Equal(5, Assert.Single(Measurements("orders.worker.outbox.pending")).Value);
    }

    public void Dispose()
    {
        _listener.Dispose();
        _meterFactory.Dispose();
    }

    private List<(long Value, Dictionary<string, object?> Tags)> Measurements(string instrument)
    {
        lock (_measurements)
        {
            return _measurements.TryGetValue(instrument, out var list) ? [.. list] : [];
        }
    }
}
