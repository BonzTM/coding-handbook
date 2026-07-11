using Orders.Core.Idempotency;

namespace Orders.UnitTests.Fakes;

/// <summary>
/// Hand-rolled in-memory <see cref="IIdempotencyRunner"/> honoring the port's
/// contract (execute-and-record, replay, in-flight, fingerprint mismatch) so
/// transport tests can exercise the IdempotencyFilter's status mapping without
/// a database. Transactional atomicity and TTL takeover are properties of the
/// real PostgreSQL runner and are proven by the integration suite.
/// </summary>
public sealed class InMemoryIdempotencyRunner : IIdempotencyRunner
{
    private sealed record Entry(string Fingerprint, StoredResponse? Response, bool InFlight);

    private readonly Lock _gate = new();
    private readonly Dictionary<(string Tenant, string Route, string Key), Entry> _entries = [];

    public async Task<IdempotencyResult> RunAsync(
        IdempotencyScope scope,
        string requestFingerprint,
        Func<CancellationToken, Task<StoredResponse?>> operation,
        CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(scope);
        ArgumentException.ThrowIfNullOrEmpty(requestFingerprint);
        ArgumentNullException.ThrowIfNull(operation);

        (string, string, string) key = (scope.TenantId.Value, scope.Route, scope.Key);
        lock (_gate)
        {
            if (_entries.TryGetValue(key, out Entry? existing))
            {
                if (existing.InFlight)
                {
                    return new IdempotencyResult.InFlight();
                }

                if (!string.Equals(existing.Fingerprint, requestFingerprint, StringComparison.Ordinal))
                {
                    return new IdempotencyResult.FingerprintMismatch();
                }

                // A completed entry always carries its response by construction.
                return new IdempotencyResult.Replayed(existing.Response!);
            }

            _entries[key] = new Entry(requestFingerprint, Response: null, InFlight: true);
        }

        StoredResponse? response;
        try
        {
            response = await operation(cancellationToken).ConfigureAwait(false);
        }
        catch
        {
            Release(key);
            throw;
        }

        lock (_gate)
        {
            if (response is null)
            {
                // Error outcome: release the claim so a retry re-executes.
                _entries.Remove(key);
                return new IdempotencyResult.Executed(null);
            }

            _entries[key] = new Entry(requestFingerprint, response, InFlight: false);
            return new IdempotencyResult.Executed(response);
        }
    }

    /// <summary>Test seam: simulate a first request that has not completed yet.</summary>
    public void MarkInFlight(IdempotencyScope scope, string fingerprint)
    {
        ArgumentNullException.ThrowIfNull(scope);
        lock (_gate)
        {
            _entries[(scope.TenantId.Value, scope.Route, scope.Key)] =
                new Entry(fingerprint, Response: null, InFlight: true);
        }
    }

    private void Release((string, string, string) key)
    {
        lock (_gate)
        {
            _entries.Remove(key);
        }
    }
}
