namespace Orders.Grpc.Core.Identity;

/// <summary>
/// Typed tenant identifier. Every store operation is scoped by tenant; the
/// wrapper keeps a bare string from being passed where a tenant is required.
/// The value always comes from the authenticated principal, never from a
/// request body (csharp/operations/security.md).
/// </summary>
public readonly record struct TenantId
{
    public const int MaxLength = 64;

    public TenantId(string value)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(value);
        ArgumentOutOfRangeException.ThrowIfGreaterThan(value.Length, MaxLength);
        Value = value;
    }

    public string Value { get; }

    public override string ToString() => Value;
}
