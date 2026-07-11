using Microsoft.Extensions.Options;

namespace Orders.Grpc.Api.Auth;

/// <summary>
/// Authentication configuration. The SECURE default is enabled: an absent
/// `Auth` section fails startup because the bearer token is then required.
/// The committed appsettings.json opts local development out explicitly
/// (`Auth:Enabled=false`), which wires the synthetic local-dev principal so
/// the service boots offline - the same gating pattern as the keystone
/// module's JWT authentication (csharp/operations/security.md).
/// </summary>
internal sealed class AuthOptions
{
    public const string SectionName = "Auth";

    /// <summary>When false, every call runs as the synthetic local-dev principal.</summary>
    public bool Enabled { get; init; } = true;

    /// <summary>
    /// The single accepted bearer token - the dependency-free reference
    /// authenticator, mirroring the Go reference. Production swaps a JWT/JWKS
    /// validator in behind the same <see cref="IAuthenticator"/> seam (the
    /// keystone exampleservice module shows that wiring end to end).
    /// Supply it via environment (Auth__BearerToken), never in this repo.
    /// </summary>
    public string BearerToken { get; init; } = string.Empty;

    /// <summary>Tenant the static-token principal is scoped to.</summary>
    public string Tenant { get; init; } = "tenant-a";
}

/// <summary>
/// Conditional requirements DataAnnotations cannot express: when auth is
/// enabled, a non-trivial bearer token and a tenant are mandatory. Runs at
/// startup via ValidateOnStart (csharp/foundations/configuration.md).
/// </summary>
internal sealed class AuthOptionsValidator : IValidateOptions<AuthOptions>
{
    private const int MinTokenLength = 16;

    public ValidateOptionsResult Validate(string? name, AuthOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);
        if (!options.Enabled)
        {
            return ValidateOptionsResult.Success;
        }

        var failures = new List<string>(capacity: 2);
        if (string.IsNullOrWhiteSpace(options.BearerToken) || options.BearerToken.Length < MinTokenLength)
        {
            failures.Add($"Auth:BearerToken must be at least {MinTokenLength} characters when Auth:Enabled is true.");
        }

        if (string.IsNullOrWhiteSpace(options.Tenant))
        {
            failures.Add("Auth:Tenant is required when Auth:Enabled is true.");
        }

        return failures.Count == 0
            ? ValidateOptionsResult.Success
            : ValidateOptionsResult.Fail(failures);
    }
}
