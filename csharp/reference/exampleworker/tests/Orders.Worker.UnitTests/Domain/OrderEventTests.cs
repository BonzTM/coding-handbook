using System.Text.Json;

using Orders.Worker.Core.Events;
using Orders.Worker.UnitTests.Fakes;

using Xunit;

namespace Orders.Worker.UnitTests.Domain;

/// <summary>Payload invariants and the wire contract's tolerance rules.</summary>
public sealed class OrderEventTests
{
    [Fact]
    public void Validate_AcceptsWellFormedEvents()
    {
        TestEvents.Placed().Validate();
        TestEvents.Cancelled().Validate();
    }

    [Theory]
    [InlineData("", "tenant-a", "ref-1")]
    [InlineData("order-1", "", "ref-1")]
    [InlineData("order-1", "tenant-a", "")]
    public void Validate_RejectsMissingRequiredFieldsOnPlaced(string orderId, string tenantId, string reference)
    {
        var orderEvent = TestEvents.Placed(orderId, tenantId, reference);

        Assert.Throws<InvalidEventException>(orderEvent.Validate);
    }

    [Fact]
    public void Validate_RejectsUnknownEventTypes()
    {
        var orderEvent = TestEvents.Placed() with { Type = "orders.order-exploded.v1" };

        Assert.Throws<InvalidEventException>(orderEvent.Validate);
    }

    [Fact]
    public void CancelledEvent_NeedsNoExternalReference()
    {
        var orderEvent = TestEvents.Cancelled() with { ExternalReference = null };

        orderEvent.Validate();
    }

    [Fact]
    public void Deserialize_ToleratesUnknownFields()
    {
        // Additive producer changes must not break this consumer
        // (csharp/services/eventing-and-messaging.md, Schema Evolution).
        const string Payload = """
            {"type":"orders.order-placed.v1","orderId":"order-1","tenantId":"tenant-a",
             "externalReference":"ref-1","aBrandNewField":{"nested":true}}
            """;

        var orderEvent = JsonSerializer.Deserialize(Payload, OrderEventJsonContext.Default.OrderEvent);

        Assert.NotNull(orderEvent);
        Assert.Equal("order-1", orderEvent.OrderId);
        orderEvent.Validate();
    }

    [Fact]
    public void Deserialize_FailsOnMissingRequiredMembers()
    {
        const string Payload = """{"type":"orders.order-placed.v1"}""";

        Assert.Throws<JsonException>(() =>
            JsonSerializer.Deserialize(Payload, OrderEventJsonContext.Default.OrderEvent));
    }
}
