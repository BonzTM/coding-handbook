using Orders.Grpc.Core.Identity;

namespace Orders.Grpc.Core.Orders;

/// <summary>
/// The order aggregate. Constructed through <see cref="Create"/> so invariants
/// hold from birth. Unlike the keystone HTTP module - where DataAnnotations on
/// wire DTOs catch bad input before Core - proto messages carry no validation,
/// so Create validates here and reports EVERY violation structurally for the
/// transport's google.rpc.BadRequest detail. Wire DTOs are the generated proto
/// messages in Orders.Grpc.Api; this type never serializes directly
/// (csharp/foundations/serialization.md).
/// </summary>
public sealed class Order
{
    public const int MaxReferenceLength = 64;
    public const int MaxCustomerIdLength = 64;
    public const int MaxQuantity = 1_000;

    private Order(
        OrderId id,
        TenantId tenantId,
        string externalReference,
        string customerId,
        int quantity,
        DateTimeOffset createdAt)
    {
        Id = id;
        TenantId = tenantId;
        ExternalReference = externalReference;
        CustomerId = customerId;
        Quantity = quantity;
        CreatedAt = createdAt;
    }

    public OrderId Id { get; }

    public TenantId TenantId { get; }

    /// <summary>Client-supplied order reference, unique per tenant (enforced by the store).</summary>
    public string ExternalReference { get; }

    public string CustomerId { get; }

    public int Quantity { get; }

    public DateTimeOffset CreatedAt { get; }

    /// <exception cref="OrderValidationException">One or more inputs are invalid; all violations are reported.</exception>
    public static Order Create(
        TenantId tenantId,
        string externalReference,
        string customerId,
        int quantity,
        DateTimeOffset now)
    {
        var violations = Validate(externalReference, customerId, quantity);
        if (violations.Count > 0)
        {
            throw new OrderValidationException(violations);
        }

        return new Order(OrderId.New(), tenantId, externalReference, customerId, quantity, now);
    }

    /// <summary>Collects every violation instead of failing on the first, so a client can fix its request in one round trip.</summary>
    private static List<FieldViolation> Validate(string externalReference, string customerId, int quantity)
    {
        var violations = new List<FieldViolation>(capacity: 3);
        if (string.IsNullOrWhiteSpace(externalReference))
        {
            violations.Add(new FieldViolation("external_reference", "external_reference must not be empty"));
        }
        else if (externalReference.Length > MaxReferenceLength)
        {
            violations.Add(new FieldViolation(
                "external_reference", $"external_reference must be at most {MaxReferenceLength} characters"));
        }

        if (string.IsNullOrWhiteSpace(customerId))
        {
            violations.Add(new FieldViolation("customer_id", "customer_id must not be empty"));
        }
        else if (customerId.Length > MaxCustomerIdLength)
        {
            violations.Add(new FieldViolation(
                "customer_id", $"customer_id must be at most {MaxCustomerIdLength} characters"));
        }

        if (quantity is < 1 or > MaxQuantity)
        {
            violations.Add(new FieldViolation("quantity", $"quantity must be between 1 and {MaxQuantity}"));
        }

        return violations;
    }
}
