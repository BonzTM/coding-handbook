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

## Common Mistakes And Forbidden Patterns

- Storing `context.Context` in a struct.
- Starting background goroutines from handlers without a supervisor.
- Using `context.Background()` deep inside the stack instead of propagating the caller's context.
- Adding concurrency to hide slow code before measuring or simplifying the design.
- Long critical sections under a mutex that include I/O.

## Verification And Proof

```bash
go test -race ./...
```

For workers and long-lived services, add proof beyond race detection:

- shutdown tests that cancel the root context and assert bounded exit
- timeout tests for external calls
- leak checks if the repo uses `goleak` or a similar helper

If you cannot explain who owns a goroutine and how it stops, the design is not done.
