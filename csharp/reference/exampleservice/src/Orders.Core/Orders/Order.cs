namespace Orders.Core.Orders;

/// <summary>
/// The order aggregate. Constructed through <see cref="Create"/> so invariants
/// hold from birth; EF Core materializes instances through the private
/// parameterless constructor. Wire DTOs live in Orders.Api/Contracts - this
/// type never serializes directly (csharp/foundations/serialization.md).
/// </summary>
public sealed class Order
{
    public const int MaxReferenceLength = 64;
    public const int MaxCustomerIdLength = 64;
    public const int MaxQuantity = 1_000;

    private Order()
    {
        // EF Core materialization path only.
        ExternalReference = string.Empty;
        CustomerId = string.Empty;
    }

    public OrderId Id { get; private set; }

    public TenantId TenantId { get; private set; }

    /// <summary>Client-supplied order reference, unique per tenant (enforced by the database).</summary>
    public string ExternalReference { get; private set; }

    public string CustomerId { get; private set; }

    public int Quantity { get; private set; }

    public OrderStatus Status { get; private set; }

    public DateTimeOffset CreatedAt { get; private set; }

    public DateTimeOffset UpdatedAt { get; private set; }

    /// <summary>
    /// Optimistic concurrency token. Maps to PostgreSQL's xmin system column;
    /// the database bumps it on every write (csharp/services/database.md).
    /// </summary>
    public uint Version { get; private set; }

    public static Order Create(
        TenantId tenantId,
        string externalReference,
        string customerId,
        int quantity,
        DateTimeOffset now)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(externalReference);
        ArgumentOutOfRangeException.ThrowIfGreaterThan(externalReference.Length, MaxReferenceLength);
        ArgumentException.ThrowIfNullOrWhiteSpace(customerId);
        ArgumentOutOfRangeException.ThrowIfGreaterThan(customerId.Length, MaxCustomerIdLength);
        ArgumentOutOfRangeException.ThrowIfLessThan(quantity, 1);
        ArgumentOutOfRangeException.ThrowIfGreaterThan(quantity, MaxQuantity);

        return new Order
        {
            Id = OrderId.New(),
            TenantId = tenantId,
            ExternalReference = externalReference,
            CustomerId = customerId,
            Quantity = quantity,
            Status = OrderStatus.Pending,
            CreatedAt = now,
            UpdatedAt = now,
        };
    }

    public void Amend(int quantity, OrderStatus status, DateTimeOffset now)
    {
        ArgumentOutOfRangeException.ThrowIfLessThan(quantity, 1);
        ArgumentOutOfRangeException.ThrowIfGreaterThan(quantity, MaxQuantity);
        ArgumentOutOfRangeException.ThrowIfEqual((int)status, (int)OrderStatus.Unknown);

        Quantity = quantity;
        Status = status;
        UpdatedAt = now;
    }
}
