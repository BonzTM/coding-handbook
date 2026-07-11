namespace Orders.Worker.Core.Messaging;

/// <summary>
/// The seam that keeps real sleeps out of tests: production wires the
/// TimeProvider-backed implementation, tests wire a recording fake that
/// advances a FakeTimeProvider and returns immediately, so retry/backoff tests
/// are deterministic with no wall-clock waits
/// (csharp/foundations/time.md, csharp/services/eventing-and-messaging.md).
/// </summary>
public interface IRetryDelayer
{
    /// <summary>Waits for <paramref name="delay"/> or until
    /// <paramref name="cancellationToken"/> is cancelled, whichever comes
    /// first. Cancellation surfaces as <see cref="OperationCanceledException"/>
    /// so a draining consumer stops waiting immediately and nacks the message
    /// for redelivery instead of dead-lettering it.</summary>
    Task DelayAsync(TimeSpan delay, CancellationToken cancellationToken);
}
