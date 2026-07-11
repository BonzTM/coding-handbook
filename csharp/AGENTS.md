# AGENTS.md - C# Project Contract

This is the authoritative fast-path contract for autonomous agents working in a new C#/.NET repository.
Read this file first: it carries the repo-wide invariants, the change-routing table, and the verification bar. Use [maintainer-reference.md](maintainer-reference.md) when you need slower-path architecture and rationale.

## Purpose

- Use this file for repo-wide invariants, change defaults, change-to-file routing, and the verification bar.
- Use [maintainer-reference.md](maintainer-reference.md) for project maps, lifecycle guidance, test taxonomy, and troubleshooting.
- For the full catalogs, see the [recipes/README.md](recipes/README.md), [checklists/README.md](checklists/README.md), and [decisions/README.md](decisions/README.md) indexes.

## Source Of Truth

- This file is the fast path. More detailed docs refine it; they do not weaken it.
- Layout, SDK-pinning, and solution defaults live in [foundations/project-setup.md](foundations/project-setup.md).
- Project and dependency boundaries live in [foundations/solution-and-project-design.md](foundations/solution-and-project-design.md) and [decisions/framework-selection.md](decisions/framework-selection.md); reusable shared constructs (`<App>.Core`, `<App>.Api/Endpoints/`, `<App>.Infrastructure/Data/`, `<App>.Infrastructure/Clients/`, `<App>.Api/Telemetry/`, `<App>.TestUtilities`) live in [foundations/shared-constructs.md](foundations/shared-constructs.md).
- Schema, API, and data-boundary rules live in [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md); type modeling in [foundations/data-modeling.md](foundations/data-modeling.md); the JSON/wire boundary in [foundations/serialization.md](foundations/serialization.md).
- Runtime correctness rules live in [foundations/cancellation-and-async.md](foundations/cancellation-and-async.md), [foundations/errors-and-logging.md](foundations/errors-and-logging.md), and [foundations/configuration.md](foundations/configuration.md); OS-portability rules in [foundations/cross-platform.md](foundations/cross-platform.md).
- Proof expectations live in [quality/testing.md](quality/testing.md) and the relevant service or operations docs.
- Analyzer and format policy lives in [quality/linting.md](quality/linting.md); time and clock discipline in [foundations/time.md](foundations/time.md); copy-paste scaffolding and the only exact version pins in [templates/README.md](templates/README.md).
- Architecture decisions and their rationale live in [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).
- Team-process docs — [onboarding-and-handoff.md](onboarding-and-handoff.md), [CONTRIBUTING.md](CONTRIBUTING.md), [checklists/incident-response.md](checklists/incident-response.md), and the [glossary.md](glossary.md) lookup aid — serve humans running the team and the handbook; they are not needed to build or change an app.

## Fast Path

