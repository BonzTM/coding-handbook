# Maintainer Reference

Purpose: hold slower-path architecture, project-map, lifecycle, and rationale guidance that is useful but not worth loading for every task.
Audience: maintainers and agents working in C#/.NET repositories that use this handbook.
Read [AGENTS.md](AGENTS.md) first. Use this file when you need the fuller background behind the fast-path rules.

## Architecture Snapshot

This handbook assumes one solution with compiler-enforced project boundaries and a minimal public surface. The dominant shape is:

```text
repo/
  global.json
  Orders.slnx
  Directory.Build.props
  Directory.Packages.props
  nuget.config
  verify.ps1
  Makefile
  api/
  src/
    Orders.Api/
      Program.cs
      Endpoints/
      Telemetry/
    Orders.Core/
    Orders.Infrastructure/
      Data/
      Clients/
  tests/
    Orders.UnitTests/
    Orders.IntegrationTests/
```

Project references are the boundary mechanism, not folder conventions: `Orders.Core` references nothing app-level, `Orders.Api` and `Orders.Infrastructure` reference Core, and Api reaches Infrastructure only from the `Program.cs` composition root. A missing `<ProjectReference>` makes a violation a build error rather than a review comment. The bootstrap commands and full tree live in [foundations/project-setup.md](foundations/project-setup.md); the reference-direction table lives in [foundations/solution-and-project-design.md](foundations/solution-and-project-design.md).

Compiling reference modules that embody this architecture end to end are planned as phase 2; until they land, [templates/README.md](templates/README.md) is the canonical scaffolding and the only home of exact version pins.

## Two-Speed Documentation Model

- Fast path: [AGENTS.md](AGENTS.md) for invariants, the task loop, change-type-to-file-set routing, and baseline proof.
- Slow path: this file for architecture, project map, test taxonomy, lifecycle, and rationale.

Use the fast path for most tasks. Use this file when a change crosses projects, introduces new runtime behavior, or challenges an existing default.

## Project Map

| Project Area | Owns | Must Not Own |
|---|---|---|
| `src/<App>.Api/Program.cs` | config composition, DI registration, host lifetime, endpoint mapping, the explicit `--migrate` step | business rules, SQL, request-validation details |
| `src/<App>.Api/Endpoints/` | minimal-API endpoint groups: route registration, request/response DTO mapping, status mapping, endpoint filters | business rules hiding in handlers — a second Core |
| `src/<App>.Api/Telemetry/` | logging, OpenTelemetry, and health wiring behind `AddServiceTelemetry()` | business middleware, auth policy, anything beyond logging/OTel/health |
| `src/<App>.Core/` | domain types, ports (interfaces), business rules, orchestration | ASP.NET Core types, EF Core, broker SDKs, any other project in the solution |
| `src/<App>.Infrastructure/Data/` | `DbContext`, entity configuration, migrations, repositories implementing Core ports | transport concerns; leaking `IQueryable` or entities to callers |
| `src/<App>.Infrastructure/Clients/` | typed outbound HTTP clients (`IHttpClientFactory` + typed client + resilience handler) | domain decisions; per-call ad hoc retry policies |
| `api/` | authoritative wire-contract sources (`.proto`, OpenAPI) when the repo publishes them | generated outputs as the only contract source |
| `tests/<App>.UnitTests/` | fast, deterministic tests; no network, database, or Docker | integration suites that need real dependencies |
| `tests/<App>.IntegrationTests/` | Testcontainers-backed DB/broker/client tests behind the explicit `-Integration` switch | unit tests that would slow the offline inner loop |
| `<App>.TestUtilities` (only when ≥2 test projects need it) | shared fakes, builders, `FakeTimeProvider` helpers | production behavior; giant assertion DSLs |

Repos with significant async work add an eventing seam — publisher/consumer ports in Core, outbox/inbox and broker client in Infrastructure — per [services/eventing-and-messaging.md](services/eventing-and-messaging.md). The same boundary rules apply: business behavior stays in Core, while broker and delivery mechanics stay in Infrastructure.

