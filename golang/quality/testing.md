# Testing

Testing strategy for Go repos that need trustworthy behavior, not just green checkmarks.

## Default Approach

Use the stdlib `testing` package first. Keep tests close to the code they prove.

### Test Taxonomy

| Test Type | Use for | Default tools |
|---|---|---|
| unit | pure business logic, edge cases, small adapters | `testing`, `t.Run` |
| transport | handler decoding, status mapping, middleware | `httptest`, fake services |
| integration | database, migrations, external clients, multi-package behavior | real DB or protocol test server |
| contract | payload compatibility, schema validation, event examples | schema fixtures, golden examples, generated validators when used |
| fuzz | parsers, decoders, protocol edges, untrusted input | `testing.F` |
| benchmark | hot paths, serialization, allocation-sensitive code | `testing.B` |

### Table-Driven Tests

Use table-driven tests when the behavior is truly the same shape repeated across inputs. Do not force everything into a table if branching setup becomes harder to read than separate tests.

### Integration Defaults

- Database code should get real database tests.
- External HTTP clients should usually use `httptest.Server` before they use mocks.
- Contract-heavy systems may need a smaller unit suite plus one or two high-value end-to-end tests.
- Event-driven systems should add duplicate-delivery, replay, retry-exhaustion, and ordering tests where those semantics matter.

### TDD Guidance

When practical, start behavior changes with the proving test first. The goal is not ritual; it is making the contract explicit before the implementation hardens.

### Determinism And Parallelism

A test that passes or fails depending on timing, ordering, scheduling, or wall-clock time is not proof. Make every test deterministic, then make independent ones parallel.

- Call `t.Parallel()` at the top of independent unit tests and at the top of subtests that do not share mutable state. Parallelism surfaces data races under `-race` and keeps the suite fast as it grows.
- Do not call `t.Parallel()` when a test mutates process-global state (env vars, `time.Local`, registered metrics, package-level singletons) or shares a fixture another test mutates. Isolate the state instead — that is the real fix.
- Subtests capture the loop variable. On Go 1.22+ the per-iteration loop variable is safe, so `tc := tc` is no longer required; on a 1.24+ floor the alias is dead code and should not be added. Still pass table cases by value into the subtest closure and never let two parallel subtests write the same map, slice, or pointer.
- Never use real `time.Sleep`, busy-wait loops, or wall-clock deadlines to "let work happen." A sleep is either a hidden race (too short) or wasted suite time (too long). Inject a fake `Clock` and `Advance` it; see [../foundations/time.md](../foundations/time.md), where the shared `FakeClock` in `internal/testutil` lives.
- Synchronize with the thing you are waiting on, not the clock: a channel the worker closes, a `sync.WaitGroup`, an `errgroup`, or `testing/synctest` (Go 1.24+, `GOEXPERIMENT=synctest` on 1.24; stable in 1.25) to run a bubble of goroutines on a fake clock with deterministic scheduling.
- Make inputs deterministic: seed any `rand` source explicitly, sort before comparing unordered collections, and pin "now" to a fixed fixture so golden output and time-derived IDs reproduce. Map iteration order is random by contract — never assert on it.

### Leak Detection

Goroutines that outlive the test they were started in are bugs that hide until production. Detect them in the suite.

- For any package that spawns goroutines (servers, workers, pools, background loops), assert no goroutine leaks. Use `goleak.VerifyTestMain(m)` from a `TestMain` for whole-package coverage, or `defer goleak.VerifyNone(t)` for a single high-risk test. The library is routed via [../decisions/framework-selection.md](../decisions/framework-selection.md).
- A leak check proves shutdown actually stops what `Run` started: cancel the context, wait for the run loop to return, and `goleak` confirms nothing is still parked on a channel, timer, or blocking read.
- Pair leak detection with `-race`. Leaks and races are different failures — one is a goroutine that never exits, the other is concurrent unsynchronized access — and a correct concurrent type must survive both.
- Allowlist only framework goroutines you do not own (`goleak.IgnoreTopFunction`) and comment why. An open-ended allowlist defeats the check.

### Fixtures And Golden Files

Prefer small inline fixtures. Reach for files only when the input or expected output is large enough that inlining hurts readability.

- Test inputs and golden outputs live in a `testdata/` directory beside the test. The Go toolchain ignores `testdata/`, so it never affects builds.
- Compare golden output with `go-cmp` (`cmp.Diff`), not string equality — a diff localizes the mismatch instead of dumping two blobs.
- Regenerate goldens behind a flag, never by hand:

```go
var update = flag.Bool("update", false, "update golden files")

func TestRender(t *testing.T) {
	got := render(input)
	golden := filepath.Join("testdata", "render.golden")
	if *update {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(string(want), string(got)); diff != "" {
		t.Errorf("render mismatch (-want +got):\n%s", diff)
	}
}
```

