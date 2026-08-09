<!-- Copy to AGENTS.md. Replace every <placeholder>; keep this project fast path short. -->

# AGENTS.md - <project-name> Contract

Fast-path contract for agents and reviewers. The full rules live in the Python handbook at `<handbook-url-or-path>`; this file may strengthen but never weaken them.

## Purpose

- This project is a <service | worker | CLI | library>: <one-line purpose>.
- Use [README.md](README.md) for shape, config, and operation.
- Run `make verify` before claiming a change complete.

## Repo-Wide Invariants

- One PEP 621 distribution under `src/<app>/`; `requires-python = ">=3.11"`; `.python-version` is the development pin.
- `core` imports no FastAPI, Pydantic DTO, SQLAlchemy, HTTPX, broker, or telemetry SDK type; Import Linter enforces the graph.
- Pydantic validates trust boundaries; plain typed values cross into core.
- Every asyncio task is TaskGroup-owned, bounded, cancellation-safe, and awaited.
- SQLAlchemy async + asyncpg + Alembic; migrations are explicit deploy work.
- stdlib logging is configured once; libraries configure no handler beyond `NullHandler`.
- Every behavior change has tests; real PostgreSQL/broker behavior uses real integration proof.
- uv owns lock/environment/build; `uv.lock` is committed for this application.
- **Project-specific:** <additional invariant or "none">.

## Change Routing

| If changing... | Start in | Read first |
|---|---|---|
| startup/lifecycle | `src/<app>/main.py`, `__main__.py` | handbook `foundations/project-setup.md` |
| domain behavior | `src/<app>/core/` | handbook `foundations/package-design.md` |
| HTTP/gRPC | `src/<app>/api/` | handbook `services/http-services.md` or `grpc-services.md` |
| database/schema | `src/<app>/db/`, `alembic/` | handbook `services/database.md` |
| outbound calls | `src/<app>/clients/` | handbook `operations/resilience.md` |
| workers/events | `src/<app>/workers/` | handbook `services/eventing-and-messaging.md` |
| config | `src/<app>/config.py`, `.env.example`, README | handbook `foundations/configuration.md` |
| telemetry/security | boundary adapter | handbook `operations/observability.md` / `security.md` |
| dependencies/tooling | `pyproject.toml`, `uv.lock`, Makefile | handbook `decisions/framework-selection.md` |
| <project-specific area> | `<path>` | `<doc>` |

## Working Norms

- Keep changes small and preserve existing boundaries.
- Add no dependency without written need, lock diff review, typing/license/security review, and proof.
- Fix or report failed verification; do not hide it.

## Baseline Verification

```bash
make verify
```

The ordered gate is lock-check, frozen sync, format-check, lint, imports, types, test, and audit. Run narrow proof first, then the full gate locally and in CI.
