<!--
TEMPLATE -> repo root README.md
Fill every <PLACEHOLDER>. Delete sections that do not apply to the project shape,
but keep the section order so every handbook-built repo reads the same way.
The config-key table is REQUIRED: foundations/configuration.md mandates documenting
every supported config key here or in linked operator docs.
-->

# <project-name>

<one-line statement of what this project does and who consumes it.>

## Project Shape

- **Shape**: <service | worker | CLI | library> (pick one; delete the others below)
- **Entrypoint**: `src/<App>.Api/Program.cs`
- **Solution**: `<App>.slnx`
- **.NET**: 10 (LTS; SDK pinned in `global.json`, `rollForward: latestFeature`)
- **Runtime surface**: <HTTP :8080 | gRPC :8080 | queue consumer | CLI binary | NuGet package>

## Quickstart

```bash
git clone https://github.com/<org>/<repo>.git
cd <repo>
pwsh ./verify.ps1                  # restore (locked), format-check, build (warnings-as-errors), test, audit
dotnet run --project src/<App>.Api # start the process locally
```

Local secrets (never in `appsettings.json`) go through user-secrets:

```bash
dotnet user-secrets set "ConnectionStrings:Default" "<local connection string>" --project src/<App>.Api
```

`pwsh ./verify.ps1` is the single gate; it must pass before any change is considered done.
See [AGENTS.md](AGENTS.md) for the full contributor contract and verification bar.

## Configuration

Configuration is `Microsoft.Extensions.Configuration`: committed `appsettings.json`
holds non-secret defaults, environment variables override it (`:` in a config path
becomes `__` in the env-var name), and options classes are validated at startup
(`ValidateDataAnnotations` + `ValidateOnStart`). The process fails fast with an
actionable message if a required value is missing or malformed. Update this table
in the same change as any key you add or change.

| Key (env form) | Type | Required | Default | Secret | Description |
|---|---|---|---|---|---|
| `ASPNETCORE_ENVIRONMENT` | string | no | `Production` | no | Host environment: `Development`, `Staging`, `Production`. |
| `ASPNETCORE_URLS` | string | no | `http://+:8080` (container image default) | no | Listen URLs for Kestrel. |
| `Logging__LogLevel__Default` | string | no | `Information` | no | Minimum log level: `Debug`, `Information`, `Warning`, `Error`. |
| `ConnectionStrings__Default` | string | yes | — | yes | Connection string for the primary database. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | string | no | — | no | OTLP endpoint for traces/metrics/logs; telemetry export is off when unset. |
| `HostOptions__ShutdownTimeout` | timespan | no | `<set in Program.cs, e.g. 00:00:15>` | no | Grace period for draining in-flight work on SIGTERM; must stay under the platform's termination grace. |
| `<Key__Path>` | `<type>` | `<yes/no>` | `<default>` | `<yes/no>` | `<what it controls; describe failure mode if unset>` |

## Architecture

<two or three sentences: the request/work flow and the major boundaries it crosses.>

Layout follows the handbook default (`src/` + `tests/`, boundaries enforced by
project references):

- `src/<App>.Api` — thin host: `Program.cs` composition root, endpoint groups, telemetry wiring. No business logic.
- `src/<App>.Core` — domain types, ports (interfaces), business rules. References no ASP.NET or EF Core.
- `src/<App>.Infrastructure` — EF Core (`Data/`), typed outbound clients (`Clients/`). References Core.
- `tests/<App>.UnitTests`, `tests/<App>.IntegrationTests` — xUnit v3; integration tests use Testcontainers.
- `api/` — published wire contracts (`.proto`, OpenAPI) when present.

Authoritative contributor rules: [AGENTS.md](AGENTS.md).
Change routing by file area: the Change Routing table in [AGENTS.md](AGENTS.md).
Architecture decisions and their rationale: [decisions/](decisions/).

## Testing

```bash
pwsh ./verify.ps1               # full offline gate, including the unit suite
pwsh ./verify.ps1 -Integration  # adds the integration suite (requires Docker for Testcontainers)
dotnet test tests/<App>.UnitTests   # narrowest loop while iterating
```

Integration tests at the database and external-service boundaries provision real
dependencies via Testcontainers, or reuse a running instance when
`TEST_DATABASE_CONNECTION` is set (that is how CI wires its Postgres service
container). Every behavior change ships with a test that proves it.

## Deploy

- **Artifact**: <container image | published NuGet package | dotnet tool>.
- **Build**: multi-stage `Dockerfile` (`sdk:10.0` build, chiseled `aspnet` runtime); release builds stamp `VERSION`/`COMMIT`.
- **Release**: tagged `v<major>.<minor>.<patch>` (the `v` prefix is required); the release workflow re-runs `verify.ps1` and publishes the image.
- **Migrations**: applied by an explicit step (`dotnet <App>.Api.dll --migrate` Job) before rollout — never on normal startup.
- **Rollout**: <how it ships — pipeline, command, or platform — and how to roll back.>
- **Health**: `/livez` (liveness, always cheap), `/readyz` (readiness, checks dependencies) — or N/A for libraries and CLIs.

## Ownership And Support

- **Owners**: see [CODEOWNERS](.github/CODEOWNERS) for area-level ownership.
- **Team / contact**: <team name, channel, or alias>.
- **On-call / escalation**: <runbook link, pager, or "best effort, file an issue">.
- **Issues**: <issue tracker URL>.
