using Grpc.Net.Client;

using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.AspNetCore.TestHost;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.DependencyInjection.Extensions;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Time.Testing;

using Orders.Grpc.Core.Orders;
using Orders.Grpc.UnitTests.Fakes;

namespace Orders.Grpc.UnitTests.Transport;

/// <summary>
/// In-proc host for offline transport tests (csharp/quality/testing.md): the
/// REAL pipeline - the interceptor chain in its pinned order, status mapping,
/// health, reflection, serialization - driven through a GrpcChannel over the
/// TestServer handler, so no socket and no network. Extra configuration keys
/// come in through <see cref="Settings"/> BEFORE the host builds, so
/// fail-fast options validation still runs against them.
/// </summary>
public class OrdersGrpcFactory : WebApplicationFactory<Program>
{
    public OrdersGrpcFactory()
        : this(new InMemoryOrderStore())
    {
    }

    /// <summary>Error-path tests swap in a faulting store (e.g. <see cref="ThrowingOrderStore"/>).</summary>
    public OrdersGrpcFactory(IOrderStore store)
    {
        Store = store;
    }

    public IOrderStore Store { get; }

    public CapturingLoggerProvider Logs { get; } = new();

    public FakeTimeProvider Time { get; } = new(new DateTimeOffset(2026, 7, 1, 12, 0, 0, TimeSpan.Zero));

    /// <summary>Configuration overrides applied on top of appsettings.json.</summary>
    protected virtual Dictionary<string, string?> Settings { get; } = new(StringComparer.Ordinal);

    /// <summary>A channel into the in-proc server; dispose it with the test.</summary>
    public GrpcChannel CreateGrpcChannel() => GrpcChannel.ForAddress(
        Server.BaseAddress,
        new GrpcChannelOptions { HttpHandler = Server.CreateHandler() });

    protected override void ConfigureWebHost(IWebHostBuilder builder)
    {
        ArgumentNullException.ThrowIfNull(builder);

        foreach ((string key, string? value) in Settings)
        {
            builder.UseSetting(key, value);
        }

        builder.ConfigureLogging(logging => logging.AddProvider(Logs));

        // ConfigureTestServices runs AFTER Program.cs registrations, so these
        // replace the production singletons for the whole host.
        builder.ConfigureTestServices(services =>
        {
            services.RemoveAll<IOrderStore>();
            services.AddSingleton(Store);
            services.RemoveAll<TimeProvider>();
            services.AddSingleton<TimeProvider>(Time);
        });
    }
}
