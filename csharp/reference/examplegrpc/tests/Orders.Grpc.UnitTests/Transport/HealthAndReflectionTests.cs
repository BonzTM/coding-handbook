using Grpc.Health.V1;
using Grpc.Reflection.V1Alpha;

using Xunit;

namespace Orders.Grpc.UnitTests.Transport;

/// <summary>
/// Deployability plumbing (csharp/services/grpc-services.md): the standard
/// gRPC health service answers SERVING, server reflection lists the orders
/// service (what makes grpcurl work without local protos), and the HTTP
/// probes respond for the kubelet.
/// </summary>
public sealed class HealthAndReflectionTests
{
    [Fact]
    public async Task GrpcHealth_Check_ReturnsServing()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var health = new Health.HealthClient(channel);

        var response = await health.CheckAsync(
            new HealthCheckRequest(), cancellationToken: TestContext.Current.CancellationToken);

        Assert.Equal(HealthCheckResponse.Types.ServingStatus.Serving, response.Status);
    }

    [Fact]
    public async Task Reflection_ListsTheOrdersService()
    {
        using var factory = new OrdersGrpcFactory();
        using var channel = factory.CreateGrpcChannel();
        var reflection = new ServerReflection.ServerReflectionClient(channel);

        using var call = reflection.ServerReflectionInfo(
            cancellationToken: TestContext.Current.CancellationToken);
        await call.RequestStream.WriteAsync(
            new ServerReflectionRequest { ListServices = "" }, TestContext.Current.CancellationToken);
        await call.RequestStream.CompleteAsync();

        Assert.True(await call.ResponseStream.MoveNext(TestContext.Current.CancellationToken));
        var services = call.ResponseStream.Current.ListServicesResponse.Service.Select(s => s.Name).ToList();
        Assert.Contains("orders.v1.OrdersService", services);
        Assert.Contains("grpc.health.v1.Health", services);
    }

    [Fact]
    public async Task HttpProbes_LivezAndReadyz_AnswerHealthy()
    {
        using var factory = new OrdersGrpcFactory();
        using var client = factory.CreateClient();

        using var livez = await client.GetAsync(
            new Uri("/livez", UriKind.Relative), TestContext.Current.CancellationToken);
        using var readyz = await client.GetAsync(
            new Uri("/readyz", UriKind.Relative), TestContext.Current.CancellationToken);

        Assert.True(livez.IsSuccessStatusCode);
        Assert.True(readyz.IsSuccessStatusCode);
    }
}
