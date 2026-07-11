using System.Diagnostics.Metrics;

namespace Orders.Worker.UnitTests.Fakes;

/// <summary>Minimal IMeterFactory so tests get isolated Meter instances -
/// mirroring the DI factory's contract of owning and disposing what it
/// creates (csharp/operations/observability.md).</summary>
internal sealed class TestMeterFactory : IMeterFactory
{
    private readonly List<Meter> _meters = [];

    public Meter Create(MeterOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);
        var meter = new Meter(options.Name, options.Version, options.Tags, scope: this);
        _meters.Add(meter);
        return meter;
    }

    public void Dispose()
    {
        foreach (var meter in _meters)
        {
            meter.Dispose();
        }

        _meters.Clear();
    }
}
