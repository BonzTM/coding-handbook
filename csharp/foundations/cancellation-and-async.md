# Cancellation and Async

Cancellation propagation, async discipline, and task ownership rules for .NET code that behaves correctly under load and on shutdown.

## Default Approach

- Accept `CancellationToken cancellationToken` as the **last** parameter of every method that performs I/O, waits on remote work, or coordinates long-lived tasks (CA1068 enforces the position).
- Forward the token through every async call in the chain — never substitute `CancellationToken.None` or `default` deep in the stack (CA2016 flags unforwarded tokens).
- In minimal API handlers, declare a `CancellationToken` parameter; ASP.NET Core binds it to `HttpContext.RequestAborted`, so a disconnected client cancels the whole call tree.
- In workers, the `stoppingToken` passed to `BackgroundService.ExecuteAsync` is the root token; every call inside the loop takes it.
- Async end to end: an async signature means the entire path is async. Sync-over-async — `.Result`, `.Wait()`, `.GetAwaiter().GetResult()`, `Task.Run(...).Result` — is forbidden everywhere. It deadlocks under legacy contexts and starves the thread pool under load (see below).
- `async void` is forbidden — exceptions in an `async void` method crash the process with no caller to observe them. The only sanctioned use is UI event handlers, which do not exist in this stack.

### Cancellation Semantics

- Cooperative cancellation: check `cancellationToken.ThrowIfCancellationRequested()` at the top of loop bodies and between expensive non-cancellable steps.
- Give every external call a timeout budget in addition to the caller's token: link them, never replace one with the other.

```csharp
using var cts = CancellationTokenSource.CreateLinkedTokenSource(cancellationToken);
cts.CancelAfter(TimeSpan.FromSeconds(5));
var quote = await pricingClient.GetQuoteAsync(sku, cts.Token);
```

- Always dispose a `CancellationTokenSource` you create (`using`), or its timer leaks.
- `OperationCanceledException` is the expected outcome of cancellation, not an error. Let it propagate to the boundary; ASP.NET Core handles aborted requests, and the host tolerates it from `ExecuteAsync` during shutdown. Never log a cancellation caused by the caller's own token as `Error` — see [errors-and-logging.md](errors-and-logging.md).
- A must-complete write (e.g. committing an outbox row after the side effect happened) may deliberately run on `CancellationToken.None` for that one statement — scope it to the statement and comment why. "The whole method ignores the token" is never acceptable.

### ConfigureAwait Policy

- In this stack's services (Api, Infrastructure, workers under ASP.NET Core / Generic Host) there is **no `SynchronizationContext`** — continuations already resume on the thread pool, so `ConfigureAwait(false)` changes nothing. Do not sprinkle it through application code; it is noise.
- In reusable libraries — code published as a NuGet package or intended to run under UI frameworks or legacy ASP.NET — use `ConfigureAwait(false)` on every await and enforce it by enabling CA2007 as an error **in that project only** (via `.editorconfig` scoped to the library).
- `ConfigureAwait(false)` is never the fix for a deadlock. The deadlock's cause is sync-over-async; remove the blocking call.

### Primitive Selection

| Use this | When | Avoid when |
|---|---|---|
| `Channel<T>` | producer/consumer pipelines, fan-in/fan-out, handing work to a supervisor | protecting shared mutable state is the main job |
| `lock` over a private `Lock` field | short, CPU-only critical sections on in-memory state | anything inside awaits or does I/O (`await` in a `lock` body is a compile error) |
| `SemaphoreSlim(1, 1)` + `WaitAsync` | mutual exclusion around async work; bounding concurrent calls to a dependency | a channel models the workflow better |
| `Interlocked` / `Volatile` | single-word counters or flags with proven contention | the state is richer than a single word |
| `Parallel.ForEachAsync` | bounded parallel processing of a batch or stream | fire-and-forget, or per-item ordering matters |
| `Task.WhenAll` | a small, fixed set of sibling tasks that succeed or fail together | unbounded fan-out sized by user input |

`SemaphoreSlim` release is always in `finally`, and the wait takes the caller's token:

```csharp
await _gate.WaitAsync(cancellationToken);
try
{
    await RefreshCacheAsync(cancellationToken);
}
finally
{
    _gate.Release();
}
```

### Task Ownership Rules

