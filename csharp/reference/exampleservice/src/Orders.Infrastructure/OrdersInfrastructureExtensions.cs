using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

using Orders.Core.Idempotency;
using Orders.Core.Orders;
using Orders.Infrastructure.Data;
using Orders.Infrastructure.Data.Repositories;

namespace Orders.Infrastructure;

/// <summary>
/// Composition entry points for the persistence layer. Program.cs calls
/// <see cref="AddOrdersInfrastructure"/> once - the only place Api crosses the
/// Infrastructure boundary (csharp/foundations/solution-and-project-design.md).
/// </summary>
public static class OrdersInfrastructureExtensions
{
    public static IServiceCollection AddOrdersInfrastructure(
        this IServiceCollection services, IConfiguration configuration)
    {
        ArgumentNullException.ThrowIfNull(services);
        ArgumentNullException.ThrowIfNull(configuration);

        // Fail-fast options: a missing connection string kills the process at
        // startup, before listeners open (csharp/foundations/configuration.md).
        services.AddOptions<DatabaseOptions>()
            .BindConfiguration(DatabaseOptions.SectionName)
            .ValidateDataAnnotations()
            .ValidateOnStart();
        services.AddOptions<IdempotencyOptions>()
            .BindConfiguration(IdempotencyOptions.SectionName)
            .ValidateDataAnnotations()
            .ValidateOnStart();

        // Pooled DbContext instances (allocation optimization); the PHYSICAL
        // connection pool is sized by the connection string, which must carry
        // all four Npgsql limits (csharp/services/database.md).
        services.AddDbContextPool<OrdersDbContext>((provider, options) =>
        {
            string connectionString = provider.GetRequiredService<IOptions<DatabaseOptions>>().Value.Default;
            options
                .UseNpgsql(connectionString, npgsql => npgsql.EnableRetryOnFailure())
                .UseQueryTrackingBehavior(QueryTrackingBehavior.NoTracking);
        });

        services.AddScoped<IOrderRepository, PostgresOrderRepository>();
        // Scoped so the runner shares the request's DbContext with the
        // repository - that shared unit of work is what makes the domain write
        // and the idempotency record commit atomically.
        services.AddScoped<IIdempotencyRunner, PostgresIdempotencyRunner>();

        // Core domain service, composed here so Program.cs stays a wiring
        // manifest. Core itself carries no DI registration code.
        services.AddScoped<OrderService>();

        // /readyz gates traffic on the database being reachable; /livez never
        // runs this check (csharp/operations/observability.md).
        services.AddHealthChecks()
            .AddDbContextCheck<OrdersDbContext>("database", tags: ["ready"]);

        return services;
    }

    /// <summary>
    /// One-shot migration mode: apply pending EF Core migrations, then exit.
    /// Run as an init step or Job (`dotnet Orders.Api.dll --migrate`) ahead of
    /// the rollout. Migrations NEVER run on normal startup - N replicas racing
    /// to alter the schema is an outage generator (csharp/services/database.md).
    /// </summary>
    public static async Task MigrateOrdersDatabaseAsync(this IHost host, CancellationToken cancellationToken = default)
    {
        ArgumentNullException.ThrowIfNull(host);

        await using var scope = host.Services.CreateAsyncScope();
        var db = scope.ServiceProvider.GetRequiredService<OrdersDbContext>();
        var logger = scope.ServiceProvider.GetRequiredService<ILogger<OrdersDbContext>>();

        MigrationLog.Applying(logger);
        await db.Database.MigrateAsync(cancellationToken).ConfigureAwait(false);
        MigrationLog.Applied(logger);
    }
}

/// <summary>Source-generated log methods (CA1848; csharp/foundations/errors-and-logging.md).</summary>
internal static partial class MigrationLog
{
    [LoggerMessage(Level = LogLevel.Information, Message = "Applying pending EF Core migrations")]
    public static partial void Applying(ILogger logger);

    [LoggerMessage(Level = LogLevel.Information, Message = "Database schema is up to date")]
    public static partial void Applied(ILogger logger);
}
