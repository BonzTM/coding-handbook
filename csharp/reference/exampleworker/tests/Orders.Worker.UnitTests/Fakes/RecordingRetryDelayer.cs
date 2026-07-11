using Microsoft.Extensions.Time.Testing;

using Orders.Worker.Core.Messaging;

namespace Orders.Worker.UnitTests.Fakes;

/// <summary>
/// Clock-driven <see cref="IRetryDelayer"/> that NEVER sleeps: it records each
/// requested backoff so tests can assert the exact bounds, advances the
/// FakeTimeProvider by the requested delay (so time-stamped assertions hold),
/// and returns immediately. This is how the retry-then-DLQ test runs with zero
/// wall-clock waits (csharp/foundations/time.md).
/// </summary>
internal sealed class RecordingRetryDelayer(FakeTimeProvider time) : IRetryDelayer
{
    private readonly Lock _gate = new();
    private readonly List<TimeSpan> _delays = [];

    public IReadOnlyList<TimeSpan> Delays
    {
        get
        {
            lock (_gate)
            {
                return [.. _delays];
            }
        }
    }

    public Task DelayAsync(TimeSpan delay, CancellationToken cancellationToken)
    {
        cancellationToken.ThrowIfCancellationRequested();
        lock (_gate)
        {
            _delays.Add(delay);
        }

        time.Advance(delay);
        return Task.CompletedTask;
    }
}
