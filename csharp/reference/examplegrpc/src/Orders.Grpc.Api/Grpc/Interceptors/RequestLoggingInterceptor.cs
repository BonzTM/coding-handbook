using Grpc.Core;
using Grpc.Core.Interceptors;

namespace Orders.Grpc.Api.Grpc.Interceptors;

/// <summary>
/// OUTERMOST interceptor (registration order = execution order,
/// csharp/services/grpc-services.md): one access-log line per RPC with the
/// full method, the status code the client actually saw (exception mapping
/// runs inside it, so only RpcExceptions pass through here), the duration,
/// and a request id. Expected client errors log at Information; server-side
/// failures (INTERNAL/UNKNOWN/DATA_LOSS) at Error - mirroring the Go
/// reference's access-log interceptor. Tracing is NOT an interceptor: the
/// OpenTelemetry ASP.NET Core instrumentation already covers gRPC calls.
/// </summary>
internal sealed partial class RequestLoggingInterceptor(
    ILogger<RequestLoggingInterceptor> logger,
    TimeProvider time) : Interceptor
{
    private const string RequestIdHeader = "x-request-id";
    private const int MaxRequestIdLength = 64;

    public override async Task<TResponse> UnaryServerHandler<TRequest, TResponse>(
        TRequest request,
        ServerCallContext context,
        UnaryServerMethod<TRequest, TResponse> continuation)
    {
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(continuation);
        long start = time.GetTimestamp();
        try
        {
            var response = await continuation(request, context).ConfigureAwait(false);
            Log(context, StatusCode.OK, start);
            return response;
        }
        catch (RpcException rpcException)
        {
            Log(context, rpcException.StatusCode, start);
            throw;
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
        long start = time.GetTimestamp();
        try
        {
            await continuation(request, responseStream, context).ConfigureAwait(false);
            Log(context, StatusCode.OK, start);
        }
        catch (RpcException rpcException)
        {
            Log(context, rpcException.StatusCode, start);
            throw;
        }
    }

    private void Log(ServerCallContext context, StatusCode statusCode, long start)
    {
        double durationMs = time.GetElapsedTime(start).TotalMilliseconds;
        string requestId = ResolveRequestId(context);
        if (statusCode is StatusCode.Internal or StatusCode.Unknown or StatusCode.DataLoss)
        {
            LogRpcFailed(context.Method, statusCode, durationMs, requestId);
        }
        else
        {
            LogRpc(context.Method, statusCode, durationMs, requestId);
        }
    }

    /// <summary>
    /// Adopts a WELL-FORMED inbound x-request-id (bounded length and charset -
    /// an untrusted header never lands in logs raw, same rule as the keystone's
    /// request-id middleware); otherwise falls back to the connection's trace
    /// identifier.
    /// </summary>
    private static string ResolveRequestId(ServerCallContext context)
    {
        string? candidate = context.RequestHeaders.GetValue(RequestIdHeader);
        if (candidate is not null && IsWellFormedRequestId(candidate))
        {
            return candidate;
        }

        return context.GetHttpContext().TraceIdentifier;
    }

    private static bool IsWellFormedRequestId(string candidate)
        => candidate.Length is > 0 and <= MaxRequestIdLength
            && candidate.All(static c => char.IsAsciiLetterOrDigit(c) || c is '-' or '_' or '.');

    [LoggerMessage(Level = LogLevel.Information,
        Message = "rpc {Method} {StatusCode} {DurationMs:F1}ms request_id={RequestId}")]
    private partial void LogRpc(string method, StatusCode statusCode, double durationMs, string requestId);

    [LoggerMessage(Level = LogLevel.Error,
        Message = "rpc {Method} {StatusCode} {DurationMs:F1}ms request_id={RequestId}")]
    private partial void LogRpcFailed(string method, StatusCode statusCode, double durationMs, string requestId);
}
