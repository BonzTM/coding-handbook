namespace Orders.Grpc.Core.Identity;

/// <summary>
/// The authenticated caller. Resolved once by the transport's auth interceptor
/// and passed explicitly into every domain call - Core never reads ambient
/// state (csharp/foundations/shared-constructs.md). The tenant always comes
/// from here, never from a request field, so a caller cannot address another
/// tenant's data (csharp/operations/security.md).
/// </summary>
public sealed record CallerPrincipal
{
    public CallerPrincipal(string subject, TenantId tenant, IEnumerable<string> roles)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(subject);
        ArgumentNullException.ThrowIfNull(roles);
        Subject = subject;
        Tenant = tenant;
        Roles = roles.ToHashSet(StringComparer.Ordinal);
    }

    public string Subject { get; }

    public TenantId Tenant { get; }

    public IReadOnlySet<string> Roles { get; }

    public bool HasRole(string role) => Roles.Contains(role);
}
