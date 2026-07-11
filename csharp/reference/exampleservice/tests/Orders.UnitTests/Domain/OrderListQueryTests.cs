using Orders.Core.Orders;

using Xunit;

namespace Orders.UnitTests.Domain;

public sealed class OrderListQueryTests
{
    [Theory]
    [InlineData(null, OrderListQuery.DefaultPageSize)] // absent -> default
    [InlineData(0, OrderListQuery.DefaultPageSize)] // nonsense -> default
    [InlineData(-5, OrderListQuery.DefaultPageSize)]
    [InlineData(1, 1)]
    [InlineData(100, 100)]
    [InlineData(101, OrderListQuery.MaxPageSize)] // clamped, never honored raw
    [InlineData(int.MaxValue, OrderListQuery.MaxPageSize)]
    public void Constructor_ClampsPageSizeServerSide(int? requested, int expected)
    {
        Assert.Equal(expected, new OrderListQuery(requested, null).PageSize);
    }
}
