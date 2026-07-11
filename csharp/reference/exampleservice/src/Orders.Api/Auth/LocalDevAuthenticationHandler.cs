using System.Security.Claims;
using System.Text.Encodings.Web;

using Microsoft.AspNetCore.Authentication;
using Microsoft.Extensions.Options;

namespace Orders.Api.Auth;

/// <summary>
/// Local/dev authentication: every request is a synthetic principal (tenant
/// `local-dev`, reader+writer roles) so the service boots offline with no
/// identity provider. Wired ONLY when `Auth:Enabled=false`; the deny-by-default
/// authorization policies still run against it, so the pipeline shape is
/// identical to production.
/// </summary>
internal sealed class LocalDevAuthenticationHandler(
    IOptionsMonitor<AuthenticationSchemeOptions> options,
    ILoggerFactory logger,
    UrlEncoder encoder) : AuthenticationHandler<AuthenticationSchemeOptions>(options, logger, encoder)
{
    public const string SchemeName = "LocalDev";

    protected override Task<AuthenticateResult> HandleAuthenticateAsync()
    {
        var identity = new ClaimsIdentity(
            [
                new Claim(OrdersClaims.Subject, "local-dev"),
                new Claim(OrdersClaims.Tenant, "local-dev"),
                new Claim(OrdersClaims.Roles, OrdersClaims.ReaderRole),
                new Claim(OrdersClaims.Roles, OrdersClaims.WriterRole),
            ],
            authenticationType: SchemeName,
            nameType: OrdersClaims.Subject,
            roleType: OrdersClaims.Roles);
        var ticket = new AuthenticationTicket(new ClaimsPrincipal(identity), SchemeName);
        return Task.FromResult(AuthenticateResult.Success(ticket));
    }
}
