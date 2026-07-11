using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Options;

using Npgsql;

using Orders.Core.Idempotency;

namespace Orders.Infrastructure.Data;

/// <summary>
/// PostgreSQL-backed <see cref="IIdempotencyRunner"/>.
///
/// Two phases, per csharp/recipes/add-idempotent-write.md:
/// 1. CLAIM - insert the (tenant, route, key) row in InFlight state and commit
///    it. The unique constraint is the concurrency gate: a concurrent duplicate
///    observes SQLSTATE 23505 and is told the key is in flight (409).
/// 2. EXECUTE - inside ONE database transaction on the request's DbContext:
///    run the operation (the domain write joins this transaction, because the
///    repository shares the same scoped context) and flip the claim to
///    Completed with the captured response bytes. Write and record commit
///    atomically - a replay can only exist if the side effect committed, and a
///    committed side effect always has its replay record.
///
/// Failure windows are safe by construction: a crash after CLAIM but before
/// COMMIT leaves an InFlight row that answers 409 until the TTL expires, then
/// is taken over and re-executed - never a double-applied write.
/// </summary>
public sealed class PostgresIdempotencyRunner(
    OrdersDbContext db,
    TimeProvider time,
    IOptions<IdempotencyOptions> options) : IIdempotencyRunner
{
    private readonly TimeSpan _ttl = options.Value.Ttl;

    public async Task<IdempotencyResult> RunAsync(
        IdempotencyScope scope,
        string requestFingerprint,
        Func<CancellationToken, Task<StoredResponse?>> operation,
        CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(scope);
        ArgumentException.ThrowIfNullOrEmpty(requestFingerprint);
        ArgumentNullException.ThrowIfNull(operation);

        var (claim, verdict) = await ClaimAsync(scope, requestFingerprint, cancellationToken)
            .ConfigureAwait(false);
        if (verdict is not null)
        {
            return verdict;
        }

        return await ExecuteClaimedAsync(claim!, operation, cancellationToken).ConfigureAwait(false);
    }

    /// <summary>Insert the claim row, or classify the existing one.</summary>
    private async Task<(IdempotencyRecord? Claim, IdempotencyResult? Verdict)> ClaimAsync(
        IdempotencyScope scope, string fingerprint, CancellationToken cancellationToken)
    {
        var claim = NewClaim(scope, fingerprint, time.GetUtcNow());
        db.IdempotencyRecords.Add(claim);
        try
        {
            await db.SaveChangesAsync(cancellationToken).ConfigureAwait(false);
            return (claim, null);
        }
        catch (DbUpdateException ex) when (ex.InnerException is PostgresException
        {
            SqlState: PostgresErrorCodes.UniqueViolation,
            ConstraintName: Configurations.IdempotencyRecordConfiguration.ScopeIndex,
        })
        {
            db.Entry(claim).State = EntityState.Detached;
            return await ClassifyExistingAsync(scope, fingerprint, cancellationToken).ConfigureAwait(false);
        }
    }

    private async Task<(IdempotencyRecord? Claim, IdempotencyResult? Verdict)> ClassifyExistingAsync(
        IdempotencyScope scope, string fingerprint, CancellationToken cancellationToken)
    {
        var existing = await db.IdempotencyRecords
            .AsTracking()
            .SingleOrDefaultAsync(
                r => r.TenantId == scope.TenantId.Value && r.Route == scope.Route && r.Key == scope.Key,
                cancellationToken)
            .ConfigureAwait(false);
        if (existing is null)
        {
            // Reaped between our failed insert and this read; treat as in flight
            // and let the client retry - the retry will claim cleanly.
            return (null, new IdempotencyResult.InFlight());
        }

        if (IsExpired(existing))
        {
            return await TryTakeOverAsync(existing, fingerprint, cancellationToken).ConfigureAwait(false);
        }

        return existing.State switch
        {
            IdempotencyState.InFlight => (null, new IdempotencyResult.InFlight()),
            IdempotencyState.Completed when !string.Equals(existing.Fingerprint, fingerprint, StringComparison.Ordinal)
                => (null, new IdempotencyResult.FingerprintMismatch()),
            IdempotencyState.Completed => (null, new IdempotencyResult.Replayed(ToStoredResponse(existing))),
            _ => (null, new IdempotencyResult.InFlight()),
        };
    }

    /// <summary>An expired record is treated as never-seen: reset it to a fresh claim.</summary>
    private async Task<(IdempotencyRecord? Claim, IdempotencyResult? Verdict)> TryTakeOverAsync(
        IdempotencyRecord existing, string fingerprint, CancellationToken cancellationToken)
    {
        var previousCreatedAt = existing.CreatedAt;
        int taken = await db.IdempotencyRecords
            .Where(r => r.Id == existing.Id && r.CreatedAt == previousCreatedAt)
            .ExecuteUpdateAsync(
                set => set
                    .SetProperty(r => r.State, IdempotencyState.InFlight)
                    .SetProperty(r => r.Fingerprint, fingerprint)
                    .SetProperty(r => r.CreatedAt, time.GetUtcNow())
                    .SetProperty(r => r.StatusCode, (int?)null)
                    .SetProperty(r => r.ContentType, (string?)null)
                    .SetProperty(r => r.Body, (byte[]?)null)
                    .SetProperty(r => r.Location, (string?)null),
                cancellationToken)
            .ConfigureAwait(false);
        if (taken == 0)
        {
            // A concurrent request won the takeover; ours is now in flight.
            db.Entry(existing).State = EntityState.Detached;
            return (null, new IdempotencyResult.InFlight());
        }

        await db.Entry(existing).ReloadAsync(cancellationToken).ConfigureAwait(false);
        return (existing, null);
    }

    /// <summary>Run the operation and complete the claim in one transaction.</summary>
    private async Task<IdempotencyResult> ExecuteClaimedAsync(
        IdempotencyRecord claim,
        Func<CancellationToken, Task<StoredResponse?>> operation,
        CancellationToken cancellationToken)
    {
        try
        {
            // EnableRetryOnFailure is on, so the user-initiated transaction runs
            // inside the execution strategy - the whole unit retries, never a
            // torn half (csharp/services/database.md, Transaction Rules).
            var strategy = db.Database.CreateExecutionStrategy();
            return await strategy.ExecuteAsync(
                    async ct => await RunInTransactionAsync(claim, operation, ct).ConfigureAwait(false),
                    cancellationToken)
                .ConfigureAwait(false);
        }
        catch
        {
            await ReleaseClaimAsync(claim, CancellationToken.None).ConfigureAwait(false);
            throw;
        }
    }

    private async Task<IdempotencyResult> RunInTransactionAsync(
        IdempotencyRecord claim,
        Func<CancellationToken, Task<StoredResponse?>> operation,
        CancellationToken cancellationToken)
    {
        await using var transaction = await db.Database.BeginTransactionAsync(cancellationToken)
            .ConfigureAwait(false);
        var response = await operation(cancellationToken).ConfigureAwait(false);
        if (response is null)
        {
            // The operation produced an outcome that must not be recorded (a
            // domain error). Roll back and release so a retry re-executes.
            await transaction.RollbackAsync(cancellationToken).ConfigureAwait(false);
            await ReleaseClaimAsync(claim, cancellationToken).ConfigureAwait(false);
            return new IdempotencyResult.Executed(null);
        }

        ApplyCompletion(claim, response);
        await db.SaveChangesAsync(cancellationToken).ConfigureAwait(false);
        await transaction.CommitAsync(cancellationToken).ConfigureAwait(false);
        return new IdempotencyResult.Executed(response);
    }

    private async Task ReleaseClaimAsync(IdempotencyRecord claim, CancellationToken cancellationToken)
    {
        db.Entry(claim).State = EntityState.Detached;
        await db.IdempotencyRecords
            .Where(r => r.Id == claim.Id && r.State == IdempotencyState.InFlight)
            .ExecuteDeleteAsync(cancellationToken)
            .ConfigureAwait(false);
    }

    private bool IsExpired(IdempotencyRecord record)
        => record.CreatedAt + _ttl <= time.GetUtcNow();

    private static IdempotencyRecord NewClaim(
        IdempotencyScope scope, string fingerprint, DateTimeOffset now) => new()
        {
            Id = Guid.CreateVersion7(),
            TenantId = scope.TenantId.Value,
            Route = scope.Route,
            Key = scope.Key,
            Fingerprint = fingerprint,
            State = IdempotencyState.InFlight,
            CreatedAt = now,
        };

    private static void ApplyCompletion(IdempotencyRecord claim, StoredResponse response)
    {
        claim.State = IdempotencyState.Completed;
        claim.StatusCode = response.StatusCode;
        claim.ContentType = response.ContentType;
        claim.Body = response.Body.ToArray();
        claim.Location = response.Location;
    }

    private static StoredResponse ToStoredResponse(IdempotencyRecord record)
    {
        if (record.StatusCode is not { } status || record.ContentType is null || record.Body is null)
        {
            // A Completed record always carries its response; anything else is
            // corrupted state and must fail loudly, not replay garbage.
            throw new InvalidOperationException(
                $"Idempotency record {record.Id} is Completed but has no stored response.");
        }

        return new StoredResponse(status, record.ContentType, record.Body, record.Location);
    }
}
