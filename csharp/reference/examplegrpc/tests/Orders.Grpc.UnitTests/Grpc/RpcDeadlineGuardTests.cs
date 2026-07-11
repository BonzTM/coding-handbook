using Microsoft.Extensions.Time.Testing;

using Orders.Grpc.Api.Grpc;

using Xunit;

namespace Orders.Grpc.UnitTests.Grpc;

/// <summary>
/// The server's self-protection contract
/// (csharp/services/grpc-services.md): a client deadline passes through
/// untouched (the framework enforces it), and a MISSING client deadline gets
/// the configured ceiling - deterministic via FakeTimeProvider.
/// </summary>
public sealed class RpcDeadlineGuardTests
{
    private static readonly TimeSpan _ceiling = TimeSpan.FromSeconds(30);

    [Fact]
    public void Bind_WhenClientSetADeadline_PassesTheCallTokenThrough()
    {
        var time = new FakeTimeProvider();
        using var call = new CancellationTokenSource();

        using var guard = RpcDeadlineGuard.Bind(
            clientSetDeadline: true, _ceiling, time, call.Token);

        Assert.Equal(call.Token, guard.Token);
    }

    [Fact]
    public void Bind_WithoutClientDeadline_CancelsWhenTheCeilingElapses()
    {
        var time = new FakeTimeProvider();
        using var call = new CancellationTokenSource();
        using var guard = RpcDeadlineGuard.Bind(
            clientSetDeadline: false, _ceiling, time, call.Token);

        Assert.False(guard.Token.IsCancellationRequested);

        time.Advance(_ceiling + TimeSpan.FromSeconds(1));

        Assert.True(guard.Token.IsCancellationRequested);
        // The CLIENT did not cancel - that distinction is how the exception
        // interceptor maps this to DEADLINE_EXCEEDED instead of CANCELLED.
        Assert.False(call.Token.IsCancellationRequested);
    }

    [Fact]
    public void Bind_WithoutClientDeadline_ClientCancellationStillPropagates()
    {
        var time = new FakeTimeProvider();
        using var call = new CancellationTokenSource();
        using var guard = RpcDeadlineGuard.Bind(
            clientSetDeadline: false, _ceiling, time, call.Token);

        call.Cancel();

        Assert.True(guard.Token.IsCancellationRequested);
    }
}
