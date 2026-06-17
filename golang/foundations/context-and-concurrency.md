# Context and Concurrency

Context propagation, cancellation, and concurrency ownership rules for Go code that behaves correctly under load and on shutdown.

## Default Approach

- Pass `ctx context.Context` as the first argument to every function performing I/O, waiting on remote work, or coordinating long-lived tasks.
- Use `r.Context()` in HTTP handlers and request-scoped server code.
- Use `signal.NotifyContext` in `main` to create the process root context.
- Use `errgroup.Group` when sibling goroutines must fail or stop together.

### Primitive Selection

| Use this | When | Avoid when |
|---|---|---|
| channel | coordination, fan-out/fan-in, ownership transfer | protecting shared mutable state is the main job |
| `sync.Mutex` / `sync.RWMutex` | protecting in-memory shared state | the workflow is better modeled as messages |
| `sync/atomic` | simple counters or flags with proven contention needs | the state is richer than a single word |
| `errgroup.Group` | parallel tasks that share cancellation and error handling | the work is fire-and-forget or independently supervised |

### Ownership Rules

- Every goroutine needs an owner: request, worker supervisor, server, or application root.
- Every goroutine needs a stop condition: context cancellation, closed channel, or finite loop.
- Every external call inside a goroutine needs a timeout budget.

### Graceful Shutdown And Draining

The root context is cancelled by `signal.NotifyContext` on the first SIGINT/SIGTERM. Shutdown is then ordered and bounded — release nothing until the work that depends on it has drained. The canonical implementation is [templates/cmd-app-main.go.txt](../templates/cmd-app-main.go.txt) (`shutdown`); the sequence is:

1. **Flip readiness to UNREADY.** The load balancer / orchestrator stops routing new traffic on the next probe while in-flight requests keep flowing. Readiness is distinct from liveness; do not fail liveness here or the platform kills the pod mid-drain. See [observability](../operations/observability.md) for the probe split.
2. **Stop accepting new connections and drain.** Call `srv.Shutdown(ctx)` for HTTP, bounded by a **fresh** `context.WithTimeout(context.Background(), grace)` — NOT the already-cancelled root context, which would make `Shutdown` return instantly and abandon in-flight requests. `ListenAndServe` returns `http.ErrServerClosed`, which is the expected clean outcome, not an error.
3. **Cancel the root/worker context** so background loops, pollers, and `errgroup` siblings observe `<-ctx.Done()` and exit. In the template this cancellation is what triggered the shutdown supervisor in the first place; any remaining detached loops must select on the same root context.
4. **Close the DB pool and external clients** only now — once no request or worker can still be holding a connection.
5. **Flush telemetry LAST** so every step above is recorded; flushing first loses the shutdown trace.

gRPC analog: prefer `grpcServer.GracefulStop()`, which stops accepting new RPCs and waits for in-flight ones. Because `GracefulStop` ignores the grace deadline, run it in a goroutine and fall back to `grpcServer.Stop()` (hard close) when the grace timer fires:

```go
done := make(chan struct{})
go func() { grpcServer.GracefulStop(); close(done) }()
select {
case <-done:
case <-ctx.Done(): // grace exceeded
	grpcServer.Stop()
}
```

The grace period is a config key (`cfg.ShutdownGrace`, loaded per [configuration](configuration.md)) and MUST exceed in-flight work's worst-case duration AND stay under the platform's termination grace (e.g. Kubernetes `terminationGracePeriodSeconds`) so the bounded `Shutdown` wins before the platform sends SIGKILL. A second signal must hard-kill: `signal.NotifyContext`'s `stop()` (deferred in `main`) restores default signal handling, so a second SIGINT/SIGTERM terminates the process even if a drain hangs.

## Common Mistakes And Forbidden Patterns

- Storing `context.Context` in a struct.
- Starting background goroutines from handlers without a supervisor.
- Using `context.Background()` deep inside the stack instead of propagating the caller's context.
- Adding concurrency to hide slow code before measuring or simplifying the design.
- Long critical sections under a mutex that include I/O.
- Closing the DB pool or external clients before in-flight requests have drained, so requests fail on a dead connection mid-shutdown.
- Passing the already-cancelled root context to `srv.Shutdown` instead of a fresh `context.WithTimeout`; it returns immediately and drops in-flight work.
- An unbounded drain (no grace deadline, or a grace longer than the platform termination grace) so shutdown hangs until SIGKILL.
- Calling `grpcServer.Stop()` directly on signal, killing in-flight RPCs instead of `GracefulStop()` with a bounded fallback.
- Flushing telemetry first, or skipping it, so the shutdown sequence is unobservable.

## Verification And Proof

```bash
go test -race ./...
```

For workers and long-lived services, add proof beyond race detection:

- shutdown tests that cancel the root context and assert bounded exit
- a shutdown smoke test: start the server, open a slow in-flight request, send SIGTERM, assert the in-flight request completes (no dropped or connection-reset responses) and the process exits within the grace budget
- assert readiness flips to unready before `Shutdown` returns, and that the grace deadline upper-bounds total shutdown time
- timeout tests for external calls
- `make race` plus leak checks (`goleak` or similar) to prove no goroutine survives shutdown

If you cannot explain who owns a goroutine and how it stops, the design is not done.
