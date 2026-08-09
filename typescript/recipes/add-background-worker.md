# Recipe: Add Background Worker

Use this when a process owns long-running or queued work.

## Files To Touch

- `src/core/<feature>-worker.ts` for the work contract
- adapter code under `src/lib/` or `src/db/`
- `src/index.ts` lifecycle composition
- config, telemetry, readiness, and runbook
- unit and lifecycle smoke tests

## Steps

1. Name the owner, input source, queue capacity, concurrency, and overflow policy.
2. Define `start`/`run` and awaited `stop` behavior around an `AbortSignal`.
3. Check cancellation before work and between bounded chunks.
4. Bound batch size, pages, attempts, in-flight tasks, and dependency timeouts.
5. Classify permanent, transient, cancellation, and poison-item outcomes.
6. Make retryable durable effects idempotent before enabling retries.
7. Register the worker only after config and dependencies are ready.
8. On shutdown, stop intake, abort owned waits, drain within deadline, then close resources.
9. Record backlog, oldest age, latency, result, and saturation with bounded attributes.

```bash
npm test -- --runInBand src/core/<feature>-worker.test.ts
npm run build
npm run verify
```

## Invariants To Preserve

- No work starts during module import.
- Every promise is awaited, returned, or supervised with observed failure.
- Queues, concurrency, loops, retries, output, and shutdown are bounded.
- A process-level unknown failure initiates termination; it is not ignored.
- Readiness becomes false before intake stops.
- Logs never contain full jobs or sensitive payloads.

## Proof

- Controlled-promise tests assert the observed concurrency maximum.
- Tests cover pre-abort, mid-work abort, overflow, retry exhaustion, and poison input.
- Shutdown smoke proves stop-intake, drain, close, and deadline behavior.
- Load evidence proves memory remains bounded at capacity.
- `npm run verify` is green.
