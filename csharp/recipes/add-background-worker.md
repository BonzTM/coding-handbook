# Recipe: Add Background Worker

Use this when the repo needs polling, queue consumption, asynchronous retries, or any long-running loop alongside request traffic.

## Files To Touch

- a `BackgroundService` under `src/Orders.Api/Workers/` (or the owning Infrastructure subsystem when the loop is broker/DB-specific)
- `src/Orders.Api/Program.cs` — `builder.Services.AddHostedService<T>()` registration
- `src/Orders.Core/...` — the unit of work the loop invokes
- health/readiness wiring if the worker affects whether the service should take traffic
- worker tests and shutdown tests under `tests/Orders.UnitTests`

## Steps

1. Separate the work from the loop: put a `RunOnceAsync(CancellationToken)`-style method on a Core service with explicit constructor-injected dependencies; the `BackgroundService` owns only the loop, backoff, and telemetry.
2. Implement `ExecuteAsync(CancellationToken stoppingToken)` as a loop that honors `stoppingToken` everywhere — pass it into every awaited call, never swallow `OperationCanceledException` triggered by it, and exit promptly when it fires. See [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md).
3. Decide retry, backoff, and idempotency behavior before writing the happy path.
4. Decide supervision deliberately: an exception that escapes `ExecuteAsync` stops the host by default (`BackgroundServiceExceptionBehavior.StopHost`) so the orchestrator restarts the process. Contain expected per-iteration failures with a `try`/`catch` inside the loop (log once, back off, continue); let genuinely fatal states escape and kill the host — never a bare catch-all that loops a broken worker forever.
5. Make shutdown bounded and observable: the loop must return within `HostOptions.ShutdownTimeout` once `stoppingToken` fires, and log start/stop once each.
6. Decide readiness semantics: if traffic must not be accepted while the worker is down, register a health check that reflects worker state on `/readyz` — see [../operations/operability.md](../operations/operability.md). A worker whose lag is tolerable stays out of readiness and alerts on lag instead.

## Invariants To Preserve

- every background task has an owner and a stop path; no fire-and-forget `Task.Run` without a holder that observes its outcome
- `stoppingToken` flows into every awaited operation in the loop
- retries do not create duplicate external side effects without an idempotency story
- worker failures are logged once at the boundary that can act, and the crash-vs-contain decision is explicit
- readiness reflects whether the worker must be healthy before traffic is accepted
- queue-backed workers also preserve ordering, ack timing, and DLQ policy according to the documented contract

## Proof

- `dotnet test tests/Orders.UnitTests --filter <Worker>` covering the `RunOnceAsync` unit directly, no host required
- a cancellation test proving bounded shutdown: start the worker (or the host via `WebApplicationFactory`), trigger stop, assert `ExecuteAsync` completes within the shutdown budget
- a retry/backoff test for failure paths, driven by a `FakeTimeProvider` where delays are involved
- local smoke test showing start, stop, and error telemetry
- run `pwsh ./verify.ps1`

Governing doc: [cancellation-and-async.md](../foundations/cancellation-and-async.md). If the worker runs on a clock, use [add-scheduled-job.md](add-scheduled-job.md); if it consumes brokered messages, also read [add-event-consumer.md](add-event-consumer.md).
