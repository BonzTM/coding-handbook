using System.Net;
using System.Text;
using System.Text.Json;

using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Mvc.Testing;

using Xunit;

namespace Orders.IntegrationTests;

/// <summary>
/// The full stack - real pipeline, real repository, real transactional
/// idempotency runner - against the container database. This is where the
/// atomic write+record guarantee and /readyz's database check are proven
/// end to end.
/// </summary>
public sealed class ApiEndToEndTests : IAsyncLifetime
{
    private readonly PostgresFixture _postgres;
    private ContainerBackedApiFactory? _factory;

    public ApiEndToEndTests(PostgresFixture postgres)
    {
        _postgres = postgres;
    }

    public ValueTask InitializeAsync()
    {
        _factory = new ContainerBackedApiFactory(_postgres.ConnectionString);
        return ValueTask.CompletedTask;
    }

    public async ValueTask DisposeAsync()
    {
        if (_factory is not null)
        {
            await _factory.DisposeAsync();
        }
    }

    private sealed class ContainerBackedApiFactory(string connectionString)
        : WebApplicationFactory<Program>
    {
        protected override void ConfigureWebHost(IWebHostBuilder builder)
            => builder.UseSetting("ConnectionStrings:Default", connectionString);
    }

    private static HttpRequestMessage PostOrder(string payload, string key)
    {
        HttpRequestMessage request = new(HttpMethod.Post, "/orders")
        {
            Content = new StringContent(payload, Encoding.UTF8, "application/json"),
        };
        request.Headers.Add("Idempotency-Key", key);
        return request;
    }

    [Fact]
    public async Task Readyz_WithReachableDatabase_Returns200()
    {
        using HttpClient client = _factory!.CreateClient();

        HttpResponseMessage response = await client.GetAsync(
            new Uri("/readyz", UriKind.Relative), TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
    }

    [Fact]
    public async Task CreateReadAmend_FullRoundTripThroughRealDatabase()
    {
        using HttpClient client = _factory!.CreateClient();
        string reference = $"e2e-{Guid.NewGuid():N}"[..24];

        using HttpRequestMessage create = PostOrder(
            $$"""{"externalReference":"{{reference}}","customerId":"cust-1","quantity":2}""",
            $"key-{Guid.NewGuid():N}");
        HttpResponseMessage createResponse = await client.SendAsync(
            create, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Created, createResponse.StatusCode);
        using var created = JsonDocument.Parse(
            await createResponse.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));
        string orderId = created.RootElement.GetProperty("orderId").GetString()!;
        uint version = created.RootElement.GetProperty("version").GetUInt32();
        Assert.NotEqual(0u, version); // real xmin, not a default

        HttpResponseMessage getResponse = await client.GetAsync(
            new Uri($"/orders/{orderId}", UriKind.Relative), TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.OK, getResponse.StatusCode);

        using StringContent amendContent = new(
            $$"""{"quantity":9,"status":"Confirmed","version":{{version}}}""",
            Encoding.UTF8,
            "application/json");
        HttpResponseMessage putResponse = await client.PutAsync(
            new Uri($"/orders/{orderId}", UriKind.Relative),
            amendContent,
            TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.OK, putResponse.StatusCode);
        using var amended = JsonDocument.Parse(
            await putResponse.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));
        Assert.NotEqual(version, amended.RootElement.GetProperty("version").GetUInt32());

        // Retrying the amend with the stale token is a 409 from the REAL xmin gate.
        using StringContent staleContent = new(
            $$"""{"quantity":1,"status":"Shipped","version":{{version}}}""",
            Encoding.UTF8,
            "application/json");
        HttpResponseMessage staleResponse = await client.PutAsync(
            new Uri($"/orders/{orderId}", UriKind.Relative),
            staleContent,
            TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Conflict, staleResponse.StatusCode);
    }

    [Fact]
    public async Task DuplicateReference_SurfacesAs409ThroughRealUniqueConstraint()
    {
        using HttpClient client = _factory!.CreateClient();
        string reference = $"e2e-dup-{Guid.NewGuid():N}"[..24];
        string payload = $$"""{"externalReference":"{{reference}}","customerId":"cust-1","quantity":1}""";

        using HttpRequestMessage first = PostOrder(payload, $"key-{Guid.NewGuid():N}");
        HttpResponseMessage firstResponse = await client.SendAsync(
            first, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Created, firstResponse.StatusCode);

        using HttpRequestMessage second = PostOrder(payload, $"key-{Guid.NewGuid():N}");
        HttpResponseMessage secondResponse = await client.SendAsync(
            second, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Conflict, secondResponse.StatusCode);
        using var body = JsonDocument.Parse(
            await secondResponse.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));
        Assert.Equal(
            "https://orders.example/problems/duplicate-order",
            body.RootElement.GetProperty("type").GetString());
    }

    [Fact]
    public async Task IdempotentReplay_ThroughRealTransactionalRunner()
    {
        using HttpClient client = _factory!.CreateClient();
        string reference = $"e2e-idem-{Guid.NewGuid():N}"[..24];
        string key = $"key-{Guid.NewGuid():N}";
        string payload = $$"""{"externalReference":"{{reference}}","customerId":"cust-1","quantity":3}""";

        using HttpRequestMessage first = PostOrder(payload, key);
        HttpResponseMessage firstResponse = await client.SendAsync(
            first, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Created, firstResponse.StatusCode);
        string firstBody = await firstResponse.Content.ReadAsStringAsync(
            TestContext.Current.CancellationToken);

        using HttpRequestMessage retry = PostOrder(payload, key);
        HttpResponseMessage retryResponse = await client.SendAsync(
            retry, TestContext.Current.CancellationToken);

        // The stored response replays byte-for-byte: same status, same body,
        // same Location - and NO second order was created (the duplicate
        // constraint would have made a re-execution a 409).
        Assert.Equal(HttpStatusCode.Created, retryResponse.StatusCode);
        Assert.Equal("true", Assert.Single(retryResponse.Headers.GetValues("Idempotency-Replayed")));
        Assert.Equal(
            firstBody,
            await retryResponse.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));
        Assert.Equal(
            firstResponse.Headers.Location?.ToString(),
            retryResponse.Headers.Location?.ToString());

        // Same key, different payload: 422 key reuse.
        using HttpRequestMessage reuse = PostOrder(
            $$"""{"externalReference":"{{reference}}-x","customerId":"cust-1","quantity":4}""", key);
        HttpResponseMessage reuseResponse = await client.SendAsync(
            reuse, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.UnprocessableEntity, reuseResponse.StatusCode);
    }
}
