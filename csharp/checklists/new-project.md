# New Project Checklist

Bootstrap checklist for a brand-new .NET repo using this handbook.

## Repository Skeleton

- [ ] Copy [global.json](../templates/global.json) into the repo root so the SDK line is pinned (`rollForward: latestFeature`); never rely on whatever SDK happens to be installed.
- [ ] Copy [Directory.Build.props](../templates/Directory.Build.props), [Directory.Packages.props](../templates/Directory.Packages.props), and [nuget.config](../templates/nuget.config) into the repo root — nullable, implicit usings, `TreatWarningsAsErrors`, `AnalysisLevel=latest-all`, Central Package Management, `RestorePackagesWithLockFile`, and `NuGetAuditMode=all` apply to every project from day one, not per-csproj.
- [ ] Copy [.editorconfig](../templates/.editorconfig), [.gitattributes](../templates/.gitattributes), and [gitignore](../templates/gitignore) (as `.gitignore`) — style/severity live in `.editorconfig` only, and `.gitattributes` pins LF endings so the repo behaves identically on Windows, Linux, and macOS per [../foundations/cross-platform.md](../foundations/cross-platform.md).
- [ ] Create the `.slnx` solution with `src/` + `tests/` and the project split for the repo's shape per [../foundations/solution-and-project-design.md](../foundations/solution-and-project-design.md): service = `<App>.Api` (thin host) + `<App>.Core` (domain, no ASP.NET/EF references) + `<App>.Infrastructure` (EF Core, external clients); tests = `<App>.UnitTests` + `<App>.IntegrationTests`; library = one project + one test project; CLI = `<App>.Cli` + `<App>.Core`.
- [ ] Keep `Program.cs` limited to startup, wiring, and shutdown; it is the only place `<App>.Api` references `<App>.Infrastructure`.
- [ ] Enforce boundaries with project references, not discipline: `Core` references nothing app-level; `Api` and `Infrastructure` reference `Core`; `InternalsVisibleTo` only for the matching test project.
- [ ] Decide whether the repo is a service, worker, CLI, library, or a combination, then document that shape in the repo README.
- [ ] Decide which boundaries need explicit contracts in `api/` (protos, OpenAPI), transport docs, or schema sources.
- [ ] If the repo publishes or consumes messages, decide event envelope shape, idempotency policy, ordering guarantees, retry limits, and DLQ behavior up front per [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).

## Runtime Contract

- [ ] Configuration binds in one place via the options pattern with `ValidateDataAnnotations` + `ValidateOnStart`, so a bad config fails startup fast per [../foundations/configuration.md](../foundations/configuration.md); committed `appsettings.json` carries no secrets, env vars override.
- [ ] `ILogger<T>` with source-generated `[LoggerMessage]` methods is wired centrally (the `AddServiceTelemetry()` extension) and injected into runtime components per [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md).
- [ ] `IHostApplicationLifetime` and `HostOptions.ShutdownTimeout` define the shutdown contract; workers honor `stoppingToken` in `BackgroundService.ExecuteAsync` per [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md).
- [ ] Health and readiness behavior is defined for networked services: `/livez` always cheap, `/readyz` checks dependencies per [../operations/operability.md](../operations/operability.md).
- [ ] Database and external clients have explicit timeout and shutdown behavior; outbound HTTP goes through `IHttpClientFactory` typed clients with a resilience handler per [../operations/resilience.md](../operations/resilience.md).

## Proof And Delivery

- [ ] Copy the committed [verify.ps1](../templates/verify.ps1), the one-line [Makefile](../templates/Makefile) shim, and the [CI workflow](../templates/github-workflows-ci.yml) into the repo so the gate is identical locally and in CI (the workflow runs the same script on an ubuntu/windows/macos matrix).
- [ ] `pwsh ./verify.ps1` is THE baseline gate from day one — restore (locked), format-check, build (warnings-as-errors), test, audit (per [../quality/linting.md](../quality/linting.md)). This is mandatory, not an optional analyzer add-on.
- [ ] `packages.lock.json` files are generated and committed for every project; CI restores with `--locked-mode` so the dependency graph cannot drift silently.
- [ ] CI runs `pwsh ./verify.ps1` on every push and pull request, with no green build possible while it fails.
- [ ] A coverage stance is set per [../quality/testing.md](../quality/testing.md): mandatory paths (domain core, error-to-ProblemDetails mapping, request decode/validation paths) are covered and coverage is tracked rather than allowed to silently regress.
- [ ] Required configuration is documented in the README config table (key, env-var override, default, required); local-dev secrets go through `dotnet user-secrets`, never a committed file.
- [ ] Initial docs link to `csharp/AGENTS.md` or the repo's own equivalent fast-path contract.

## Verification

```powershell
pwsh ./verify.ps1   # restore (locked), format-check, build (warnings-as-errors), test, audit
```
