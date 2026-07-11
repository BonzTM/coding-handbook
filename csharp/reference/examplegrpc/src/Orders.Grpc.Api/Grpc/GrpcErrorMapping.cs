using Google.Protobuf.WellKnownTypes;
using Google.Rpc;

using Grpc.Core;

using Orders.Grpc.Core.Identity;
using Orders.Grpc.Core.Orders;

using Status = Google.Rpc.Status;

namespace Orders.Grpc.Api.Grpc;

/// <summary>
/// The ONE place domain errors become gRPC statuses - mapped once, at the
/// transport boundary (csharp/services/grpc-services.md; the HTTP twin is the
/// keystone's DomainExceptionHandler). Field-level validation failures carry a
/// google.rpc.BadRequest detail so clients map errors back to fields instead
/// of scraping a message string.
/// </summary>
internal static class GrpcErrorMapping
{
    /// <summary>
    /// Maps a known domain exception to the <see cref="RpcException"/> the
    /// client should see, or returns null for an unexpected exception - the
    /// caller (the exception-mapping interceptor) logs those once and shields
    /// them behind a generic INTERNAL.
    /// </summary>
    public static RpcException? Map(Exception exception)
    {
        ArgumentNullException.ThrowIfNull(exception);
        return exception switch
        {
            OrderValidationException validation => ToRpcException(validation),
            InvalidCursorException cursor => new RpcException(
                new global::Grpc.Core.Status(StatusCode.InvalidArgument, cursor.Message)),
            OrderNotFoundException notFound => new RpcException(
                new global::Grpc.Core.Status(StatusCode.NotFound, notFound.Message)),
            DuplicateOrderException duplicate => new RpcException(
                new global::Grpc.Core.Status(StatusCode.AlreadyExists, duplicate.Message)),
            PermissionDeniedException denied => new RpcException(
                new global::Grpc.Core.Status(StatusCode.PermissionDenied, denied.Message)),
            _ => null,
        };
    }

    /// <summary>
    /// INVALID_ARGUMENT with a google.rpc.BadRequest detail listing every
    /// {field, description} violation. Core carries the violations
    /// structurally; the proto detail is built here so Core never references
    /// Google.Rpc (csharp/services/grpc-services.md, Error Details).
    /// </summary>
    private static RpcException ToRpcException(OrderValidationException exception)
    {
        var badRequest = new BadRequest();
        foreach (var violation in exception.Violations)
        {
            badRequest.FieldViolations.Add(new BadRequest.Types.FieldViolation
            {
                Field = violation.Field,
                Description = violation.Description,
            });
        }

        return new Status
        {
            Code = (int)Code.InvalidArgument,
            Message = exception.Message,
            Details = { Any.Pack(badRequest) },
        }.ToRpcException();
    }
}
