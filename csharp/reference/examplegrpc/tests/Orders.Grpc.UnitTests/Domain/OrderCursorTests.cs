using Orders.Grpc.Core.Orders;

using Xunit;

namespace Orders.Grpc.UnitTests.Domain;

/// <summary>
/// The cursor is opaque but strict: it round-trips exactly, and every
/// malformed shape is rejected (mapped to INVALID_ARGUMENT at the transport),
/// never guessed at.
/// </summary>
public sealed class OrderCursorTests
{
    [Fact]
    public void EncodeThenDecode_RoundTrips()
    {
        var cursor = new OrderCursor(
            new DateTimeOffset(2026, 7, 1, 12, 0, 0, TimeSpan.Zero), Guid.CreateVersion7());

        Assert.True(OrderCursor.TryDecode(cursor.Encode(), out var decoded));
        Assert.Equal(cursor, decoded);
    }

    [Theory]
    [InlineData("")]
    [InlineData("not-base64!!!")]
    [InlineData("aGVsbG8")] // "hello": decodes but has no separator
    [InlineData("MjAyNnwwMTk3")] // two parts but neither parses
    public void TryDecode_MalformedToken_ReturnsFalse(string token)
    {
        Assert.False(OrderCursor.TryDecode(token, out _));
    }

    [Fact]
    public void TryDecode_OverlongToken_ReturnsFalse()
    {
        // Bounded input: a token longer than the cap is rejected before any
        // decoding work.
        Assert.False(OrderCursor.TryDecode(new string('A', 129), out _));
    }
}
