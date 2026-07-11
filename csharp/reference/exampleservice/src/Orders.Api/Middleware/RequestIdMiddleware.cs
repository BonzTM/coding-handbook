namespace Orders.Api.Middleware;

/// <summary>
/// Outermost middleware: adopts a well-formed inbound X-Request-Id (so a
/// caller's correlation id survives the hop) or keeps the framework-generated
/// TraceIdentifier, and echoes the id on the response. It runs BEFORE the
/// exception handler so even an unhandled-exception 500 carries requestId in
/// its ProblemDetails body (csharp/services/http-services.md).
/// </summary>
internal sealed class RequestIdMiddleware(RequestDelegate next)
{
    public const string HeaderName = "X-Request-Id";

    private const int MaxLength = 64;

    public Task InvokeAsync(HttpContext context)
    {
        ArgumentNullException.ThrowIfNull(context);

        string inbound = context.Request.Headers[HeaderName].ToString();
        if (IsWellFormed(inbound))
        {
            context.TraceIdentifier = inbound;
        }

        // OnStarting, not an eager header write: the exception handler CLEARS
        // the response (headers included) before rewriting it as a 500, and the
        // request id must survive that path too.
        context.Response.OnStarting(static state =>
        {
            var httpContext = (HttpContext)state;
            httpContext.Response.Headers[HeaderName] = httpContext.TraceIdentifier;
            return Task.CompletedTask;
        }, context);
        return next(context);
    }

    /// <summary>Bounded length and charset - an untrusted header never lands in logs raw.</summary>
    internal static bool IsWellFormed(string candidate)
    {
        if (candidate.Length is 0 or > MaxLength)
        {
            return false;
        }

        foreach (char c in candidate)
        {
            bool ok = char.IsAsciiLetterOrDigit(c) || c is '-' or '_' or '.' or ':' or '+' or '/';
            if (!ok)
            {
                return false;
            }
        }

        return true;
    }
}
