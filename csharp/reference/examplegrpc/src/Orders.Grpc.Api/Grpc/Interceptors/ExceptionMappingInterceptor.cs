using Grpc.Core;
using Grpc.Core.Interceptors;

namespace Orders.Grpc.Api.Grpc.Interceptors;

/// <summary>
/// INNERMOST interceptor (csharp/services/grpc-services.md): everything above
/// it observes a proper <see cref="RpcException"/>, never a raw domain
/// exception. Known domain exceptions map through GrpcErrorMapping (validation
/// failures carry a google.rpc.BadRequest detail); cancellation maps to
/// CANCELLED or DEADLINE_EXCEEDED; anything unexpected is logged ONCE with its
/// stack and shielded behind a generic INTERNAL so internals never leak
/// (csharp/foundations/errors-and-logging.md).
/// </summary>
internal sealed partial class ExceptionMappingInterceptor(
    ILogger<ExceptionMappingInterceptor> logger) : Interceptor
{
    public override async Task<TResponse> UnaryServerHandler<TRequest, TResponse>(
        TRequest request,
        ServerCallContext context,
        UnaryServerMethod<TRequest, TResponse> continuation)
    {
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(continuation);
        try
        {
            return await continuation(request, context).ConfigureAwait(false);
        }
        catch (Exception exception) when (exception is not RpcException)
        {
            throw MapToRpcException(exception, context);
        }
    }

    public override async Task ServerStreamingServerHandler<TRequest, TResponse>(
        TRequest request,
        IServerStreamWriter<TResponse> responseStream,
        ServerCallContext context,
        ServerStreamingServerMethod<TRequest, TResponse> continuation)
    {
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(continuation);
        try
        {
            await continuation(request, responseStream, context).ConfigureAwait(false);
        }
        catch (Exception exception) when (exception is not RpcException)
        {
            throw MapToRpcException(exception, context);
        }
    }

    private RpcException MapToRpcException(Exception exception, ServerCallContext context)
    {
        if (exception is OperationCanceledException)
        {
            // Client cancellation (or an expired client deadline - the
            // framework cancels context.CancellationToken for both) maps to
            // CANCELLED; a cancellation the CLIENT did not cause can only be
            // the server-side deadline ceiling (RpcDeadlineGuard).
            return context.CancellationToken.IsCancellationRequested
                ? new RpcException(new Status(StatusCode.Cancelled, "call cancelled"))
                : new RpcException(new Status(
                    StatusCode.DeadlineExceeded, "server-side maximum RPC duration exceeded"));
        }

        var mapped = GrpcErrorMapping.Map(exception);
        if (mapped is not null)
        {
            return mapped;
        }

        // Unexpected: log once with the stack, shield the client.
        LogUnexpected(exception, context.Method);
        return new RpcException(new Status(StatusCode.Internal, "internal error"));
    }

    [LoggerMessage(Level = LogLevel.Error,
        Message = "unexpected exception in {Method}")]
    private partial void LogUnexpected(Exception exception, string method);
}
