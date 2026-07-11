using Orders.Grpc.Core.Identity;

namespace Orders.Grpc.Api.Auth;

/// <summary>
/// Verifies a bearer token and resolves the caller's <see cref="CallerPrincipal"/>.
/// This is the seam a production build implements with a JWT/JWKS validator
/// (see the keystone exampleservice module); this reference ships a
/// static-token and a synthetic local-dev implementation, mirroring the Go
/// reference's Authenticator interface.
/// </summary>
internal interface IAuthenticator
{
    /// <summary>Returns the principal, or null when the token is missing or invalid.</summary>
    ValueTask<CallerPrincipal?> AuthenticateAsync(string? bearerToken, CancellationToken cancellationToken);
}
