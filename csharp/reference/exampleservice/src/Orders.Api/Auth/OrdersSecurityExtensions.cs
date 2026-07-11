using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.Authorization;
using Microsoft.Extensions.Options;

using Orders.Api.Telemetry;

namespace Orders.Api.Auth;

/// <summary>
/// Authentication + authorization wiring (csharp/operations/security.md).
///
/// AuthN: JWT bearer validated against the issuer's JWKS (fetched and refreshed
/// via OIDC discovery - key material is never embedded). Issuer and audience
/// are pinned; signature algorithms are allowlisted so `alg=none` and
/// HS/RS-confusion tokens are rejected. Config-gated: `Auth:Enabled=false`
/// swaps in the local-dev synthetic principal for offline development.
///
/// AuthZ: deny by default - a fallback policy requires an authenticated user on
/// every endpoint; anonymous access is an explicit, reviewable opt-in
/// (the health probes). Capability policies map to roles in one place.
/// </summary>
internal static class OrdersSecurityExtensions
{
    public static WebApplicationBuilder AddOrdersSecurity(this WebApplicationBuilder builder)
    {
        var auth = builder.Configuration.GetSection(AuthOptions.SectionName).Get<AuthOptions>()
            ?? new AuthOptions();

        builder.Services.AddOptions<AuthOptions>()
            .BindConfiguration(AuthOptions.SectionName)
            .ValidateOnStart();
        builder.Services.AddSingleton<IValidateOptions<AuthOptions>, AuthOptionsValidator>();

        AddAuthentication(builder, auth);
        AddAuthorization(builder.Services);
        return builder;
    }

    private static void AddAuthentication(WebApplicationBuilder builder, AuthOptions auth)
    {
        if (!auth.Enabled)
        {
            builder.Services
                .AddAuthentication(LocalDevAuthenticationHandler.SchemeName)
                .AddScheme<Microsoft.AspNetCore.Authentication.AuthenticationSchemeOptions, LocalDevAuthenticationHandler>(
                    LocalDevAuthenticationHandler.SchemeName, configureOptions: null);
            return;
        }

        builder.Services
            .AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
            .AddJwtBearer(options =>
            {
                options.Authority = auth.Authority;   // OIDC discovery + JWKS live here
                options.MapInboundClaims = false;      // keep raw JWT claim names
                options.TokenValidationParameters.ValidIssuer = auth.Authority;
                options.TokenValidationParameters.ValidAudience = auth.Audience;
                options.TokenValidationParameters.NameClaimType = OrdersClaims.Subject;
                options.TokenValidationParameters.RoleClaimType = OrdersClaims.Roles;
                options.TokenValidationParameters.RequireExpirationTime = true;
                // Allowlist, never a denylist: asymmetric algorithms only.
                options.TokenValidationParameters.ValidAlgorithms =
                    ["RS256", "RS384", "RS512", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512"];
                options.Events = new JwtBearerEvents
                {
                    // Audit every authentication failure. The 401 body stays
                    // uniform (never says WHICH check failed); the audit record
                    // carries the reason for operators.
                    OnChallenge = context =>
                    {
                        var audit = context.HttpContext.RequestServices.GetRequiredService<AuditLogger>();
                        audit.AuthenticationFailed(
                            reason: context.Error ?? "missing_token",
                            requestId: context.HttpContext.TraceIdentifier);
                        return Task.CompletedTask;
                    },
                };
            });
    }

    private static void AddAuthorization(IServiceCollection services)
    {
        services.AddAuthorizationBuilder()
            .SetFallbackPolicy(new AuthorizationPolicyBuilder().RequireAuthenticatedUser().Build())
            .AddPolicy(OrdersPolicies.Read, policy => policy.RequireRole(OrdersClaims.ReaderRole))
            .AddPolicy(OrdersPolicies.Write, policy => policy.RequireRole(OrdersClaims.WriterRole));

        // Wraps the default result handler to audit authorization denials.
        services.AddSingleton<IAuthorizationMiddlewareResultHandler, AuditingAuthorizationResultHandler>();
    }
}

/// <summary>Capability-named policies (csharp/operations/security.md).</summary>
internal static class OrdersPolicies
{
    public const string Read = "orders:read";
    public const string Write = "orders:write";
}
