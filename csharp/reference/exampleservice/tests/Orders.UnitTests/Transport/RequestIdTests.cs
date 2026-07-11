using System.Net;
using System.Text.Json;

using Orders.Api.Middleware;

using Xunit;

namespace Orders.UnitTests.Transport;

/// <summary>
/// The request-id contract (csharp/services/http-services.md): a well-formed
/// inbound X-Request-Id is adopted and echoed - including inside ProblemDetails
/// bodies - and a malformed one is replaced, never propagated raw.
/// </summary>
public sealed class RequestIdTests : IClassFixture<OrdersApiFactory>
{
    private readonly OrdersApiFactory _factory;

    public RequestIdTests(OrdersApiFactory factory)
    {
        _factory = factory;
    }

    [Fact]
    public async Task WellFormedInboundRequestId_IsAdoptedAndEchoed()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = new(HttpMethod.Get, $"/orders/{Guid.CreateVersion7()}");
        request.Headers.Add(RequestIdMiddleware.HeaderName, "caller-correlation-42");

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        Assert.Equal(
            "caller-correlation-42",
            Assert.Single(response.Headers.GetValues(RequestIdMiddleware.HeaderName)));

        // The 404 ProblemDetails carries the SAME id, joining wire and logs.
        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
        using var body = JsonDocument.Parse(
            await response.Content.ReadAsStringAsync(TestContext.Current.CancellationToken));
        Assert.Equal("caller-correlation-42", body.RootElement.GetProperty("requestId").GetString());
    }

    [Fact]
    public async Task MalformedInboundRequestId_IsReplacedWithGeneratedId()
    {
        using HttpClient client = _factory.CreateClient();
        using HttpRequestMessage request = new(HttpMethod.Get, "/livez");
        request.Headers.Add(RequestIdMiddleware.HeaderName, "bad id with spaces!");

        HttpResponseMessage response = await client.SendAsync(request, TestContext.Current.CancellationToken);

        string echoed = Assert.Single(response.Headers.GetValues(RequestIdMiddleware.HeaderName));
        Assert.NotEqual("bad id with spaces!", echoed);
        Assert.False(string.IsNullOrEmpty(echoed));
    }

    [Fact]
    public async Task AbsentRequestId_GetsGeneratedIdOnResponse()
    {
        using HttpClient client = _factory.CreateClient();

        HttpResponseMessage response = await client.GetAsync(
            new Uri("/livez", UriKind.Relative), TestContext.Current.CancellationToken);

        string echoed = Assert.Single(response.Headers.GetValues(RequestIdMiddleware.HeaderName));
        Assert.False(string.IsNullOrEmpty(echoed));
    }

    [Theory]
    [InlineData("abc-123", true)]
    [InlineData("trace:span/1+x", true)]
    [InlineData("", false)]
    [InlineData("has space", false)]
    [InlineData("emoji-☃", false)]
    public void IsWellFormed_BoundsCharsetAndLength(string candidate, bool expected)
    {
        Assert.Equal(expected, RequestIdMiddleware.IsWellFormed(candidate));
    }

    [Fact]
    public void IsWellFormed_OversizedId_IsRejected()
    {
        Assert.False(RequestIdMiddleware.IsWellFormed(new string('a', 65)));
    }
}
