using Orders.Worker.Core.Events;
using Orders.Worker.Core.Orders;

namespace Orders.Worker.UnitTests.Fakes;

/// <summary>Hand-rolled processor whose per-call outcome is scripted: result i
/// applies to call i, the last result repeats (csharp/quality/testing.md -
/// fakes over mocks).</summary>
internal sealed class ScriptedProcessor(params Exception?[] results) : IOrderEventProcessor
{
    private readonly Lock _gate = new();
    private readonly List<OrderEvent> _seen = [];
    private int _calls;

    public int Calls
    {
        get
        {
            lock (_gate)
            {
                return _calls;
            }
        }
    }

    public IReadOnlyList<OrderEvent> Seen
    {
        get
        {
            lock (_gate)
            {
                return [.. _seen];
            }
        }
    }

    public Task ProcessAsync(OrderEvent orderEvent, CancellationToken cancellationToken)
    {
        Exception? result;
        lock (_gate)
        {
            _seen.Add(orderEvent);
            int index = Math.Min(_calls, results.Length - 1);
            _calls++;
            result = results.Length == 0 ? null : results[index];
        }

        return result is null ? Task.CompletedTask : Task.FromException(result);
    }
}

/// <summary>Processor delegating to a test-supplied callback, for outcomes a
/// script cannot express (cancel-mid-call, throw the caller's token, ...).</summary>
internal sealed class CallbackProcessor(Func<OrderEvent, CancellationToken, Task> callback) : IOrderEventProcessor
{
    private int _calls;

    public int Calls => Volatile.Read(ref _calls);

    public Task ProcessAsync(OrderEvent orderEvent, CancellationToken cancellationToken)
    {
        Interlocked.Increment(ref _calls);
        return callback(orderEvent, cancellationToken);
    }
}

/// <summary>
/// Blocks the first Process call until released, so a test can hold a message
/// in-flight, trigger the host drain, and assert the in-flight message still
/// finishes and settles - no message lost. It deliberately IGNORES the
/// cancellation token: it simulates a side effect that cannot be abandoned
/// halfway, which is exactly what the drain budget exists for.
/// </summary>
internal sealed class BlockingProcessor : IOrderEventProcessor
{
    private readonly TaskCompletionSource _started = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private readonly TaskCompletionSource _release = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private int _processed;

    /// <summary>Completes when the first message is being processed (in-flight).</summary>
    public Task Started => _started.Task;

    public int Processed => Volatile.Read(ref _processed);

    public void Release() => _release.TrySetResult();

    public async Task ProcessAsync(OrderEvent orderEvent, CancellationToken cancellationToken)
    {
        _started.TrySetResult();
        await _release.Task.ConfigureAwait(false);
        Interlocked.Increment(ref _processed);
    }
}