- Run `go test ./... -run TestRender -update` to refresh, then review the diff in code review like any other change — a golden update is a contract change, not a formality.
- Keep fixtures small, named for the case they prove, and free of secrets or PII. A 4,000-line fixture nobody can read is worse than three focused ones.

### Assertions And Comparison

Reach for the lightest tool that gives a clear failure message.

- Stdlib `testing` first: `if got != want { t.Errorf(...) }` for scalars and simple cases. No dependency, no indirection.
- `go-cmp` (`cmp.Diff`) for structs, slices, and maps — it prints a minimal diff. Configure semantics explicitly with `cmpopts` (`EquateApproxTime`, `EquateEmpty`, `SortSlices`, `IgnoreFields`) rather than pre-massaging values. For `time.Time`, prefer `Equal` or `cmpopts.EquateApproxTime`; raw `cmp` compares monotonic readings and `*Location` (see [../foundations/time.md](../foundations/time.md)).
- `testify/require` where it sharpens signal — `require.NoError`, `require.ErrorIs`, table assertions that should halt on first failure. Use `require` (stops the test) for preconditions and `assert` (continues) only when collecting multiple independent checks helps.
- All three are routed via [../decisions/framework-selection.md](../decisions/framework-selection.md); pick per repo, do not mix three assertion styles in one package.
- Assert on errors by identity and type: `errors.Is`/`errors.As`, not substring matching on `err.Error()`. Message text is not a contract; sentinel errors and types are.
- Avoid opaque assertion DSLs and deep matcher chains. A reader should see exactly what is compared without learning a sublanguage.

### Coverage Policy

Coverage is a floor and a map of untested code, never a target. A high number with the error paths untested is a lie.

- Mandatory paths MUST be covered: the domain core, every error and status-mapping branch, and every decode/parse path that accepts untrusted input. These are where defects ship and where regressions hide.
- Vanity paths are not chased: generated code, `main` wiring, trivial getters, and `String()` methods do not need contrived tests to lift a number.
- Measure with atomic counters so parallel and race runs are accurate:

```bash
go test -covermode=atomic -coverprofile=cover.out ./...
go tool cover -func=cover.out      # per-function summary
```

- Merge unit and integration coverage with the binary coverage format. Build instrumented test binaries (or run integration suites) writing to `GOCOVERDIR`, then merge:

```bash
GOCOVERDIR=./covdata go test -covermode=atomic ./...   # unit, emitting raw profiles
# run integration suite / instrumented binary, also writing to ./covdata
go tool covdata textfmt -i=./covdata -o=cover.out       # merge into one profile
go tool covdata percent -i=./covdata                    # combined percentage
```

- CI posture is no-regression / ratchet: coverage may not drop below the recorded baseline, and the baseline only moves up. A PR that adds an untested branch fails the gate.
- Coverage runs via `make cover`, which is NOT part of `make verify`. `make verify` is the correctness gate (tidy/fmt-check/lint/vet/test/race/vuln/build); coverage is a separate, explicit step so it never silently gates merges on a number. See [../operations/ci-and-release.md](../operations/ci-and-release.md).
- Reference: the example service ([../reference/exampleservice/](../reference/exampleservice/)) sits at 69.7% overall, with `core`, `api/http`, and `config` well covered and the uncovered remainder being `main` wiring and generated code — a deliberate shape, not a target missed.

### Test Organization And Naming

- Name tests `TestXxx` where `Xxx` is the function or behavior under test; name subtests for the case (`t.Run("expired token", ...)`), so `-run 'TestAuth/expired_token'` selects exactly one case.
- Use an external `_test` package (`package foo_test`) for black-box tests that exercise the exported API the way callers do. It prevents tests from reaching into unexported internals and keeps the public contract honest. Keep internal `package foo` tests only for genuinely unexported logic that cannot be reached otherwise.
- Write `Example_` and `ExampleXxx` functions for runnable documentation: they appear in godoc and fail the suite if the `// Output:` comment drifts. They are the cheapest defense against stale docs. The godoc side of this contract lives in [../foundations/style-and-review.md](../foundations/style-and-review.md).
- Separate fast unit runs from real-DB/integration runs explicitly. Either gate integration tests behind a build tag (`//go:build integration`) or skip them under `-short`:

```go
func TestRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	// ... real DB ...
}
```

  Run fast tests with `go test -short ./...` in the inner loop; run the full suite (including integration) in `make verify` and CI. Pick one mechanism per repo and apply it consistently.

### Property-Based And Mutation Testing

A stance, not a mandate.