- Every task needs an owner that awaits it: the request, a `Task.WhenAll` group, a channel consumer, or a hosted service.
- Every long-lived loop needs a stop condition: the stopping token, a completed channel reader, or a finite bound.
- Fire-and-forget (`_ = DoWorkAsync()`) is forbidden: exceptions vanish, cancellation never reaches it, and shutdown cannot drain it. When work must outlive the request, enqueue it to a bounded `Channel<T>` consumed by a `BackgroundService` — the supervisor. See [../recipes/add-background-worker.md](../recipes/add-background-worker.md).
- A `BackgroundService` whose `ExecuteAsync` throws stops the host by default (`BackgroundServiceExceptionBehavior.StopHost`) — a crashed worker must be loud, not silently absent. Keep it that way.
- `ExecuteAsync` must reach its first `await` quickly; a long synchronous prologue delays host startup because hosted services start sequentially. `await Task.Yield()` first if setup is heavy.

### Channels For Producer/Consumer

Bounded by default — an unbounded channel is an unbounded queue is an OOM. Choose `FullMode` deliberately: `Wait` applies backpressure to producers; `DropOldest`/`DropWrite` shed load for lossy telemetry-style streams.

```csharp
public sealed class EmailQueue
{
    private readonly Channel<ConfirmationEmail> _channel =
        Channel.CreateBounded<ConfirmationEmail>(new BoundedChannelOptions(capacity: 1_000)
        {
            FullMode = BoundedChannelFullMode.Wait,
            SingleReader = true,
        });

    public ValueTask EnqueueAsync(ConfirmationEmail email, CancellationToken cancellationToken) =>
        _channel.Writer.WriteAsync(email, cancellationToken);

    public IAsyncEnumerable<ConfirmationEmail> ReadAllAsync(CancellationToken cancellationToken) =>
        _channel.Reader.ReadAllAsync(cancellationToken);
}
```

The supervisor consumes it, logs failures once, and keeps going:

```csharp
public sealed class EmailDispatcher(EmailQueue queue, IEmailSender sender, ILogger<EmailDispatcher> logger)
    : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        await foreach (var email in queue.ReadAllAsync(stoppingToken))
        {
            try
            {
                await sender.SendAsync(email, stoppingToken);
            }
            catch (Exception ex) when (ex is not OperationCanceledException)
            {
                DispatcherLog.SendFailed(logger, email.OrderId, ex);
            }
        }
    }
}

internal static partial class DispatcherLog
{
    [LoggerMessage(Level = LogLevel.Error, Message = "Sending confirmation for order {OrderId} failed; continuing")]
    public static partial void SendFailed(ILogger logger, Guid orderId, Exception exception);
}
```

In-memory channels lose their contents on crash. Work that must survive a restart goes through the outbox instead — see [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).

### Bounded Parallelism And Streams

- `Parallel.ForEachAsync` for parallel batch work: set `MaxDegreeOfParallelism` explicitly (a deliberate bound, not the incidental processor count) and pass the token through `ParallelOptions` **and** to the per-item body via its `ct` parameter.

```csharp
var options = new ParallelOptions { MaxDegreeOfParallelism = 8, CancellationToken = cancellationToken };
await Parallel.ForEachAsync(orderIds, options, async (id, ct) =>
    await reconciler.ReconcileAsync(id, ct));
```

- Never `Task.WhenAll` over a collection whose size the caller controls — that is unbounded fan-out against your own dependencies.
- Stream large result sets with `IAsyncEnumerable<T>` instead of materializing lists. Async iterators annotate the token with `[EnumeratorCancellation]`; consumers attach theirs with `WithCancellation`:

```csharp
public async IAsyncEnumerable<Order> StreamOpenOrdersAsync(
    [EnumeratorCancellation] CancellationToken cancellationToken = default)
{
    await foreach (var order in _db.Orders.Where(o => o.Status == OrderStatus.Open)
        .AsAsyncEnumerable().WithCancellation(cancellationToken))
    {
        yield return order;
    }
}
```

### Thread-Pool Starvation

The failure mode that replaces Go's "goroutine leak" headline: a blocked thread-pool thread (sync-over-async, long `lock`, blocking I/O) holds a worker that queued continuations need. Under load, every thread ends up blocked waiting for continuations that have no thread to run on. Symptoms: latency spikes with low CPU, thread count climbing by ~1–2/second as the pool injects threads, requests timing out in bursts. The fix is never "more threads"; it is removing the blocking call. Detect it live with `dotnet-counters monitor` (`threadpool-queue-length`, `threadpool-thread-count`) — see [../operations/operability.md](../operations/operability.md).

### Graceful Shutdown And Draining

SIGTERM (or Ctrl+C) triggers `IHostApplicationLifetime.ApplicationStopping`, and the host runs an ordered, bounded drain. The canonical wiring is [templates/program-main.cs.txt](../templates/program-main.cs.txt); the sequence is:

