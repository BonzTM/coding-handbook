# Async And Cancellation

Ownership, cancellation, concurrency, and shutdown rules for reliable Node.js processes.

## Default Approach

Every asynchronous operation has an owner, a bound, a cancellation path, and an observed outcome.

### AbortSignal Is The Cancellation Contract

Accept `AbortSignal` as a named option on every I/O operation, long-running computation, retry loop, queue wait, and worker entry point. Pass the caller's signal through rather than creating an unrelated controller in each layer.

```ts
type LoadOptions = Readonly<{
  signal: AbortSignal;
}>;

async function loadWidget(id: WidgetId, options: LoadOptions): Promise<Widget> {
  return repository.get(id, { signal: options.signal });
}
```

Check `signal.throwIfAborted()` before expensive synchronous work and between bounded chunks. APIs that support signals receive the same composed signal. Code that cannot cancel a dependency must still stop awaiting it, bound its resource use, and document the limitation.

Cancellation is not failure noise. Preserve the abort reason where safe, classify it separately from timeout and dependency failure, and do not retry caller cancellation.

### Deadlines And Timeouts

Every external operation has a timeout chosen from the caller's remaining budget. Compose caller cancellation and a timeout signal with `AbortSignal.any`; create deadlines at the boundary that owns the service-level budget.

```ts
const timeout = AbortSignal.timeout(config.dependencyTimeoutMs);
const signal = AbortSignal.any([requestSignal, timeout]);
const response = await fetch(url, { signal });
```

Keep composition inside the outbound adapter and check failure explicitly:

```ts
export async function fetchJson(
  url: URL,
  callerSignal: AbortSignal,
  timeoutMs: number,
): Promise<unknown> {
  const signal = AbortSignal.any([
    callerSignal,
    AbortSignal.timeout(timeoutMs),
  ]);
  const response = await fetch(url, { signal });
  if (!response.ok) throw new Error(`dependency returned ${response.status}`);
  const body: unknown = await response.json();
  return body;
}
```

Do not layer independent timeouts that can exceed the top-level deadline. A timeout is an operational policy, not a magic constant hidden in an adapter. Record it in configuration, telemetry, and runbooks where operators need it.

Clear manually created timers in `finally`. Prefer platform signal helpers because they bind timer and cancellation lifetime safely.

### Promise Ownership

Every promise is awaited, returned, or deliberately owned by a lifecycle object that observes rejection and can stop the work. Type-aware lint must reject floating promises.

`void promise` is not ownership. Use it only at an adapter boundary that attaches a rejection handler and transfers work to a documented supervisor. Event handlers that cannot return promises wrap async work and surface failure to the owning component or process.

Use `Promise.all` only when all work may start together, failure semantics are understood, and fan-out is already bounded. `Promise.allSettled` is for deliberate partial-result policy; every rejection still needs classification and action.

### Structured Lifetimes

Child work must not outlive its request, job, component, or process owner. The owner creates the controller, starts children, aborts them when its scope ends, and awaits their settlement before reporting completion.

For React effects, create a controller inside the effect and abort it in cleanup. Prefer TanStack Query's supplied signal for query functions. Never use an `isMounted` boolean as a substitute for canceling the underlying operation.

For services, the process lifecycle owns listeners, consumers, timers, pools, exporters, and worker threads. Avoid module-level background work that begins on import.

### Bounded Concurrency

Never map an unbounded collection directly into `Promise.all`. Establish a maximum input count and a concurrency limit. Prefer a fixed worker pool or an existing owned limiter; adding a dependency requires the framework-selection review.

Each queue has a capacity and an overload policy: wait within a deadline, reject, shed, or persist elsewhere. An unbounded in-memory queue is a memory leak with delayed symptoms.

Batch size, page count, retry count, and worker count are explicit configuration with validated upper bounds. Pagination loops stop on a maximum page or item count even if a dependency returns a broken cursor.

A fixed worker pool bounds input and active work:

```ts
const MAX_BATCH = 1_000;
const MAX_WORKERS = 32;

export async function runBounded(
  tasks: ReadonlyArray<() => Promise<void>>,
  concurrency: number,
): Promise<void> {
  if (tasks.length > MAX_BATCH) throw new RangeError("batch too large");
  if (!Number.isInteger(concurrency) || concurrency < 1 || concurrency > MAX_WORKERS) {
    throw new RangeError("invalid concurrency");
  }
  let nextIndex = 0;
  async function runWorker(): Promise<void> {
    for (let attempt = 0; attempt < MAX_BATCH; attempt += 1) {
      const task = tasks[nextIndex];
      nextIndex += 1;
      if (task === undefined) return;
      await task();
    }
  }
  const workers = Array.from({ length: Math.min(concurrency, tasks.length) }, runWorker);
  await Promise.all(workers);
}
```

The hard batch bound also bounds each worker loop. Add sibling cancellation when the work accepts a signal.

### Error Propagation

Awaited work preserves typed failures and `cause`. Do not catch merely to log and rethrow; log once at the boundary that can act. A task group defines whether one failure cancels siblings, whether partial results are permitted, and which error is returned.

Attach process-level handlers for `uncaughtException` and `unhandledRejection` only to record a last-resort fatal event, initiate best-effort bounded shutdown where safe, and terminate. Continuing after an unknown process-level failure is forbidden.

The final handler owns its shutdown promise and always terminates:

```ts
process.once("unhandledRejection", (reason: unknown) => {
  logger.fatal({ err: reason }, "unhandled promise rejection");
  lifetime.abort(reason);
  void shutdown().then(
    () => process.exit(1),
    () => process.exit(1),
  );
});
```

### Graceful Shutdown And Drain

