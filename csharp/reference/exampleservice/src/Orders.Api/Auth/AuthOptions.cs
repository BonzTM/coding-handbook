using Microsoft.Extensions.Options;

namespace Orders.Api.Auth;

/// <summary>
/// Authentication configuration. The SECURE default is enabled: an absent
/// `Auth` section fails startup because Authority/Audience are then required.
/// The committed appsettings.json opts local development out explicitly
/// (`Auth:Enabled=false`) - production configuration must set
/// `Auth__Enabled=true` plus Authority and Audience.
/// </summary>
internal sealed class AuthOptions
{
    public const string SectionName = "Auth";

    /// <summary>When false, the local-dev scheme authenticates a synthetic principal.</summary>
    public bool Enabled { get; init; } = true;

    /// <summary>OIDC issuer; discovery + JWKS live under it. Also the pinned `iss`.</summary>
    public string Authority { get; init; } = string.Empty;

    /// <summary>The pinned `aud`; a token minted for another API must not work here.</summary>
    public string Audience { get; init; } = string.Empty;
}

/// <summary>
/// Conditional requirements DataAnnotations cannot express: when auth is
/// enabled, Authority and Audience are mandatory. Runs at startup via
/// ValidateOnStart (csharp/foundations/configuration.md).
/// </summary>
internal sealed class AuthOptionsValidator : IValidateOptions<AuthOptions>
{
    public ValidateOptionsResult Validate(string? name, AuthOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);
        if (!options.Enabled)
        {
            return ValidateOptionsResult.Success;
        }

        var failures = new List<string>(capacity: 2);
        if (string.IsNullOrWhiteSpace(options.Authority))
        {
            failures.Add("Auth:Authority is required when Auth:Enabled is true.");
        }

        if (string.IsNullOrWhiteSpace(options.Audience))
        {
            failures.Add("Auth:Audience is required when Auth:Enabled is true.");
        }

        return failures.Count == 0
            ? ValidateOptionsResult.Success
            : ValidateOptionsResult.Fail(failures);
    }
}
