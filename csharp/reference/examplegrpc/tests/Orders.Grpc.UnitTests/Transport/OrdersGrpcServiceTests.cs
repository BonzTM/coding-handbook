using Google.Protobuf.WellKnownTypes;
using Google.Rpc;

using Grpc.Core;

using Orders.Grpc.Api.V1;

using Xunit;

namespace Orders.Grpc.UnitTests.Transport;

/// <summary>
/// The wire contract through the real pipeline: CRUD round trips, the
/// google.rpc.BadRequest validation detail, keyset pagination tokens, and the
/// server-streaming method - all through the generated client over an in-proc
/// GrpcChannel (csharp/services/grpc-services.md, Verification And Proof).
/// Each test boots its own host so store state never leaks between tests.
/// </summary>
public sealed class OrdersGrpcServiceTests
{
    [Fact]
    public async Task CreateThenGet_RoundTripsTheOrder()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var created = await client.CreateOrderAsync(
            new CreateOrderRequest { ExternalReference = "ord-1001", CustomerId = "cust-42", Quantity = 3 },
            cancellationToken: TestContext.Current.CancellationToken);

        Assert.Equal("ord-1001", created.Order.ExternalReference);
        Assert.Equal("cust-42", created.Order.CustomerId);
        Assert.Equal(3, created.Order.Quantity);
        // Time is an input: the wire timestamp is exactly the fake clock.
        Assert.Equal(Timestamp.FromDateTimeOffset(factory.Time.GetUtcNow()), created.Order.CreatedAt);

        var fetched = await client.GetOrderAsync(
            new GetOrderRequest { Id = created.Order.Id },
            cancellationToken: TestContext.Current.CancellationToken);

        Assert.Equal(created.Order, fetched.Order);
    }

    [Fact]
    public async Task Create_WithInvalidInput_ReturnsInvalidArgumentWithBadRequestDetail()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var exception = await Assert.ThrowsAsync<RpcException>(async () =>
            await client.CreateOrderAsync(
                new CreateOrderRequest { ExternalReference = "", CustomerId = "", Quantity = 0 },
                cancellationToken: TestContext.Current.CancellationToken));

        Assert.Equal(StatusCode.InvalidArgument, exception.StatusCode);

        // Decode the typed detail, never the message string.
        var badRequest = exception.GetRpcStatus()?.GetDetail<BadRequest>();
        Assert.NotNull(badRequest);
        var fields = badRequest.FieldViolations.Select(v => v.Field).ToList();
        Assert.Equal(["external_reference", "customer_id", "quantity"], fields);
    }

    [Fact]
    public async Task Create_DuplicateReference_ReturnsAlreadyExists()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);
        var request = new CreateOrderRequest { ExternalReference = "ord-1", CustomerId = "cust-1", Quantity = 1 };
        _ = await client.CreateOrderAsync(request, cancellationToken: TestContext.Current.CancellationToken);

        var exception = await Assert.ThrowsAsync<RpcException>(async () =>
            await client.CreateOrderAsync(request, cancellationToken: TestContext.Current.CancellationToken));

        Assert.Equal(StatusCode.AlreadyExists, exception.StatusCode);
    }

    [Fact]
    public async Task Get_UnknownId_ReturnsNotFound()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var exception = await Assert.ThrowsAsync<RpcException>(async () =>
            await client.GetOrderAsync(
                new GetOrderRequest { Id = Guid.CreateVersion7().ToString("D") },
                cancellationToken: TestContext.Current.CancellationToken));

        Assert.Equal(StatusCode.NotFound, exception.StatusCode);
    }

    [Fact]
    public async Task Get_MalformedId_ReturnsInvalidArgumentWithFieldViolation()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var exception = await Assert.ThrowsAsync<RpcException>(async () =>
            await client.GetOrderAsync(
                new GetOrderRequest { Id = "not-a-uuid" },
                cancellationToken: TestContext.Current.CancellationToken));

        Assert.Equal(StatusCode.InvalidArgument, exception.StatusCode);
        var badRequest = exception.GetRpcStatus()?.GetDetail<BadRequest>();
        Assert.NotNull(badRequest);
        var violation = Assert.Single(badRequest.FieldViolations);
        Assert.Equal("id", violation.Field);
    }

    [Fact]
    public async Task List_MalformedPageToken_ReturnsInvalidArgument()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var exception = await Assert.ThrowsAsync<RpcException>(async () =>
            await client.ListOrdersAsync(
                new ListOrdersRequest { PageSize = 10, PageToken = "not-a-cursor!!!" },
                cancellationToken: TestContext.Current.CancellationToken));

        Assert.Equal(StatusCode.InvalidArgument, exception.StatusCode);
    }

    [Fact]
    public async Task List_WalksAllPagesThroughNextPageToken()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);
        var createdIds = await SeedAsync(factory, client, count: 5);

        var seen = new List<string>();
        string pageToken = "";
        // Bounded walk: at most one page per seeded order.
        for (int page = 0; page < createdIds.Count; page++)
        {
            var response = await client.ListOrdersAsync(
                new ListOrdersRequest { PageSize = 2, PageToken = pageToken },
                cancellationToken: TestContext.Current.CancellationToken);
            seen.AddRange(response.Orders.Select(o => o.Id));
            pageToken = response.NextPageToken;
            if (pageToken.Length == 0)
            {
                break;
            }
        }

        Assert.Equal(createdIds, seen);
    }

    [Fact]
    public async Task StreamOrders_StreamsEveryOrderInKeysetOrder()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);
        var createdIds = await SeedAsync(factory, client, count: 4);

        using var call = client.StreamOrders(
            new StreamOrdersRequest(), cancellationToken: TestContext.Current.CancellationToken);

        var streamed = new List<string>();
        await foreach (var message in call.ResponseStream.ReadAllAsync(TestContext.Current.CancellationToken))
        {
            streamed.Add(message.Order.Id);
        }

        Assert.Equal(createdIds, streamed);
    }

    private static async Task<List<string>> SeedAsync(
        OrdersGrpcFactory factory, OrdersService.OrdersServiceClient client, int count)
    {
        var ids = new List<string>(count);
        for (int i = 0; i < count; i++)
        {
            var created = await client.CreateOrderAsync(
                new CreateOrderRequest { ExternalReference = $"ord-{i:D3}", CustomerId = "cust-1", Quantity = i + 1 },
                cancellationToken: TestContext.Current.CancellationToken);
            ids.Add(created.Order.Id);
            // Distinct timestamps keep the keyset order unambiguous.
            factory.Time.Advance(TimeSpan.FromSeconds(1));
        }

        return ids;
    }
}
