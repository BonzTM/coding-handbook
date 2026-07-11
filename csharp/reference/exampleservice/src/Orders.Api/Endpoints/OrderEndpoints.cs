using System.Security.Claims;

using Microsoft.AspNetCore.Http.HttpResults;

using Orders.Api.Auth;
using Orders.Api.Contracts;
using Orders.Api.Idempotency;
using Orders.Api.RateLimiting;
using Orders.Api.Telemetry;
using Orders.Core.Orders;

namespace Orders.Api.Endpoints;

/// <summary>
/// The orders endpoint group (csharp/services/http-services.md). Handlers are
/// translation layers: bind → (validation already ran) → call core → map →
/// TypedResults. Domain errors surface as exceptions and are mapped to
/// ProblemDetails in ONE place (DomainExceptionHandler), so the unions here
/// list the statuses each handler produces itself.
/// </summary>
internal static class OrderEndpoints
{
    public static IEndpointRouteBuilder MapOrderEndpoints(this IEndpointRouteBuilder routes)
    {
        ArgumentNullException.ThrowIfNull(routes);

        var group = routes.MapGroup("/orders")
            .WithTags("Orders")
            .RequireRateLimiting(OrdersRateLimitingExtensions.PerClientPolicy)
            .AddEndpointFilter<TenantScopeFilter>();

        group.MapPost("/", CreateOrder)
            .WithName("CreateOrder")
            .RequireAuthorization(OrdersPolicies.Write)
            .AddEndpointFilter<IdempotencyFilter>();

        group.MapGet("/{id:guid}", GetOrder)
            .WithName("GetOrder")
            .RequireAuthorization(OrdersPolicies.Read);

        group.MapGet("/", ListOrders)
            .WithName("ListOrders")
            .RequireAuthorization(OrdersPolicies.Read);

        group.MapPut("/{id:guid}", UpdateOrder)
            .WithName("UpdateOrder")
            .RequireAuthorization(OrdersPolicies.Write);

        group.MapDelete("/{id:guid}", DeleteOrder)
            .WithName("DeleteOrder")
            .RequireAuthorization(OrdersPolicies.Write);

        return routes;
    }

    private static async Task<Created<OrderResponse>> CreateOrder(
        CreateOrderRequest request,
        ClaimsPrincipal user,
        HttpContext http,
        OrderService orders,
        OrdersMetrics metrics,
        AuditLogger audit,
        CancellationToken cancellationToken)
    {
        var tenant = user.RequiredTenant();
        var order = await orders.CreateAsync(
            tenant, request.ExternalReference, request.CustomerId, request.Quantity, cancellationToken);

        metrics.OrderCreated();
        audit.OrderCreated(user.Actor(), tenant.Value, order.Id.ToString(), http.TraceIdentifier);
        return TypedResults.Created($"/orders/{order.Id}", order.ToResponse());
    }

    private static async Task<Ok<OrderResponse>> GetOrder(
        Guid id,
        ClaimsPrincipal user,
        OrderService orders,
        CancellationToken cancellationToken)
    {
        var order = await orders.GetAsync(user.RequiredTenant(), new OrderId(id), cancellationToken);
        return TypedResults.Ok(order.ToResponse());
    }

    private static async Task<Results<Ok<OrderListResponse>, ValidationProblem>> ListOrders(
        int? pageSize,
        string? cursor,
        ClaimsPrincipal user,
        OrderService orders,
        CancellationToken cancellationToken)
    {
        OrderCursor? decoded = null;
        if (!string.IsNullOrEmpty(cursor))
        {
            if (!OrderCursor.TryDecode(cursor, out var parsed))
            {
                return TypedResults.ValidationProblem(new Dictionary<string, string[]>(StringComparer.Ordinal)
                {
                    ["cursor"] = ["The cursor is malformed; pass back the nextCursor value verbatim."],
                });
            }

            decoded = parsed;
        }

        var page = await orders.ListAsync(
            user.RequiredTenant(), new OrderListQuery(pageSize, decoded), cancellationToken);
        return TypedResults.Ok(page.ToResponse());
    }

    private static async Task<Ok<OrderResponse>> UpdateOrder(
        Guid id,
        UpdateOrderRequest request,
        ClaimsPrincipal user,
        HttpContext http,
        OrderService orders,
        AuditLogger audit,
        CancellationToken cancellationToken)
    {
        var tenant = user.RequiredTenant();
        var order = await orders.AmendAsync(
            tenant, new OrderId(id), request.Version, request.Quantity, request.Status, cancellationToken);

        audit.OrderAmended(user.Actor(), tenant.Value, order.Id.ToString(), http.TraceIdentifier);
        return TypedResults.Ok(order.ToResponse());
    }

    private static async Task<NoContent> DeleteOrder(
        Guid id,
        ClaimsPrincipal user,
        HttpContext http,
        OrderService orders,
        AuditLogger audit,
        CancellationToken cancellationToken)
    {
        var tenant = user.RequiredTenant();
        await orders.DeleteAsync(tenant, new OrderId(id), cancellationToken);

        audit.OrderDeleted(user.Actor(), tenant.Value, new OrderId(id).ToString(), http.TraceIdentifier);
        return TypedResults.NoContent();
    }
}
