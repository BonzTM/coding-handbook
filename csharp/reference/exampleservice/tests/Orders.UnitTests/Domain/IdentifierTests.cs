using Orders.Core.Idempotency;
using Orders.Core.Orders;

using Xunit;

namespace Orders.UnitTests.Domain;

public sealed class IdentifierTests
{
    [Fact]
    public void OrderIdNew_GeneratesVersion7Guids()
    {
        var id = OrderId.New();
        Assert.Equal(7, id.Value.Version);
    }

    [Theory]
    [InlineData("0197b0c0-2f9d-7c7a-8b1e-3f2a4d5e6f70", true)]
    [InlineData("not-a-guid", false)]
    [InlineData("", false)]
    [InlineData(null, false)]
    public void OrderIdTryParse_Candidate_ReportsValidity(string? candidate, bool valid)
    {
        Assert.Equal(valid, OrderId.TryParse(candidate, out _));
    }

    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public void TenantId_BlankValue_Throws(string value)
    {
        Assert.ThrowsAny<ArgumentException>(() => new TenantId(value));
    }

    [Fact]
    public void TenantId_OverMaxLength_Throws()
    {
        string value = new('t', TenantId.MaxLength + 1);
        Assert.Throws<ArgumentOutOfRangeException>(() => new TenantId(value));
    }

    [Fact]
    public void IdempotencyScope_ValidatesRouteAndKeyBounds()
    {
        var tenant = new TenantId("tenant-a");

        Assert.ThrowsAny<ArgumentException>(() => new IdempotencyScope(tenant, "", "key"));
        Assert.ThrowsAny<ArgumentException>(() => new IdempotencyScope(tenant, "POST /orders", ""));
        Assert.Throws<ArgumentOutOfRangeException>(() =>
            new IdempotencyScope(tenant, "POST /orders", new string('k', IdempotencyScope.MaxKeyLength + 1)));
    }
}
