using Grpc.Core;
using Grpc.Health.V1;

using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.TestHost;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.DependencyInjection.Extensions;

using Orders.Grpc.Api.Auth;
using Orders.Grpc.Api.V1;
using Orders.Grpc.Core.Identity;

using Xunit;

namespace Orders.Grpc.UnitTests.Transport;

/// <summary>
/// The auth interceptor's contract: with `Auth:Enabled=true` a missing or
/// wrong bearer token is UNAUTHENTICATED (generic message - never which check
/// failed), the configured token succeeds, health stays reachable without a
/// token (the kubelet analog), and a principal without the writer role is
/// PERMISSION_DENIED from the domain, not the interceptor.
/// </summary>
public sealed class AuthTests
{
    private const string ValidToken = "test-bearer-token-0123456789abcdef";

    [Fact]
    public async Task AuthEnabled_MissingToken_IsUnauthenticated()
    {
        using var factory = new AuthEnabledFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var exception = await Assert.ThrowsAsync<RpcException>(async () =>
            await client.GetOrderAsync(
                new GetOrderRequest { Id = Guid.CreateVersion7().ToString("D") },
                cancellationToken: TestContext.Current.CancellationToken));

        Assert.Equal(StatusCode.Unauthenticated, exception.StatusCode);
        Assert.Equal("invalid or missing bearer token", exception.Status.Detail);
    }

    [Fact]
    public async Task AuthEnabled_WrongToken_IsUnauthenticated()
    {
        using var factory = new AuthEnabledFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var exception = await Assert.ThrowsAsync<RpcException>(async () =>
            await client.CreateOrderAsync(
                new CreateOrderRequest { ExternalReference = "ord-1", CustomerId = "cust-1", Quantity = 1 },
                new Metadata { { "authorization", "Bearer wrong-token-wrong-token" } },
                cancellationToken: TestContext.Current.CancellationToken));

        Assert.Equal(StatusCode.Unauthenticated, exception.StatusCode);
    }

    [Fact]
    public async Task AuthEnabled_ConfiguredToken_Succeeds()
    {
        using var factory = new AuthEnabledFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var created = await client.CreateOrderAsync(
            new CreateOrderRequest { ExternalReference = "ord-1", CustomerId = "cust-1", Quantity = 1 },
            new Metadata { { "authorization", $"Bearer {ValidToken}" } },
            cancellationToken: TestContext.Current.CancellationToken);

        Assert.Equal("ord-1", created.Order.ExternalReference);
    }

    [Fact]
    public async Task AuthEnabled_HealthService_NeedsNoToken()
    {
        using var factory = new AuthEnabledFactory();
        using var channel = factory.CreateGrpcChannel();
        var health = new Health.HealthClient(channel);

        var response = await health.CheckAsync(
            new HealthCheckRequest(), cancellationToken: TestContext.Current.CancellationToken);

        Assert.Equal(HealthCheckResponse.Types.ServingStatus.Serving, response.Status);
    }

    [Fact]
    public async Task PrincipalWithoutWriterRole_CreateIsPermissionDenied()
    {
        using var factory = new ReaderOnlyPrincipalFactory();
        using var channel = factory.CreateGrpcChannel();
        var client = new OrdersService.OrdersServiceClient(channel);

        var exception = await Assert.ThrowsAsync<RpcException>(async () =>
            await client.CreateOrderAsync(
                new CreateOrderRequest { ExternalReference = "ord-1", CustomerId = "cust-1", Quantity = 1 },
                cancellationToken: TestContext.Current.CancellationToken));

        Assert.Equal(StatusCode.PermissionDenied, exception.StatusCode);
    }

    /// <summary>Production-shaped host: static-token auth on, everything else defaults.</summary>
    private sealed class AuthEnabledFactory : OrdersGrpcFactory
    {
        protected override Dictionary<string, string?> Settings { get; } = new(StringComparer.Ordinal)
        {
            ["Auth:Enabled"] = "true",
            ["Auth:BearerToken"] = ValidToken,
            ["Auth:Tenant"] = "tenant-a",
        };
    }

    /// <summary>Swaps the authenticator seam for one resolving a reader-only principal.</summary>
    private sealed class ReaderOnlyPrincipalFactory : OrdersGrpcFactory
    {
        protected override void ConfigureWebHost(IWebHostBuilder builder)
        {
            base.ConfigureWebHost(builder);
            builder.ConfigureTestServices(services =>
            {
                services.RemoveAll<IAuthenticator>();
                services.AddSingleton<IAuthenticator>(new StubAuthenticator(new CallerPrincipal(
                    "reader", new TenantId("tenant-a"), [OrderRoles.Reader])));
            });
        }
    }

    private sealed class StubAuthenticator(CallerPrincipal principal) : IAuthenticator
    {
        public ValueTask<CallerPrincipal?> AuthenticateAsync(
            string? bearerToken, CancellationToken cancellationToken)
            => ValueTask.FromResult<CallerPrincipal?>(principal);
    }
}
