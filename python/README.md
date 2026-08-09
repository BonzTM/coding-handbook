# Python Project Handbook

This handbook is the default engineering contract for new Python repositories. It is not a Python language tutorial. It exists to make services, workers, CLIs, and libraries converge on the same structure, runtime behavior, dependency posture, and proof of correctness.

## Start Here

- Humans: read this file, then follow the reading path for your project shape.
- Agents: read [AGENTS.md](AGENTS.md) first (it includes the change-routing table), then the relevant topical docs and recipes.
- Default assumptions unless a repo says otherwise:
  - Python 3.11 is the compatibility floor; development and CI use a current stable pin from `.python-version`
  - one PEP 621 distribution package per repo, with `pyproject.toml` at root and a `src/<app>/` plus `tests/` layout
  - thin `__main__.py` and `[project.scripts]` entry points
  - FastAPI on uvicorn, Pydantic v2 at trust boundaries, SQLAlchemy 2.0 with asyncpg, stdlib `logging`, pytest, and asyncio first
  - env-driven configuration through pydantic-settings with fail-fast startup validation
  - lock-check, frozen sync, format-check, lint, import-boundary check, strict types, test, and audit as baseline proof, all wrapped by `make verify`

## Reading Paths

| If you are building... | Read in this order |
|---|---|
| HTTP service | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/package-design.md](foundations/package-design.md) -> [foundations/configuration.md](foundations/configuration.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [services/http-services.md](services/http-services.md) -> [operations/observability.md](operations/observability.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-http-endpoint.md](recipes/add-http-endpoint.md); verified exemplar: [reference/exampleservice/](reference/exampleservice/) |
| Server-rendered web app | the HTTP service path above, inserting [services/web-apps.md](services/web-apps.md) after [services/http-services.md](services/http-services.md) — Jinja2 templates, static assets, server-side sessions, CSRF, and browser security headers layer on the same service skeleton |
| gRPC service | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/package-design.md](foundations/package-design.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [services/grpc-services.md](services/grpc-services.md) -> [foundations/errors-and-logging.md](foundations/errors-and-logging.md) -> [services/database.md](services/database.md) -> [operations/observability.md](operations/observability.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-grpc-method.md](recipes/add-grpc-method.md) |
| Background worker | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md) -> [foundations/configuration.md](foundations/configuration.md) -> [operations/observability.md](operations/observability.md) -> [operations/security.md](operations/security.md) -> [recipes/add-background-worker.md](recipes/add-background-worker.md) |
| Event-driven service or async worker | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [services/eventing-and-messaging.md](services/eventing-and-messaging.md) -> [services/database.md](services/database.md) -> [operations/observability.md](operations/observability.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-event-publisher.md](recipes/add-event-publisher.md) -> [recipes/add-event-consumer.md](recipes/add-event-consumer.md) |
| CLI tool | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/style-and-review.md](foundations/style-and-review.md) -> [foundations/configuration.md](foundations/configuration.md) -> [decisions/framework-selection.md](decisions/framework-selection.md) -> [recipes/add-cli-command.md](recipes/add-cli-command.md) |
| Library | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/package-design.md](foundations/package-design.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [foundations/typing-discipline.md](foundations/typing-discipline.md) -> [quality/testing.md](quality/testing.md) -> [foundations/errors-and-logging.md](foundations/errors-and-logging.md) -> [checklists/release.md](checklists/release.md) -> [recipes/release-library-version.md](recipes/release-library-version.md) |

Every shape also adopts [quality/linting.md](quality/linting.md) and the committed [templates/](templates/) scaffolding, runs `make verify` as its proof gate, and follows [foundations/data-modeling.md](foundations/data-modeling.md), [foundations/serialization.md](foundations/serialization.md), and [foundations/time.md](foundations/time.md) for type, wire-shape, and clock decisions. Networked services additionally follow [operations/resilience.md](operations/resilience.md), [operations/deployment.md](operations/deployment.md), and [operations/operability.md](operations/operability.md).

Cross-cutting concerns apply across shapes: [services/caching.md](services/caching.md) and [foundations/configuration.md](foundations/configuration.md) (feature flags) affect most services, and [foundations/git-workflow.md](foundations/git-workflow.md) governs commits, branches, and changelog discipline everywhere.

## Non-Negotiables

- Keep one installable package under `src/<app>/`; tests live outside `src/`, and import paths never depend on the checkout root.
- Keep `__main__.py` boring. It delegates to a composition root that wires config, logging, dependencies, signals, and shutdown; neither holds business rules.
- Keep `core/` framework-free. It imports nothing from `api/`, `db/`, `clients/`, or `telemetry/`; `lint-imports` enforces the direction.
- Type public and cross-layer boundaries. Mypy strict mode is the gate; libraries ship `py.typed`.
- Validate untrusted input with Pydantic at the boundary, then map it into plain typed domain objects. Domain code never imports FastAPI or Pydantic models.
- Every asyncio task has an owner. Use `TaskGroup`, explicit deadlines, bounded concurrency, and re-raise `CancelledError` after cleanup.
- Create one lifespan-owned `httpx.AsyncClient` with explicit timeouts; never use blocking HTTP or file/process calls on the event loop.
- Use SQLAlchemy 2.0 plus asyncpg and Alembic; parameterize SQL, test against real PostgreSQL, and never auto-apply migrations on normal startup.
- Configure stdlib `logging` once at the composition root. Libraries configure no handlers beyond `NullHandler`; log once at the boundary that can act.
- Commit `uv.lock` for applications. Do not commit secrets, unpinned application dependency graphs, or new framework defaults without explicit justification.

## Default Stack

| Concern | Default | Reach for something else when |
|---|---|---|
| Environment and lock workflow | `uv`, committed `uv.lock` for applications | an existing organization workflow mandates pip-tools |
| Packaging | PEP 621 `pyproject.toml`, `src/` layout, `uv_build` for pure-Python distributions | extension modules or unusual build hooks require Hatchling or a specialized backend |
| HTTP service | FastAPI on uvicorn, thin routers, lifespan-owned resources | Starlette or Litestar is mandated or measured needs justify an ADR; Django is for a different application shape |
| Validation and config | Pydantic v2 at trust boundaries; pydantic-settings for startup config | plain stdlib parsing is sufficient for a small library or CLI |
| Data access | SQLAlchemy 2.0 async + asyncpg; Alembic migrations | the repo does not persist data, or an ADR selects a different datastore |
| Outbound HTTP | one shared `httpx.AsyncClient` with explicit `httpx.Timeout` | a synchronous CLI has no async runtime |
| Logging | stdlib `logging` configured with `dictConfig`; JSON for services | required sinks or context propagation justify structlog through an ADR |
| Metrics and tracing | prometheus-client and OpenTelemetry | the org mandates another interoperable backend |
| Testing | pytest, pytest-asyncio strict mode, hand-rolled fakes at `Protocol` seams, real-boundary integration tests | Hypothesis is ADR-approved for parser-heavy input spaces |
| CLI parsing | stdlib `argparse` | Typer or Click is justified by a substantial command tree and UX requirements |

The full table, escalation rules, and forbidden packages live in [decisions/framework-selection.md](decisions/framework-selection.md).

## Handbook Map

- [AGENTS.md](AGENTS.md) - fast-path contract and change routing for autonomous agents and reviewers
- [maintainer-reference.md](maintainer-reference.md) - architecture, rationale, and deeper guidance
- `foundations/` - package layout, contracts, config, errors, asyncio, typing, time, data modeling, serialization, [shared constructs](foundations/shared-constructs.md), and [git-workflow.md](foundations/git-workflow.md)
- `quality/` - pytest strategy, Ruff/mypy/import-boundary policy, coverage, and proof commands
- `services/` - HTTP, gRPC, server-rendered web, messaging, database, and cache guidance
- `operations/` - telemetry, security, resilience, deployment, operability/SLOs, data handling/PII, CI, and releases
- `decisions/` ([README.md](decisions/README.md)) - ADRs plus dependency and framework selection rules
- `checklists/` ([README.md](checklists/README.md)) and `recipes/` ([README.md](recipes/README.md)) - executable startup, review, release, handoff, and implementation guidance
- `templates/` ([README.md](templates/README.md)) - committed copy-paste scaffolding (`pyproject.toml`, Makefile, `.python-version`, CI workflows, Dockerfile, source skeletons, and project docs); exact version pins live here and in the reference exemplar, not prose
- `reference/` - [exampleservice](reference/exampleservice/), a complete `make verify`-green FastAPI+Postgres sidecar composing the handbook patterns end to end
- Team process: [onboarding-and-handoff.md](onboarding-and-handoff.md), [checklists/incident-response.md](checklists/incident-response.md), [glossary.md](glossary.md), and [CONTRIBUTING.md](CONTRIBUTING.md)

## What This Handbook Optimizes For

- code that still looks obvious six months later
- boundaries that make testing and refactoring cheaper
- runtime behavior that is safe under load and easy to debug
- defaults that keep agents from inventing new architecture every task
- minimal dependency surface unless there is a clear return on complexity

## Where To Go Next

- New repo bootstrap: [checklists/new-project.md](checklists/new-project.md)
- Resolve WHAT decisions before a build: [checklists/spec-intake.md](checklists/spec-intake.md)
- Active agent work and change routing: [AGENTS.md](AGENTS.md)
- Choose a third-party package: [decisions/framework-selection.md](decisions/framework-selection.md)
- Record an architecture decision: [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md)
- Copy-paste scaffolding: [templates/README.md](templates/README.md)
- Take over or hand off a project: [onboarding-and-handoff.md](onboarding-and-handoff.md)
- Lint and type policy: [quality/linting.md](quality/linting.md)
- Make a networked service production-ready: [operations/resilience.md](operations/resilience.md), [operations/deployment.md](operations/deployment.md), [operations/operability.md](operations/operability.md)
- Implement a specific change: [recipes/README.md](recipes/README.md)
- Look up a handbook term: [glossary.md](glossary.md)
- Change the handbook itself: [CONTRIBUTING.md](CONTRIBUTING.md), [foundations/git-workflow.md](foundations/git-workflow.md)
- Copy the complete worked example: [reference/exampleservice/](reference/exampleservice/)
