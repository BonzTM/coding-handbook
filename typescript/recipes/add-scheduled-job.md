# Recipe: Add Scheduled Job

Use this when work runs on a recurring or delayed schedule.

## Files To Touch

- `src/core/<name>-job.ts`
- scheduler adapter under `src/lib/`
- `src/index.ts` composition and shutdown
- config, telemetry, runbook, and tests
- durable schedule storage when execution must survive restart

## Steps

1. Define the schedule, IANA zone, missed-run, overlap, duplicate, and restart semantics.
2. Keep calendar calculation separate from job execution.
3. Inject a `Clock` for time decisions; use timers only in the scheduler adapter.
4. Persist intent instead of using a long in-memory timer when durability matters.
5. Cap runs, batch size, concurrency, retries, and execution deadline.
6. Prevent overlap or make overlapping execution idempotent by contract.
7. Pass the process `AbortSignal`; remove timers and drain current work on stop.
8. Emit low-cardinality job/result/latency and missed-run telemetry.
9. Document safe manual trigger, diagnosis, and replay in the runbook.

```bash
npm test -- --runInBand src/core/<name>-job.test.ts
npm run lint
npm run verify
```

## Invariants To Preserve

- UTC instants are used for storage and transport.
- Local calendar rules always name a zone and daylight-saving policy.
- Tests use a fixed clock or Jest fake timers, never real sleeps.
- A timer has one owner and one cleanup path.
- Missed and duplicate execution cannot silently multiply durable effects.
- Shutdown never waits without a deadline.

## Proof

- Tests cover boundary instant, missed run, overlap, duplicate, and restart.
- DST gaps and folds are tested when local schedules matter.
- Fake timers are restored and no open handles remain.
- A shutdown smoke test proves cancellation and bounded drain.
- `npm run verify` is green.
