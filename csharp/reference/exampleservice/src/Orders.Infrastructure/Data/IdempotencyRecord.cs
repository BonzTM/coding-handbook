namespace Orders.Infrastructure.Data;

/// <summary>
/// Storage row for one idempotency key: the claim (InFlight) and, once the
/// domain write commits, the captured response (Completed). Uniqueness of
/// (TenantId, Route, Key) is the concurrency gate - two simultaneous first
/// uses race on the database constraint, not on application logic
/// (csharp/recipes/add-idempotent-write.md).
/// </summary>
public sealed class IdempotencyRecord
{
    public Guid Id { get; set; }

    public string TenantId { get; set; } = string.Empty;

    public string Route { get; set; } = string.Empty;

    public string Key { get; set; } = string.Empty;

    /// <summary>SHA-256 (hex) of the canonical request; detects key reuse for a different request.</summary>
    public string Fingerprint { get; set; } = string.Empty;

    public IdempotencyState State { get; set; }

    public int? StatusCode { get; set; }

    public string? ContentType { get; set; }

    // EF Core storage row: bytea maps to byte[] by design, and EF needs the
    // settable property. The wire-facing StoredResponse wraps it in
    // ReadOnlyMemory<byte>; this row type never crosses the Core boundary.
#pragma warning disable CA1819 // Properties should not return arrays
    public byte[]? Body { get; set; }
#pragma warning restore CA1819

    public string? Location { get; set; }

    public DateTimeOffset CreatedAt { get; set; }
}

public enum IdempotencyState
{
    Unknown = 0,
    InFlight = 1,
    Completed = 2,
}
