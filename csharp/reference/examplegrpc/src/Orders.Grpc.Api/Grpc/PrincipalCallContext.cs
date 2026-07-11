using Grpc.Core;

using Orders.Grpc.Core.Identity;

namespace Orders.Grpc.Api.Grpc;

/// <summary>
/// Bridges the authenticated <see cref="CallerPrincipal"/> from the auth
/// interceptor to the service method through <see cref="ServerCallContext"/>'s
/// per-call UserState. The service passes it into Core explicitly - Core never
/// reads call state (csharp/foundations/shared-constructs.md).
/// </summary>
internal static class PrincipalCallContext
{
    private const string Key = "Orders.Grpc.Principal";

    public static void SetPrincipal(this ServerCallContext context, CallerPrincipal principal)
    {
        ArgumentNullException.ThrowIfNull(context);
        ArgumentNullException.ThrowIfNull(principal);
        context.UserState[Key] = principal;
    }

    /// <summary>
    /// Returns the principal the auth interceptor resolved. A missing
    /// principal means the interceptor chain was not applied - a wiring bug,
    /// surfaced as INTERNAL rather than an unauthenticated bypass.
    /// </summary>
    public static CallerPrincipal GetPrincipal(this ServerCallContext context)
    {
        ArgumentNullException.ThrowIfNull(context);
        if (context.UserState.TryGetValue(Key, out object? value) && value is CallerPrincipal principal)
        {
            return principal;
        }

        throw new RpcException(new Status(StatusCode.Internal, "internal error"));
    }
}
