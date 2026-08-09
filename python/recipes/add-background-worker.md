# Recipe: Add Background Worker

Use this when a process polls, consumes, retries, or performs long-running asynchronous work.

## Files To Touch

- `src/<app>/workers/<worker>.py`
- lifespan/composition root and readiness/telemetry wiring
- worker, cancellation, and shutdown tests

## Steps

1. Give the worker explicit dependencies and one `async run()` entrypoint.
2. Separate one unit of work from the loop; classify retryable, terminal, and cancellation outcomes first.
3. Start it inside the root `asyncio.TaskGroup`; do not use an unowned `create_task`.
4. Bound external calls with `asyncio.timeout()` and concurrency with a semaphore or bounded queue.
5. On cancellation, stop intake, finish/settle owned work within the grace budget, close resources, and re-raise `CancelledError`.

## Invariants To Preserve

- Every task has one owner, one cancellation source, and one awaited completion path.
- Retries are bounded, jittered, and safe for the side effect.
- Blocking calls run through `asyncio.to_thread` or a bounded executor.
- Readiness and telemetry reflect a worker whose health is required.

## Proof

```bash
uv run pytest tests/workers -k '<worker>'
uv run pytest tests/workers -k 'cancel or shutdown or retry'
make verify
```

Prove no leaked tasks after cancellation and bounded drain under failure. Governing doc: [concurrency and asyncio](../foundations/concurrency-and-asyncio.md).
