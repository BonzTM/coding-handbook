using Orders.Core.Orders;

using Xunit;

namespace Orders.UnitTests.Domain;

public sealed class OrderCursorTests
{
    [Fact]
    public void EncodeThenDecode_RoundTripsExactly()
    {
        var original = new OrderCursor(
            new DateTimeOffset(2026, 7, 1, 12, 30, 15, TimeSpan.Zero).AddTicks(1234567),
            Guid.CreateVersion7());

        string token = original.Encode();
        Assert.True(OrderCursor.TryDecode(token, out OrderCursor decoded));
        Assert.Equal(original, decoded);
    }

    [Fact]
    public void Encode_IsUrlSafe()
    {
        string token = new OrderCursor(DateTimeOffset.UnixEpoch, Guid.CreateVersion7()).Encode();
        Assert.DoesNotContain('+', token);
        Assert.DoesNotContain('/', token);
        Assert.DoesNotContain('=', token);
    }

    [Theory]
    [InlineData(null)]
    [InlineData("")]
    [InlineData("not-base64!!")]
    [InlineData("aGVsbG8")] // valid base64url, wrong payload shape
    [InlineData("MjAyNi0wNy0wMXxub3QtYS1ndWlk")] // "2026-07-01|not-a-guid"
    public void TryDecode_MalformedToken_IsRejected(string? token)
    {
        Assert.False(OrderCursor.TryDecode(token, out _));
    }

    [Fact]
    public void TryDecode_OversizedToken_IsRejected()
    {
        string token = new('A', 129); // bounded input: length cap before decoding
        Assert.False(OrderCursor.TryDecode(token, out _));
    }
}
