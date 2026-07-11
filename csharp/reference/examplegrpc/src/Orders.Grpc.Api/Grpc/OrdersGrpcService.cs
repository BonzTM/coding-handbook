using Google.Protobuf.WellKnownTypes;

using Grpc.Core;

using Microsoft.Extensions.Options;

using Orders.Grpc.Api.Telemetry;
using Orders.Grpc.Api.V1;
using Orders.Grpc.Core.Orders;

using DomainOrder = Orders.Grpc.Core.Orders.Order;
using WireOrder = Orders.Grpc.Api.V1.Order;

namespace Orders.Grpc.Api.Grpc;

/// <summary>
/// The gRPC transport adapter over Orders.Grpc.Core - thin by contract
/// (csharp/services/grpc-services.md): each method decodes the request, calls
/// exactly one Core method with a cancellation token that combines client
/// cancellation, the client deadline, and the server's own ceiling
/// (RpcDeadlineGuard), and builds the response. Domain errors are raised as
/// typed exceptions in Core and mapped once by the exception-mapping
/// interceptor; no mapping and no domain logic here.
/// </summary>
internal sealed class OrdersGrpcService(
    OrderService orders,
    OrdersGrpcMetrics metrics,
    IOptions<ServerOptions> serverOptions,
    TimeProvider time) : OrdersService.OrdersServiceBase
{
    public override async Task<CreateOrderResponse> CreateOrder(
        CreateOrderRequest request, ServerCallContext context)
    {
        ArgumentNullException.ThrowIfNull(request);
        ArgumentNullException.ThrowIfNull(context);
        var principal = context.GetPrincipal();
        using var guard = RpcDeadlineGuard.Bind(context, serverOptions.Value.MaxRpcDuration, time);
        var order = await orders.CreateAsync(
            principal, request.ExternalReference, request.CustomerId, request.Quantity, guard.Token)
            .ConfigureAwait(false);
        metrics.OrderCreated();
        return new CreateOrderResponse { Order = ToWire(order) };
    }

    public override async Task<GetOrderResponse> GetOrder(
        GetOrderRequest request, ServerCallContext context)
    {
        ArgumentNullException.ThrowIfNull(request);
        ArgumentNullException.ThrowIfNull(context);
        if (!OrderId.TryParse(request.Id, out var id))
        {
            // Transport-shape validation (the id is a UUID on the wire) stays at
            // the boundary, but the violation is carried structurally so the
            // interceptor renders the same google.rpc.BadRequest detail.
            throw new OrderValidationException([new FieldViolation("id", "id must be a UUID")]);
        }

        var principal = context.GetPrincipal();
        using var guard = RpcDeadlineGuard.Bind(context, serverOptions.Value.MaxRpcDuration, time);
        var order = await orders.GetAsync(principal, id, guard.Token).ConfigureAwait(false);
        return new GetOrderResponse { Order = ToWire(order) };
    }

    public override async Task<ListOrdersResponse> ListOrders(
        ListOrdersRequest request, ServerCallContext context)
    {
        ArgumentNullException.ThrowIfNull(request);
        ArgumentNullException.ThrowIfNull(context);
        OrderCursor? cursor = null;
        if (!string.IsNullOrEmpty(request.PageToken))
        {
            if (!OrderCursor.TryDecode(request.PageToken, out var decoded))
            {
                throw new InvalidCursorException();
            }

            cursor = decoded;
        }

        var principal = context.GetPrincipal();
        using var guard = RpcDeadlineGuard.Bind(context, serverOptions.Value.MaxRpcDuration, time);
        var page = await orders.ListAsync(
            principal, new OrderListQuery(request.PageSize, cursor), guard.Token)
            .ConfigureAwait(false);

        var response = new ListOrdersResponse { NextPageToken = page.NextCursor ?? string.Empty };
        foreach (var order in page.Items)
        {
            response.Orders.Add(ToWire(order));
        }

        return response;
    }

    public override async Task StreamOrders(
        StreamOrdersRequest request,
        IServerStreamWriter<StreamOrdersResponse> responseStream,
        ServerCallContext context)
    {
        ArgumentNullException.ThrowIfNull(responseStream);
        ArgumentNullException.ThrowIfNull(context);
        var principal = context.GetPrincipal();
        using var guard = RpcDeadlineGuard.Bind(context, serverOptions.Value.MaxRpcDuration, time);
        await foreach (var order in orders.StreamAsync(principal, guard.Token).ConfigureAwait(false))
        {
            await responseStream
                .WriteAsync(new StreamOrdersResponse { Order = ToWire(order) }, guard.Token)
                .ConfigureAwait(false);
            metrics.OrderStreamed();
        }
    }

    /// <summary>Renders a domain order as its wire DTO. The tenant id never goes on the wire.</summary>
    private static WireOrder ToWire(DomainOrder order) => new()
    {
        Id = order.Id.ToString(),
        ExternalReference = order.ExternalReference,
        CustomerId = order.CustomerId,
        Quantity = order.Quantity,
        CreatedAt = Timestamp.FromDateTimeOffset(order.CreatedAt),
    };
}
