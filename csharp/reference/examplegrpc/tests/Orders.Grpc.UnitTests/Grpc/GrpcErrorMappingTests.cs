using Google.Rpc;

using Grpc.Core;

using Orders.Grpc.Api.Grpc;
using Orders.Grpc.Core.Identity;
using Orders.Grpc.Core.Orders;

using Xunit;

namespace Orders.Grpc.UnitTests.Grpc;

/// <summary>
/// Table test for the single domain-to-status mapping
/// (csharp/recipes/add-grpc-method.md): every mapped code, the BadRequest
/// detail on validation failures, and null for unexpected exceptions (which
/// the interceptor shields as INTERNAL).
/// </summary>
public sealed class GrpcErrorMappingTests
{
    [Fact]
    public void Map_OrderNotFound_IsNotFound()
    {
        var mapped = GrpcErrorMapping.Map(new OrderNotFoundException(OrderId.New()));

        Assert.NotNull(mapped);
        Assert.Equal(StatusCode.NotFound, mapped.StatusCode);
    }

    [Fact]
    public void Map_DuplicateOrder_IsAlreadyExists()
    {
        var mapped = GrpcErrorMapping.Map(new DuplicateOrderException("ord-1"));

        Assert.NotNull(mapped);
        Assert.Equal(StatusCode.AlreadyExists, mapped.StatusCode);
    }

    [Fact]
    public void Map_PermissionDenied_IsPermissionDenied()
    {
        var mapped = GrpcErrorMapping.Map(new PermissionDeniedException("subject", OrderRoles.Writer));

        Assert.NotNull(mapped);
        Assert.Equal(StatusCode.PermissionDenied, mapped.StatusCode);
    }

    [Fact]
    public void Map_InvalidCursor_IsInvalidArgument()
    {
        var mapped = GrpcErrorMapping.Map(new InvalidCursorException());

        Assert.NotNull(mapped);
        Assert.Equal(StatusCode.InvalidArgument, mapped.StatusCode);
    }

    [Fact]
    public void Map_Validation_IsInvalidArgumentWithBadRequestDetail()
    {
        var exception = new OrderValidationException(
        [
            new FieldViolation("external_reference", "external_reference must not be empty"),
            new FieldViolation("quantity", "quantity must be between 1 and 1000"),
        ]);

        var mapped = GrpcErrorMapping.Map(exception);

        Assert.NotNull(mapped);
        Assert.Equal(StatusCode.InvalidArgument, mapped.StatusCode);

        // Decode exactly the way a client does: GetRpcStatus() then the typed
        // detail - never by scraping the message string
        // (csharp/services/grpc-services.md, Verification And Proof).
        var rpcStatus = mapped.GetRpcStatus();
        Assert.NotNull(rpcStatus);
        var badRequest = rpcStatus.GetDetail<BadRequest>();
        Assert.NotNull(badRequest);
        Assert.Collection(
            badRequest.FieldViolations,
            v => Assert.Equal(("external_reference", "external_reference must not be empty"), (v.Field, v.Description)),
            v => Assert.Equal(("quantity", "quantity must be between 1 and 1000"), (v.Field, v.Description)));
    }

    [Fact]
    public void Map_UnexpectedException_ReturnsNullSoTheInterceptorShieldsIt()
    {
        Assert.Null(GrpcErrorMapping.Map(new InvalidOperationException("secret internals")));
    }
}
