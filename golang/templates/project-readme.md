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
- **Entrypoint**: `cmd/<app>/main.go`
- **Module**: `github.com/<org>/<repo>`
- **Go**: 1.24+ (toolchain pinned in `go.mod`)
- **Runtime surface**: <HTTP :8080 | gRPC :9090 | queue consumer | CLI binary | importable package>

## Quickstart

```bash
git clone https://github.com/<org>/<repo>.git
cd <repo>
make verify           # tidy, format, lint, vet, tests, race, vuln scan, build
cp .env.example .env  # then edit values for your environment
go run ./cmd/<app>    # start the process locally
```

`make verify` is the single gate; it must pass before any change is considered done.
See [AGENTS.md](AGENTS.md) for the full contributor contract and verification bar.

## Configuration

All configuration is loaded once in `internal/config` from environment variables
(with flag overrides where noted) and validated at startup. The process fails fast
with an actionable message if a required value is missing or malformed. Commit changes
to `.env.example` alongside any change to this table.

| Key | Type | Required | Default | Secret | Description |
|---|---|---|---|---|---|
| `LOG_LEVEL` | string | no | `info` | no | `slog` level: `debug`, `info`, `warn`, `error`. |
| `HTTP_ADDR` | string | no | `:8080` | no | Listen address for the HTTP server. |
| `HTTP_READ_TIMEOUT` | duration | no | `15s` | no | Maximum duration for reading a request, including the body. |
| `SHUTDOWN_GRACE` | duration | no | `15s` | no | Grace period for in-flight work during shutdown (maps to `cfg.ShutdownGrace`). |
| `DB_DSN` | string | yes | — | yes | DSN for the primary database. |
| `DB_MAX_OPEN_CONNS` | int | no | `25` | no | Maximum open connections in the pool. |
| `<KEY>` | `<type>` | `<yes/no>` | `<default>` | `<yes/no>` | `<what it controls; describe failure mode if unset>` |

## Architecture

<two or three sentences: the request/work flow and the major boundaries it crosses.>

Layout follows the handbook default (`cmd/` + `internal/`):

- `cmd/<app>` — process wiring, signal handling, shutdown; no business logic.
- `internal/core` — domain logic and the interfaces consumed from the outside.
- `internal/api/http`, `internal/api/grpc` — transport adapters only.
- `internal/db` — queries, migrations, transactions, storage mapping.
- `internal/config` — loading, defaults, validation, fail-fast startup.
- `internal/telemetry` — logging, metrics, tracing, health primitives.
- `api/` — published wire contracts (`.proto`, OpenAPI) when present.

Authoritative contributor rules: [AGENTS.md](AGENTS.md).
Change routing by file area: the Change Routing table in [AGENTS.md](AGENTS.md).
Architecture decisions and their rationale: [decisions/](decisions/).

## Testing

```bash
make test     # go test ./...
make race     # go test -race ./...
make cover    # coverage profile + report
```

Integration tests at the database and external-service boundaries require
<describe: e.g. a local Postgres via `docker compose up -d db`, or testcontainers>.
Every behavior change ships with a test that proves it.

## Deploy

- **Artifact**: <container image | static binary | published Go module tag>.
- **Build**: <`make build` for the dev check; release builds add `-trimpath`; e.g. `docker build` / goreleaser>.
- **Release**: tagged `v<major>.<minor>.<patch>` (the `v` prefix is required); see [decisions/](decisions/) and the release checklist.
- **Rollout**: <how it ships — pipeline, command, or platform — and how to roll back.>
- **Health**: <`/livez`, `/readyz`, or N/A for libraries and CLIs>.

## Ownership And Support

- **Owners**: see [CODEOWNERS](.github/CODEOWNERS) for area-level ownership.
- **Team / contact**: <team name, channel, or alias>.
- **On-call / escalation**: <runbook link, pager, or "best effort, file an issue">.
- **Issues**: <issue tracker URL>.
