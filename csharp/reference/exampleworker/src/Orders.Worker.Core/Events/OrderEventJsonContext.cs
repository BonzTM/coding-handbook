using System.Text.Json.Serialization;

namespace Orders.Worker.Core.Events;

/// <summary>
/// Source-generated serializer context for the event payload - the trust
/// boundary serializes through a closed, reviewable set of types, never
/// reflection (csharp/foundations/serialization.md). Unknown fields are
/// skipped on read by System.Text.Json's default, which is this contract's
/// explicit policy: consumers tolerate additive producer changes
/// (csharp/services/eventing-and-messaging.md, Schema Evolution).
/// </summary>
[JsonSourceGenerationOptions(PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase)]
[JsonSerializable(typeof(OrderEvent))]
public sealed partial class OrderEventJsonContext : JsonSerializerContext;
