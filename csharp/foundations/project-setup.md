# Project Setup

Default repository shape, SDK policy, and bootstrap expectations for new .NET projects.

## Default Approach

Start with one solution, the `src/` plus `tests/` split, and the three-project service shape: `<App>.Api` (thin host), `<App>.Core` (domain, no ASP.NET or EF references), `<App>.Infrastructure` (EF Core, external clients). Repo-wide build policy lives in root MSBuild files, not in individual `.csproj` files. Libraries use a single project plus one test project; CLIs use `<App>.Cli` plus `<App>.Core`; going smaller than that requires a note in the project README.

### Bootstrap Commands

```bash
mkdir orders && cd orders
dotnet new sln --name Orders        # the SDK emits Orders.slnx
dotnet new web      -o src/Orders.Api
dotnet new classlib -o src/Orders.Core
dotnet new classlib -o src/Orders.Infrastructure
dotnet new classlib -o tests/Orders.UnitTests
dotnet new classlib -o tests/Orders.IntegrationTests
dotnet sln Orders.slnx add src/Orders.Api/Orders.Api.csproj \
  src/Orders.Core/Orders.Core.csproj \
  src/Orders.Infrastructure/Orders.Infrastructure.csproj \
  tests/Orders.UnitTests/Orders.UnitTests.csproj \
  tests/Orders.IntegrationTests/Orders.IntegrationTests.csproj
```

Then copy `global.json`, `Directory.Build.props`, `Directory.Packages.props`, `nuget.config`, `.editorconfig`, `.gitattributes`, `gitignore` (rename to `.gitignore`), `verify.ps1`, and `Makefile` from [../templates/README.md](../templates/README.md). Exact SDK and package versions live only in those templates — never in prose docs and never hand-typed. Rewrite the generated test projects to the xunit.v3-on-Microsoft.Testing.Platform shape from the templates (see [../quality/testing.md](../quality/testing.md)); the `dotnet new` output is a starting skeleton, not the contract.

### Preferred Tree

```text
orders/
  global.json
  Orders.slnx
  Directory.Build.props
  Directory.Packages.props
  nuget.config
  .editorconfig
  .gitattributes
  .gitignore
  LICENSE
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

### What Goes Where

- `src/Orders.Api/Program.cs`: config loading, DI composition, host lifetime, endpoint mapping — nothing else
- `src/Orders.Api/Endpoints/`: minimal-API endpoint groups; transport adapters only
- `src/Orders.Api/Telemetry/`: logging, OpenTelemetry, and health wiring behind `AddServiceTelemetry()`
- `src/Orders.Core/`: domain types, ports (interfaces), business rules; references nothing app-level
- `src/Orders.Infrastructure/Data/`: `DbContext`, migrations, repositories
- `src/Orders.Infrastructure/Clients/`: typed outbound HTTP clients
- `tests/Orders.UnitTests/`: fast, deterministic tests; no network, database, or Docker
- `tests/Orders.IntegrationTests/`: Testcontainers-backed tests behind the explicit `-Integration` switch
- `api/`: `.proto`, OpenAPI, or other wire-contract definitions
- `verify.ps1`: the one canonical gate; `Makefile` is a one-line shim delegating to it

If a repo publishes external APIs or schemas, `api/` holds the authoritative contract sources rather than generated outputs alone. The ownership rules for each shared construct are in [shared-constructs.md](shared-constructs.md); the boundary rules between the three projects are in [solution-and-project-design.md](solution-and-project-design.md).

### Thin Program.cs

`Program.cs` uses top-level statements and wires configuration, logging, DI, and lifetime — it delegates everything else to extension methods owned by the project that owns the behavior:

```csharp
var builder = WebApplication.CreateBuilder(args);

builder.Services.AddProblemDetails();
builder.Services.AddServiceTelemetry(builder.Configuration);
builder.Services.AddOrdersCore();
builder.Services.AddOrdersInfrastructure(builder.Configuration);

var app = builder.Build();

app.UseExceptionHandler();
app.MapHealthEndpoints();   // /livez and /readyz
app.MapOrdersEndpoints();

if (args.Contains("--migrate", StringComparer.Ordinal))
{
    await app.MigrateOrdersDatabaseAsync();
    return;
}