1. **Readiness flips to unhealthy.** Register a readiness check that fails once `ApplicationStopping` has fired, so the orchestrator stops routing new traffic while in-flight requests keep flowing. Liveness must keep passing or the platform kills the pod mid-drain. See [../operations/observability.md](../operations/observability.md) for the `/livez` vs `/readyz` split.
2. **Hosted services stop** in reverse registration order: `stoppingToken` is cancelled, loops observe it, finish the current item, and exit. Channel producers stop first, then consumers drain what remains or persist it via the outbox.
3. **Kestrel stops accepting new connections and drains in-flight requests**, bounded by `HostOptions.ShutdownTimeout` (default 30 s). Set it explicitly to your drain budget.
4. **DI disposes singletons in reverse creation order** — DbContext pools, `IHttpClientFactory` handlers, and the OpenTelemetry providers. Telemetry flushes last, on provider disposal, so the shutdown itself is recorded.

```csharp
builder.Services.Configure<HostOptions>(o => o.ShutdownTimeout = TimeSpan.FromSeconds(25));
```

The budget must exceed in-flight work's worst-case duration AND stay under the platform's termination grace (e.g. Kubernetes `terminationGracePeriodSeconds`, default 30 s) so the bounded drain finishes before SIGKILL. The platform's SIGKILL is the hard backstop for a hung drain. The budget is configuration, loaded per [configuration.md](configuration.md).

## Common Mistakes And Forbidden Patterns

- `.Result`, `.Wait()`, `.GetAwaiter().GetResult()`, or `Task.Run(...).Result` anywhere — sync-over-async is forbidden, not discouraged.
- `async void` methods.
- `_ = DoWorkAsync()` fire-and-forget from a handler instead of enqueueing to a supervised channel.
- Accepting a `CancellationToken` and not forwarding it to every awaited call (CA2016).
- Passing `CancellationToken.None` mid-stack "because it's almost done" without a scoped, commented reason.
- Creating a `CancellationTokenSource` without disposing it, or replacing the caller's token with a fresh one instead of linking.
- Unbounded channels, or `Task.WhenAll` fan-out sized by user input.
- `await` while holding a `SemaphoreSlim` without `try`/`finally` release, or long I/O inside a `lock`.
- `lock (this)`, `lock (typeof(T))`, or locking on strings — lock only on a private `Lock`/`object` field.
- Catching `OperationCanceledException` and logging it as an error during shutdown, or swallowing it so the loop never exits.
- A `BackgroundService` that ignores `stoppingToken` (e.g. `Task.Delay(...)` without the token) so shutdown hangs until the timeout kills the drain budget.
- Adding parallelism to hide slow code before measuring or simplifying the design.

## Verification And Proof

```powershell
pwsh ./verify.ps1
```

Be honest about what the gate can and cannot prove: **.NET has no Go-style race detector.** Nothing in the toolchain proves the absence of data races, so race safety is by construction plus targeted proof:

- Construction: immutable messages, ownership transfer through channels, single-writer rules, and the primitive table above. If shared mutable state exists, name its guard.
- Analyzer hygiene (part of the gate via `AnalysisLevel=latest-all` — see [../quality/linting.md](../quality/linting.md)): CA2016 (forward tokens), CA1849 (no blocking calls in async methods), CA2012 (ValueTask misuse), CA2008 (tasks without an explicit scheduler), CA1068 (token position). These enforce async hygiene; they do **not** detect races.
- Stress tests for suspicious state: hammer it with `Parallel.ForEachAsync` across many iterations and assert invariants afterward; run them in the CI matrix where scheduling differs across OSes.
- Shutdown tests: cancel the stopping token and assert bounded exit; a smoke test that opens a slow in-flight request, sends SIGTERM, and asserts the request completes and the process exits within the drain budget.
- Timeout tests for every external call path.
- Live diagnosis: `dotnet-counters` for starvation signals; Visual Studio's Concurrency Visualizer (Windows-only) or PerfView traces for investigating suspected races and contention — investigation tools, not gates.

If you cannot explain who owns a task and how it stops, the design is not done.

## Where To Go Next

- [errors-and-logging.md](errors-and-logging.md) — how cancellation and failures surface at the boundary.
- [time.md](time.md) — timers, delays, and scheduling through `TimeProvider`.
- [../recipes/add-background-worker.md](../recipes/add-background-worker.md) — the supervised worker recipe.
- [../operations/resilience.md](../operations/resilience.md) — timeout, retry, and bulkhead budgets for outbound calls.
