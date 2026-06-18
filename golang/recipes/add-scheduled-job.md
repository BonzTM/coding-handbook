# Recipe: Add Scheduled Job

Use this when work must run on a clock — every N minutes, hourly, or on a calendar schedule (nightly, weekdays at 02:00) — as opposed to a generic event/queue loop (use [add-background-worker.md](add-background-worker.md) for that).

## Files To Touch

- the owning subsystem package (e.g. `internal/scheduler` if the repo grows a dedicated scheduler, otherwise the worker/core package that owns the job) — the job logic + the timing loop
- `cmd/<app>/main.go` or `internal/runtime` wiring to own the goroutine under the root context
- `internal/telemetry` for the run-count, duration, and last-success instruments
- the job's idempotency / leader-election support in `internal/db` if it runs across replicas (see [../services/database.md](../services/database.md))
- scheduler tests (fake-clock schedule, overlap/skip, clean shutdown) and a leader-lock integration test where applicable

## Steps

1. **Pick the timing primitive.** For a fixed interval use a single stdlib `time.Ticker` created once and `Stop`ped on exit. For a *calendar* schedule (e.g. "weekdays 02:00 America/New_York") you need a cron-expression library — route the specific pick to [../decisions/framework-selection.md](../decisions/framework-selection.md) (default: stdlib `time.Ticker`; reach for a cron library only when the schedule is genuinely calendar-shaped). Compute the next fire time from an injected `Clock`, never `time.Now()` — see [../foundations/time.md](../foundations/time.md).
2. **Define the job as a unit, separate from the loop.** `RunOnce(ctx context.Context) error` holds the work; the scheduler holds only the timing. This keeps the job directly testable without driving the clock.
3. **Own the goroutine under the root context with a clean stop.** Start the scheduler from the supervised owner (`errgroup.Group` in `main`/`runtime`), `select` on `ctx.Done()` plus the ticker, and return when the context is cancelled. One owner, one stop path. See [../foundations/context-and-concurrency.md](../foundations/context-and-concurrency.md).
4. **Prevent overlapping runs.** A slow run must not be re-entered by the next tick. Use a non-blocking guard — `atomic.Bool` (CompareAndSwap) or `sync.Mutex.TryLock` — and *skip* (or queue at most one) when a run is already in flight; emit a "skipped, still running" metric/log so the overrun is visible.
5. **Decide missed-run policy on restart, deliberately.** A process that was down across a fire time has missed it. Choose explicitly: catch up once (run immediately on start if overdue) or wait for the next scheduled tick. Document the choice; do not let it be an accident of `Ticker` semantics.
6. **Make each run idempotent.** The same logical run executing twice (overlap that slipped through, a replay, a catch-up plus a normal tick) must not double-apply side effects. Key the work by a derived run identity (date bucket, watermark, or upsert) so re-execution converges.
7. **Fire once across replicas.** Multiple replicas each hold a scheduler; without coordination the job fires N times. Gate `RunOnce` behind leader election or a Postgres advisory lock (`pg_try_advisory_xact_lock` inside the run's transaction) so exactly one replica executes a given fire — cross-link [../services/database.md](../services/database.md). The non-leader replicas tick, fail to acquire, skip cleanly, and stay warm to take over.
8. **Bound each run.** Derive a per-run context with `context.WithTimeout` (a config key per [../foundations/configuration.md](../foundations/configuration.md), parsed as `time.Duration`) so one wedged run cannot hold the guard forever or starve subsequent fires.
9. **Make it observable.** Add a run counter labeled by outcome (`success`/`error`/`skipped`), a duration histogram, and a last-success timestamp gauge so an alert can fire on staleness. Labels stay low-cardinality — outcome only, never run IDs or timestamps as labels. See [add-metric.md](add-metric.md) and [../operations/observability.md](../operations/observability.md). Log each run start/finish once with `slog` including the run identity and elapsed duration.

## Invariants To Preserve

- single owner and a single stop path; the loop exits on `ctx.Done()` within the grace budget.
- no overlapping execution within a replica (guard) and no double execution across replicas (leader election / advisory lock).
- every run is idempotent: re-execution of the same logical fire converges, no duplicated side effects.
- each run is bounded by a timeout; a wedged run cannot block the schedule indefinitely.
- missed-run-on-restart behavior is an explicit, documented decision — not emergent.
- the job is observable: run count by outcome, duration, last-success time, and one structured log per run.
- the clock is injected; no `time.Now()` / `time.Sleep` in scheduler or job code.

## Proof

- `go test -race ./internal/scheduler/...` driving the schedule with the **fake clock** from `internal/testutil` (`Advance` past a fire time, assert `RunOnce` ran exactly once) — no real sleeps; cite [../foundations/time.md](../foundations/time.md).
- an overlap/skip test: hold one run open, `Advance` past the next tick, assert the second fire is skipped (guard held) and the `skipped` metric incremented, not the `success` count.
- a shutdown test: cancel the root context mid-loop and assert the scheduler goroutine returns within the grace budget (pair with `goleak` to prove no surviving goroutine).
- where multi-replica: an integration test against a real Postgres proving two contending schedulers result in exactly one successful `RunOnce` per fire (advisory lock held), per [../services/database.md](../services/database.md).
- `make verify` green (vet, test, race).

If the job mutates schema or seeds data, the change ships through a migration — see [add-migration.md](add-migration.md). The compiling wiring and telemetry patterns to mirror live in [../reference/exampleservice/](../reference/exampleservice/).
