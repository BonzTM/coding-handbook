# Recipe: Add Scheduled Job

Use this when work must run on a clock — every N minutes, hourly, or on a calendar schedule (nightly, weekdays at 02:00) — as opposed to a generic event/queue loop (use [add-background-worker.md](add-background-worker.md) for that).

## Files To Touch

- the job unit in `src/Orders.Core/...` — `RunOnceAsync(CancellationToken)` holding the work
- the timing loop as a `BackgroundService` under `src/Orders.Api/Workers/`, registered in `Program.cs`
- telemetry wiring for the run-count, duration, and last-success instruments
- the job's idempotency / cross-replica lock support in `src/Orders.Infrastructure/Data` if it runs across replicas (see [../services/database.md](../services/database.md))
- scheduler tests (fake-time schedule, overrun, clean shutdown) under `tests/Orders.UnitTests`, and a lock integration test under `tests/Orders.IntegrationTests` where applicable

## Steps

1. **Pick the timing primitive.** For a fixed interval use one `PeriodicTimer` constructed with the injected `TimeProvider` (`new PeriodicTimer(period, timeProvider)`), created once in `ExecuteAsync` and disposed with it (`using`). For a *calendar* schedule (e.g. "weekdays 02:00 America/New_York") compute the next fire time from `timeProvider.GetUtcNow()` plus `TimeZoneInfo.FindSystemTimeZoneById` with the IANA ID, and delay with `Task.Delay(until, timeProvider, stoppingToken)`; adopting a cron/scheduler library (Quartz, Hangfire) is NOT the default — route that pick to [../decisions/framework-selection.md](../decisions/framework-selection.md). Never `DateTime.Now`, never `Task.Delay` without the `TimeProvider` overload — see [../foundations/time.md](../foundations/time.md).
2. **Define the job as a unit, separate from the loop.** `RunOnceAsync(CancellationToken)` in Core holds the work; the `BackgroundService` holds only the timing. This keeps the job directly testable without driving the clock.
3. **Own the loop under the host with a clean stop.** `await timer.WaitForNextTickAsync(stoppingToken)` in a `while` loop; it throws `OperationCanceledException` when `stoppingToken` fires — let that end the loop. One owner, one stop path. See [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md).
4. **Prevent overlapping runs.** Awaiting `RunOnceAsync` inside the loop means a replica never overlaps itself, and `PeriodicTimer` coalesces ticks that fire during a slow run rather than queueing them. Make the overrun visible: when a run exceeds the period, emit a "run overran period" metric/log instead of silently skipping fires.
5. **Decide missed-run policy on restart, deliberately.** A process that was down across a fire time has missed it. Choose explicitly: catch up once (run immediately on start if overdue, keyed off durable last-success state) or wait for the next tick. Document the choice; do not let it be an accident of timer semantics.
6. **Make each run idempotent.** The same logical run executing twice (a catch-up plus a normal tick, a replay, a replica failover) must not double-apply side effects. Key the work by a derived run identity (date bucket, watermark, or upsert) so re-execution converges.
7. **Fire once across replicas.** Every replica hosts the scheduler; without coordination the job fires N times. Gate `RunOnceAsync` behind a Postgres advisory lock (`SELECT pg_try_advisory_xact_lock(...)` inside the run's transaction, via `Database.ExecuteSqlAsync`) so exactly one replica executes a given fire — cross-link [../services/database.md](../services/database.md). Non-winning replicas tick, fail to acquire, skip cleanly, and stay warm to take over.
8. **Bound each run.** Derive a per-run token with `CancellationTokenSource.CreateLinkedTokenSource(stoppingToken)` plus `CancelAfter` from a validated config key (a `TimeSpan` option per [../foundations/configuration.md](../foundations/configuration.md)) so one wedged run cannot starve subsequent fires.
9. **Make it observable.** Add a run counter tagged by outcome (`success`/`error`/`skipped`), a duration histogram, and a last-success timestamp gauge so an alert can fire on staleness. Tags stay low-cardinality — outcome only, never run IDs or timestamps as tags. See [add-metric.md](add-metric.md) and [../operations/observability.md](../operations/observability.md). Log each run start/finish once via `[LoggerMessage]` including the run identity and elapsed duration.

## Invariants To Preserve

- single owner and a single stop path; the loop exits on `stoppingToken` within the shutdown budget.
- no overlapping execution within a replica (awaited run) and no double execution across replicas (advisory lock / leader election).
- every run is idempotent: re-execution of the same logical fire converges, no duplicated side effects.
- each run is bounded by a timeout; a wedged run cannot block the schedule indefinitely.
- missed-run-on-restart behavior is an explicit, documented decision — not emergent.
- the job is observable: run count by outcome, duration, last-success time, and one structured log per run.
- the clock is injected: no `DateTime.Now`/`DateTimeOffset.Now`, no raw `Task.Delay`, no `Stopwatch.StartNew` in scheduler or job code — everything flows through `TimeProvider`.

## Proof

- `dotnet test tests/Orders.UnitTests --filter <Job>` driving the schedule with `FakeTimeProvider` (from `Microsoft.Extensions.TimeProvider.Testing`): `Advance(period)` past a fire time, assert `RunOnceAsync` ran exactly once — no real sleeps; cite [../foundations/time.md](../foundations/time.md).
- an overrun test: hold one run open past the next tick via the fake clock, assert no second concurrent run started and the overrun metric/log fired.
- a shutdown test: stop the host mid-loop and assert `ExecuteAsync` returns within the shutdown budget.
- where multi-replica: an integration test against a real PostgreSQL proving two contending schedulers result in exactly one successful `RunOnceAsync` per fire (advisory lock held), behind the `-Integration` switch, per [../services/database.md](../services/database.md).
- run `pwsh ./verify.ps1` — green.

Governing doc: [time.md](../foundations/time.md). If the job mutates schema or seeds data, the change ships through a migration — see [add-migration.md](add-migration.md).
