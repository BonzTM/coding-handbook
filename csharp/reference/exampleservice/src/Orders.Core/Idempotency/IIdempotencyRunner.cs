namespace Orders.Core.Idempotency;

/// <summary>
/// Port for idempotent write execution, implemented in Orders.Infrastructure.
///
/// Contract (csharp/recipes/add-idempotent-write.md):
/// - First use claims the scope, runs <c>operation</c>, and commits the domain
///   write and the completed idempotency record ATOMICALLY (one transaction).
/// - A duplicate of a completed key with the same fingerprint replays the
///   stored response without re-running the operation.
/// - A duplicate while the first request is still open reports
///   <see cref="IdempotencyResult.InFlight"/>; exactly one execution wins.
/// - The same key with a different fingerprint reports
///   <see cref="IdempotencyResult.FingerprintMismatch"/>.
/// - Records expire after the configured TTL; an expired key is treated as
///   never-seen.
/// - When <c>operation</c> returns null (an outcome that must not be recorded,
///   e.g. a domain error) or throws, the claim is released and nothing commits.
/// </summary>
public interface IIdempotencyRunner
{
    Task<IdempotencyResult> RunAsync(
        IdempotencyScope scope,
        string requestFingerprint,
        Func<CancellationToken, Task<StoredResponse?>> operation,
        CancellationToken cancellationToken);
}
