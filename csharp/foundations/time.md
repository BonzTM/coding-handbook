# Time

Time handling rules that keep behavior deterministic in tests, correct across zones, and free of clock-related bugs.

## Default Approach

Time is an input, not an ambient fact. Inject `TimeProvider`, store UTC, measure spans with timestamps, and schedule with cancellation-aware timers.

### Inject TimeProvider; Never Call DateTime.Now In Core Code

Domain and core code take time as a dependency. `TimeProvider` is the standard library's clock abstraction — do not define your own `IClock`.

```csharp
public sealed class QuoteExpiryPolicy(TimeProvider time)
{
    private static readonly TimeSpan Lifetime = TimeSpan.FromMinutes(30);

    public bool IsExpired(Quote quote) =>
        time.GetUtcNow() - quote.IssuedAt >= Lifetime;
}
```

- Constructors accept `TimeProvider` like any other dependency; the composition root registers the real one once: `builder.Services.AddSingleton(TimeProvider.System);`.
- No code outside the composition root calls `DateTime.Now`, `DateTime.UtcNow`, `DateTimeOffset.Now`, or `DateTimeOffset.UtcNow` directly. `DateTime.Now` in particular reads the server's local zone — deployment-dependent and never a contract.
- `TimeProvider` covers the whole surface: `GetUtcNow()` for instants, `GetTimestamp()`/`GetElapsedTime()` for durations, `CreateTimer` for timers, and it plugs into `Task.Delay` and `PeriodicTimer` overloads. One injected dependency replaces every ambient time API.

### DateTimeOffset By Default; DateOnly/TimeOnly For Calendar Values

- Instants are `DateTimeOffset`. `DateTime` carries a `Kind` flag that comparison and arithmetic silently ignore — two `DateTime` values compare by ticks even when one is UTC and one is local, and `Kind == Unspecified` values survive round-trips undetected. `DateTimeOffset` makes the offset part of the value and compares instants correctly.
- Calendar dates without a time are `DateOnly` (birth dates, invoice dates); wall-clock times without a date are `TimeOnly` (business hours). Do not smuggle them through midnight-anchored `DateTime` values.
- Spans are `TimeSpan`, never raw `int` seconds or milliseconds. Config binds human-readable strings (`"00:00:30"`) to `TimeSpan` properties — see [configuration.md](configuration.md). Wire and storage fields that must be integers name the unit (`TimeoutMillis`) and convert at the boundary.

### Store And Transmit UTC; Localize Only At The Edge

- Persist, log, and serialize instants in UTC (`DateTimeOffset` with zero offset). PostgreSQL storage uses `timestamptz`, and Npgsql requires UTC values for it — see [../services/database.md](../services/database.md).
- Serialize machine-to-machine timestamps as ISO 8601 / RFC 3339, which `System.Text.Json` produces for `DateTimeOffset` by default. Reject naive local timestamps at the trust boundary — see [serialization.md](serialization.md).
- Convert to a zone only at the presentation boundary, with an explicit `TimeZoneInfo` chosen deliberately (usually from user or tenant data), never from the server's environment.

### Time Zones: IANA IDs Everywhere

- Look up zones with `TimeZoneInfo.FindSystemTimeZoneById("America/Chicago")` using IANA IDs. Since .NET 6, ICU maps IANA IDs on Windows too, so one ID set works on every OS — see [cross-platform.md](cross-platform.md).
- Store IANA IDs, never Windows zone names (`"Central Standard Time"`) and never raw UTC offsets — an offset is not a zone; it cannot represent DST.
- Converting a stored Windows ID at an integration boundary uses `TimeZoneInfo.TryConvertWindowsIdToIanaId`; new data never writes Windows IDs.
- Container images must ship ICU and tzdata for this to work — see [../operations/deployment.md](../operations/deployment.md) for the image choice.

### Measure Durations With Timestamps, Not Wall-Clock Subtraction

Subtracting two `GetUtcNow()` readings measures the wall clock, which NTP can step backwards or forwards mid-measurement. Durations use the monotonic timestamp API:

```csharp
var start = time.GetTimestamp();
await handler.HandleAsync(order, cancellationToken);
var elapsed = time.GetElapsedTime(start);
```

- `TimeProvider.GetTimestamp()`/`GetElapsedTime(start)` is the injectable form; `Stopwatch.GetTimestamp()`/`Stopwatch.GetElapsedTime(start)` is the static equivalent for code that has no `TimeProvider` (composition root, benchmarks).
- Never persist or serialize a timestamp reading — it is meaningless outside the process. Persist instants (`GetUtcNow()`), measure with timestamps.

### Timers And Delays Go Through TimeProvider

