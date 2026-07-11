using Grpc.Core;

namespace Orders.Grpc.Api.Grpc;

/// <summary>
/// Server-side self-protection against unbounded work
/// (csharp/services/grpc-services.md): when the client set a deadline, the
/// framework already enforces it through <c>context.CancellationToken</c> and
/// this guard is a pass-through. When the client set NONE, the guard imposes
/// the configured ceiling so a single RPC cannot run forever - the C# twin of
/// the Go reference's deadline-guard interceptor. Always dispose (it owns the
/// timer); always pass <see cref="Token"/> into Core and I/O calls.
/// </summary>
internal sealed class RpcDeadlineGuard : IDisposable
{
    private readonly CancellationTokenSource? _ceiling;
    private readonly CancellationTokenSource? _linked;

    private RpcDeadlineGuard(CancellationTokenSource? ceiling, CancellationTokenSource? linked, CancellationToken token)
    {
        Token = token;
        _ceiling = ceiling;
        _linked = linked;
    }

    /// <summary>The token to pass into Core: client cancellation, client deadline, or the server ceiling - whichever fires first.</summary>
    public CancellationToken Token { get; }

    public static RpcDeadlineGuard Bind(ServerCallContext context, TimeSpan maxRpcDuration, TimeProvider time)
    {
        ArgumentNullException.ThrowIfNull(context);
        // ServerCallContext.Deadline is UTC; MaxValue means "no client deadline".
        bool clientSetDeadline = context.Deadline != DateTime.MaxValue;
        return Bind(clientSetDeadline, maxRpcDuration, time, context.CancellationToken);
    }

    /// <summary>Transport-free core so tests can drive it with a FakeTimeProvider.</summary>
    internal static RpcDeadlineGuard Bind(
        bool clientSetDeadline,
        TimeSpan maxRpcDuration,
        TimeProvider time,
        CancellationToken callToken)
    {
        ArgumentOutOfRangeException.ThrowIfLessThanOrEqual(maxRpcDuration, TimeSpan.Zero);
        ArgumentNullException.ThrowIfNull(time);
        if (clientSetDeadline)
        {
            return new RpcDeadlineGuard(ceiling: null, linked: null, callToken);
        }

        var ceiling = new CancellationTokenSource(maxRpcDuration, time);
        var linked = CancellationTokenSource.CreateLinkedTokenSource(callToken, ceiling.Token);
        return new RpcDeadlineGuard(ceiling, linked, linked.Token);
    }

    public void Dispose()
    {
        _linked?.Dispose();
        _ceiling?.Dispose();
    }
}
