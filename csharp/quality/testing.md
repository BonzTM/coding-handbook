# Testing

Testing strategy for .NET repos that need trustworthy behavior, not just green checkmarks.

## Default Approach

Use xUnit v3 running on Microsoft.Testing.Platform, with plain xUnit `Assert`. Tests live in the `tests/` projects the solution layout mandates — `<App>.UnitTests` and `<App>.IntegrationTests` (see [../foundations/solution-and-project-design.md](../foundations/solution-and-project-design.md)) — and the test projects copy their setup (`UseMicrosoftTestingPlatformRunner`, runner config) from the reference module's test projects ([../reference/exampleservice/tests/](../reference/exampleservice/tests/)), not from memory.

### Test Taxonomy

| Test Type | Use for | Default tools |
|---|---|---|
| unit | pure domain logic in `<App>.Core`, edge cases, small adapters | xUnit `[Fact]`/`[Theory]`, plain `Assert` |
| transport | endpoint routing, request decoding, status mapping, endpoint filters, `ProblemDetails` shape | `WebApplicationFactory`, hand-rolled fakes behind Core ports |
| integration | database, EF Core migrations, broker, external clients | Testcontainers, behind the explicit `-Integration` switch |
| contract | wire payload compatibility, golden JSON shapes, event examples | golden files, `JsonSerializerContext` round-trips ([../foundations/serialization.md](../foundations/serialization.md)) |
| property | parsers, encoders, invariants over large input spaces | FsCheck via ADR |
| benchmark | hot paths, serialization, allocation-sensitive code | BenchmarkDotNet via ADR |

### Theories And Data-Driven Tests

Use `[Theory]` when the behavior is truly the same shape repeated across inputs: `[InlineData]` for scalar cases, `[MemberData]` with a strongly-typed `TheoryData<T>` for structured cases. Do not force everything into a theory — if per-case setup branches, separate `[Fact]`s read better than a theory full of flags.

```csharp
public sealed class OrderNumberTests
{
    [Theory]
    [InlineData("ORD-000123", true)]
    [InlineData("ord-000123", false)] // prefix is case-sensitive by contract
    [InlineData("ORD-", false)]
    [InlineData("", false)]
    public void TryParse_Candidate_ReportsValidity(string candidate, bool valid)
    {
        Assert.Equal(valid, OrderNumber.TryParse(candidate, out _));
    }
}
```

### Integration Defaults

