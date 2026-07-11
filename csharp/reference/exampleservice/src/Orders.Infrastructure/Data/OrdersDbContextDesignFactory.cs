using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Design;

namespace Orders.Infrastructure.Data;

/// <summary>
/// Design-time factory for `dotnet ef migrations add`. Scaffolding only needs
/// the model, never a live database, so the connection string is a placeholder;
/// it keeps design-time tooling from booting the real host (and its fail-fast
/// config validation). Runtime wiring lives in OrdersInfrastructureExtensions.
/// </summary>
internal sealed class OrdersDbContextDesignFactory : IDesignTimeDbContextFactory<OrdersDbContext>
{
    public OrdersDbContext CreateDbContext(string[] args)
    {
        var options = new DbContextOptionsBuilder<OrdersDbContext>()
            .UseNpgsql("Host=localhost;Database=orders_design;Username=design;Password=design")
            .Options;
        return new OrdersDbContext(options);
    }
}
