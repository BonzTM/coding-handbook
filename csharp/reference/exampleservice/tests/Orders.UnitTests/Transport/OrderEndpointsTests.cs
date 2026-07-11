using System.Net;
using System.Text;
using System.Text.Json;

using Orders.UnitTests.Fakes;

using Xunit;

namespace Orders.UnitTests.Transport;

/// <summary>
/// Endpoint contract tests through the real pipeline: status codes, wire
/// shapes, ProblemDetails (RFC 9457) with requestId, field-level validation
/// errors, and cursor pagination (csharp/services/http-services.md).
/// </summary>
public sealed class OrderEndpointsTests : IClassFixture<OrdersApiFactory>
{
    private readonly OrdersApiFactory _factory;

    public OrderEndpointsTests(OrdersApiFactory factory)
    {
        _factory = factory;
    }

    private static StringContent Json(string payload)
        => new(payload, Encoding.UTF8, "application/json");

    private static async Task<JsonDocument> ReadJson(HttpResponseMessage response)
        => JsonDocument.Parse(await response.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));

    private static async Task<JsonDocument> CreateOrder(HttpClient client, string reference, int quantity = 1)
    {
        using HttpRequestMessage request = new(HttpMethod.Post, "/orders");
        request.Headers.Add("Idempotency-Key", $"key-{reference}-{Guid.NewGuid():N}");
        request.Content = Json($$"""{"externalReference":"{{reference}}","customerId":"cust-1","quantity":{{quantity}}}""");
        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        return await ReadJson(response);
    }

    [Fact]
    public async Task CreateOrder_Valid_Returns201WithLocationAndBody()
    {
        // Own host: the timestamp assertion needs a clock no other test advances.
        using var factory = new OrdersApiFactory();
        using HttpClient client = factory.CreateClient();
        using HttpRequestMessage request = new(HttpMethod.Post, "/orders");
        request.Headers.Add("Idempotency-Key", $"key-{Guid.NewGuid():N}");
        request.Content = Json("""{"externalReference":"ref-create","customerId":"cust-9","quantity":3}""");

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Created, response.StatusCode);
        using JsonDocument body = await ReadJson(response);
        string orderId = body.RootElement.GetProperty("orderId").GetString()!;
        Assert.Equal($"/orders/{orderId}", response.Headers.Location?.ToString());
        Assert.Equal("ref-create", body.RootElement.GetProperty("externalReference").GetString());
        Assert.Equal("cust-9", body.RootElement.GetProperty("customerId").GetString());
        Assert.Equal(3, body.RootElement.GetProperty("quantity").GetInt32());
        Assert.Equal("Pending", body.RootElement.GetProperty("status").GetString()); // enum as string
        Assert.Equal("2026-07-01T12:00:00+00:00", body.RootElement.GetProperty("createdAt").GetString());
    }

    [Fact]
    public async Task CreateOrder_InvalidBody_Returns400WithFieldErrors()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = new(HttpMethod.Post, "/orders");
        request.Headers.Add("Idempotency-Key", $"key-{Guid.NewGuid():N}");
        request.Content = Json("""{"externalReference":"","customerId":"cust-1","quantity":0}""");

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        Assert.Equal("application/problem+json", response.Content.Headers.ContentType?.MediaType);
        using JsonDocument body = await ReadJson(response);
        JsonElement errors = body.RootElement.GetProperty("errors");
        Assert.Contains(errors.EnumerateObject(), e =>
            string.Equals(e.Name, "externalReference", StringComparison.OrdinalIgnoreCase));
        Assert.Contains(errors.EnumerateObject(), e =>
            string.Equals(e.Name, "quantity", StringComparison.OrdinalIgnoreCase));
    }

    [Fact]
    public async Task CreateOrder_DuplicateReference_Returns409Problem()
    {
        using HttpClient client = _factory.CreateClient();
        using JsonDocument _ = await CreateOrder(client, "ref-dup");

        using HttpRequestMessage request = new(HttpMethod.Post, "/orders");
        request.Headers.Add("Idempotency-Key", $"key-{Guid.NewGuid():N}");
        request.Content = Json("""{"externalReference":"ref-dup","customerId":"cust-1","quantity":1}""");
        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
        using JsonDocument body = await ReadJson(response);
        Assert.Equal(
            "https://orders.example/problems/duplicate-order",
            body.RootElement.GetProperty("type").GetString());
        Assert.False(string.IsNullOrEmpty(body.RootElement.GetProperty("requestId").GetString()));
    }

    [Fact]
    public async Task GetOrder_Unknown_Returns404ProblemWithRequestId()
    {
        using HttpClient client = _factory.CreateClient();

        HttpResponseMessage response = await client.GetAsync(
            new Uri($"/orders/{Guid.CreateVersion7()}", UriKind.Relative),
            TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
        Assert.Equal("application/problem+json", response.Content.Headers.ContentType?.MediaType);
        using JsonDocument body = await ReadJson(response);
        Assert.Equal(
            "https://orders.example/problems/order-not-found",
            body.RootElement.GetProperty("type").GetString());
        Assert.False(string.IsNullOrEmpty(body.RootElement.GetProperty("requestId").GetString()));
    }

    [Fact]
    public async Task GetOrder_AfterCreate_RoundTrips()
    {
        using HttpClient client = _factory.CreateClient();
        using JsonDocument created = await CreateOrder(client, "ref-get");
        string orderId = created.RootElement.GetProperty("orderId").GetString()!;

        HttpResponseMessage response = await client.GetAsync(
            new Uri($"/orders/{orderId}", UriKind.Relative), TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        using JsonDocument body = await ReadJson(response);
        Assert.Equal(orderId, body.RootElement.GetProperty("orderId").GetString());
        Assert.Equal("ref-get", body.RootElement.GetProperty("externalReference").GetString());
    }

    [Fact]
    public async Task ListOrders_PagesWithOpaqueCursor()
    {
        // Own host: the page-count assertions need an order inventory no other
        // test writes to.
        using var factory = new OrdersApiFactory();
        using HttpClient client = factory.CreateClient();
        for (int i = 0; i < 3; i++)
        {
            factory.Time.Advance(TimeSpan.FromSeconds(1)); // distinct CreatedAt
            using JsonDocument _ = await CreateOrder(client, $"ref-page-{i}");
        }

        HttpResponseMessage firstResponse = await client.GetAsync(
            new Uri("/orders?pageSize=2", UriKind.Relative), TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.OK, firstResponse.StatusCode);
        using JsonDocument first = await ReadJson(firstResponse);
        Assert.Equal(2, first.RootElement.GetProperty("items").GetArrayLength());
        string cursor = first.RootElement.GetProperty("nextCursor").GetString()!;
        Assert.False(string.IsNullOrEmpty(cursor));

        HttpResponseMessage secondResponse = await client.GetAsync(
            new Uri($"/orders?pageSize=2&cursor={cursor}", UriKind.Relative),
            TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.OK, secondResponse.StatusCode);
        using JsonDocument second = await ReadJson(secondResponse);
        Assert.Equal(1, second.RootElement.GetProperty("items").GetArrayLength());
        Assert.Equal(JsonValueKind.Null, second.RootElement.GetProperty("nextCursor").ValueKind);
    }

    [Fact]
    public async Task ListOrders_MalformedCursor_Returns400ValidationProblem()
    {
        using HttpClient client = _factory.CreateClient();

        HttpResponseMessage response = await client.GetAsync(
            new Uri("/orders?cursor=%20not-a-cursor%20", UriKind.Relative),
            TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        using JsonDocument body = await ReadJson(response);
        Assert.True(body.RootElement.GetProperty("errors").TryGetProperty("cursor", out _));
    }

    [Fact]
    public async Task UpdateOrder_MatchingVersion_Returns200()
    {
        using HttpClient client = _factory.CreateClient();
        using JsonDocument created = await CreateOrder(client, "ref-update");
        string orderId = created.RootElement.GetProperty("orderId").GetString()!;
        uint version = created.RootElement.GetProperty("version").GetUInt32();

        using StringContent amend = Json($$"""{"quantity":7,"status":"Confirmed","version":{{version}}}""");
        HttpResponseMessage response = await client.PutAsync(
            new Uri($"/orders/{orderId}", UriKind.Relative), amend, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        using JsonDocument body = await ReadJson(response);
        Assert.Equal(7, body.RootElement.GetProperty("quantity").GetInt32());
        Assert.Equal("Confirmed", body.RootElement.GetProperty("status").GetString());
    }

    [Fact]
    public async Task UpdateOrder_StaleVersion_Returns409VersionConflict()
    {
        using HttpClient client = _factory.CreateClient();
        using JsonDocument created = await CreateOrder(client, "ref-stale");
        string orderId = created.RootElement.GetProperty("orderId").GetString()!;
        uint stale = created.RootElement.GetProperty("version").GetUInt32() + 41;

        using StringContent amend = Json($$"""{"quantity":7,"status":"Confirmed","version":{{stale}}}""");
        HttpResponseMessage response = await client.PutAsync(
            new Uri($"/orders/{orderId}", UriKind.Relative), amend, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
        using JsonDocument body = await ReadJson(response);
        Assert.Equal(
            "https://orders.example/problems/version-conflict",
            body.RootElement.GetProperty("type").GetString());
    }

    [Fact]
    public async Task UpdateOrder_UnknownStatus_Returns400()
    {
        using HttpClient client = _factory.CreateClient();
        using JsonDocument created = await CreateOrder(client, "ref-badstatus");
        string orderId = created.RootElement.GetProperty("orderId").GetString()!;

        using StringContent amend = Json("""{"quantity":7,"status":"Unknown","version":0}""");
        HttpResponseMessage response = await client.PutAsync(
            new Uri($"/orders/{orderId}", UriKind.Relative), amend, TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    [Fact]
    public async Task DeleteOrder_Existing_Returns204ThenGetReturns404()
    {
        using HttpClient client = _factory.CreateClient();
        using JsonDocument created = await CreateOrder(client, "ref-delete");
        string orderId = created.RootElement.GetProperty("orderId").GetString()!;

        HttpResponseMessage deleteResponse = await client.DeleteAsync(
            new Uri($"/orders/{orderId}", UriKind.Relative), TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.NoContent, deleteResponse.StatusCode);

        HttpResponseMessage getResponse = await client.GetAsync(
            new Uri($"/orders/{orderId}", UriKind.Relative), TestContext.Current.CancellationToken);
        Assert.Equal(HttpStatusCode.NotFound, getResponse.StatusCode);
    }

    [Fact]
    public async Task Livez_IsAnonymousAndAlwaysHealthy()
    {
        using HttpClient client = _factory.CreateClient();

        HttpResponseMessage response = await client.GetAsync(
            new Uri("/livez", UriKind.Relative), TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
    }

    [Fact]
    public async Task CreateOrder_EmitsAuditRecordOnDedicatedCategory()
    {
        using HttpClient client = _factory.CreateClient();
        using JsonDocument created = await CreateOrder(client, "ref-audit");
        string orderId = created.RootElement.GetProperty("orderId").GetString()!;

        CapturedLogRecord record = Assert.Single(
            _factory.Logs.AuditRecords(),
            r => r.Message.Contains("order.create", StringComparison.Ordinal)
                && r.Message.Contains(orderId, StringComparison.Ordinal));
        Assert.Contains("local-dev@local-dev", record.Message, StringComparison.Ordinal);
        Assert.Contains("result=success", record.Message, StringComparison.Ordinal);
    }
}
