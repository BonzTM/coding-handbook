using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

using Orders.Core.Idempotency;
using Orders.Core.Orders;

namespace Orders.Infrastructure.Data.Configurations;

internal sealed class IdempotencyRecordConfiguration : IEntityTypeConfiguration<IdempotencyRecord>
{
    public const string ScopeIndex = "ix_idempotency_tenant_route_key";

    private const int FingerprintLength = 64; // SHA-256 as lowercase hex
    private const int MaxContentTypeLength = 128;
    private const int MaxLocationLength = 256;

    public void Configure(EntityTypeBuilder<IdempotencyRecord> builder)
    {
        builder.ToTable("idempotency_keys");

        builder.HasKey(r => r.Id);
        builder.Property(r => r.Id).ValueGeneratedNever();

        builder.Property(r => r.TenantId).HasMaxLength(TenantId.MaxLength).IsRequired();
        builder.Property(r => r.Route).HasMaxLength(IdempotencyScope.MaxRouteLength).IsRequired();
        builder.Property(r => r.Key).HasMaxLength(IdempotencyScope.MaxKeyLength).IsRequired();
        builder.Property(r => r.Fingerprint).HasMaxLength(FingerprintLength).IsRequired();

        builder.Property(r => r.State)
            .HasConversion<string>()
            .HasMaxLength(16)
            .IsRequired();

        builder.Property(r => r.ContentType).HasMaxLength(MaxContentTypeLength);
        builder.Property(r => r.Location).HasMaxLength(MaxLocationLength);

        // The concurrency gate for duplicate first-uses: exactly one insert per
        // (tenant, route, key) wins; the loser observes SQLSTATE 23505.
        builder.HasIndex(r => new { r.TenantId, r.Route, r.Key })
            .IsUnique()
            .HasDatabaseName(ScopeIndex);
    }
}
