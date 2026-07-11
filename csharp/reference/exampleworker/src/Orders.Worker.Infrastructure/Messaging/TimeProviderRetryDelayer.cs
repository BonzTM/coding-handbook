using Orders.Worker.Core.Messaging;

namespace Orders.Worker.Infrastructure.Messaging;

/// <summary>
/// Production <see cref="IRetryDelayer"/>: the TimeProvider-aware
/// <see cref="Task.Delay(TimeSpan, TimeProvider, CancellationToken)"/>
/// overload, so the wait is cancellable (a draining consumer stops
/// immediately) and tests that go through the real delayer can drive it with a
/// FakeTimeProvider (csharp/foundations/time.md).
/// </summary>
internal sealed class TimeProviderRetryDelayer(TimeProvider time) : IRetryDelayer
{
    public Task DelayAsync(TimeSpan delay, CancellationToken cancellationToken)
    {
        if (delay <= TimeSpan.Zero)
        {
            cancellationToken.ThrowIfCancellationRequested();
            return Task.CompletedTask;
        }

        return Task.Delay(delay, time, cancellationToken);
    }
}
