# Recipe: Add Scheduled Job

Use this for fixed-interval work; use a platform CronJob for independent calendar schedules.

## Files To Touch

- `src/<app>/workers/<job>.py` for `run_once` and the interval loop
- composition, telemetry, and optional PostgreSQL lease/advisory-lock code
- deterministic timing, overlap, replica, and shutdown tests

## Steps

1. Keep `run_once()` separate from scheduling and inject clock, sleep, and jitter seams.
2. Run an owned loop under the root TaskGroup: execute, compute bounded interval+jitter, await injected sleep, and exit on cancellation.
3. Prevent overlap. Across replicas, use a platform singleton or PostgreSQL coordination; otherwise every replica runs the job.
4. Make a logical run idempotent and choose missed-run behavior explicitly.
5. Bound every run with `asyncio.timeout()` and emit success/error/skipped count, duration, and last-success time.

## Invariants To Preserve

- No real sleeps in tests; no calendar scheduler dependency for a fixed interval.
- Jitter is bounded and cannot make the interval negative or unbounded.
- One run cannot overlap itself or hold the schedule forever.
- Cancellation is re-raised after cleanup and completes within platform grace.

## Proof

```bash
uv run pytest tests/workers -k 'schedule or overlap or shutdown'
uv run pytest -m integration -k '<job>_singleton'
make verify
```

Use the integration command only when replica coordination is required. Governing docs: [time](../foundations/time.md) and [deployment](../operations/deployment.md).
