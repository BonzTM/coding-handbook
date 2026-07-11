using Orders.Core.Orders;

namespace Orders.Core.Idempotency;

/// <summary>
/// The lookup key for an idempotent write: (tenant, route, client key) - never
/// the bare header value, so two tenants reusing the same key string cannot
/// collide or read each other's stored response
/// (csharp/recipes/add-idempotent-write.md, Step 2).
/// </summary>
public sealed record IdempotencyScope
{
    public const int MaxKeyLength = 128;
    public const int MaxRouteLength = 128;

    public IdempotencyScope(TenantId tenantId, string route, string key)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(route);
        ArgumentOutOfRangeException.ThrowIfGreaterThan(route.Length, MaxRouteLength);
        ArgumentException.ThrowIfNullOrWhiteSpace(key);
        ArgumentOutOfRangeException.ThrowIfGreaterThan(key.Length, MaxKeyLength);

        TenantId = tenantId;
        Route = route;
        Key = key;
    }

    public TenantId TenantId { get; }

    public string Route { get; }

    public string Key { get; }
}
