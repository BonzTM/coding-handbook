<!--
TEMPLATE -> repo root AGENTS.md
Fill every <PLACEHOLDER>. This is the per-PROJECT fast path. It mirrors the C#
handbook's AGENTS.md two-speed model but is scoped to one repo and points back to
the handbook for the full rules. Keep it short; do not restate handbook prose here.
-->

# AGENTS.md - <project-name> Contract

Fast-path contract for autonomous agents and reviewers working in this repository.
Read this file first. It is scoped to this project; the full engineering rules live
in the C# handbook (`<link or path to the C# handbook AGENTS.md>`). Where this file
is silent, the handbook governs. This file may add constraints but never weakens it.

## Purpose

- This project is a <service | worker | CLI | library>: <one-line purpose>.
- Use this file for project-specific invariants, change routing, and the verification bar.
- Use [README.md](README.md) for project shape, config keys, and how to run it.
- Use the C# handbook for the complete, cross-project rules and rationale.

## Repo-Wide Invariants

- **Single solution**: `<App>.slnx` at the root; SDK pinned by `global.json` (.NET 10 LTS, `rollForward: latestFeature`).
- **Layout**: `src/` + `tests/` per the handbook; projects are `<App>.Api`, `<App>.Core`, `<App>.Infrastructure` plus their test projects.
- **Thin Program.cs**: `src/<App>.Api/Program.cs` is the composition root — config, DI, telemetry, endpoint mapping, shutdown. No business logic.
- **Boundaries by project reference**: `Core` references no app-level project and no ASP.NET/EF packages; `Api` and `Infrastructure` reference `Core`; `Api` touches `Infrastructure` types only in the composition root. `InternalsVisibleTo` only for the matching test project.
- **Nullable + warnings**: `<Nullable>enable</Nullable>` everywhere, `TreatWarningsAsErrors=true`, `AnalysisLevel=latest-all` (set once in `Directory.Build.props`; never disabled per-project).
- **Async discipline**: a `CancellationToken` flows through every I/O path; no `.Result`/`.Wait()`/`GetAwaiter().GetResult()` on async work.
- **Errors**: exceptions cross boundaries as RFC 9457 `ProblemDetails` on the wire; log once, via `ILogger<T>` + `[LoggerMessage]`, at the boundary that can act.
- **Config**: options pattern with `ValidateDataAnnotations` + `ValidateOnStart`, fail-fast at startup; every key documented in the [README.md](README.md) table.
- **Persistence**: EF Core + Npgsql in `Infrastructure/Data/`; migrations applied only by the explicit `--migrate` step, never on normal startup; real integration tests (Testcontainers) at the DB boundary.
- **Time**: inject `TimeProvider`; never `DateTime.Now`.
- **Testing**: xUnit v3, plain `Assert`; hand-rolled fakes first, NSubstitute when a fake is too costly; Moq and FluentAssertions are forbidden. Every behavior change ships with a test.
- **Dependencies**: versions live only in `Directory.Packages.props` (Central Package Management); lock files committed; explicit rationale per non-trivial dependency.
- **Project-specific**: <invariants unique to this repo — domain rules, external system, SLAs — or "none beyond the above".>

## Change Routing

Route the change; do not guess where code belongs. Owners per area: [CODEOWNERS](.github/CODEOWNERS).

| If you are changing... | Start in | Read first |
|---|---|---|
| process startup, wiring, shutdown | `src/<App>.Api/Program.cs` | handbook `foundations/project-setup.md` |
| domain logic or rules | `src/<App>.Core/` | handbook `foundations/solution-and-project-design.md` |
| HTTP endpoints, filters, routes | `src/<App>.Api/Endpoints/` | handbook `services/http-services.md` |
| gRPC methods, protos, interceptors | `src/<App>.Api/`, `api/` | handbook `services/grpc-services.md` |
| queries, schema, migrations | `src/<App>.Infrastructure/Data/` | handbook `services/database.md` |
| outbound HTTP clients | `src/<App>.Infrastructure/Clients/` | handbook `recipes/add-external-client.md` |
| config keys or startup defaults | options classes, `appsettings.json`, [README.md](README.md) | handbook `foundations/configuration.md` |
| logging, metrics, traces, health | `src/<App>.Api/Telemetry/` | handbook `operations/observability.md` |
| auth, secrets, input boundaries | the boundary project | handbook `operations/security.md` |
| CI, build, release, containers | `.github/workflows/`, `verify.ps1`, `Dockerfile` | handbook `operations/ci-and-release.md` |
| a new dependency or framework | `Directory.Packages.props` + the consuming project | handbook `decisions/framework-selection.md` |
| <project-specific area> | `<path>` | `<doc>` |

## Working Norms

- Prefer small, reviewable changes; match the repo's current shape, do not invent new architecture.
- Do not bypass boundaries: endpoints do not query the DbContext, repositories carry no transport logic, `Program.cs` holds no business rules.
- When adding a dependency, document why the BCL or an existing dependency is insufficient.
- Write the proving test before claiming success whenever practical.
- If verification fails, fix it or report it clearly. Do not claim the change is done.

## Baseline Verification

```powershell
pwsh ./verify.ps1
```

`pwsh ./verify.ps1` is the single ordered gate: restore (locked), format-check,
build (warnings-as-errors), test, audit. CI runs the same script on an
ubuntu/windows/macos matrix. Run the narrowest meaningful tests first, then this
gate. A change is not done until `pwsh ./verify.ps1` passes locally and in CI.
