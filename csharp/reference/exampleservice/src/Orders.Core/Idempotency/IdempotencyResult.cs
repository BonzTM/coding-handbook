namespace Orders.Core.Idempotency;

/// <summary>
/// Outcome of running an operation under an idempotency key. Closed set - the
/// transport maps each case to its documented status
/// (csharp/recipes/add-idempotent-write.md).
/// </summary>
public abstract record IdempotencyResult
{
    private IdempotencyResult()
    {
    }

    /// <summary>
    /// The operation ran. <see cref="Response"/> is the stored response when the
    /// operation succeeded and was recorded; null when the operation declined to
    /// be recorded (an error outcome), in which case the claim was released so a
    /// retry re-executes.
    /// </summary>
    public sealed record Executed(StoredResponse? Response) : IdempotencyResult;

    /// <summary>A completed record with a matching fingerprint - replay it byte-identically.</summary>
    public sealed record Replayed(StoredResponse Response) : IdempotencyResult;

    /// <summary>The first request with this key has not finished - the caller retries later (409).</summary>
    public sealed record InFlight : IdempotencyResult;

    /// <summary>The key was reused for a different request - a client bug (422).</summary>
    public sealed record FingerprintMismatch : IdempotencyResult;
}
