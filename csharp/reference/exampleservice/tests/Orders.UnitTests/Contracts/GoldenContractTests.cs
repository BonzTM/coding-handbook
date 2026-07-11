using System.Runtime.CompilerServices;
using System.Text.Json;

using Orders.Api.Contracts;
using Orders.Core.Orders;

using Xunit;

namespace Orders.UnitTests.Contracts;

/// <summary>
/// Golden-file contract tests (csharp/quality/testing.md): the committed JSON
/// in TestData/ IS the wire contract. A failing diff here is a contract
/// change - regenerate deliberately with UPDATE_GOLDEN=1 and review the diff
/// like an API change (csharp/foundations/contracts-and-compatibility.md).
/// </summary>
public sealed class GoldenContractTests
{
    [Fact]
    public void OrderResponse_MatchesGoldenFile()
    {
        OrderResponse response = SampleOrder();

        string actual = JsonSerializer.Serialize(response, OrdersJsonContext.Default.OrderResponse);

        AssertMatchesGolden("order-response.json", actual);
    }

    [Fact]
    public void OrderListResponse_MatchesGoldenFile()
    {
        var response = new OrderListResponse
        {
            Items = [SampleOrder()],
            NextCursor = new OrderCursor(
                new DateTimeOffset(2026, 7, 1, 12, 0, 0, TimeSpan.Zero),
                Guid.Parse("0197b0c0-2f9d-7c7a-8b1e-3f2a4d5e6f70")).Encode(),
        };

        string actual = JsonSerializer.Serialize(response, OrdersJsonContext.Default.OrderListResponse);

        AssertMatchesGolden("order-list-response.json", actual);
    }

    [Fact]
    public void CreateOrderRequest_RoundTripsThroughWireContext()
    {
        const string Wire = """{"externalReference":"ref-1","customerId":"cust-1","quantity":3}""";

        CreateOrderRequest? request = JsonSerializer.Deserialize(
            Wire, OrdersJsonContext.Default.CreateOrderRequest);

        Assert.NotNull(request);
        Assert.Equal("ref-1", request.ExternalReference);
        Assert.Equal(
            Wire, JsonSerializer.Serialize(request, OrdersJsonContext.Default.CreateOrderRequest));
    }

    private static OrderResponse SampleOrder() => new()
    {
        OrderId = "0197b0c0-2f9d-7c7a-8b1e-3f2a4d5e6f70",
        ExternalReference = "ref-0042",
        CustomerId = "cust-7",
        Quantity = 3,
        Status = OrderStatus.Confirmed,
        Version = 41,
        CreatedAt = new DateTimeOffset(2026, 7, 1, 12, 0, 0, TimeSpan.Zero),
        UpdatedAt = new DateTimeOffset(2026, 7, 1, 12, 30, 0, TimeSpan.Zero),
    };

    private static void AssertMatchesGolden(
        string fileName, string actual, [CallerFilePath] string thisFile = "")
    {
        // Resolve TestData/ next to this source file so UPDATE_GOLDEN rewrites
        // the COMMITTED contract, not a bin/ copy.
        string goldenPath = Path.Combine(Path.GetDirectoryName(thisFile)!, "..", "TestData", fileName);
        if (Environment.GetEnvironmentVariable("UPDATE_GOLDEN") == "1")
        {
            File.WriteAllText(goldenPath, actual + '\n'); // LF: goldens are byte-stable across OSes
        }

        Assert.True(File.Exists(goldenPath), $"Golden file missing: {goldenPath} (run with UPDATE_GOLDEN=1)");
        string expected = File.ReadAllText(goldenPath).TrimEnd('\r', '\n');
        Assert.Equal(expected, actual);
    }
}
