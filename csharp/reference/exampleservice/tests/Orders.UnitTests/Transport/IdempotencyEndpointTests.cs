using System.Net;
using System.Text;
using System.Text.Json;

using Orders.Api.Idempotency;
using Orders.Core.Idempotency;
using Orders.Core.Orders;

using Xunit;

namespace Orders.UnitTests.Transport;

/// <summary>
/// The Idempotency-Key contract on POST /orders through the real filter
/// (csharp/recipes/add-idempotent-write.md): required key, byte-identical
/// replay, in-flight 409, key-reuse 422. Storage semantics use the in-memory
/// runner; the transactional runner is proven by the integration suite.
/// </summary>
public sealed class IdempotencyEndpointTests : IClassFixture<OrdersApiFactory>
{
    private const string OrdersRoute = "POST /orders";

    private readonly OrdersApiFactory _factory;

    public IdempotencyEndpointTests(OrdersApiFactory factory)
    {
        _factory = factory;
    }

    private static HttpRequestMessage PostOrder(string payload, string? key)
    {
        HttpRequestMessage request = new(HttpMethod.Post, "/orders")
        {
            Content = new StringContent(payload, Encoding.UTF8, "application/json"),
        };
        if (key is not null)
        {
            request.Headers.Add("Idempotency-Key", key);
        }

        return request;
    }

    [Fact]
    public async Task Post_WithoutIdempotencyKey_Returns400Problem()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = PostOrder(
            """{"externalReference":"ref-nokey","customerId":"cust-1","quantity":1}""", key: null);

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        using var body = JsonDocument.Parse(
            await response.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));
        Assert.Equal(
            "https://orders.example/problems/idempotency-key-missing",
            body.RootElement.GetProperty("type").GetString());
    }

    [Fact]
    public async Task Post_MalformedIdempotencyKey_Returns400()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = PostOrder(
            """{"externalReference":"ref-badkey","customerId":"cust-1","quantity":1}""", key: "no spaces allowed");

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    [Fact]
    public async Task Post_RetryWithSameKeyAndBody_ReplaysFirstResponseByteForByte()
    {
        using HttpClient client = _factory.CreateClient();
        string key = $"key-replay-{Guid.NewGuid():N}";
        const string Payload = """{"externalReference":"ref-replay","customerId":"cust-1","quantity":2}""";

        using HttpRequestMessage first = PostOrder(Payload, key);
        HttpResponseMessage firstResponse = await client.SendAsync(first, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Created, firstResponse.StatusCode);
        Assert.False(firstResponse.Headers.Contains(StoredResponseResult.ReplayHeaderName));
        string firstBody = await firstResponse.Content.ReadAsStringAsync(TestContext.Current.CancellationToken);

        using HttpRequestMessage retry = PostOrder(Payload, key);
        HttpResponseMessage retryResponse = await client.SendAsync(retry, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Created, retryResponse.StatusCode);
        Assert.Equal(
            firstResponse.Headers.Location?.ToString(),
            retryResponse.Headers.Location?.ToString());
        Assert.Equal("true", Assert.Single(
            retryResponse.Headers.GetValues(StoredResponseResult.ReplayHeaderName)));
        string retryBody = await retryResponse.Content.ReadAsStringAsync(TestContext.Current.CancellationToken);
        Assert.Equal(firstBody, retryBody); // byte-identical replay, no second create
    }

    [Fact]
    public async Task Post_SameKeyDifferentBody_Returns422KeyReuse()
    {
        using HttpClient client = _factory.CreateClient();
        string key = $"key-reuse-{Guid.NewGuid():N}";

        using HttpRequestMessage first = PostOrder(
            """{"externalReference":"ref-reuse","customerId":"cust-1","quantity":1}""", key);
        HttpResponseMessage firstResponse = await client.SendAsync(first, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Created, firstResponse.StatusCode);

        using HttpRequestMessage second = PostOrder(
            """{"externalReference":"ref-reuse-other","customerId":"cust-1","quantity":9}""", key);
        HttpResponseMessage secondResponse = await client.SendAsync(second, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.UnprocessableEntity, secondResponse.StatusCode);
        using var body = JsonDocument.Parse(
            await secondResponse.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));
        Assert.Equal(
            "https://orders.example/problems/idempotency-key-reuse",
            body.RootElement.GetProperty("type").GetString());
    }

    [Fact]
    public async Task Post_WhileFirstRequestInFlight_Returns409()
    {
        using HttpClient client = _factory.CreateClient();
        string key = $"key-inflight-{Guid.NewGuid():N}";
        _factory.Idempotency.MarkInFlight(
            new IdempotencyScope(new TenantId("local-dev"), OrdersRoute, key), "fingerprint");

        using HttpRequestMessage request = PostOrder(
            """{"externalReference":"ref-inflight","customerId":"cust-1","quantity":1}""", key);
        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
        using var body = JsonDocument.Parse(
            await response.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));
        Assert.Equal(
            "https://orders.example/problems/idempotency-in-flight",
            body.RootElement.GetProperty("type").GetString());
    }

    [Fact]
    public async Task Post_DomainErrorOutcome_IsNotRecordedSoRetryReExecutes()
    {
        using HttpClient client = _factory.CreateClient();
        string key = $"key-error-{Guid.NewGuid():N}";
        const string Duplicate = """{"externalReference":"ref-error-dup","customerId":"cust-1","quantity":1}""";

        // Seed the duplicate with a DIFFERENT key so the second create conflicts.
        using HttpRequestMessage seed = PostOrder(Duplicate, $"key-seed-{Guid.NewGuid():N}");
        HttpResponseMessage seedResponse = await client.SendAsync(seed, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Created, seedResponse.StatusCode);

        using HttpRequestMessage conflicting = PostOrder(Duplicate, key);
        HttpResponseMessage conflictResponse = await client.SendAsync(
            conflicting, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Conflict, conflictResponse.StatusCode);

        // The 409 was NOT recorded against the key: retrying the key with a
        // now-valid body executes normally instead of replaying the error.
        using HttpRequestMessage retry = PostOrder(
            """{"externalReference":"ref-error-fresh","customerId":"cust-1","quantity":1}""", key);
        HttpResponseMessage retryResponse = await client.SendAsync(retry, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Created, retryResponse.StatusCode);
    }
}
