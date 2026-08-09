# New Project Checklist

Bootstrap checklist for a new Python repo using this handbook.

## Repository Skeleton

- [ ] Run [spec intake](spec-intake.md); record project shape, bounded MVP, defaults, and required ADRs before scaffolding.
- [ ] Copy `pyproject.toml`, `.python-version`, `uv.lock` workflow, `Makefile`, `.gitignore`, `.editorconfig`, and `.dockerignore` mappings from [templates](../templates/README.md); replace every `<placeholder>`.
- [ ] Set `requires-python = ">=3.11"`, pin a current stable interpreter in `.python-version`, declare PEP 621 metadata, and select `uv_build` for the default pure-Python package.
- [ ] Create one installable package under `src/<app>/`, tests under `tests/`, and `src/<app>/py.typed` when publishing a library.
- [ ] Create only required owners: `core/`, `api/http/`, `db/`, `clients/`, `config.py` or `config/`, `telemetry/`, and `workers/`; no empty architecture theater.
- [ ] Add a thin `[project.scripts]` callable and delegating `__main__.py` for executable shapes.
- [ ] Declare Import Linter contracts so `core` cannot import adapter packages.
- [ ] Document whether the repo is a service, worker, CLI, library, or named combination in the README.

## Runtime Contract

- [ ] Pydantic settings loads and validates config once at composition; `.env.example` lists safe placeholders only.
- [ ] `logging.config.dictConfig` configures service JSON once; libraries add at most `NullHandler`.
- [ ] FastAPI services use `create_app()` plus lifespan ownership; workers/CLIs have one composition-owned root.
- [ ] Every task is owned, cancellation propagates, blocking work leaves the event loop, and shutdown drains within a configured bound.
- [ ] Every HTTP/database/broker/subprocess operation has explicit timeout, concurrency, and resource-close behavior.
- [ ] Networked shapes define `/livez`, `/readyz`, `/metrics`, structured logs, and OpenTelemetry propagation.
- [ ] Database shapes use SQLAlchemy async + asyncpg and Alembic with migrations outside normal startup.

## Proof And Delivery

- [ ] Copy [CI](../templates/github-workflows-ci.yml) so local and CI both invoke exactly `make verify`.
- [ ] Commit `uv.lock` for applications and review its initial graph; libraries document their lock policy for development/release proof.
- [ ] Configure Ruff, strict mypy, Import Linter, pytest, strict pytest-asyncio, coverage, and pip-audit per the templates.
- [ ] Add a Docker integration lane that runs `uv run pytest -m integration` against real required dependencies.
- [ ] Add `SECURITY.md`, `CODEOWNERS`, project contribution/agent docs, changelog, and `docs/runbook.md` where the project shape requires them.
- [ ] Prove the installed package/console entry point from outside the checkout; tests do not pass through accidental flat-layout imports.
- [ ] Initial documentation links to `python/AGENTS.md` or carries an equivalent copied fast-path contract.

## Verification

```bash
uv lock --check
uv sync --frozen
make verify
uv run pytest -m integration
uv build --no-sources
```
