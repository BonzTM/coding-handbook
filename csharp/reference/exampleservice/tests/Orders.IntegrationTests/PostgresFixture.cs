using Microsoft.EntityFrameworkCore;

using Orders.Infrastructure.Data;

using Testcontainers.PostgreSql;

using Xunit;

[assembly: AssemblyFixture(typeof(Orders.IntegrationTests.PostgresFixture))]

namespace Orders.IntegrationTests;

/// <summary>
/// One disposable PostgreSQL container for the whole assembly
/// (csharp/quality/testing.md, Integration Defaults): started once, the EF
/// Core migrations applied once, then every test class works against the real
/// provider - real SQLSTATE 23505, real xmin, real row-value pagination.
/// Tests isolate on unique tenants/references, not separate databases.
/// </summary>
public sealed class PostgresFixture : IAsyncLifetime
{
    private readonly PostgreSqlContainer _container =
        new PostgreSqlBuilder("postgres:16-alpine") // same image as docker-compose.yml
            .Build();

    public string ConnectionString => _container.GetConnectionString();

    /// <summary>A fresh context per unit of work - never shared across tests.</summary>
    public OrdersDbContext CreateContext()
    {
        DbContextOptions<OrdersDbContext> options = new DbContextOptionsBuilder<OrdersDbContext>()
            .UseNpgsql(ConnectionString, npgsql => npgsql.EnableRetryOnFailure())
            .UseQueryTrackingBehavior(QueryTrackingBehavior.NoTracking)
            .Options;
        return new OrdersDbContext(options);
    }

    public async ValueTask InitializeAsync()
    {
        await _container.StartAsync(TestContext.Current.CancellationToken);
        await using OrdersDbContext db = CreateContext();
        // The same migration path `dotnet Orders.Api.dll --migrate` runs.
        await db.Database.MigrateAsync(TestContext.Current.CancellationToken);
    }

    public async ValueTask DisposeAsync() => await _container.DisposeAsync();
}
