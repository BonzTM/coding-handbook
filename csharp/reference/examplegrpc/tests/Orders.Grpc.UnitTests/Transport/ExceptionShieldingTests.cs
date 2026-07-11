using Grpc.Core;

using Microsoft.Extensions.Logging;

using Orders.Grpc.Api.V1;
using Orders.Grpc.UnitTests.Fakes;

using Xunit;

namespace Orders.Grpc.UnitTests.Transport;

/// <summary>
/// The exception-mapping interceptor's shielding and cancellation contracts,
/// driven deterministically through the real transport with a faulting store:
/// unexpected exceptions become a generic INTERNAL (logged once server-side,
/// never leaked), and a cancellation the client did NOT cause maps to
/// DEADLINE_EXCEEDED - the server-side ceiling's signature.
/// </summary>
public sealed class ExceptionShieldingTests
{
    [Fact]
    public async Task UnexpectedStoreException_IsShieldedAsGenericInternal_AndLoggedOnce()
    {
        using var factory = new OrdersGrpcFactory(
            new ThrowingOrderStore(() => new InvalidOperationException("secret internals")));
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var exception = await Assert.ThrowsAsync<RpcException>(async () =>
            await client.GetOrderAsync(
                new GetOrderRequest { Id = Guid.CreateVersion7().ToString("D") },
                cancellationToken: TestContext.Current.CancellationToken));

        Assert.Equal(StatusCode.Internal, exception.StatusCode);
        Assert.Equal("internal error", exception.Status.Detail);
        Assert.DoesNotContain("secret", exception.ToString(), StringComparison.OrdinalIgnoreCase);

        // Logged once, server-side, with the method that failed.
        var unexpected = Assert.Single(factory.Logs.Snapshot(), record =>
            record.Level == LogLevel.Error
            && record.Message.Contains("unexpected exception", StringComparison.Ordinal));
        Assert.Contains("/orders.v1.OrdersService/GetOrder", unexpected.Message, StringComparison.Ordinal);
    }

    [Fact]
    public async Task CancellationTheClientDidNotCause_MapsToDeadlineExceeded()
    {
        // The store throws OperationCanceledException while the CLIENT's call
        // token is not cancelled - exactly what the RpcDeadlineGuard ceiling
        // produces when a client sent no deadline.
        using var factory = new OrdersGrpcFactory(
            new ThrowingOrderStore(() => new OperationCanceledException()));
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var exception = await Assert.ThrowsAsync<RpcException>(async () =>
            await client.GetOrderAsync(
                new GetOrderRequest { Id = Guid.CreateVersion7().ToString("D") },
                cancellationToken: TestContext.Current.CancellationToken));

        Assert.Equal(StatusCode.DeadlineExceeded, exception.StatusCode);
    }
}