app.Run();
```

Migrations run only behind the explicit `--migrate` step — never automatically on normal startup (see [../services/database.md](../services/database.md)). A copy-ready version of this file is `program-main.cs.txt` in [../templates/README.md](../templates/README.md).

## SDK Pinning And Solution Policy

- Pin the SDK with `global.json` at the repo root from day one; copy it from the template. `rollForward: latestFeature` lets patch and feature-band updates flow without editing the file while still refusing a different major version.
- LTS releases only. Building on an STS release requires an ADR (see [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).
- One solution per repo, in `.slnx` format. Keep a `.sln` only for tooling that cannot read `.slnx`, and generate it — do not maintain two solution files by hand.
- `Directory.Build.props` sets the target framework, `Nullable`, `ImplicitUsings`, analyzers, and warnings-as-errors once for every project. Individual `.csproj` files do not override repo-wide policy.
- Stay current on supported patch releases; do not let a new repo start on a stale SDK line.

### License And Headers

- Every repo has a top-level `LICENSE` from day one. The license is a deliberate choice — the org default for internal/proprietary code, or a per-project OSI license for anything published — and a non-default pick is ADR-worthy (see [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).
- Default header policy: rely on the `LICENSE` file plus repository provenance; do not add per-file copyright headers. Add SPDX headers (`// SPDX-License-Identifier: <id>`) only when org policy or a regulatory/compliance context requires them — then apply them uniformly and enforce them in `verify.ps1`, not by hand.
- Review third-party license obligations when adding a NuGet package: confirm the license expression in the package metadata is compatible with how you ship, and capture the check in the dependency decision (see [../decisions/framework-selection.md](../decisions/framework-selection.md) and [../operations/security.md](../operations/security.md)). A copyleft or attribution-required dependency is a decision, not an accident discovered at release.

## Tool Dependencies

Track repo-local CLI tools with a tool manifest so CI and local environments resolve the same versions:

```bash
dotnet new tool-manifest            # creates .config/dotnet-tools.json
dotnet tool install dotnet-ef
dotnet tool restore                 # in bootstrap and CI, before any tool runs
```

The manifest commits exact tool versions; run tools as `dotnet ef ...` (or `dotnet tool run <tool>`). Use the manifest for CLI tools only — do not smuggle runtime dependencies into it. Global tool installs (`--global`) are a per-machine convenience, never a build prerequisite: everything `verify.ps1` needs must come from the SDK, the manifest, or the lock-file-pinned package graph.

## Build Defaults

- Central Package Management is mandatory: every package version lives in `Directory.Packages.props`; `.csproj` files carry `<PackageReference Include="..." />` with no `Version` attribute.
- Lock files are mandatory: `RestorePackagesWithLockFile=true` (set once in `Directory.Build.props`), `packages.lock.json` committed per project, CI restores with `--locked-mode`.
- `nuget.config` commits the feed list — nuget.org only by default — with package source mapping, so restores never silently pull from ambient machine-level feeds.
- NuGetAudit runs on restore (`NuGetAuditMode=all`) and high/critical advisories fail the gate.
- Warnings are errors everywhere: `TreatWarningsAsErrors=true`, `AnalysisLevel=latest-all`, `EnforceCodeStyleInBuild=true`, with `.editorconfig` as the single style and severity source (see [../quality/linting.md](../quality/linting.md)).
- Release artifacts come from `dotnet build`/`publish` in Release configuration; that is what `verify.ps1` builds and what the Dockerfile publishes. Debug is for local iteration only.
- Do not cargo-cult Native AOT or ReadyToRun into day-one repos; both are ADR-gated optimizations adopted after a measured need (see [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).

## Common Mistakes And Forbidden Patterns

- No `global.json`, so developers and CI silently build with different SDKs.
- `Version` attributes in `.csproj` files overriding or bypassing Central Package Management.
- A `Program.cs` that grows into the service instead of delegating to extension methods in the owning project.
- A `.sln` and `.slnx` maintained side by side and drifting apart.
- `Orders.Common` or `Orders.Shared` projects that quietly become the real application.
- Exact versions copied into prose docs instead of living in `templates/`.
- Build scripts that require undeclared global tools, hidden shell state, or bash-only steps in the verify path (see [cross-platform.md](cross-platform.md)).
- No `LICENSE` file, or a license copied in by reflex without deciding whether it fits how the code ships.
- Per-file copyright headers added ad hoc when no policy requires them, or SPDX headers applied to some files and not others.
- Pulling in a package without checking its license obligations against how the artifact is distributed.

## Verification And Proof

```powershell
pwsh ./verify.ps1    # restore (locked), format-check, build (warnings-as-errors), test, audit
dotnet tool restore
Get-Item LICENSE, global.json, Directory.Build.props, Directory.Packages.props, nuget.config
```

Proof is complete when `pwsh ./verify.ps1` is green, the tree matches the intended boundaries rather than a temporary prototype layout, every project appears in the `.slnx`, restore succeeds in locked mode, a deliberate `LICENSE` is present, and the header policy (none, or SPDX everywhere) is applied uniformly.
