using System.Globalization;
using System.Threading.RateLimiting;

namespace Orders.Api.RateLimiting;

/// <summary>
/// Built-in RateLimiter middleware with the two mandatory non-defaults: reject
/// with 429 (the default 503 is wrong - clients cannot tell "slow down" from
/// "I am down") and always send Retry-After so well-behaved clients back off
/// (csharp/services/http-services.md, Rate Limiting). Partitioned per
/// authenticated identity first, client address as fallback - one noisy tenant
/// must not exhaust everyone's budget.
/// </summary>
internal static class OrdersRateLimitingExtensions
{
    public const string PerClientPolicy = "per-client";

    public static WebApplicationBuilder AddOrdersRateLimiting(this WebApplicationBuilder builder)
    {
        var limits = builder.Configuration.GetSection(RateLimitingOptions.SectionName).Get<RateLimitingOptions>()
            ?? new RateLimitingOptions();

        builder.Services.AddOptions<RateLimitingOptions>()
            .BindConfiguration(RateLimitingOptions.SectionName)
            .ValidateDataAnnotations()
            .ValidateOnStart();

        builder.Services.AddRateLimiter(options =>
        {
            options.RejectionStatusCode = StatusCodes.Status429TooManyRequests;
            options.OnRejected = (context, _) =>
            {
                var retryAfter = context.Lease.TryGetMetadata(MetadataName.RetryAfter, out var fromLease)
                    ? fromLease
                    : limits.ReplenishmentPeriod;
                context.HttpContext.Response.Headers.RetryAfter =
                    Math.Max(1, (int)retryAfter.TotalSeconds).ToString(CultureInfo.InvariantCulture);
                return ValueTask.CompletedTask;
            };
            options.AddPolicy(PerClientPolicy, context =>
                RateLimitPartition.GetTokenBucketLimiter(
                    context.User.Identity?.Name
                        ?? context.Connection.RemoteIpAddress?.ToString()
                        ?? "anonymous",
                    _ => new TokenBucketRateLimiterOptions
                    {
                        TokenLimit = limits.TokenLimit,
                        TokensPerPeriod = limits.TokensPerPeriod,
                        ReplenishmentPeriod = limits.ReplenishmentPeriod,
                        QueueLimit = 0,
                    }));
        });

        return builder;
    }
}
