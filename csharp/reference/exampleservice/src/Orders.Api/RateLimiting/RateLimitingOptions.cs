using System.ComponentModel.DataAnnotations;

namespace Orders.Api.RateLimiting;

/// <summary>
/// Per-client token-bucket limits. Limits live in config, not literals
/// (csharp/services/http-services.md, Rate Limiting).
/// </summary>
internal sealed class RateLimitingOptions
{
    public const string SectionName = "RateLimiting";

    [Range(1, 1_000_000)]
    public int TokenLimit { get; init; } = 100;

    [Range(1, 1_000_000)]
    public int TokensPerPeriod { get; init; } = 100;

    [Range(typeof(TimeSpan), "00:00:00.100", "00:10:00")]
    public TimeSpan ReplenishmentPeriod { get; init; } = TimeSpan.FromSeconds(1);
}
