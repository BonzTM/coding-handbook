namespace Orders.Api.ErrorHandling;

/// <summary>
/// Stable RFC 9457 `type` URIs - the machine-readable signal a client branches
/// on, independent of the HTTP status. Documented wire contract: add types
/// additively, never repurpose one (csharp/foundations/serialization.md).
/// </summary>
internal static class ProblemTypes
{
    private const string Base = "https://orders.example/problems/";

    public const string OrderNotFound = Base + "order-not-found";
    public const string DuplicateOrder = Base + "duplicate-order";
    public const string VersionConflict = Base + "version-conflict";
    public const string TenantMissing = Base + "tenant-missing";
    public const string IdempotencyKeyMissing = Base + "idempotency-key-missing";
    public const string IdempotencyInFlight = Base + "idempotency-in-flight";
    public const string IdempotencyKeyReuse = Base + "idempotency-key-reuse";
}