- Database code gets real database tests: Testcontainers (`Testcontainers.PostgreSql`) starts a disposable Postgres, the fixture applies the EF Core migrations, and the tests exercise the real provider — not the InMemory provider, which lies about translation, transactions, and constraints (see [../services/database.md](../services/database.md)).
- Integration tests live in `<App>.IntegrationTests` and run only behind an explicit switch — `pwsh ./verify.ps1 -Integration` locally, a dedicated CI job with Docker in the matrix — because Docker is not guaranteed on every dev machine. The default `pwsh ./verify.ps1` gate runs unit tests only. When the switch is passed and Docker is missing, the run fails loudly; it never silently skips. See [../operations/ci-and-release.md](../operations/ci-and-release.md).
- External HTTP clients get tests against an in-proc stub server or a hand-rolled `HttpMessageHandler` fake before anyone reaches for a mock of the typed client interface.
- Event-driven systems add duplicate-delivery, replay, retry-exhaustion, and ordering tests where those semantics matter (see [Eventing-Specific Proof](#eventing-specific-proof) and [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md)).

```csharp
public sealed class PostgresFixture : IAsyncLifetime
{
    private readonly PostgreSqlContainer _container = new PostgreSqlBuilder().Build();

    public string ConnectionString => _container.GetConnectionString();

    public async ValueTask InitializeAsync()
    {
        await _container.StartAsync(TestContext.Current.CancellationToken);
        await using var db = OrdersDbContextFactory.Create(ConnectionString);
        await db.Database.MigrateAsync(TestContext.Current.CancellationToken);
    }

    public async ValueTask DisposeAsync() => await _container.DisposeAsync();
}

[CollectionDefinition(nameof(PostgresFixture))]
public sealed class PostgresCollection : ICollectionFixture<PostgresFixture>;
```

One container per collection, migrations applied once, each test isolating its own data (unique keys or per-test schema) — not one container per test, which turns the suite into a Docker stress test.

### Test Doubles

The default test double is a hand-rolled fake implementing a port that `<App>.Core` already declares (see [../foundations/shared-constructs.md](../foundations/shared-constructs.md)). A fake is a real, simple implementation — a `Dictionary` behind a lock, a fixed clock, an in-memory outbox. It survives refactors because it asserts on observable behavior, not on which methods were called in what order.

- Reach for NSubstitute only when a fake is genuinely too costly — a wide third-party interface you do not own, or an interface churning too fast to keep a fake current. It is the second choice, not the default.
- **Moq is forbidden** (the SponsorLink incident showed the project will ship surprising payloads; NSubstitute covers every legitimate use). **FluentAssertions is forbidden** (post-v8 commercial licensing; plain xUnit `Assert` plus `Assert.Equivalent` covers the need). Neither gets an ADR path.
- Over-specified substitutes (`Received(1)` on every call) couple the test to the implementation: the test fails on harmless refactors and passes on real regressions. Assert outcomes, not call scripts.
- Shared fakes and builders live in the `<App>.TestUtilities` project once two or more test projects need them.

### TDD Guidance

When practical, start behavior changes with the proving test first. The goal is not ritual; it is making the contract explicit before the implementation hardens.

### Determinism And Parallelism

A test that passes or fails depending on timing, ordering, scheduling, culture, or wall-clock time is not proof. Make every test deterministic first; xUnit's defaults then give you parallelism for free.

- xUnit v3 runs test **collections** in parallel and tests within a collection serially; by default every test class is its own collection. Leave that on. Tests that must share expensive state (a database container) join a named collection via `ICollectionFixture<T>`, which also serializes them against each other.
- Never disable assembly-wide parallelism (`parallelizeTestCollections: false` in `xunit.runner.json`) to paper over shared mutable state. Isolate the state instead — that is the real fix.
- Do not mutate process-global state in parallel tests: environment variables, `CultureInfo.CurrentCulture`, static singletons, the current directory. A test that truly must touch one gets its own collection marked `[CollectionDefinition(..., DisableParallelization = true)]` with a comment saying why.
- Never use `Task.Delay`, `Thread.Sleep`, or wall-clock polling to "let async work finish." A sleep is either a hidden race (too short) or wasted suite time (too long). Synchronize with the thing you are waiting on — a `TaskCompletionSource` the fake completes, the worker's returned `Task`, `Channel` completion — and drive time through `FakeTimeProvider`.
- Pass `TestContext.Current.CancellationToken` into every async call a test makes, so a hung test is cancelled by the runner instead of wedging the suite.
- Make inputs deterministic: seed `Random` explicitly (or take an injected seed), sort before comparing unordered collections, and pin "now" via `FakeTimeProvider` so time-derived output reproduces.
- The suite must pass on ubuntu, windows, and macos — the CI matrix runs all three. That means `Path.Combine` for paths, `StringComparison.Ordinal` and `CultureInfo.InvariantCulture` for machine-facing comparisons, `ReplaceLineEndings("\n")` before comparing multi-line text, and file names that match case exactly (see [../foundations/cross-platform.md](../foundations/cross-platform.md)).

### Clock Control

`TimeProvider` is injected everywhere per [../foundations/time.md](../foundations/time.md); tests substitute `FakeTimeProvider` (from `Microsoft.Extensions.TimeProvider.Testing`) and advance it explicitly. `Advance` fires timers and `Task.Delay` continuations scheduled against the provider, so expiry logic is proven in microseconds of real time:

```csharp
[Fact]
public async Task IsExpired_ReservationPastTtl_ReturnsTrue()
{
    var clock = new FakeTimeProvider(DateTimeOffset.Parse("2026-07-10T00:00:00Z", CultureInfo.InvariantCulture));
    var service = new ReservationService(new FakeReservationStore(), clock);
    await service.ReserveAsync("order-1", TestContext.Current.CancellationToken);

    clock.Advance(TimeSpan.FromMinutes(16)); // reservation TTL is 15 minutes

    Assert.True(await service.IsExpiredAsync("order-1", TestContext.Current.CancellationToken));
}
```

A test that constructs `DateTime.Now`, `DateTimeOffset.UtcNow`, or `Stopwatch` directly is a determinism bug even if it currently passes.

### Leak Detection

.NET has no `goleak`; be honest about what replaces it. The leak classes are undisposed `IDisposable`/`IAsyncDisposable` resources, background `Task`s that outlive their owner, timers, and `CancellationTokenRegistration`s — and the defense is layered, not a single assertion:

- **Static analysis first**: the dispose-tracking analyzers (`CA2000`, `CA1001`, `CA2213`) run in every build via [linting.md](linting.md) and catch the mechanical cases at compile time.
- **Deterministic disposal in tests**: every fixture and test disposes what it opens — `using`/`await using` on clients, responses, containers, and `CancellationTokenSource`s. This is also a cross-platform requirement: Windows locks open files, so a leaked handle turns into a flaky delete on the CI windows leg.
- **Shutdown proof for owned background work**: any type that owns long-lived work (a `BackgroundService`, an outbox dispatcher, a consumer loop) MUST have a test that starts it, stops it, and asserts bounded completion. This is the .NET counterpart of a goleak shutdown test — it proves cancellation actually stops what `ExecuteAsync` started:

```csharp
[Fact]
public async Task StopAsync_RunningDispatcher_CompletesWithinBudget()
{
    var worker = new OutboxDispatcher(new FakeOutboxStore(), new FakeTimeProvider(), NullLogger<OutboxDispatcher>.Instance);
    await worker.StartAsync(TestContext.Current.CancellationToken);

    using var budget = new CancellationTokenSource(TimeSpan.FromSeconds(5));
    await worker.StopAsync(budget.Token);

    Assert.True(worker.ExecuteTask is { IsCompletedSuccessfully: true });
}
```

- **Soak-time observation for the rest**: slow leaks that no unit assertion can see are caught by watching `dotnet-counters` during soak runs — threadpool queue length, timer count, gen 2 heap size, working set — for monotonic growth, plus `dotnet-gcdump` when a curve rises. The `System.Diagnostics` metrics the service already emits ([../operations/observability.md](../operations/observability.md)) show the same curves in production.

### Fixtures, Builders And Golden Files

Prefer small inline fixtures. Reach for files only when the input or expected output is large enough that inlining hurts readability.

- Per-class setup is the constructor plus `IDisposable`/`IAsyncLifetime`; shared expensive setup is `IClassFixture<T>` (per class) or `ICollectionFixture<T>` (across classes). Async setup implements `IAsyncLifetime` — never `async void` helpers or `.GetAwaiter().GetResult()` in constructors.
- Test-data builders live in `<App>.TestUtilities`: a static factory returning a valid aggregate, with optional mutation hooks — `TestOrders.Canonical(o => o.Status = OrderStatus.Shipped)`. Use a builder when tests need many variants of a valid aggregate; object initializers on the DTO remain right for small, simple inputs.
- Small, stable golden payloads belong inline: a short raw string literal (`"""..."""`) compared with `Assert.Equal` pins a wire shape with zero machinery.
- Large or frequently-regenerated outputs move to golden files: they live in a `TestData/` directory beside the test (copied to output via a `<Content CopyToOutputDirectory="PreserveNewest">` item), are regenerated behind an explicit environment variable — never by hand — and are line-ending-normalized before comparison so the ubuntu/windows legs agree:

```csharp
[Fact]
public async Task Serialize_CanonicalOrder_MatchesGolden()
{
    var got = JsonSerializer.Serialize(TestOrders.Canonical(), OrdersJsonContext.Default.OrderResponse);
    var path = Path.Combine(AppContext.BaseDirectory, "TestData", "order-response.golden.json");

    if (Environment.GetEnvironmentVariable("UPDATE_GOLDEN") == "1")
    {
        await File.WriteAllTextAsync(path, got, TestContext.Current.CancellationToken);
    }

    var want = await File.ReadAllTextAsync(path, TestContext.Current.CancellationToken);
    Assert.Equal(want.ReplaceLineEndings("\n"), got.ReplaceLineEndings("\n"));
}
```

- Regenerating writes to the build-output copy; copy the diff back to the source `TestData/` file and review it in code review like any other change — a golden update is a contract change, not a formality.
- This plain-golden-file approach is the default. The Verify snapshot library is an ADR-level adoption via [../decisions/framework-selection.md](../decisions/framework-selection.md) — justified when snapshot volume is high enough that received/verified tooling pays for the dependency, not before.
- Keep fixtures small, named for the case they prove, and free of secrets or PII. A 4,000-line fixture nobody can read is worse than three focused ones.

### Assertions And Comparison

Plain xUnit `Assert` is the only assertion vocabulary. No fluent DSLs, no matcher chains — a reader should see exactly what is compared without learning a sublanguage.

- `Assert.Equal` for scalars, strings, and collections (it compares sequences element-wise); `Assert.Equivalent` for structural comparison of object graphs where reference identity does not matter — this covers the case people reach for FluentAssertions to solve.
- Assert on exceptions by type and properties, never by message substring: `var ex = await Assert.ThrowsAsync<OrderNotFoundException>(...)` then assert on `ex.OrderId`. Message text is not a contract; exception types and typed error results are (see [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md)).
- At the HTTP boundary, assert the `ProblemDetails` contract — status code, `application/problem+json` content type, and the fields that matter — not raw body substrings.
- `Assert.Collection` when both order and per-element shape are the contract; sort first and `Assert.Equal` when order is not.

### Coverage Policy

Coverage is a floor and a map of untested code, never a target. A high number with the error paths untested is a lie.

- Mandatory paths MUST be covered: the domain core, every error and status-mapping branch, and every parse/validation path that accepts untrusted input. These are where defects ship and where regressions hide.
- Vanity paths are not chased: `Program.cs` wiring, EF Core migrations, source-generated code, and trivial property bags do not need contrived tests to lift a number.
- Collect with `Microsoft.Testing.Extensions.CodeCoverage` — the coverage extension native to Microsoft.Testing.Platform, so it needs no separate collector process (coverlet is not the default here):

```bash
dotnet test --project tests/Orders.UnitTests -- --coverage --coverage-output-format cobertura
```

- Posture is no-regression / ratchet: coverage may not drop below the recorded baseline, and the baseline only moves up. The ratchet is a review policy, not automation — the baseline is recorded in the repo's docs and reviewers hold the line.
- Coverage is a separate, explicit step, NOT part of `pwsh ./verify.ps1`. The gate is correctness — restore (locked), format-check, build (warnings-as-errors), test, audit — and coverage never silently gates merges on a number. See [../operations/ci-and-release.md](../operations/ci-and-release.md).

### Test Organization And Naming

The naming convention is **`Method_Scenario_Expectation`** — `Cancel_AlreadyShipped_ReturnsConflict`, `TryParse_EmptyInput_ReturnsFalse`. It is chosen over given/when/then because it leads with the member under test, so failures group and grep by API surface; do not mix the two styles in one repo.

- One test class per type or endpoint under test, named `<Subject>Tests` (`OrderServiceTests`, `CancelOrderEndpointTests`), in a namespace mirroring the source project's.
- Test the public contract the way callers use it. `InternalsVisibleTo` is granted only to the matching test project (the boundary rule from [../foundations/solution-and-project-design.md](../foundations/solution-and-project-design.md)) and is for genuinely internal logic that cannot be reached otherwise — not a license to test private implementation details.
- Fast/slow separation is structural, not attribute-based: unit tests live in `<App>.UnitTests` (always in the gate), integration tests in `<App>.IntegrationTests` (behind `-Integration`). No `[Trait("Category", ...)]` filtering games — the project split is the filter, and it cannot silently drift.
- Do not expose private helpers only to satisfy tests; if a function is hard to test, refactor until the behavior is reachable through a narrow public boundary.

### Property-Based And Mutation Testing

A stance, not a mandate.

- Reach for property-based testing (FsCheck, routed via [../decisions/framework-selection.md](../decisions/framework-selection.md)) when the input space is large and an invariant should hold across all of it: parsers and encoders (round-trip `Parse(Format(x)) == x`), normalizers (idempotence), comparers (ordering laws), state machines. Keep any failing case FsCheck finds as a plain `[Fact]` regression test once shrunk.
- Mutation testing with Stryker.NET is optional signal for critical projects: it mutates the code and checks the tests catch it, exposing assertions that never actually constrain behavior. Run `dotnet stryker` occasionally against `<App>.Core` or a security-sensitive project, not on every build — it is far too slow for the gate.
- Both are ADR-level adoptions. Do not bolt property tests onto code whose behavior a small theory proves completely.

### Transport Tests With WebApplicationFactory

In-proc HTTP tests via `Microsoft.AspNetCore.Mvc.Testing` prove routing, binding, validation, endpoint filters, status mapping, and the `ProblemDetails` envelope — with fakes substituted at the Core ports so no real infrastructure runs. `Program.cs` exposes itself to the factory with `public partial class Program;` (see the [program-main template](../templates/program-main.cs.txt)).

```csharp
public sealed class OrdersApiFactory : WebApplicationFactory<Program>
{
    protected override void ConfigureWebHost(IWebHostBuilder builder)
    {
        builder.ConfigureServices(services =>
        {
            services.RemoveAll<IOrderStore>();
            services.AddSingleton<IOrderStore, FakeOrderStore>();
            services.AddSingleton<TimeProvider>(new FakeTimeProvider());
        });
    }
}

public sealed class GetOrderEndpointTests(OrdersApiFactory factory) : IClassFixture<OrdersApiFactory>
{
    [Fact]
    public async Task GetOrder_UnknownId_Returns404ProblemDetails()
    {
        using var client = factory.CreateClient();

        using var response = await client.GetAsync(
            $"/orders/{Guid.Empty}", TestContext.Current.CancellationToken);

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
        Assert.Equal("application/problem+json", response.Content.Headers.ContentType?.MediaType);
    }
}
```

These are unit-speed tests and belong in `<App>.UnitTests` (or a `Transport/` folder within it) — the factory hosts the app in-process with `TestServer`, no sockets, no Docker. Every error branch of an endpoint gets one: validation failure (400 with `errors` map), not-found (404), conflict (409), each asserting the ProblemDetails contract from [../services/http-services.md](../services/http-services.md).

### Benchmarks And Profiling

Benchmarks are proof for performance claims, never vibes.

- BenchmarkDotNet is the benchmark harness, routed via [../decisions/framework-selection.md](../decisions/framework-selection.md). Benchmarks live in a separate console project outside `tests/`, always run in Release, always with `[MemoryDiagnoser]` so allocations are part of the result.
- Never benchmark with `Stopwatch` loops in a unit test — no warmup, no statistics, JIT and tiering noise. And never let a "performance test" with a wall-clock assertion into the gate; it is flaky by construction.
- Benchmarks run on demand (before/after a performance-relevant change), not in `pwsh ./verify.ps1`. Compare against a baseline run on the same machine, not a number from prose.
- Profile with `dotnet-trace` (CPU), `dotnet-counters` (live metrics), and `dotnet-gcdump` (heap) against a Release build. Production profiling is a deploy-time decision made with the operations owner, not a code default.

### Load And Soak Testing

Functional tests prove correctness; load and soak tests prove the service survives sustained traffic and produce the numbers SLOs are written against.

- Load test when capacity is a contract: before setting or revising an SLO, before a launch, or when a change alters the hot path. Drive realistic request mixes and measure latency percentiles and error rate at the target throughput.
- Soak test for leaks and slow degradation: run sustained load for hours and watch the `dotnet-counters` curves (threadpool queue, gen 2 heap, working set, timer count) for monotonic growth. Flat curves prove no leak; a rising curve points at the same disposal/ownership bugs [Leak Detection](#leak-detection) catches in the small.
- These runs feed the capacity and headroom numbers that [../operations/operability.md](../operations/operability.md) and [../operations/resilience.md](../operations/resilience.md) depend on. Record the results next to the SLO so the budget has evidence behind it.
- Keep load harnesses out of the test projects. They are operational tooling run against a deployed instance, never part of `pwsh ./verify.ps1`. The driver tool is per-repo choice via ADR.

### End-To-End And Smoke

A handful of end-to-end tests prove the assembled system works; they never replace unit coverage.

- Keep them few — one or two high-value paths, not a parallel suite. Every behavior they touch is already unit-proven; e2e only proves the wiring.
- They exercise the real published binary (or container image) against real local dependencies via the repo's [compose stack](../templates/docker-compose.yml): start the service, hit real endpoints including `/livez` and `/readyz`, assert real responses.
- They live in `<App>.IntegrationTests` (or a small `e2e/` script the release checklist runs), excluded from the default gate, and run in CI's integration lane or pre-release.
- The shutdown smoke test — SIGTERM drains in-flight requests within `HostOptions.ShutdownTimeout` ([../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md)) — is one of these, and the release-time smoke items in [../checklists/release.md](../checklists/release.md) and [../checklists/rollout-and-slo-readiness.md](../checklists/rollout-and-slo-readiness.md) run the same checks against the release artifact.

### Eventing-Specific Proof

- Contract tests prove payload shape, metadata, and compatibility expectations ([../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md)).
- Duplicate-delivery tests prove idempotency of consumers and the inbox.
- Replay and out-of-order tests prove the real ordering contract instead of an assumed one.
- DLQ / parked-message tests prove terminal failures stop retrying and preserve operator context.
- The outbox dispatcher gets the shutdown proof from [Leak Detection](#leak-detection) plus a crash-between-write-and-publish test. See [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).

## Common Mistakes And Forbidden Patterns

- Moq or FluentAssertions anywhere in the dependency graph — both are forbidden outright, with no ADR path.
- Substituting repositories, HTTP clients, and workers so aggressively that no boundary behavior is exercised, or scripting `Received(...)` on every call instead of asserting outcomes.
- The EF Core InMemory provider standing in for database proof — it does not translate queries, enforce constraints, or run transactions like Postgres does.
- `Task.Delay`/`Thread.Sleep` or wall-clock polling to "let async work finish" instead of `FakeTimeProvider` plus a synchronization signal.
- `DateTime.Now`/`UtcNow` in tests or code under test instead of an injected `TimeProvider`.
- Disabling xUnit parallelism assembly-wide to hide shared mutable state, or mutating env vars/culture/statics in a parallel test.
- Blocking on async in tests: `.Result`, `.Wait()`, `.GetAwaiter().GetResult()` — use `async Task` tests and `IAsyncLifetime`.
- `async void` test helpers or event handlers; failures vanish instead of failing the test.
- Asserting on exception message substrings, dictionary ordering, or unsorted sequences instead of types, properties, and sorted comparisons.
- Hand-editing golden files instead of regenerating behind `UPDATE_GOLDEN`, or comparing them without line-ending normalization (breaks the windows CI leg).
- Types that own background work with no shutdown test proving bounded completion after cancellation.
- Integration tests that silently `Skip` when Docker is absent — behind the explicit switch they must fail, or the CI job proves nothing.
- Leaked handles in tests (`HttpResponseMessage`, streams, containers, `CancellationTokenSource` without `using`) — flaky file-lock failures on Windows are the symptom.
- Treating the coverage number as the deliverable, or letting it ratchet down to merge a change.

## Verification And Proof

```powershell
pwsh ./verify.ps1                # the gate: restore (locked), format-check, build (warnings-as-errors), test, audit
pwsh ./verify.ps1 -Integration   # adds the Testcontainers suite (requires Docker)
dotnet test --project tests/Orders.UnitTests                                        # fast inner loop
dotnet test --project tests/Orders.UnitTests -- --coverage --coverage-output-format cobertura  # coverage, separate from the gate
```

Add the following when relevant:

- a shutdown-proof test for every type that owns background work, asserting bounded completion after `StopAsync`
- `FakeTimeProvider` for every time-dependent behavior; no real sleeps anywhere in the suite
- `WebApplicationFactory` transport tests for every endpoint's error branches, asserting the ProblemDetails contract
- Testcontainers integration tests for every repository and migration, in CI's Docker-enabled job
- golden files in `TestData/`, regenerated behind `UPDATE_GOLDEN`, reviewed as contract changes
- property-based tests (FsCheck, via ADR) for parsers, encoders, and invariants over large input spaces
- an occasional `dotnet stryker` run over `<App>.Core` when assertion quality is in doubt
- BenchmarkDotNet comparisons when performance is part of the change; load and soak runs when an SLO or the hot path changes
- contract, duplicate-delivery, replay, and DLQ tests when the repo publishes or consumes messages

Testing is done when the chosen proof matches the risk of the change, not when a single unit suite turns green.

## Where To Go Next

- Static analysis that backstops the suite: [linting.md](linting.md)
- Time injection contract: [../foundations/time.md](../foundations/time.md)
- Cancellation and shutdown semantics under test: [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md)
- How the gate and the integration lane run in CI: [../operations/ci-and-release.md](../operations/ci-and-release.md)
