using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Authorization.Policy;

using Orders.Api.Telemetry;

namespace Orders.Api.Auth;

/// <summary>
/// Decorates the framework's authorization result handling so every DENIAL
/// emits an audit event - a denial is as important to record as a success
/// (csharp/operations/security.md, Audit Logging). The response itself stays
/// the framework's: 403, rendered as ProblemDetails by UseStatusCodePages.
/// </summary>
internal sealed class AuditingAuthorizationResultHandler(AuditLogger audit)
    : IAuthorizationMiddlewareResultHandler
{
    private readonly AuthorizationMiddlewareResultHandler _inner = new();

    public Task HandleAsync(
        RequestDelegate next,
        HttpContext context,
        AuthorizationPolicy policy,
        PolicyAuthorizationResult authorizeResult)
    {
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(authorizeResult);

        if (authorizeResult.Forbidden)
        {
            string tenant = context.User.TryGetTenant(out var t) ? t.Value : "unknown";
            audit.AuthorizationDenied(
                actor: context.User.Actor(),
                tenant: tenant,
                action: $"{context.Request.Method} {context.Request.Path}",
                requestId: context.TraceIdentifier);
        }

        return _inner.HandleAsync(next, context, policy, authorizeResult);
    }
}
