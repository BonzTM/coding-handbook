# Concurrency and Asyncio

Cancellation, structured concurrency, and task-ownership rules for Python code that behaves correctly under load and on shutdown.

## Default Approach

Use asyncio for concurrent I/O in services and workers. Keep simple CLIs and libraries synchronous unless their boundary is already async.

### Structured Concurrency

Use `asyncio.TaskGroup` for sibling work that succeeds or fails as one scope. Exiting the context awaits every child; the first non-cancellation failure cancels siblings and failures emerge as an `ExceptionGroup`. This behavior is defined by the [Python 3.11 asyncio task documentation](https://docs.python.org/3.11/library/asyncio-task.html#task-groups).

Every task has an owner: request, TaskGroup, worker supervisor, or application root. Bare fire-and-forget is forbidden. If `asyncio.create_task()` is necessary for independently supervised work, retain a strong reference, observe its result, cancel it during shutdown, and await it; the event loop keeps only weak task references, as the [`create_task` documentation](https://docs.python.org/3.11/library/asyncio-task.html#asyncio.create_task) warns.

### Deadlines And Cancellation

- Wrap bounded work in `asyncio.timeout(...)`; propagate a caller's shorter budget rather than replacing it.
- Cleanup belongs in `finally`. If `CancelledError` is caught, re-raise it after cleanup. TaskGroup and timeout use cancellation internally and can misbehave when it is swallowed.
- `asyncio.shield()` requires a comment naming the must-complete operation, owner, and outer deadline. Shield only the smallest operation and still retain/await its task.
- A retry consumes one overall deadline; it does not grant a fresh unbounded timeout per attempt.

### Bound All Fan-Out

Use `asyncio.Semaphore` around dependency calls and bounded `asyncio.Queue(maxsize=...)` for producer/consumer pipelines. User-sized input never becomes unbounded tasks or `gather` arguments. State the concurrency limit in config, cap it to a safe range, and emit saturation telemetry.

### Blocking Work In Async Paths

No blocking network, database, filesystem, process, or sleep call runs on the event loop. Use an async-native client. Move a small unavoidable blocking function through `asyncio.to_thread`; use `loop.run_in_executor` only when selecting and owning an executor explicitly. Ruff `ASYNC` rules enforce common violations.

CPU-bound work does not belong in threads expecting parallel speedup. Use a bounded `ProcessPoolExecutor`, an external worker, or a synchronous design, and include serialization/cancellation cost in the decision.

### Async Iterators And Resources

Consume async context managers with `async with`. When stopping an async generator early and its owner does not already close it, use `contextlib.aclosing()` so cleanup runs in the same context. Never leave an HTTP response stream, DB result, generator, or task for garbage collection to close.

### Locks And Shared State

Prefer immutable messages and single-owner state. Use `asyncio.Lock` only for short event-loop critical sections; never hold it across external I/O. An asyncio lock does not protect access from threads or processes. Use a queue when the workflow is ownership transfer rather than shared mutation.

### Graceful Shutdown And Draining

The composition root installs loop signal handlers for SIGINT/SIGTERM and cancels the root task. Shutdown is ordered and bounded:

1. Flip readiness to unready; liveness remains healthy.
2. Stop accepting new work and drain in-flight HTTP/message work under a fresh shutdown timeout.
3. Cancel and await worker/root task groups.
4. Close database engines and external clients after their users stop.
5. Flush telemetry last.

The configured grace stays below the platform termination grace. A second signal restores hard termination. Signal portability and platform behavior belong in the deployment proof, not hidden in a helper.

### When Not To Use Asyncio

A command that performs sequential local work stays synchronous. A public library does not force asyncio on callers unless its core job is async I/O. CPU-heavy algorithms use processes or a separate service after measurement. Concurrency is not a treatment for slow design.

## Common Mistakes And Forbidden Patterns

- Unowned `create_task`, never-awaited coroutine, or task whose exception is never retrieved.
- Swallowed `CancelledError`, broad exception handling that catches cancellation intent, or unbounded shield.
- `gather` or task creation sized directly by external input.
- Blocking calls or `time.sleep()` in an async path.
- A semaphore acquired without guaranteed release, or a lock held across I/O.
- Closing clients/engines before dependent work drains; flushing telemetry before shutdown completes.
- An unbounded drain or grace longer than the platform termination window.

## Verification And Proof

```bash
PYTHONASYNCIODEBUG=1 uv run pytest -W error
uv run ruff check .
make verify
```

Add deterministic tests for sibling failure, cancellation propagation, deadline expiry, semaphore/queue bounds, early async-generator close, and shutdown ordering. A process smoke test sends SIGTERM during a slow request, proves readiness changes first, the request drains, resources close, and the process exits within grace. No pending-task or unclosed-resource warning is acceptable.

Related: [time](time.md), [resilience](../operations/resilience.md), and [background-worker recipe](../recipes/add-background-worker.md).
