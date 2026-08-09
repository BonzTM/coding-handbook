<!-- Copy to README.md. Replace every <placeholder>; delete non-applicable shape rows. -->

# <project-name>

<one-line statement of purpose and consumers.>

## Project Shape

- **Shape:** <service | worker | CLI | library>
- **Entrypoint:** `src/<app>/__main__.py`
- **Distribution/import package:** `<distribution-name>` / `<app>`
- **Python:** `>=3.11`; development pin in `.python-version`
- **Runtime surface:** <HTTP :8080 | queue consumer | CLI | importable library>

## Quickstart

```bash
git clone https://github.com/<org>/<repo>.git
cd <repo>
cp .env.example .env
uv sync --frozen
make verify
uv run <app>
```

`make verify` is the single gate. See [AGENTS.md](AGENTS.md) for routing and invariants.

## Configuration

Settings are loaded once with pydantic-settings and validated before resources/listeners open. Update this table, `.env.example`, and deployment configuration together.

| Key | Type | Required | Default | Secret | Description |
|---|---|---|---|---|---|
| `APP_ENV` | enum | no | `local` | no | Runtime environment name. |
| `LOG_LEVEL` | enum | no | `INFO` | no | Process log threshold. |
| `LOG_JSON` | bool | no | `false` | no | JSON service logs when true. |
| `HTTP_HOST` | string | no | `127.0.0.1` | no | Bind host; deployments use `0.0.0.0`. |
| `HTTP_PORT` | int | no | `8080` | no | HTTP port. |
| `DATABASE_URL` | URL | yes | — | yes | SQLAlchemy async PostgreSQL URL. |
| `<KEY>` | `<type>` | `<yes/no>` | `<default>` | `<yes/no>` | `<purpose and failure mode>` |

## Architecture

<two or three sentences describing the main flow and boundaries.>

- `src/<app>/core/` — domain values, use cases, and consumer-owned Protocol ports.
- `src/<app>/api/` — HTTP/gRPC transport adapters and boundary DTOs.
- `src/<app>/db/` — SQLAlchemy mappings, repositories, sessions, and Alembic integration.
- `src/<app>/clients/`, `workers/`, `telemetry/` — outward adapters with explicit ownership.
- `tests/` — unit tests plus marked real-boundary integration tests.

## Testing

```bash
make test
make test-integration
make verify
```

Integration prerequisites: <Docker/PostgreSQL/broker setup and required environment keys>.

## Deploy

- **Artifact:** <digest-addressed container | Python distribution>.
- **Migrations:** `uv run alembic upgrade head` in an explicit pre-rollout job, never normal startup.
- **Rollout/rollback:** <pipeline and rollback procedure>.
- **Health:** </livez, /readyz, /metrics | N/A>.

## Ownership And Support

- **Owners:** [CODEOWNERS](.github/CODEOWNERS)
- **Team/contact:** <team and channel>
- **On-call:** <runbook/pager or support policy>
- **Issues:** <issue tracker URL>
