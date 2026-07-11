using Orders.Worker.Core.Messaging;

using Xunit;

namespace Orders.Worker.UnitTests.Messaging;

/// <summary>Exponential ceiling growth, the max cap, and full-jitter bounds -
/// the numbers the retry loop's delays are proven against.</summary>
public sealed class BackoffPolicyTests
{
    private static readonly TimeSpan _base = TimeSpan.FromMilliseconds(100);
    private static readonly TimeSpan _max = TimeSpan.FromSeconds(1);

    [Theory]
    [InlineData(1, 100)]
    [InlineData(2, 200)]
    [InlineData(3, 400)]
    [InlineData(4, 800)]
    [InlineData(5, 1000)] // capped at max
    [InlineData(6, 1000)]
    public void Ceiling_DoublesPerAttemptAndCapsAtMax(int attempt, int expectedMilliseconds)
    {
        var policy = new BackoffPolicy(_base, _max);

        Assert.Equal(TimeSpan.FromMilliseconds(expectedMilliseconds), policy.Ceiling(attempt));
    }

    [Fact]
    public void Ceiling_ClampsAttemptsBelowOne()
    {
        var policy = new BackoffPolicy(_base, _max);

        Assert.Equal(_base, policy.Ceiling(0));
        Assert.Equal(_base, policy.Ceiling(-5));
    }

    [Fact]
    public void NextDelay_WithPinnedJitter_IsDeterministic()
    {
        Assert.Equal(
            TimeSpan.FromMilliseconds(200),
            new BackoffPolicy(_base, _max, () => 1.0).NextDelay(2));
        Assert.Equal(
            TimeSpan.FromMilliseconds(100),
            new BackoffPolicy(_base, _max, () => 0.5).NextDelay(2));
        Assert.Equal(TimeSpan.Zero, new BackoffPolicy(_base, _max, () => 0.0).NextDelay(2));
    }

    [Fact]
    public void NextDelay_OutOfRangeJitterIsClamped()
    {
        Assert.Equal(_base, new BackoffPolicy(_base, _max, () => 42.0).NextDelay(1));
        Assert.Equal(TimeSpan.Zero, new BackoffPolicy(_base, _max, () => -1.0).NextDelay(1));
    }

    [Fact]
    public void NextDelay_DefaultJitter_IsAlwaysBoundedByTheCeiling()
    {
        var policy = new BackoffPolicy(_base, _max);
        for (int attempt = 1; attempt <= 6; attempt++)
        {
            for (int i = 0; i < 100; i++)
            {
                var delay = policy.NextDelay(attempt);
                Assert.InRange(delay, TimeSpan.Zero, policy.Ceiling(attempt));
            }
        }
    }
}