The composition root owns shutdown. On `SIGTERM` or `SIGINT`:

1. mark readiness false so new traffic stops arriving;
2. stop accepting new HTTP requests, messages, and scheduled work;
3. abort the process-lifetime signal;
4. allow in-flight work to settle within a fixed drain deadline;
5. close consumers, servers, database pools, workers, and telemetry exporters;
6. exit zero for a clean signal-driven stop, nonzero for fatal startup or runtime failure.

Signal handling is idempotent. A second signal may shorten the wait, but cleanup still must not run concurrently twice. Shutdown never waits forever; when the drain deadline expires, record remaining work without sensitive payloads and terminate.

Queue consumers stop fetching before they wait for current handlers. A message is acknowledged only after the durable effect required by its contract. Work interrupted before that point remains eligible for redelivery.

HTTP shutdown rejects or stops new connections and permits in-flight requests only within the deadline. Coordinate server keep-alive behavior with the orchestrator's termination grace period.

Track owned work and race its settlement against the process deadline:

```ts
const lifetime = new AbortController();
const inFlight = new Set<Promise<void>>();

export function track(task: Promise<void>): Promise<void> {
  inFlight.add(task);
  return task.finally(() => {
    inFlight.delete(task);
  });
}

async function shutdown(): Promise<void> {
  lifetime.abort(new Error("process stopping"));
  const drain = Promise.allSettled([...inFlight]);
  let timer: NodeJS.Timeout | undefined;
  const deadline = new Promise<never>((_, reject) => {
    timer = setTimeout(() => reject(new Error("drain deadline exceeded")), 10_000);
  });
  try {
    await Promise.race([drain, deadline]);
  } finally {
    if (timer !== undefined) clearTimeout(timer);
  }
}

process.once("SIGTERM", () => {
  void shutdown().then(() => process.exit(0), () => process.exit(1));
});
```

Production composition stops intake and closes resources around this drain. Callers await `track(task)` so task rejection remains observed.

### Event-Loop Discipline

The Node event loop must remain responsive. Avoid synchronous filesystem, crypto, compression, parsing, or child-process calls on request and consumer paths. Break CPU work into bounded chunks only when yielding preserves correctness and latency; otherwise isolate it.

Measure event-loop delay for services where saturation is a material risk. High delay is a capacity or blocking-work signal, not something to hide with longer timeouts.

Microtask chains can starve timers and I/O. Do not recursively schedule promises, `queueMicrotask`, `process.nextTick`, or zero-delay timers without a fixed iteration bound and cancellation check.

Streams require backpressure. Await writes or pipeline completion, propagate abort, cap buffered bytes, and release resources on all paths. Do not convert an unbounded stream into one in-memory buffer.

### Worker Threads

Use `worker_threads` only for measured CPU-bound JavaScript that harms event-loop latency. They do not make I/O faster and do not remove the need for bounded queues or cancellation.

Define a small serializable message protocol, validate messages on both sides, cap worker count and queued jobs, and transfer large buffers when ownership permits. Never pass secrets or unbounded objects merely because structured clone accepts them.

The owner handles worker startup failure, message validation failure, nonzero exit, timeout, cancellation, and replacement. Termination is awaited and bounded. A worker crash fails or requeues its owned job according to an explicit idempotency policy.

### Child Processes And External Work

Prefer library APIs to shell commands. When a subprocess is required, use argument arrays without a shell, constrain executable and working directory, bound output, pass a minimal environment, attach an abort signal, and await exit. Kill escalation and cleanup are part of the adapter contract.

Network, database, broker, and subprocess operations all need cancellation, timeout, concurrency, and response-size bounds. Retrying does not remove those requirements.

### Testing Async Behavior

Use controlled promises, fake clocks, local servers, and explicit lifecycle hooks. Never use arbitrary sleep to wait for work. A test should observe the event that establishes completion: a returned promise, closed stream, acknowledged message, worker response, or shutdown result.

Test pre-aborted signals, abort during work, timeout, sibling failure, partial completion policy, concurrency maximum, queue overflow, second shutdown signal, and drain expiry. Restore fake timers and global handlers after every test.

## Common Mistakes And Forbidden Patterns

- Floating promises, `void` used as a lint escape, or `.catch(() => undefined)`.
- `new Promise(async ...)`, async array callbacks whose results are ignored, or promise constructors around promise APIs.
- Unbounded `Promise.all`, queues, retries, pages, streams, or worker creation.
- A fresh `AbortController` that severs caller cancellation.
- Retrying aborts, validation failures, authorization failures, or non-idempotent work.
- Real sleeps in tests or production polling without a fixed bound and backoff.
- Synchronous CPU or filesystem work on request and consumer paths.
- Process handlers that log an unhandled failure and continue serving.
- Shutdown that closes the database before in-flight handlers finish.
- Worker threads used for ordinary I/O or without a bounded job queue.

## Verification And Proof

- Type-aware lint reports zero floating or misused promises.
- Tests prove pre-abort, mid-operation abort, timeout classification, and resource cleanup.
- Concurrency tests assert the observed maximum never exceeds configuration.
- Load tests exercise queue capacity and overload policy without unbounded memory growth.
- Shutdown smoke tests stop intake, drain owned work, close resources, and meet the orchestrator grace period.
- A forced drain timeout terminates predictably and reports unfinished work safely.
- Event-loop delay and worker-queue behavior are observable under representative CPU load.
- Process-level uncaught failure tests or a controlled smoke harness prove nonzero termination.
- Every long-lived task has one named owner and one awaited stop path.

Related: [time.md](time.md), [errors-and-logging.md](errors-and-logging.md), [../operations/resilience.md](../operations/resilience.md), and [../operations/deployment.md](../operations/deployment.md).