- Reach for property-based testing (e.g. `rapid`) when the input space is large and an invariant should hold across all of it: parsers and encoders (round-trip `decode(encode(x)) == x`), normalizers (idempotence), comparators (ordering laws), and state machines. A property test plus a generator finds edge cases tables never enumerate. Keep a failing seed as a regression unit test once found.
- Mutation testing is optional signal for critical packages: it mutates the code and checks your tests catch it, exposing assertions that never actually constrain behavior. Run it occasionally on the domain core or a security-sensitive package, not on every build.
- Both tools are routed via [../decisions/framework-selection.md](../decisions/framework-selection.md). Adopt them where the invariant or the risk justifies the cost; do not bolt property tests onto code whose behavior a small table proves completely.

### Load And Soak Testing

Functional tests prove correctness; load and soak tests prove the service survives sustained traffic and produce the numbers SLOs are written against.

- Load test when capacity is a contract: before setting or revising an SLO, before a launch, or when a change alters the hot path (new dependency, larger payloads, added serialization). Drive realistic request mixes and measure latency percentiles and error rate at the target throughput.
- Soak test for leaks and slow degradation: run sustained load for hours and watch goroutine count, heap, and RSS for monotonic growth. Flat curves prove no leak; a rising goroutine count points at the same shutdown/ownership bugs `goleak` catches in the small.
- These runs feed the capacity and headroom numbers that [../operations/operability.md](../operations/operability.md) and [../operations/resilience.md](../operations/resilience.md) depend on — saturation thresholds, autoscaling targets, timeout and concurrency limits. Record the results next to the SLO so the budget has evidence behind it.
- Keep load harnesses out of `go test`. They are operational tooling, run against a deployed instance or a dedicated environment, not part of `make verify`.

### Eventing-Specific Proof

- Contract tests prove payload shape, metadata, and compatibility expectations.
- Duplicate-delivery tests prove idempotency.
- Replay and out-of-order tests prove the real ordering contract instead of an assumed one.
- DLQ or parked-message tests prove terminal failures stop retrying and preserve operator context.

## Common Mistakes And Forbidden Patterns

- Mocking repositories, HTTP clients, and workers so aggressively that no boundary behavior is exercised.
- Mocking brokers so heavily that replay, settlement, duplicate delivery, or exhaustion behavior is never exercised.
- Treating coverage percentage as the goal instead of edge-case confidence.
- Huge shared fixtures that make failures opaque.
- Benchmarks without stable inputs or with hidden setup inside the hot loop.
- Fuzz targets that mutate shared state or depend on external side effects.
- `time.Sleep` or wall-clock waits in tests to "let async work finish" instead of an injected clock plus a synchronization signal.
- `t.Parallel()` on a test that mutates process-global state or shares a fixture another test writes.
- Asserting on map iteration order, unsorted slice order, or `err.Error()` substrings instead of `errors.Is`/`errors.As`.
- Comparing golden output with string equality instead of `cmp.Diff`, or hand-editing golden files instead of regenerating behind `-update`.
- Packages that spawn goroutines but never assert against leaks with `goleak`.
- Mixing three assertion styles in one package, or hiding comparisons behind an opaque matcher DSL.
- Treating the coverage number as the deliverable, or letting it ratchet down to merge a change.
- Running real-DB or integration tests in the fast inner loop because they were never gated behind a build tag or `-short`.

## Verification And Proof

```bash
go test ./...                  # fast feedback
go test -short ./...            # fast unit subset, integration skipped
go test -race ./...            # races and leaks under concurrency
make verify                    # the gate: includes test + race
make cover                     # coverage, separate from the gate
```

Add the following when relevant:

- `t.Parallel()` on independent tests so `-race` exercises real concurrency
- `goleak.VerifyTestMain` (or `VerifyNone`) for packages that spawn goroutines, paired with `-race`
- fuzz targets for parsing or untrusted input paths
- event handlers and decoders that accept untrusted payloads are good fuzz targets when the decode path is non-trivial
- golden files in `testdata/` compared with `cmp.Diff`, regenerated behind `-update`
- integration suites against a real database or test server, gated behind `//go:build integration` or `testing.Short()`
- `go test -covermode=atomic -coverprofile=cover.out ./...`, merging unit and integration coverage via `GOCOVERDIR` plus `go tool covdata`, against a no-regression baseline
- property-based tests (e.g. `rapid`) for parsers, encoders, and invariants over large input spaces
- benchmarks compared with a stable baseline when performance is part of the change
- load and soak runs feeding capacity numbers when an SLO or the hot path changes
- contract validation plus replay, duplicate-delivery, and DLQ tests when the repo publishes or consumes messages

Testing is done when the chosen proof matches the risk of the change, not when a single unit suite turns green.
