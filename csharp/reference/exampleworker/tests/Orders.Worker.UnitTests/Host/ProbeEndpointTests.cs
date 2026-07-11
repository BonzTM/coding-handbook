using System.Net;

using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.Extensions.DependencyInjection;

using Orders.Worker.Infrastructure.Messaging;

using Xunit;

namespace Orders.Worker.UnitTests.Host;

/// <summary>
/// The probe contract on the real host (WebApplicationFactory boots Program
/// with its BackgroundServices): /livez is always cheap and green; /readyz
/// reflects consumer subscription + broker connectivity and sheds work when
/// either is gone (csharp/operations/observability.md).
/// </summary>
public sealed class ProbeEndpointTests
{
    [Fact]
    public async Task Livez_IsHealthyImmediately()
    {
        using var factory = new WebApplicationFactory<Program>();
        using var client = factory.CreateClient();

        using var response = await client.GetAsync(
            new Uri("/livez", UriKind.Relative), TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
    }

    [Fact]
    public async Task Readyz_BecomesHealthyOnceTheConsumerSubscribes()
    {
        using var factory = new WebApplicationFactory<Program>();
        using var client = factory.CreateClient();

        var status = await PollReadyzAsync(client, until: HttpStatusCode.OK);

        Assert.Equal(HttpStatusCode.OK, status);
    }

    [Fact]
    public async Task Readyz_ShedsWhenTheBrokerReportsUnhealthy()
    {
        using var factory = new WebApplicationFactory<Program>();
        using var client = factory.CreateClient();
        _ = await PollReadyzAsync(client, until: HttpStatusCode.OK);

        factory.Services.GetRequiredService<InMemoryBroker>().SetHealthy(false);
        using var response = await client.GetAsync(
            new Uri("/readyz", UriKind.Relative), TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.ServiceUnavailable, response.StatusCode);
    }

    /// <summary>The consumer flips readiness asynchronously after startup;
    /// poll with a fixed bound instead of asserting a race.</summary>
    private static async Task<HttpStatusCode> PollReadyzAsync(HttpClient client, HttpStatusCode until)
    {
        HttpStatusCode status = HttpStatusCode.ServiceUnavailable;
        for (int i = 0; i < 100; i++)
        {
            using var response = await client.GetAsync(
                new Uri("/readyz", UriKind.Relative), TestContext.Current.CancellationToken);
            status = response.StatusCode;
            if (status == until)
            {
                return status;
            }

            await Task.Delay(TimeSpan.FromMilliseconds(50), TestContext.Current.CancellationToken);
        }

        return status;
    }
}