## Lifecycle Model

For services and workers, the normal process lifecycle is:

1. `WebApplication.CreateBuilder(args)` composes configuration (`appsettings.json`, environment JSON, user secrets in Development, env vars, args — last wins).
2. Options bind with `ValidateDataAnnotations` plus `ValidateOnStart`, so invalid config kills the process before any listener opens.
3. DI registration runs through per-project extension methods (`AddServiceTelemetry()`, `AddOrdersCore()`, `AddOrdersInfrastructure()`).
4. `app.Build()` then endpoint mapping wire the transport; the explicit `--migrate` argument runs migrations and exits instead of serving.
5. `app.Run()` owns the lifetime: SIGTERM and Ctrl+C flow through `IHostApplicationLifetime`, on all OSes, with no hand-rolled signal handling.
6. On shutdown, Kestrel drains in-flight requests and `BackgroundService.ExecuteAsync` loops observe `stoppingToken`, all bounded by `HostOptions.ShutdownTimeout`; DI disposes registered resources deterministically.

If a repository shape does not fit this lifecycle, it should document the exception explicitly. The full rules live in [foundations/cancellation-and-async.md](foundations/cancellation-and-async.md) and [foundations/shared-constructs.md](foundations/shared-constructs.md).

## Test Taxonomy

| Test Type | Default Location | What It Proves |
|---|---|---|
| unit tests | `tests/<App>.UnitTests`, mirroring the source project structure | Core business rules and edge cases, in-memory and offline |
| transport tests | `tests/<App>.UnitTests` using `WebApplicationFactory` with fakes behind Core ports | routing, request decoding, status mapping, endpoint filters, `ProblemDetails` shape |
| repository integration tests | `tests/<App>.IntegrationTests` with Testcontainers | real SQL, transactions, and committed EF Core migrations against the real provider |
| external client tests | unit project with a stub server or hand-rolled `HttpMessageHandler` fake | request construction, timeout handling, response mapping |
| contract tests | golden files plus `JsonSerializerContext` round-trips | wire payload compatibility and unknown-field policy |
| benchmarks | BenchmarkDotNet project, ADR-gated | allocation and throughput characteristics of measured hot paths |

The important principle is not "more tests". It is "the right tests at the right boundary". A fake-backed repository test does not replace a real migration or transaction test, and the EF InMemory provider is not a database. The full strategy, including determinism, clock control, and leak detection, is [quality/testing.md](quality/testing.md).

## Runtime Contracts Worth Remembering

- Every background task must have an owner, a `stoppingToken`-driven stop condition, and a shutdown proof.
- Every external call must take a `CancellationToken` and a timeout budget.
- Every network-facing component should have a clear readiness story (`/readyz`) distinct from plain liveness (`/livez`).
- Every non-trivial feature should add telemetry where operators will actually need it.
- Every dependency added today becomes part of tomorrow's debugging and patch surface — and of the lockfile diff someone must review.

## Contract Surfaces

- HTTP and gRPC boundaries should have an obvious source of truth for payload shapes and error semantics; errors on the wire are RFC 9457 `ProblemDetails` per [foundations/serialization.md](foundations/serialization.md).
- Database schema, migration order, and compatibility expectations are data contracts, not incidental implementation details; expand/contract is the destructive-change pattern per [services/database.md](services/database.md).
- Event payloads should have explicit envelopes, versioning rules, and idempotency expectations when they cross process boundaries.
- Generated code is never the only contract source; the `.proto`, OpenAPI document, or migration history remains authoritative.
- A library's public API surface is a contract: package validation against the previous release keeps "breaking" from being discovered by a consumer per [operations/ci-and-release.md](operations/ci-and-release.md).

