using Orders.Grpc.Core.Identity;
using Orders.Grpc.Core.Orders;

using Xunit;

namespace Orders.Grpc.UnitTests.Domain;

/// <summary>
/// Aggregate invariants: a valid order is fully initialized from birth, and a
/// bad one reports EVERY violation - the transport turns that list into one
/// google.rpc.BadRequest detail so a client fixes its request in one round
/// trip.
/// </summary>
public sealed class OrderTests
{
    private static readonly TenantId _tenant = new("tenant-a");
    private static readonly DateTimeOffset _now = new(2026, 7, 1, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public void Create_WithValidInput_InitializesEveryField()
    {
        var order = Order.Create(_tenant, "ord-1001", "cust-42", 3, _now);

        Assert.NotEqual(default, order.Id);
        Assert.Equal(_tenant, order.TenantId);
        Assert.Equal("ord-1001", order.ExternalReference);
        Assert.Equal("cust-42", order.CustomerId);
        Assert.Equal(3, order.Quantity);
        Assert.Equal(_now, order.CreatedAt);
    }

    [Fact]
    public void Create_WithEverythingInvalid_ReportsAllViolations()
    {
        var exception = Assert.Throws<OrderValidationException>(
            () => Order.Create(_tenant, "", "", 0, _now));

        Assert.Equal(3, exception.Violations.Count);
        Assert.Contains(exception.Violations, v => v.Field == "external_reference");
        Assert.Contains(exception.Violations, v => v.Field == "customer_id");
        Assert.Contains(exception.Violations, v => v.Field == "quantity");
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-1)]
    [InlineData(Order.MaxQuantity + 1)]
    public void Create_WithOutOfRangeQuantity_ReportsQuantityViolation(int quantity)
    {
        var exception = Assert.Throws<OrderValidationException>(
            () => Order.Create(_tenant, "ord-1", "cust-1", quantity, _now));

        var violation = Assert.Single(exception.Violations);
        Assert.Equal("quantity", violation.Field);
    }

    [Fact]
    public void Create_WithOverlongReference_ReportsViolation()
    {
        string reference = new('x', Order.MaxReferenceLength + 1);

        var exception = Assert.Throws<OrderValidationException>(
            () => Order.Create(_tenant, reference, "cust-1", 1, _now));

        var violation = Assert.Single(exception.Violations);
        Assert.Equal("external_reference", violation.Field);
    }
}
