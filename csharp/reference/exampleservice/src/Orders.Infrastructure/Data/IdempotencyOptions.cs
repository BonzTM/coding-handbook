using System.ComponentModel.DataAnnotations;

namespace Orders.Infrastructure.Data;

/// <summary>
/// Idempotency-key dedupe window. A key replayed after <see cref="Ttl"/> is
/// treated as never-seen and re-executes, so the TTL must exceed any client's
/// realistic retry window; document it where clients can find it
/// (csharp/recipes/add-idempotent-write.md, Step 8).
/// </summary>
public sealed class IdempotencyOptions
{
    public const string SectionName = "Idempotency";

    [Range(typeof(TimeSpan), "00:01:00", "30.00:00:00")]
    public TimeSpan Ttl { get; init; } = TimeSpan.FromHours(24);
}