Event delivery rules are operational contracts too: whether delivery is at-least-once, what ordering is guaranteed, when retries stop, and what happens at dead-letter boundaries should be written down before a queue-backed feature is considered done. See [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Dependency Rationale

- BCL-and-platform-first keeps onboarding, debugging, and long-term maintenance cheaper; the runtime plus `Microsoft.Extensions.*` cover more ground than most ecosystems' standard libraries.
- EF Core is the default because it is the platform's supported data stack; the discipline (one `DbContext` per service, no repository frameworks over it, Dapper only for measured hot paths) is what keeps it honest.
- `ILogger<T>` with `[LoggerMessage]` is the default because it is standard, structured, source-generated, and every serious sink consumes it.
- OpenTelemetry over `Meter`/`ActivitySource` is acceptable because it is a de facto interoperability standard, not framework lock-in.
- Moq, FluentAssertions, AutoMapper, and MediatR-for-simple-dispatch are forbidden for licensing and traceability reasons documented in [decisions/framework-selection.md](decisions/framework-selection.md); reflection-heavy DI containers and web micro-frameworks are exceptions that need evidence, not preferences.

## Common Failure Modes

| Symptom | Likely Cause | First Fix |
|---|---|---|
| endpoints know too much about storage | business rules leaked out of `<App>.Core` | move orchestration into Core service methods behind ports |
| `Program.cs` grows with every feature | composition and domain logic are mixed | extract per-project `Add*` extension methods |
| tests pass but deploys fail | no real integration coverage for DB or external boundaries | add Testcontainers tests and a startup smoke check |
| workers die mid-write or hang on shutdown | `stoppingToken` ignored or tasks without owners | honor cancellation in `ExecuteAsync`; bound the drain with `HostOptions.ShutdownTimeout` |
| requests stall under load | thread-pool starvation from sync-over-async | remove `.Result`/`.Wait()`; make the path async end to end |
| metrics backend churns or explodes | high-cardinality tags | collapse tags to stable, finite values |
| config bugs show up in production only | options bound without `ValidateOnStart` or lazy `IConfiguration` reads | validate every required field at startup through typed options |
| CI green on Linux, broken on Windows | path separators, culture-implicit parsing, undisposed file handles | apply [foundations/cross-platform.md](foundations/cross-platform.md); keep all three matrix legs |

## Primary Sources Behind These Defaults

- .NET release cadence and support policy: `https://learn.microsoft.com/dotnet/core/releases-and-support`
- minimal APIs: `https://learn.microsoft.com/aspnet/core/fundamentals/minimal-apis`
- generic host and hosted services: `https://learn.microsoft.com/dotnet/core/extensions/generic-host`
- options pattern and validation: `https://learn.microsoft.com/dotnet/core/extensions/options`
- high-performance logging (`LoggerMessage`): `https://learn.microsoft.com/dotnet/core/extensions/logger-message-generator`
- EF Core migrations: `https://learn.microsoft.com/ef/core/managing-schemas/migrations/`
- `System.Text.Json` source generation: `https://learn.microsoft.com/dotnet/standard/serialization/system-text-json/source-generation`
- .NET observability with OpenTelemetry: `https://learn.microsoft.com/dotnet/core/diagnostics/observability-with-otel`
- `TimeProvider` and testable time: `https://learn.microsoft.com/dotnet/standard/datetime/timeprovider-overview`
- Central Package Management and NuGetAudit: `https://learn.microsoft.com/nuget/consume-packages/central-package-management`, `https://learn.microsoft.com/nuget/concepts/auditing-packages`
- RFC 9457 Problem Details: `https://www.rfc-editor.org/rfc/rfc9457`

## Related Docs

- Fast path and change routing: [AGENTS.md](AGENTS.md)
- Project layout: [foundations/project-setup.md](foundations/project-setup.md)
- Project boundaries: [foundations/solution-and-project-design.md](foundations/solution-and-project-design.md)
- Contracts and compatibility: [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md)
- Event delivery guidance: [services/eventing-and-messaging.md](services/eventing-and-messaging.md)
- Proof and testing: [quality/testing.md](quality/testing.md)
