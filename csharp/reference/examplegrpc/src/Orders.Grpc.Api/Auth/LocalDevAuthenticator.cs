using Orders.Grpc.Core.Identity;

namespace Orders.Grpc.Api.Auth;

/// <summary>
/// Local/dev authentication: every call resolves to a synthetic principal
/// (tenant `local-dev`, reader + writer) so the service boots offline with no
/// identity provider. Wired ONLY when `Auth:Enabled=false`; the interceptor
/// chain and the domain's role checks still run against it, so the pipeline
/// shape is identical to production.
/// </summary>
internal sealed class LocalDevAuthenticator : IAuthenticator
{
    private static readonly CallerPrincipal _localDevPrincipal = new(
        subject: "local-dev",
        tenant: new TenantId("local-dev"),
        roles: [OrderRoles.Reader, OrderRoles.Writer]);

    public ValueTask<CallerPrincipal?> AuthenticateAsync(string? bearerToken, CancellationToken cancellationToken)
        => ValueTask.FromResult<CallerPrincipal?>(_localDevPrincipal);
}
