using System.Security.Claims;

using Orders.Core.Orders;

namespace Orders.Api.Auth;

/// <summary>
/// The claim names and roles this service understands. Raw JWT claim names -
/// MapInboundClaims is off, so nothing is renamed to legacy SOAP-era URIs.
/// </summary>
internal static class OrdersClaims
{
    public const string Subject = "sub";
    public const string Tenant = "tenant";
    public const string Roles = "roles";

    public const string ReaderRole = "orders.reader";
    public const string WriterRole = "orders.writer";

    /// <summary>
    /// Tenant comes from the authenticated principal, never a request body
    /// (csharp/operations/security.md). Absent or malformed → no tenant.
    /// </summary>
    public static bool TryGetTenant(this ClaimsPrincipal principal, out TenantId tenantId)
    {
        tenantId = default;
        string? value = principal.FindFirstValue(Tenant);
        if (string.IsNullOrWhiteSpace(value) || value.Length > TenantId.MaxLength)
        {
            return false;
        }

        tenantId = new TenantId(value);
        return true;
    }

    public static TenantId RequiredTenant(this ClaimsPrincipal principal)
        => principal.TryGetTenant(out var tenant)
            ? tenant
            : throw new InvalidOperationException(
                "No tenant claim on an authorized request; TenantScopeFilter must run first.");

    public static string Actor(this ClaimsPrincipal principal)
        => principal.FindFirstValue(Subject) ?? "anonymous";
}
