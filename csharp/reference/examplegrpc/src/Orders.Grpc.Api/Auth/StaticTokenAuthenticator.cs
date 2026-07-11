using System.Security.Cryptography;
using System.Text;

using Microsoft.Extensions.Options;

using Orders.Grpc.Core.Identity;

namespace Orders.Grpc.Api.Auth;

/// <summary>
/// Accepts exactly the configured bearer token and resolves it to a fixed
/// principal (reader + writer in the configured tenant). Comparison is
/// constant-time so the token cannot be guessed byte by byte. Wired when
/// `Auth:Enabled=true`; production swaps in a JWT/JWKS validator behind the
/// same seam.
/// </summary>
internal sealed class StaticTokenAuthenticator : IAuthenticator
{
    private readonly byte[] _expectedToken;
    private readonly CallerPrincipal _principal;

    public StaticTokenAuthenticator(IOptions<AuthOptions> options)
    {
        ArgumentNullException.ThrowIfNull(options);
        _expectedToken = Encoding.UTF8.GetBytes(options.Value.BearerToken);
        _principal = new CallerPrincipal(
            subject: "static-token-client",
            tenant: new TenantId(options.Value.Tenant),
            roles: [OrderRoles.Reader, OrderRoles.Writer]);
    }

    public ValueTask<CallerPrincipal?> AuthenticateAsync(string? bearerToken, CancellationToken cancellationToken)
    {
        if (string.IsNullOrEmpty(bearerToken))
        {
            return ValueTask.FromResult<CallerPrincipal?>(null);
        }

        byte[] presented = Encoding.UTF8.GetBytes(bearerToken);
        bool valid = CryptographicOperations.FixedTimeEquals(presented, _expectedToken);
        return ValueTask.FromResult(valid ? _principal : null);
    }
}