1. Read this file and identify the project shape from [README.md](README.md). For a brand-new build, run [checklists/spec-intake.md](checklists/spec-intake.md) first — it says which WHAT decisions to take from the spec, which to ask about, and which defaults apply when the spec is silent.
2. Route the change through the [Change Routing](#change-routing) table below; do not guess where code belongs.
3. Read the relevant foundations doc before editing code in a new area.
4. Implement with the repo defaults unless the repo has already documented an exception.
5. Prove the change with the narrowest meaningful tests first, then the repo-wide baseline.

## Repo-Wide Invariants

- **Single solution**: one `.slnx` at the repo root, SDK pinned by `global.json`. Extra solutions require explicit architectural review.
- **Three-project shape**: `src/` plus `tests/`; a service is `<App>.Api` (thin host), `<App>.Core` (domain), `<App>.Infrastructure` (EF Core, external clients) per [foundations/project-setup.md](foundations/project-setup.md). Fewer projects for tiny tools needs a project-README note; more needs review.
- **Boundary by reference**: `<App>.Core` references nothing app-level; Api and Infrastructure reference Core; Api touches Infrastructure only in the `Program.cs` composition root. `InternalsVisibleTo` only for the matching test project.
- **Thin Program.cs**: `Program.cs` wires config, logging, DI, telemetry, endpoints, and shutdown. Business logic belongs in `<App>.Core`.
- **Cancellation discipline**: `CancellationToken` flows through every async and I/O path. No sync-over-async (`.Result`, `.Wait()`, `GetAwaiter().GetResult()` on incomplete tasks); no fire-and-forget tasks without an owner.
- **Nullable everywhere**: `<Nullable>enable</Nullable>` set once in `Directory.Build.props`, never disabled in new code.
- **Errors**: exceptions for exceptional states, RFC 9457 `ProblemDetails` on the wire, and log once at the boundary that can act.
- **Logging**: use `ILogger<T>` with structured templates and source-generated `[LoggerMessage]`; no static or ambient loggers in reusable code.
- **Persistence**: default to EF Core with one `DbContext` per service; migrations apply only behind an explicit `--migrate` step, never on normal startup.
- **Testing**: every behavior change needs tests. DB and external boundaries need real integration coverage (Testcontainers), not only fakes.
- **Observability**: new networked behavior adds logs, metrics, and health/readiness behavior where appropriate.
- **Cross-platform**: everything in `verify.ps1` passes on ubuntu, windows, and macos; PowerShell 7 is the only blessed script runtime; no bash-only steps in the verify path.
- **Dependency posture**: BCL and `Microsoft.Extensions.*` first, explicit rationale for every non-trivial package, Central Package Management plus committed `packages.lock.json`, no floating versions.

## Change Routing

Use this when you know what kind of change you are making but not the file set. Start Here is what you read and touch first; Also Update is the sync surface the change normally drags along; Verify Or Confirm is the proof.

| Change Type | Start Here | Also Update | Verify Or Confirm |
|---|---|---|---|
| SDK pin, repo layout, solution membership | `global.json`, `src/`, `tests/`, [foundations/project-setup.md](foundations/project-setup.md) | `Directory.Build.props`, CI workflow, tool manifest, onboarding docs | `dotnet restore --locked-mode`, `dotnet build` via `pwsh ./verify.ps1` |
| HTTP or gRPC contract shape, schema, compatibility, generated stubs | `api/**`, transport project, [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md), [recipes/deprecate-and-remove-contract.md](recipes/deprecate-and-remove-contract.md) | clients, generated-code policy, release notes, tests | schema or generation check, compatibility review |
| Event or message producers, consumers, payload contracts | eventing seam, `api/**`, [services/eventing-and-messaging.md](services/eventing-and-messaging.md) (## Outbox And Inbox Patterns), [recipes/add-event-publisher.md](recipes/add-event-publisher.md), [recipes/add-event-consumer.md](recipes/add-event-consumer.md) | outbox or inbox storage, telemetry, release notes, tests, DLQ policy | contract tests, replay/idempotency tests, integration against real broker path |
| HTTP endpoint, endpoint filter, request or response shape | `src/<App>.Api/Endpoints/`, [services/http-services.md](services/http-services.md), [recipes/add-http-endpoint.md](recipes/add-http-endpoint.md), [recipes/add-http-middleware.md](recipes/add-http-middleware.md) | `<App>.Core`, request validation, telemetry, route-group registration, tests | `WebApplicationFactory` tests, smoke test, readiness behavior |
| Server-rendered pages, Razor Pages, static assets, sessions, antiforgery | `src/<App>.Api/Pages/`, [services/web-apps.md](services/web-apps.md), [services/http-services.md](services/http-services.md) | layout and partials, session/auth cookie config, security headers, PRG flows, tests | antiforgery negative test, session cookie-flag assertions, XSS probe |
| gRPC proto or server method | `api/**`, `src/<App>.Api/`, [services/grpc-services.md](services/grpc-services.md), [recipes/add-grpc-method.md](recipes/add-grpc-method.md) | Grpc.Tools generation policy, interceptors, error mapping, docs | proto lint or generate check, service tests, `grpcurl` |
| Business logic or domain rules | `src/<App>.Core/`, [foundations/solution-and-project-design.md](foundations/solution-and-project-design.md) | unit tests, transport adapters if contracts changed | targeted unit tests, relevant integration coverage |
| Errors, ProblemDetails mapping, structured logging | [foundations/errors-and-logging.md](foundations/errors-and-logging.md), [foundations/serialization.md](foundations/serialization.md) (### Error Responses) | exception middleware, status mapping, log fields at the acting boundary | error-mapping tests; log-once review at the boundary |
| Code review heuristics, comments, naming, API shape | [foundations/style-and-review.md](foundations/style-and-review.md) | XML doc comments, exported-contract docs, review checklist | `dotnet format --verify-no-changes`, public surface reads as a contract |
| DB queries, schema, transaction behavior | `src/<App>.Infrastructure/Data/`, [services/database.md](services/database.md) (### Migrations), [recipes/add-database-feature.md](recipes/add-database-feature.md), [recipes/add-migration.md](recipes/add-migration.md) | entity configurations, repositories, callers, observability | migration apply test, real DB integration tests (Testcontainers) |
| Configuration keys or startup defaults | options classes, [foundations/configuration.md](foundations/configuration.md), [recipes/add-config-key.md](recipes/add-config-key.md) | `appsettings.json`, startup wiring, docs, tests | options-binding tests, startup smoke test with a value removed |
| Background workers, schedulers, shutdown paths | worker classes, [foundations/cancellation-and-async.md](foundations/cancellation-and-async.md) (### Graceful Shutdown And Draining), [recipes/add-background-worker.md](recipes/add-background-worker.md), [recipes/add-scheduled-job.md](recipes/add-scheduled-job.md) | `Program.cs` registration, telemetry, readiness semantics | shutdown test proving `stoppingToken` stops the work, leak check per [quality/testing.md](quality/testing.md) (### Leak Detection) |
| Async correctness: token flow, task ownership, channels, parallelism | [foundations/cancellation-and-async.md](foundations/cancellation-and-async.md) | callers missing `CancellationToken`, task owners, bounded queues | analyzer-clean build; deterministic async tests, no `Task.Delay` sleeps |
| Cross-platform behavior: paths, line endings, culture, scripts, signals | [foundations/cross-platform.md](foundations/cross-platform.md) | `.gitattributes`, `.editorconfig`, `verify.ps1`, culture-sensitive call sites | `pwsh ./verify.ps1` green on the ubuntu/windows/macos CI matrix |
| External HTTP or gRPC client | `src/<App>.Infrastructure/Clients/`, [recipes/add-external-client.md](recipes/add-external-client.md), [operations/resilience.md](operations/resilience.md) | port in `<App>.Core`, resilience handler, telemetry, security checks | stub-server or `HttpMessageHandler`-fake tests, timeout tests |
| Logging, metrics, traces, health endpoints | `src/<App>.Api/Telemetry/`, [operations/observability.md](operations/observability.md), [recipes/add-metric.md](recipes/add-metric.md) | endpoints, repositories, workers, dashboards or alerts if repo has them | `/livez`, `/readyz`, metric tag-cardinality and log schema review |
| Security-sensitive boundary | boundary code, [operations/security.md](operations/security.md) | validation, auth policies, secret handling, release docs if exposure changed | audit stage of `pwsh ./verify.ps1`, targeted negative tests, [checklists/security-review.md](checklists/security-review.md) |
| Audit logging of security-relevant actions | [operations/security.md](operations/security.md) (### Audit Logging) | audit sink/retention, actor and resource fields, telemetry, access controls on the log | audit events emitted at the action boundary; tamper-evidence and retention honored |
| Data classification, PII, retention, compliance | [operations/data-handling.md](operations/data-handling.md) | field-level classification, log/metric redaction, retention jobs, export/delete paths | PII never in logs/metrics; retention and deletion paths exercised by tests |
| Pre-build WHAT decisions (scope, acceptance, spec intake) | [checklists/spec-intake.md](checklists/spec-intake.md) | the task's acceptance criteria, open questions or defaulted assumptions, downstream design docs | spec intake resolved (answered, asked, or defaulted-and-disclosed) before code; acceptance criteria agreed and testable |
| Idempotent HTTP writes (Idempotency-Key, safe retries) | [recipes/add-idempotent-write.md](recipes/add-idempotent-write.md), [services/http-services.md](services/http-services.md) (### Idempotent Writes) | idempotency-key storage, dedupe window, replay response shape, tests | duplicate requests collapse to one effect; replays return the stored result |
| CLI commands or options | `src/<App>.Cli/`, [recipes/add-cli-command.md](recipes/add-cli-command.md) | config loading, help output, README if user-visible | command tests, `--help`, build and smoke run |
| Build, CI, containers, release automation | `.github/workflows/`, `verify.ps1`, Dockerfile, [operations/ci-and-release.md](operations/ci-and-release.md) | version stamping (MinVer), changelog, release packaging | CI dry run on all three matrix legs, container smoke test |
| New dependency or framework | [decisions/framework-selection.md](decisions/framework-selection.md), [recipes/bump-dependency.md](recipes/bump-dependency.md) | `Directory.Packages.props`, `packages.lock.json`, the caller project, proof docs | written rationale, lockfile diff understood, audit stage clean |
| Analyzer config, severities, format policy | `.editorconfig`, `Directory.Build.props`, [quality/linting.md](quality/linting.md) | CI stages, suppression justifications | `dotnet format --verify-no-changes` clean, `pwsh ./verify.ps1` green |
| Time, clocks, timeouts, scheduling | [foundations/time.md](foundations/time.md) | injected `TimeProvider`, `FakeTimeProvider` in tests, callers reaching for `DateTime.Now` | deterministic tests with `FakeTimeProvider`; no sleeps |
| Domain type modeling: enums, typed IDs, optional values, collections | [foundations/data-modeling.md](foundations/data-modeling.md) | the wire shape (serialization), DB mapping, validation | round-trip tests; construction-time validation tests |
| Serialization / JSON wire shape: contexts, DTOs, unknown fields | [foundations/serialization.md](foundations/serialization.md) | transport DTOs, `JsonSerializerContext`, contract compatibility, golden tests | round-trip + golden tests; decoder handles unknown fields per policy |
| Non-obvious or hard-to-reverse architecture decision | [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md) | the project's `decisions/` directory, any superseded ADR, README/onboarding links | ADR recorded with status, alternatives, and consequences before the change merges |
| Project ownership transfer or onboarding | [onboarding-and-handoff.md](onboarding-and-handoff.md), [checklists/handoff.md](checklists/handoff.md) | CODEOWNERS, secrets/rotation docs, on-call, deploy access, open decisions | handoff checklist complete; new owner runs `pwsh ./verify.ps1` and a deploy dry-run unaided |
| New-repo scaffolding (verify.ps1, CI, Program.cs, configs, project docs) | [templates/README.md](templates/README.md), [checklists/new-project.md](checklists/new-project.md) | the copied artifacts and their `<placeholder>` values | `pwsh ./verify.ps1` green on the fresh repo |
| Outbound resilience: timeouts, retries, backoff, breakers, rate limits | [operations/resilience.md](operations/resilience.md) | typed clients, eventing consumers, telemetry for retries/breakers | load test shows graceful shedding; a timeout on every external call |
| Containerization, Dockerfile, runtime limits, rollout | [operations/deployment.md](operations/deployment.md), [templates/Dockerfile](templates/Dockerfile) | CI image build, resource limits, probes, shutdown grace period | image builds chiseled + non-root; probes wired; `pwsh ./verify.ps1` green |
| SLOs, alerting, runbooks, on-call | [operations/operability.md](operations/operability.md), [templates/runbook.md](templates/runbook.md) | dashboards/alerts (repo-specific), error budgets, on-call docs | each service has SLOs, symptom alerts, and a current runbook |
| Caching layer, invalidation, cache keys, stampede control | [services/caching.md](services/caching.md) | callers, telemetry for hit/miss, invalidation paths, config keys | cache hit/miss metrics; correctness under invalidation; no stale-read regressions |
| Feature flags, gating, rollout toggles | [foundations/configuration.md](foundations/configuration.md) (### Feature Flags) | typed accessor, default state, cleanup of stale flags, tests | flag default and override tested; stale flags scheduled for removal |
| Git workflow, branching, commits, CHANGELOG | [foundations/git-workflow.md](foundations/git-workflow.md), [operations/ci-and-release.md](operations/ci-and-release.md) (### Changelog Policy) | CHANGELOG, release notes, PR template | commit and branch conventions followed; changelog entry present where required |
| Tagging or publishing a library/package version | [recipes/release-library-version.md](recipes/release-library-version.md), [checklists/release.md](checklists/release.md) | CHANGELOG, semver tag, package validation baseline, consumers | `v1.2.3` tag is canonical; package validation passes; `pwsh ./verify.ps1` green |
| Reusable shared construct (`<App>.Core` ports, `Endpoints/`, `Data/`, `Clients/`, `Telemetry/`, `TestUtilities`) | [foundations/shared-constructs.md](foundations/shared-constructs.md) | the consuming projects and `Program.cs` wiring | each construct states exactly what it owns; `Program.cs` reads top-to-bottom; no junk-drawer growth |

## High-Value Boundaries

- `Program.cs` owns composition and process lifetime; it should not hold business rules.
- `src/<App>.Core/` owns domain behavior and the ports (interfaces) consumed from the outside; it references nothing app-level.
- `src/<App>.Api/Endpoints/` translates transport concerns into Core calls; an endpoint that touches a `DbContext` or typed client directly has broken the boundary even though it compiles.
- `src/<App>.Infrastructure/Data/` owns the `DbContext`, migrations, repositories, and storage-specific mapping.
- `src/<App>.Infrastructure/Clients/` owns typed outbound HTTP clients and their resilience handlers.
- `src/<App>.Api/Telemetry/` owns logging, OpenTelemetry, and health wiring behind `AddServiceTelemetry()`.
- `api/` owns published wire-contract definitions such as `.proto` or OpenAPI sources when the repo uses them.
- The eventing seam owns broker interaction, settlement, retry policy, and payload mapping rather than scattering those concerns through endpoints or `Program.cs`.

## Proof Hints

- Transport changes usually need both narrow `WebApplicationFactory` tests and one manual or scripted smoke test.
- Persistence changes usually need real database integration tests (Testcontainers); fakes are not enough.
- New background work usually needs a shutdown test proving cancellation stops what `ExecuteAsync` started.
- Config changes are not done until invalid input fails startup fast and documented keys stay in sync with `appsettings.json`.
- Dependency changes are not done until the rationale is explicit and the `packages.lock.json` diff is understood.

## Working Norms

- Prefer small, reviewable changes over broad cleanup.
- Do not introduce new architecture because it feels cleaner; match the repo's current shape unless the task is explicitly architectural.
- Do not bypass boundaries: endpoints do not query the database directly, repositories do not contain transport logic, and `Program.cs` does not absorb business rules.
- When adding a dependency, document why the BCL, ASP.NET Core, or an existing dependency is insufficient.
- When behavior changes, write the failing or proving test before claiming success whenever practical.
- If verification fails, fix it or report it clearly. Do not claim the change is done.

## Baseline Verification

| Goal | Command | Expectation |
|---|---|---|
| dependency-graph hygiene | `dotnet restore --locked-mode` | committed `packages.lock.json` matches the project graph; drift fails, it is not rewritten |
| format | `dotnet format --verify-no-changes` | no diff; `.editorconfig` is the single style source |
| compile safety and static analysis | `dotnet build -c Release -warnaserror` | zero warnings — built-in analyzers (`AnalysisLevel=latest-all`) run inside the build (policy in [quality/linting.md](quality/linting.md)) |
| functional confidence | `dotnet test` (unit projects; xUnit v3 on Microsoft.Testing.Platform) | all unit projects pass, offline |
| supply-chain check | vulnerable-package audit (`dotnet list package --vulnerable --include-transitive` + NuGetAudit on restore) | no findings |
| integration boundaries | `pwsh ./verify.ps1 -Integration` | Testcontainers-backed suites pass where Docker is available (explicit switch; CI runs it on ubuntu) |
| file-specific correctness | targeted tests from the relevant recipe or service doc | pass with expected assertions |

The committed [templates/verify.ps1](templates/verify.ps1) wraps the full gate — restore (locked), format-check, build (warnings-as-errors), test, audit — as `pwsh ./verify.ps1`; run it so local and CI stay identical. The [templates/Makefile](templates/Makefile) is a one-line shim delegating to the same script, and CI runs it on an ubuntu/windows/macos matrix per [operations/ci-and-release.md](operations/ci-and-release.md). Use broader tools such as benchmarks or load tests when the repo or task calls for them.

## Slow Path Docs

- Architecture and project map: [maintainer-reference.md](maintainer-reference.md)
- Startup and solution layout: [foundations/project-setup.md](foundations/project-setup.md)
- Runtime correctness: [foundations/cancellation-and-async.md](foundations/cancellation-and-async.md), [foundations/errors-and-logging.md](foundations/errors-and-logging.md), [foundations/configuration.md](foundations/configuration.md)
- OS portability: [foundations/cross-platform.md](foundations/cross-platform.md)
- Proof and verification: [quality/testing.md](quality/testing.md)
