using System.Net;

using Xunit;

namespace Orders.UnitTests.Transport;

/// <summary>
/// Rate limiting rejects with 429 + Retry-After - never the framework's 503
/// default (csharp/services/http-services.md, Rate Limiting). Uses a dedicated
/// factory with a tiny token bucket so the test exhausts it deterministically.
/// </summary>
public sealed class RateLimitingTests : IClassFixture<RateLimitingTests.TightLimitFactory>
{
    public sealed class TightLimitFactory : OrdersApiFactory
    {
        protected override Dictionary<string, string?> Settings { get; } = new(StringComparer.Ordinal)
        {
            ["RateLimiting:TokenLimit"] = "2",
            ["RateLimiting:TokensPerPeriod"] = "1",
            ["RateLimiting:ReplenishmentPeriod"] = "00:10:00",
        };
    }

    private readonly TightLimitFactory _factory;

    public RateLimitingTests(TightLimitFactory factory)
    {
        _factory = factory;
    }

    [Fact]
    public async Task ExhaustedBucket_Returns429WithRetryAfter()
    {
        using HttpClient client = _factory.CreateClient();
        Uri list = new("/orders", UriKind.Relative);

        HttpResponseMessage first = await client.GetAsync(list, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.OK, first.StatusCode);
        HttpResponseMessage second = await client.GetAsync(list, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.OK, second.StatusCode);

        HttpResponseMessage third = await client.GetAsync(list, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.TooManyRequests, third.StatusCode);
        string retryAfter = Assert.Single(third.Headers.GetValues("Retry-After"));
        Assert.True(int.Parse(retryAfter, System.Globalization.CultureInfo.InvariantCulture) >= 1);
    }
}