Every delay and timer takes the injected provider and a cancellation token, so tests control the clock and shutdown is never blocked by a sleep:

- `await Task.Delay(backoff, time, cancellationToken);` — the overload that accepts a `TimeProvider`.
- Recurring work uses `PeriodicTimer` constructed with the provider, in a loop that honors `stoppingToken`:

```csharp
public sealed class ReconciliationJob(TimeProvider time, IOrderReconciler reconciler)
    : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        using var timer = new PeriodicTimer(TimeSpan.FromMinutes(5), time);
        while (await timer.WaitForNextTickAsync(stoppingToken))
        {
            await reconciler.RunOnceAsync(stoppingToken);
        }
    }
}
```

`WaitForNextTickAsync` throws `OperationCanceledException` when `stoppingToken` fires — the expected shutdown path, handled by the host. `PeriodicTimer` does not stack missed ticks, so a slow iteration never causes a burst. Full recipe: [../recipes/add-scheduled-job.md](../recipes/add-scheduled-job.md).

- Raw `System.Threading.Timer` and `Task.Run` sleep-loops are forbidden for scheduling; they dodge both the provider and the stopping token.

### Tests Are Deterministic And Sleepless

Drive time with `FakeTimeProvider` (from the `Microsoft.Extensions.TimeProvider.Testing` package — pinned in [templates/Directory.Packages.props](../templates/Directory.Packages.props)). Advance time explicitly; never sleep.

```csharp
[Fact]
public void Quote_expires_after_thirty_minutes()
{
    var time = new FakeTimeProvider(new DateTimeOffset(2026, 7, 10, 12, 0, 0, TimeSpan.Zero));
    var policy = new QuoteExpiryPolicy(time);
    var quote = new Quote(IssuedAt: time.GetUtcNow());

    time.Advance(TimeSpan.FromMinutes(29));
    Assert.False(policy.IsExpired(quote));

    time.Advance(TimeSpan.FromMinutes(1));
    Assert.True(policy.IsExpired(quote));
}
```

- `Advance` also fires due timers and completes `Task.Delay`/`PeriodicTimer` waits created from the fake — scheduling logic is testable without wall-clock waits.
- A `Task.Delay` or `Thread.Sleep` in a test is either a hidden race or wasted wall-clock time. Replace it with `FakeTimeProvider` plus a synchronization signal (awaited task, `TaskCompletionSource`).
- Make "now" an explicit fixture value so golden output and time-derived data are reproducible. Shared helpers live in `Orders.TestUtilities` — see [../quality/testing.md](../quality/testing.md).

## Common Mistakes And Forbidden Patterns

- `DateTime.Now`, `DateTime.UtcNow`, or `DateTimeOffset.UtcNow` in domain or core code instead of an injected `TimeProvider`.
- Using `DateTime` for instants — `Kind` mismatches compare silently wrong; use `DateTimeOffset`.
- Storing or transmitting local-zone timestamps; deciding business logic from the server's local zone.
- Storing Windows time zone names or bare UTC offsets instead of IANA IDs.
- Measuring elapsed time by subtracting wall-clock readings instead of `GetTimestamp`/`GetElapsedTime`.
- Raw `int` seconds/milliseconds for spans instead of `TimeSpan`.
- `Thread.Sleep` or un-faked `Task.Delay` in tests to wait for time-dependent behavior.
- Scheduling loops with `Task.Delay` chains or `System.Threading.Timer` instead of `PeriodicTimer` with the provider and `stoppingToken`.
- Defining a custom `IClock` interface when `TimeProvider` is the standard abstraction.

## Verification And Proof

```powershell
pwsh ./verify.ps1
```

Time handling is done when:

- `TimeProvider` is injected everywhere time matters, and a search finds no `DateTime.Now`, `DateTime.UtcNow`, or `DateTimeOffset.UtcNow` outside the composition root.
- instants are `DateTimeOffset`, stored and transmitted in UTC, localized only at presentation with an explicit IANA zone.
- durations are measured with `GetTimestamp`/`GetElapsedTime` and carried as `TimeSpan`.
- scheduled loops use `PeriodicTimer` with the provider and exit promptly on `stoppingToken`.
- tests advance `FakeTimeProvider` instead of sleeping, run fast, and are deterministic across runs and OSes (the CI matrix proves the latter).

## Where To Go Next

- [cancellation-and-async.md](cancellation-and-async.md) — timeouts, cancellation, and scheduling loop ownership.
- [cross-platform.md](cross-platform.md) — IANA IDs and ICU across operating systems.
- [configuration.md](configuration.md) — binding `TimeSpan` values from configuration.
- [../quality/testing.md](../quality/testing.md) — where `FakeTimeProvider` helpers live and how tests prove time-dependent behavior.
