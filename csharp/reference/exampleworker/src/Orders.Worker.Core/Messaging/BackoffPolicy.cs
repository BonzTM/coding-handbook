namespace Orders.Worker.Core.Messaging;

/// <summary>
/// Bounded exponential backoff with FULL jitter, per
/// csharp/services/eventing-and-messaging.md (Retries And Dead-Letter
/// Behavior): the un-jittered ceiling doubles each attempt up to
/// <paramref name="maxDelay"/>, and the actual delay is a uniform draw in
/// [0, ceiling]. The jitter source is injectable so tests can pin it and
/// assert exact delays; the policy itself never sleeps - the caller awaits the
/// returned duration through <see cref="IRetryDelayer"/> and the injected
/// <see cref="TimeProvider"/>.
/// </summary>
public sealed class BackoffPolicy(TimeSpan baseDelay, TimeSpan maxDelay, Func<double>? jitterSource = null)
{
    private readonly Func<double> _jitterSource = jitterSource ?? DefaultJitter;

    /// <summary>The un-jittered exponential ceiling for a 1-based attempt:
    /// baseDelay * 2^(attempt-1), capped at maxDelay. Attempts below 1 clamp
    /// to 1, so a caller bug can never produce a negative or unbounded wait.</summary>
    public TimeSpan Ceiling(int attempt)
    {
        if (attempt < 1)
        {
            attempt = 1;
        }

        var ceiling = baseDelay;
        for (int i = 1; i < attempt; i++)
        {
            if (ceiling >= maxDelay)
            {
                return maxDelay;
            }

            ceiling *= 2;
        }

        return ceiling > maxDelay ? maxDelay : ceiling;
    }

    /// <summary>The jittered wait for a 1-based attempt: a uniform draw in
    /// [0, <see cref="Ceiling"/>]. Always bounded by the exponential ceiling,
    /// which the retry-then-DLQ test asserts.</summary>
    public TimeSpan NextDelay(int attempt)
    {
        var ceiling = Ceiling(attempt);
        if (ceiling <= TimeSpan.Zero)
        {
            return TimeSpan.Zero;
        }

        double fraction = Math.Clamp(_jitterSource(), 0d, 1d);
        return TimeSpan.FromTicks((long)(ceiling.Ticks * fraction));
    }

    private static double DefaultJitter()
    {
        // Jitter spreads retry storms; it is not security-sensitive, so the
        // shared PRNG is the right tool (same call the Go reference makes with
        // its gosec annotation).
#pragma warning disable CA5394 // Do not use insecure randomness
        return Random.Shared.NextDouble();
#pragma warning restore CA5394
    }
}
