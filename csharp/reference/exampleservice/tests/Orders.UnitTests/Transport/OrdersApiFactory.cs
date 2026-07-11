using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.AspNetCore.TestHost;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.DependencyInjection.Extensions;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Time.Testing;

using Orders.Core.Idempotency;
using Orders.Core.Orders;
using Orders.UnitTests.Fakes;

namespace Orders.UnitTests.Transport;

/// <summary>
/// In-proc host for offline transport tests (csharp/quality/testing.md):
/// the REAL pipeline - middleware order, auth (local-dev scheme), endpoint
/// filters, ProblemDetails mapping, serialization - with hand-rolled fakes
/// swapped in behind the Core ports so no database or network is touched.
/// Extra configuration keys come in through <see cref="Settings"/> BEFORE the
/// host builds, so fail-fast options validation still runs against them.
/// </summary>
public class OrdersApiFactory : WebApplicationFactory<Program>
{
    public InMemoryOrderRepository Repository { get; } = new();

    public InMemoryIdempotencyRunner Idempotency { get; } = new();

    public CapturingLoggerProvider Logs { get; } = new();

    public FakeTimeProvider Time { get; } = new(new DateTimeOffset(2026, 7, 1, 12, 0, 0, TimeSpan.Zero));

    /// <summary>Configuration overrides applied on top of appsettings.json.</summary>
    protected virtual Dictionary<string, string?> Settings { get; } = new(StringComparer.Ordinal);

    protected override void ConfigureWebHost(IWebHostBuilder builder)
    {
        ArgumentNullException.ThrowIfNull(builder);

        foreach ((string key, string? value) in Settings)
        {
            builder.UseSetting(key, value);
        }

        builder.ConfigureLogging(logging => logging.AddProvider(Logs));

        // ConfigureTestServices runs AFTER Program.cs registrations, so these
        // replace the PostgreSQL adapters for the whole host.
        builder.ConfigureTestServices(services =>
        {
            services.RemoveAll<IOrderRepository>();
            services.AddSingleton<IOrderRepository>(Repository);
            services.RemoveAll<IIdempotencyRunner>();
            services.AddSingleton<IIdempotencyRunner>(Idempotency);
            services.RemoveAll<TimeProvider>();
            services.AddSingleton<TimeProvider>(Time);
        });
    }
}
