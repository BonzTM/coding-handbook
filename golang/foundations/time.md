# Time

Time handling rules that keep behavior deterministic in tests, correct across zones, and free of clock-related bugs.

## Default Approach

Time is an input, not an ambient fact. Inject it, store it in UTC, measure spans with `time.Duration`, and cancel with context.

### Inject A Clock; Never Call time.Now() In Core Code

Domain and core code must take time as a dependency. Define a small `Clock` interface, wire the real clock in `main`, and pass a controllable fake in tests.

```go
// Clock is the source of time for code that must not read the wall clock directly.
type Clock interface {
	Now() time.Time
}

// systemClock is the production implementation, wired in main.
type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }
```

- Constructors accept a `Clock` (or a tighter interface) as an explicit dependency, like any other.
- `main` wires `systemClock{}`. No core package imports a global clock and no core package calls `time.Now()`, `time.Since`, or `time.Until` directly.
- Keep the interface minimal. Add `Until`, `After`, or a ticker factory only when a type actually needs them; a one-method `Now()` covers most code.
- The fake lives in `internal/testutil` so every package shares one controllable clock. See [../quality/testing.md](../quality/testing.md).

The fake exposes deterministic control:

```go
// FakeClock is a controllable Clock for tests. Advance time explicitly; never sleep.
type FakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func NewFakeClock(t time.Time) *FakeClock { return &FakeClock{now: t} }

func (c *FakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *FakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}
```

### Store And Transmit UTC; Localize Only At The Edge

- Persist, log, and serialize timestamps in UTC. Call `t.UTC()` before storing or marshaling.
- Convert to a location only at the presentation boundary (rendering for a human, formatting a report), and only with an explicit `*time.Location` you loaded deliberately.
- Never rely on the server's local time zone for business decisions. `time.Local` depends on deployment environment and is not a contract.
- Format machine-to-machine timestamps with `time.RFC3339Nano`; parse with the matching layout. Reject naive local timestamps at the edge.

### Compare Time Correctly: The Monotonic-Clock Gotcha

A `time.Time` from `time.Now()` carries two readings: a wall-clock component and a monotonic component. `Sub`, `Since`, and `Until` use the monotonic reading, so elapsed-time math stays correct even if the wall clock is adjusted. But the monotonic reading is fragile:

- It is stripped by round-tripping through any serialization (JSON, protobuf, database driver, `gob`).
- It is stripped by `Round`, `Truncate`, `UTC`, `Local`, `In`, and `AddDate`.
- A value built from `time.Date`, parsed from a string, or read back from storage has no monotonic reading at all.

This breaks `==`. The `==` operator compares all fields including the monotonic reading and the `*Location` pointer, so a value and its round-tripped copy are not `==` even when they name the same instant. Always compare instants with `Equal`:

```go
a := clock.Now()
b := a.Round(0) // strips the monotonic reading; same instant, different struct

a == b        // false: monotonic readings differ
a.Equal(b)    // true: same instant
```

- Compare instants with `t1.Equal(t2)`. Order with `Before`/`After`.
- Never use `==` on `time.Time` and never use it as a map key.
- To deliberately drop the monotonic reading (for stable logs or before storing), use `t.Round(0)`.
- In test assertions prefer `Equal`; reach for `cmp.Diff` only with `cmpopts.EquateApproxTime` or after stripping monotonic readings, since `cmp` defaults to field-wise comparison.

### Measure Spans With time.Duration

- Durations, timeouts, intervals, and ages are `time.Duration`, never raw `int` seconds or milliseconds. The type carries its unit and prevents unit-mismatch bugs.
- Config and flags parse human strings (`"30s"`, `"5m"`) into `time.Duration` via `time.ParseDuration` or the `flag` duration var. See [configuration.md](configuration.md).
- API and storage layers that must use integers should name the unit explicitly (`TimeoutMillis`) and convert at the boundary, not leak bare numbers inward.

### Cancel With Context, Not Sleeps Or Timers You Forget

Deadlines and timeouts are a cancellation concern. Express them through `context`, which propagates cancellation across the call tree; do not invent ad hoc timer plumbing.

- Bound external calls with `context.WithTimeout` / `context.WithDeadline` and always `defer cancel()`.
- In long-lived `select` loops, do not use `time.After`: each iteration allocates a timer that is not collected until it fires, leaking under high loop rates. Use a single `time.NewTimer`/`time.NewTicker` created once, `Stop` it on exit, and `Reset` per iteration, or select on `ctx.Done()`.

See [context-and-concurrency.md](context-and-concurrency.md) for the full propagation and ownership rules.

### Tests Are Deterministic And Sleepless

- Drive time with the fake clock. To test timeout, expiry, or scheduling behavior, call `Advance`; never `time.Sleep` to "wait for" something.
- A `time.Sleep` in a test is either a hidden race or wasted wall-clock time. Replace it with an injected clock plus a synchronization signal (channel, `sync.WaitGroup`, or `errgroup`).
- Make "now" an explicit fixture so golden output and time-derived IDs are reproducible.

## Common Mistakes And Forbidden Patterns

- Calling `time.Now()`, `time.Since`, or `time.Until` inside business or core logic instead of taking a `Clock`.
- Storing or transmitting local-zone timestamps; depending on `time.Local` for a business decision.
- Comparing `time.Time` with `==`, or using it as a map key, on values that may carry or lack a monotonic reading.
- Passing spans as raw `int` seconds/milliseconds instead of `time.Duration`.
- `time.Sleep` in tests to wait for asynchronous work or timer expiry.
- `time.After` inside a long-lived `select` loop: leaks timers until they fire.
- Creating a context timeout without `defer cancel()`, leaking the timer and goroutine.

## Verification And Proof

```bash
make verify   # the full gate; runs vet, test, and race
make race     # determinism check on its own during a tight loop
```

Time handling is done when:

- a `Clock` is injected everywhere time matters, and `grep` finds no `time.Now()` / `time.Since` / `time.Until` in core packages.
- timestamps are stored and transmitted in UTC, localized only at presentation.
- instants are compared with `Equal`/`Before`/`After`, never `==`.
- tests advance the fake clock instead of sleeping, run fast, and are deterministic across runs.
- `go test -race` is clean.

## Where To Go Next

- [context-and-concurrency.md](context-and-concurrency.md) — deadlines, timeouts, and goroutine ownership.
- [configuration.md](configuration.md) — parsing durations from env and flags.
- [../quality/testing.md](../quality/testing.md) — where the fake clock lives and how tests prove time-dependent behavior.
- [style-and-review.md](style-and-review.md) — the baseline idiom this doc deepens.
