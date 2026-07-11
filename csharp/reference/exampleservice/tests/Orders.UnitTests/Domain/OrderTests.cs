using Orders.Core.Orders;

using Xunit;

namespace Orders.UnitTests.Domain;

public sealed class OrderTests
{
    private static readonly TenantId _tenant = new("tenant-a");
    private static readonly DateTimeOffset _now = new(2026, 7, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void Create_ValidInput_SetsInvariantsFromBirth()
    {
        var order = Order.Create(_tenant, "ref-1", "cust-1", 3, _now);

        Assert.NotEqual(Guid.Empty, order.Id.Value);
        Assert.Equal(_tenant, order.TenantId);
        Assert.Equal("ref-1", order.ExternalReference);
        Assert.Equal("cust-1", order.CustomerId);
        Assert.Equal(3, order.Quantity);
        Assert.Equal(OrderStatus.Pending, order.Status);
        Assert.Equal(_now, order.CreatedAt);
        Assert.Equal(_now, order.UpdatedAt);
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void Create_BlankExternalReference_Throws(string reference)
    {
        Assert.ThrowsAny<ArgumentException>(() => Order.Create(_tenant, reference, "cust-1", 1, _now));
    }

    [Fact]
    public void Create_ExternalReferenceOverMaxLength_Throws()
    {
        string reference = new('r', Order.MaxReferenceLength + 1);
        Assert.Throws<ArgumentOutOfRangeException>(() => Order.Create(_tenant, reference, "cust-1", 1, _now));
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-1)]
    [InlineData(Order.MaxQuantity + 1)]
    public void Create_QuantityOutOfRange_Throws(int quantity)
    {
        Assert.Throws<ArgumentOutOfRangeException>(() => Order.Create(_tenant, "ref-1", "cust-1", quantity, _now));
    }

    [Fact]
    public void Amend_ValidInput_UpdatesQuantityStatusAndTimestamp()
    {
        var order = Order.Create(_tenant, "ref-1", "cust-1", 1, _now);
        DateTimeOffset later = _now.AddMinutes(5);

        order.Amend(9, OrderStatus.Confirmed, later);

        Assert.Equal(9, order.Quantity);
        Assert.Equal(OrderStatus.Confirmed, order.Status);
        Assert.Equal(later, order.UpdatedAt);
        Assert.Equal(_now, order.CreatedAt); // creation time never moves
    }

    [Fact]
    public void Amend_UnknownStatus_Throws()
    {
        var order = Order.Create(_tenant, "ref-1", "cust-1", 1, _now);
        Assert.Throws<ArgumentOutOfRangeException>(() => order.Amend(1, OrderStatus.Unknown, _now));
    }
}
