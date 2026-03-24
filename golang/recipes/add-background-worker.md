# Recipe: Add Background Worker

Use this when the repo needs polling, scheduled work, queue consumption, or asynchronous retries.

## Files To Touch

- a worker package under `internal/worker` or the owning subsystem
- `cmd/<app>/main.go` or `internal/runtime` wiring
- telemetry or readiness code if the worker affects service state
- worker tests and shutdown tests

## Steps

1. Define a worker struct with explicit dependencies and a `Run(ctx context.Context) error`-style entrypoint.
2. Implement the loop around `select` on `ctx.Done()` plus the work trigger.
3. Decide retry, backoff, and idempotency behavior before writing the happy path.
4. Start the worker from a supervised owner such as `errgroup.Group`.
5. Make shutdown bounded and observable.

## Invariants To Preserve

- every goroutine has an owner and a stop path
- retries do not create duplicate external side effects without an idempotency story
- worker failures are logged and surfaced at the right boundary
- readiness reflects whether the worker must be healthy before traffic is accepted
- queue-backed workers also preserve ordering, ack timing, and DLQ policy according to the documented contract

## Proof

- `go test -race` for the worker package or owning package
- cancellation test proving bounded shutdown
- retry or backoff test for failure paths
- local smoke test showing start, stop, and error telemetry

If the worker consumes brokered messages, also read [add-event-consumer.md](add-event-consumer.md).
