# C# Project Handbook

This handbook is the default engineering contract for new C#/.NET repositories. It is not a C# language tutorial. It exists to make services, workers, CLIs, and libraries converge on the same structure, runtime behavior, dependency posture, and proof of correctness.

## Start Here

- Humans: read this file, then follow the reading path for your project shape.
- Agents: read [AGENTS.md](AGENTS.md) first (it includes the change-routing table), then the relevant topical docs and recipes.
- Default assumptions unless a repo says otherwise:
  - one solution per repo, in `.slnx` format, SDK pinned by `global.json`
  - `src/` plus `tests/` as the default layout; a service is exactly three projects — `<App>.Api` (thin host), `<App>.Core` (domain, no ASP.NET or EF references), `<App>.Infrastructure` (EF Core, external clients)
  - thin `Program.cs`
  - ASP.NET Core Minimal APIs, EF Core with Npgsql, `ILogger<T>`, `System.Text.Json`, and xUnit v3 first
  - `appsettings.json` plus env-var overrides, bound to validated options with fail-fast startup
  - restore (locked), format-check, build (warnings-as-errors), test, and vulnerable-package audit as baseline proof, all wrapped by `pwsh ./verify.ps1`

## Reading Paths

| If you are building... | Read in this order |
|---|---|
| HTTP service | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/solution-and-project-design.md](foundations/solution-and-project-design.md) -> [foundations/configuration.md](foundations/configuration.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [services/http-services.md](services/http-services.md) -> [operations/observability.md](operations/observability.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-http-endpoint.md](recipes/add-http-endpoint.md); compiling exemplar: [reference/exampleservice/](reference/exampleservice/) |
| Server-rendered web app | the HTTP service path above, inserting [services/web-apps.md](services/web-apps.md) after [services/http-services.md](services/http-services.md) — Razor Pages, static assets, sessions, antiforgery, and browser security headers layer on the same service skeleton |
| gRPC service | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/solution-and-project-design.md](foundations/solution-and-project-design.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [services/grpc-services.md](services/grpc-services.md) -> [foundations/errors-and-logging.md](foundations/errors-and-logging.md) -> [services/database.md](services/database.md) -> [operations/observability.md](operations/observability.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-grpc-method.md](recipes/add-grpc-method.md); compiling exemplar: [reference/examplegrpc/](reference/examplegrpc/) |
| Background worker | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/cancellation-and-async.md](foundations/cancellation-and-async.md) -> [foundations/configuration.md](foundations/configuration.md) -> [operations/observability.md](operations/observability.md) -> [operations/security.md](operations/security.md) -> [recipes/add-background-worker.md](recipes/add-background-worker.md) |
| Event-driven service or async worker | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [services/eventing-and-messaging.md](services/eventing-and-messaging.md) -> [services/database.md](services/database.md) -> [operations/observability.md](operations/observability.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-event-publisher.md](recipes/add-event-publisher.md) -> [recipes/add-event-consumer.md](recipes/add-event-consumer.md); compiling exemplar: [reference/exampleworker/](reference/exampleworker/) |
| CLI tool | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/style-and-review.md](foundations/style-and-review.md) -> [foundations/configuration.md](foundations/configuration.md) -> [decisions/framework-selection.md](decisions/framework-selection.md) -> [recipes/add-cli-command.md](recipes/add-cli-command.md) |
| Library | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/solution-and-project-design.md](foundations/solution-and-project-design.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [foundations/style-and-review.md](foundations/style-and-review.md) -> [quality/testing.md](quality/testing.md) -> [foundations/errors-and-logging.md](foundations/errors-and-logging.md) -> [checklists/release.md](checklists/release.md) -> [recipes/release-library-version.md](recipes/release-library-version.md) |

Three complete, compiling, `pwsh ./verify.ps1`-green reference modules live under `reference/`: [exampleservice](reference/exampleservice/) (HTTP+Postgres, full enterprise stack), [examplegrpc](reference/examplegrpc/) (gRPC), and [exampleworker](reference/exampleworker/) (event-driven worker). Copy the one matching your shape to bootstrap a new repo; [templates/](templates/) holds the same scaffolding piecemeal.

Every shape also adopts [quality/linting.md](quality/linting.md), [foundations/cross-platform.md](foundations/cross-platform.md), and the committed [templates/](templates/) scaffolding, runs `pwsh ./verify.ps1` as its proof gate, and follows [foundations/data-modeling.md](foundations/data-modeling.md), [foundations/serialization.md](foundations/serialization.md), and [foundations/time.md](foundations/time.md) for type, wire-shape, and clock decisions. Networked services additionally follow [operations/resilience.md](operations/resilience.md), [operations/deployment.md](operations/deployment.md), and [operations/operability.md](operations/operability.md).

Cross-cutting concerns apply across shapes: [services/caching.md](services/caching.md) and [foundations/configuration.md](foundations/configuration.md) (feature flags) affect most services, and [foundations/git-workflow.md](foundations/git-workflow.md) governs commits, branches, and changelog discipline everywhere.

## Non-Negotiables

- Keep `Program.cs` boring. It wires config, logging, DI, telemetry, endpoints, and shutdown; it does not hold business logic.
- Pass `CancellationToken` explicitly through every async and I/O path. No sync-over-async: `.Result`, `.Wait()`, and `GetAwaiter().GetResult()` on incomplete tasks are forbidden.
- `<Nullable>enable</Nullable>` everywhere, set once in `Directory.Build.props`; never disabled in new code.
- Project references define the boundaries: `<App>.Core` references nothing app-level; `<App>.Api` reaches `<App>.Infrastructure` only from the `Program.cs` composition root.
- Use `ILogger<T>` with structured message templates and source-generated `[LoggerMessage]`. Log once at the boundary that can act.
- Errors on the wire are RFC 9457 `ProblemDetails`; never a bare error string.
- Inject `TimeProvider`; never read `DateTime.Now` or `DateTime.UtcNow` in logic.
- Use real integration tests (Testcontainers) for database and external-service boundaries; do not mock everything by default.
- Keep metric tags low-cardinality; request IDs and user IDs never belong in metrics.
- Do not commit floating package versions, real secrets, or framework-heavy defaults without explicit justification; versions live in `Directory.Packages.props` plus `packages.lock.json`.

## Default Stack

| Concern | Default | Reach for something else when |
|---|---|---|
| Web framework | ASP.NET Core Minimal APIs with `TypedResults`, route groups, endpoint filters | filter pipelines or model-binding volume justify MVC controllers, via ADR |
| Server-rendered web | Razor Pages on the same host | interactivity genuinely outgrows pages plus progressive enhancement — Blazor via ADR |
| Data access | EF Core, one `DbContext` per service, Npgsql/PostgreSQL provider | measured hot paths justify Dapper via ADR; SQL Server when the org mandates it |
| Migrations | EF Core migrations behind an explicit `--migrate` step, never on normal startup | almost never |
| Configuration | `appsettings.json` (committed, secret-free) plus env-var overrides; options pattern with `ValidateDataAnnotations` + `ValidateOnStart` | an operator workflow requires extra config files, via ADR |
| Dependency injection | built-in `Microsoft.Extensions.DependencyInjection`, constructor injection | almost never; a third-party container requires an ADR |
| Logging | `ILogger<T>` + source-generated `[LoggerMessage]` | Serilog via ADR only when its sinks are required — still consumed as `ILogger<T>` |
| Metrics and tracing | OpenTelemetry over `Meter`/`ActivitySource`, OTLP exporter | the org scrapes Prometheus — use the OTel Prometheus exporter |
| Serialization | `System.Text.Json` with source-generated `JsonSerializerContext` at trust boundaries | never Newtonsoft.Json in new code without an ADR |
| Validation | built-in minimal-API validation (`AddValidation` + DataAnnotations on request DTOs) | rule complexity outgrows attributes — FluentValidation via ADR |
| Testing | xUnit v3 on Microsoft.Testing.Platform, plain `Assert`, hand-rolled fakes, `WebApplicationFactory`, Testcontainers | a fake is too costly — NSubstitute; Moq and FluentAssertions are forbidden |
| Messaging | repo-owned publisher/consumer seam with outbox/inbox; broker client confined behind it | documented delivery semantics justify MassTransit/Wolverine via ADR |
| Caching | `HybridCache` in-process with bounded size and TTL | the working set or cross-instance sharing demands Redis/Valkey as the backend |
| CLI parsing | `System.CommandLine` (GA `SetAction` API) | almost never |
| Time | injected `TimeProvider`; `DateTimeOffset`/`DateOnly`/`TimeOnly` on storage and wire | never |

The full table, escalation rules, and forbidden packages live in [decisions/framework-selection.md](decisions/framework-selection.md).

## Handbook Map

- [AGENTS.md](AGENTS.md) - fast-path contract and change routing for autonomous agents and reviewers
- [maintainer-reference.md](maintainer-reference.md) - architecture, rationale, and deeper guidance
- `foundations/` - solution layout, project boundaries, contracts, config, errors, [cancellation and async](foundations/cancellation-and-async.md), time, data modeling, serialization, [cross-platform rules](foundations/cross-platform.md), [shared constructs](foundations/shared-constructs.md), and [git-workflow.md](foundations/git-workflow.md) (commits, branches, changelog)
- `quality/` - test strategy, analyzers/format policy, and proof commands
- `services/` - transport, eventing, persistence, [caching.md](services/caching.md), and [web-apps.md](services/web-apps.md) guidance for HTTP, gRPC, server-rendered web, messaging, database, and cache work
- `operations/` - telemetry, security, audit logging, resilience, deployment, operability/SLOs, data handling/PII, CI, releases, and runtime expectations
- `decisions/` ([README.md](decisions/README.md)) - architecture decision records (ADRs) plus dependency and framework selection rules
- `checklists/` ([README.md](checklists/README.md)) and `recipes/` ([README.md](recipes/README.md)) - executable startup, review, release, handoff, and implementation guidance
- `templates/` ([README.md](templates/README.md)) - committed copy-paste scaffolding (`global.json`, `Directory.Build.props`, `Directory.Packages.props`, `verify.ps1`, `Program.cs`, CI workflows, Dockerfile, project README/AGENTS/CODEOWNERS, ADR template) — exact version pins live here and in the reference modules, nowhere else
- `reference/` - three complete, compiling, `verify.ps1`-green services that compose the handbook patterns end to end: [exampleservice](reference/exampleservice/) (HTTP+Postgres, full enterprise stack), [examplegrpc](reference/examplegrpc/) (gRPC), and [exampleworker](reference/exampleworker/) (event-driven worker); copy the one matching your shape to bootstrap a new repo
- Team process (human-facing; not read during app builds): [onboarding-and-handoff.md](onboarding-and-handoff.md) for ownership transfer, [checklists/incident-response.md](checklists/incident-response.md) for on-call, [glossary.md](glossary.md) as a term lookup, and [CONTRIBUTING.md](CONTRIBUTING.md) for changing the handbook itself

## What This Handbook Optimizes For

- code that still looks obvious six months later
- boundaries that make testing and refactoring cheaper
- runtime behavior that is safe under load and easy to debug
- defaults that keep agents from inventing new architecture every task
- minimal dependency surface unless there is a clear return on complexity

## Where To Go Next

- New repo bootstrap: [checklists/new-project.md](checklists/new-project.md)
- Resolving WHAT decisions before a build (take from spec, ask, or default): [checklists/spec-intake.md](checklists/spec-intake.md)
- Active agent work: [AGENTS.md](AGENTS.md)
- Routing a change quickly: [AGENTS.md](AGENTS.md) (## Change Routing)
- Choosing third-party libraries: [decisions/framework-selection.md](decisions/framework-selection.md)
- Recording an architecture decision: [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md)
- Copy-paste scaffolding for a new repo: [templates/README.md](templates/README.md)
- Taking over or handing off a project: [onboarding-and-handoff.md](onboarding-and-handoff.md)
- Analyzer and format policy: [quality/linting.md](quality/linting.md)
- Making a networked service production-ready: [operations/resilience.md](operations/resilience.md), [operations/deployment.md](operations/deployment.md), [operations/operability.md](operations/operability.md)
- Implementing a specific change step by step: [recipes/README.md](recipes/README.md)
- Looking up a handbook term: [glossary.md](glossary.md)
- Changing the handbook itself, commits, or the changelog: [CONTRIBUTING.md](CONTRIBUTING.md), [foundations/git-workflow.md](foundations/git-workflow.md)
- A complete worked example to copy: [reference/exampleservice/](reference/exampleservice/) (HTTP+Postgres), [reference/examplegrpc/](reference/examplegrpc/) (gRPC), or [reference/exampleworker/](reference/exampleworker/) (event-driven) — all `pwsh ./verify.ps1`-green
