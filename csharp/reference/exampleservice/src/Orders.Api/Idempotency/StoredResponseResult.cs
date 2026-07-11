using Orders.Core.Idempotency;

namespace Orders.Api.Idempotency;

/// <summary>
/// Writes a captured response verbatim: exact status, content type, Location,
/// and body bytes. Used for BOTH the first execution and replays so the two
/// are byte-identical by construction
/// (csharp/recipes/add-idempotent-write.md, Byte-identical replay).
/// </summary>
internal sealed class StoredResponseResult(StoredResponse stored, bool replayed) : IResult
{
    public const string ReplayHeaderName = "Idempotency-Replayed";

    public async Task ExecuteAsync(HttpContext httpContext)
    {
        ArgumentNullException.ThrowIfNull(httpContext);

        var response = httpContext.Response;
        response.StatusCode = stored.StatusCode;
        response.ContentType = stored.ContentType;
        response.ContentLength = stored.Body.Length;
        if (stored.Location is { } location)
        {
            response.Headers.Location = location;
        }

        if (replayed)
        {
            // Observability marker only; the body must not change.
            response.Headers[ReplayHeaderName] = "true";
        }

        await response.Body.WriteAsync(stored.Body, httpContext.RequestAborted).ConfigureAwait(false);
    }
}
