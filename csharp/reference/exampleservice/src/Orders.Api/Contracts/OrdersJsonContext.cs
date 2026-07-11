using System.Text.Json.Serialization;

namespace Orders.Api.Contracts;

/// <summary>
/// The source-generated serialization context for this API's wire surface.
/// Adding a type to the wire is a visible diff here; naming policy and
/// enum-as-string are declared once, never per call site
/// (csharp/foundations/serialization.md).
/// </summary>
[JsonSourceGenerationOptions(
    PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase,
    UseStringEnumConverter = true)]
[JsonSerializable(typeof(CreateOrderRequest))]
[JsonSerializable(typeof(UpdateOrderRequest))]
[JsonSerializable(typeof(OrderResponse))]
[JsonSerializable(typeof(OrderListResponse))]
internal sealed partial class OrdersJsonContext : JsonSerializerContext;
