using Orders.Api.ErrorHandling;
using Orders.Api.Telemetry;

namespace Orders.Api.Auth;

/// <summary>
/// Group-level guard: every /orders request must carry a tenant claim. An
/// authenticated principal without one gets a 403 ProblemDetails (audited) -
/// handlers downstream can then read the tenant unconditionally.
/// </summary>
internal sealed class TenantScopeFilter(AuditLogger audit) : IEndpointFilter
{
    public async ValueTask<object?> InvokeAsync(EndpointFilterInvocationContext context, EndpointFilterDelegate next)
    {
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(next);

        var http = context.HttpContext;
        if (!http.User.TryGetTenant(out _))
        {
            audit.AuthorizationDenied(
                actor: http.User.Actor(),
                tenant: "missing",
                action: $"{http.Request.Method} {http.Request.Path}",
                requestId: http.TraceIdentifier);
            return TypedResults.Problem(
                statusCode: StatusCodes.Status403Forbidden,
                type: ProblemTypes.TenantMissing,
                title: "No tenant on the authenticated principal.");
        }

        return await next(context).ConfigureAwait(false);
    }
}
