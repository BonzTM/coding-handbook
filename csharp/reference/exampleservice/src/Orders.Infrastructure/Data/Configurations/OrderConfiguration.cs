using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

using Orders.Core.Orders;

namespace Orders.Infrastructure.Data.Configurations;

/// <summary>
/// Mapping for the orders aggregate: explicit lengths on every string column,
/// enum stored as text, xmin as the concurrency token, a unique constraint per
/// tenant on the external reference, and the composite keyset index that backs
/// cursor pagination (csharp/services/database.md).
/// </summary>
internal sealed class OrderConfiguration : IEntityTypeConfiguration<Order>
{
    public const string TenantExternalReferenceIndex = "ix_orders_tenant_external_reference";

    public void Configure(EntityTypeBuilder<Order> builder)
    {
        builder.ToTable("orders");

        builder.HasKey(o => o.Id);
        builder.Property(o => o.Id).ValueGeneratedNever();

        builder.Property(o => o.TenantId)
            .HasMaxLength(TenantId.MaxLength)
            .IsRequired();

        builder.Property(o => o.ExternalReference)
            .HasMaxLength(Order.MaxReferenceLength)
            .IsRequired();

        builder.Property(o => o.CustomerId)
            .HasMaxLength(Order.MaxCustomerIdLength)
            .IsRequired();

        // Enum as text: "shipped" in a psql session is debuggable, 3 is not
        // (csharp/foundations/serialization.md applies the same rule on the wire).
        builder.Property(o => o.Status)
            .HasConversion<string>()
            .HasMaxLength(32)
            .IsRequired();

        // PostgreSQL xmin: zero-schema-cost concurrency token, bumped by every
        // write (csharp/services/database.md, Concurrency Tokens).
        builder.Property(o => o.Version).IsRowVersion();

        // Uniqueness the database enforces; the repository translates SQLSTATE
        // 23505 on this constraint to DuplicateOrderException.
        builder.HasIndex(o => new { o.TenantId, o.ExternalReference })
            .IsUnique()
            .HasDatabaseName(TenantExternalReferenceIndex);

        // Keyset-pagination index over the stable (tenant, created_at, id) order.
        builder.HasIndex(o => new { o.TenantId, o.CreatedAt, o.Id })
            .HasDatabaseName("ix_orders_tenant_created_id");
    }
}
