using Microsoft.Extensions.Logging;

using Orders.Grpc.Api.V1;

using Xunit;

namespace Orders.Grpc.UnitTests.Transport;

/// <summary>
/// The access-log contract (outermost interceptor): one line per RPC carrying
/// the full method, the FINAL status the client saw, and the request id - a
/// well-formed inbound x-request-id is adopted, a malformed one is replaced,
/// never propagated raw.
/// </summary>
public sealed class RequestLoggingTests
{
    [Fact]
    public async Task SuccessfulRpc_LogsOneLineWithMethodStatusAndAdoptedRequestId()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        _ = await client.CreateOrderAsync(
            new CreateOrderRequest { ExternalReference = "ord-1", CustomerId = "cust-1", Quantity = 1 },
            new global::Grpc.Core.Metadata { { "x-request-id", "caller-correlation-42" } },
            cancellationToken: TestContext.Current.CancellationToken);

        var line = Assert.Single(factory.Logs.Snapshot(), record =>
            record.Message.StartsWith("rpc /orders.v1.OrdersService/CreateOrder", StringComparison.Ordinal));
        Assert.Equal(LogLevel.Information, line.Level);
        Assert.Contains(" OK ", line.Message, StringComparison.Ordinal);
        Assert.Contains("request_id=caller-correlation-42", line.Message, StringComparison.Ordinal);
    }

    [Fact]
    public async Task MalformedInboundRequestId_IsReplaced_NeverLoggedRaw()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        _ = await client.CreateOrderAsync(
            new CreateOrderRequest { ExternalReference = "ord-1", CustomerId = "cust-1", Quantity = 1 },
            new global::Grpc.Core.Metadata { { "x-request-id", "bad id with spaces!" } },
            cancellationToken: TestContext.Current.CancellationToken);

        var line = Assert.Single(factory.Logs.Snapshot(), record =>
            record.Message.StartsWith("rpc /orders.v1.OrdersService/CreateOrder", StringComparison.Ordinal));
        Assert.DoesNotContain("bad id with spaces!", line.Message, StringComparison.Ordinal);
        Assert.Contains("request_id=", line.Message, StringComparison.Ordinal);
    }
}
