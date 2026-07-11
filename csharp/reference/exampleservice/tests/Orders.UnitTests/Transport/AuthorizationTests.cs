using System.Net;
using System.Text;
using System.Text.Json;

using Xunit;

namespace Orders.UnitTests.Transport;

/// <summary>
/// AuthN/AuthZ contract against the real JWT/JWKS pipeline
/// (csharp/operations/security.md): deny-by-default 401s, role-based 403s,
/// tenant scoping from claims (cross-tenant reads are 404, never data), and
/// audit events for failures and denials.
/// </summary>
public sealed class AuthorizationTests : IClassFixture<JwksApiFactory>
{
    private static readonly string[] _readerWriter = ["orders.reader", "orders.writer"];
    private static readonly string[] _readerOnly = ["orders.reader"];
    private static readonly string[] _noRoles = [];

    private readonly JwksApiFactory _factory;

    public AuthorizationTests(JwksApiFactory factory)
    {
        _factory = factory;
    }

    private static StringContent OrderPayload(string reference)
        => new(
            $$"""{"externalReference":"{{reference}}","customerId":"cust-1","quantity":1}""",
            Encoding.UTF8,
            "application/json");

    private static async Task<string> CreateOrderAs(
        HttpClient client, string token, string reference)
    {
        using HttpRequestMessage request = new(HttpMethod.Post, "/orders");
        request.Headers.Authorization = JwksApiFactory.Bearer(token);
        request.Headers.Add("Idempotency-Key", $"key-{Guid.NewGuid():N}");
        request.Content = OrderPayload(reference);
        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        using var body = JsonDocument.Parse(
            await response.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));
        return body.RootElement.GetProperty("orderId").GetString()!;
    }

    [Fact]
    public async Task Request_WithoutToken_Is401AndAudited()
    {
        using HttpClient client = _factory.CreateClient();

        HttpResponseMessage response = await client.GetAsync(
            new Uri("/orders", UriKind.Relative), TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
        Assert.Contains(
            _factory.Logs.AuditRecords(),
            r => r.Message.Contains("result=failure", StringComparison.Ordinal));
    }

    [Fact]
    public async Task Request_WithUnsignedToken_Is401()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = new(HttpMethod.Get, "/orders");
        request.Headers.Authorization = JwksApiFactory.Bearer(
            _factory.CreateToken("alice", "tenant-a", _readerWriter, sign: false));

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task Request_WithWrongAudience_Is401()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = new(HttpMethod.Get, "/orders");
        request.Headers.Authorization = JwksApiFactory.Bearer(
            _factory.CreateToken("alice", "tenant-a", _readerWriter, audience: "some-other-api"));

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task ValidToken_WithReaderRole_CanList()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = new(HttpMethod.Get, "/orders");
        request.Headers.Authorization = JwksApiFactory.Bearer(
            _factory.CreateToken("alice", "tenant-a", _readerOnly));

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
    }

    [Fact]
    public async Task ReaderWithoutWriterRole_CannotCreate_403Audited()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = new(HttpMethod.Post, "/orders");
        request.Headers.Authorization = JwksApiFactory.Bearer(
            _factory.CreateToken("mallory", "tenant-a", _readerOnly));
        request.Headers.Add("Idempotency-Key", $"key-{Guid.NewGuid():N}");
        request.Content = OrderPayload("ref-denied");

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
        Assert.Contains(
            _factory.Logs.AuditRecords(),
            r => r.Message.Contains("mallory", StringComparison.Ordinal)
                && r.Message.Contains("result=denied", StringComparison.Ordinal));
    }

    [Fact]
    public async Task TokenWithoutRoles_CannotRead_403()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = new(HttpMethod.Get, "/orders");
        request.Headers.Authorization = JwksApiFactory.Bearer(
            _factory.CreateToken("nobody", "tenant-a", _noRoles));

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
    }

    [Fact]
    public async Task TokenWithoutTenantClaim_Is403TenantMissing()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = new(HttpMethod.Get, "/orders");
        request.Headers.Authorization = JwksApiFactory.Bearer(
            _factory.CreateToken("alice", tenant: "", _readerWriter));

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
        using var body = JsonDocument.Parse(
            await response.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));
        Assert.Equal(
            "https://orders.example/problems/tenant-missing",
            body.RootElement.GetProperty("type").GetString());
    }

    [Fact]
    public async Task CrossTenantRead_IsIndistinguishableFrom404()
    {
        using HttpClient client = _factory.CreateClient();
        string orderId = await CreateOrderAs(
            client, _factory.CreateToken("alice", "tenant-a", _readerWriter), "ref-cross-tenant");

        using HttpRequestMessage request = new(HttpMethod.Get, $"/orders/{orderId}");
        request.Headers.Authorization = JwksApiFactory.Bearer(
            _factory.CreateToken("eve", "tenant-b", _readerWriter));
        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);

        // The owner still sees it.
        using HttpRequestMessage ownerRequest = new(HttpMethod.Get, $"/orders/{orderId}");
        ownerRequest.Headers.Authorization = JwksApiFactory.Bearer(
            _factory.CreateToken("alice", "tenant-a", _readerWriter));
        HttpResponseMessage ownerResponse = await client.SendAsync(
            ownerRequest, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.OK, ownerResponse.StatusCode);
    }
}
