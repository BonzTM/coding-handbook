using Grpc.Core;
using Grpc.Core.Interceptors;

using Orders.Grpc.Api.Auth;
using Orders.Grpc.Core.Identity;

namespace Orders.Grpc.Api.Grpc.Interceptors;

/// <summary>
/// SECOND in the chain (after logging, before exception mapping,
/// csharp/services/grpc-services.md): resolves the bearer token to a
/// <see cref="CallerPrincipal"/> via the <see cref="IAuthenticator"/> seam and
/// rejects before any handler work. Registered globally so a future service
/// is authenticated BY DEFAULT (fail closed); the health and reflection
/// services are the two explicit exemptions - the kubelet and grpcurl
/// discovery carry no token, exactly as the keystone's probes are the only
/// AllowAnonymous endpoints. The 16 (UNAUTHENTICATED) message never says
/// which check failed; the reason is logged server-side only.
/// </summary>
internal sealed partial class AuthInterceptor(
    IAuthenticator authenticator,
    ILogger<AuthInterceptor> logger) : Interceptor
{
    private const string AuthorizationHeader = "authorization";
    private const string BearerPrefix = "Bearer ";

    /// <summary>Unauthenticated service prefixes - the full method is "/package.Service/Method".</summary>
    private static readonly string[] _exemptMethodPrefixes =
    [
        "/grpc.health.v1.Health/",
        "/grpc.reflection.",
    ];

    public override async Task<TResponse> UnaryServerHandler<TRequest, TResponse>(
        TRequest request,
        ServerCallContext context,
        UnaryServerMethod<TRequest, TResponse> continuation)
    {
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(continuation);
        if (!IsExempt(context.Method))
        {
            context.SetPrincipal(await AuthenticateAsync(context).ConfigureAwait(false));
        }

        return await continuation(request, context).ConfigureAwait(false);
    }

    public override async Task ServerStreamingServerHandler<TRequest, TResponse>(
        TRequest request,
        IServerStreamWriter<TResponse> responseStream,
        ServerCallContext context,
        ServerStreamingServerMethod<TRequest, TResponse> continuation)
    {
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(continuation);
        if (!IsExempt(context.Method))
        {
            context.SetPrincipal(await AuthenticateAsync(context).ConfigureAwait(false));
        }

        await continuation(request, responseStream, context).ConfigureAwait(false);
    }

    private async Task<CallerPrincipal> AuthenticateAsync(ServerCallContext context)
    {
        string? token = BearerToken(context.RequestHeaders.GetValue(AuthorizationHeader));
        var principal = await authenticator
            .AuthenticateAsync(token, context.CancellationToken)
            .ConfigureAwait(false);
        if (principal is null)
        {
            LogAuthenticationFailed(context.Method, token is null ? "missing token" : "invalid token");
            throw new RpcException(new Status(StatusCode.Unauthenticated, "invalid or missing bearer token"));
        }

        return principal;
    }

    private static bool IsExempt(string fullMethod)
        => _exemptMethodPrefixes.Any(prefix => fullMethod.StartsWith(prefix, StringComparison.Ordinal));

    private static string? BearerToken(string? authorization)
    {
        if (string.IsNullOrEmpty(authorization))
        {
            return null;
        }

        return authorization.StartsWith(BearerPrefix, StringComparison.OrdinalIgnoreCase)
            ? authorization[BearerPrefix.Length..]
            : authorization;
    }

    // The reason stays server-side; the wire message is deliberately generic.
    [LoggerMessage(Level = LogLevel.Warning,
        Message = "authentication failed for {Method}: {Reason}")]
    private partial void LogAuthenticationFailed(string method, string reason);
}
