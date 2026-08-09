# AGENTS.md - Python Project Contract

This is the authoritative fast-path contract for autonomous agents working in a new Python repository.
Read this file first: it carries the repo-wide invariants, the change-routing table, and the verification bar. Use [maintainer-reference.md](maintainer-reference.md) when you need slower-path architecture and rationale.

## Purpose

- Use this file for repo-wide invariants, change defaults, change-to-file routing, and the verification bar.
- Use [maintainer-reference.md](maintainer-reference.md) for package maps, lifecycle guidance, test taxonomy, and troubleshooting.
- For the full catalogs, see the [recipes/README.md](recipes/README.md), [checklists/README.md](checklists/README.md), and [decisions/README.md](decisions/README.md) indexes.

## Source Of Truth

- This file is the fast path. More detailed docs refine it; they do not weaken it.
- Layout, Python pinning, PEP 621 metadata, and uv defaults live in [foundations/project-setup.md](foundations/project-setup.md).
- Package and dependency boundaries live in [foundations/package-design.md](foundations/package-design.md) and [decisions/framework-selection.md](decisions/framework-selection.md); reusable constructs (`core`, `api/http`, `db`, `clients`, `config`, `telemetry`, `workers`, and test utilities) live in [foundations/shared-constructs.md](foundations/shared-constructs.md).
- Schema, API, and data-boundary rules live in [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md); type modeling in [foundations/data-modeling.md](foundations/data-modeling.md); the JSON/wire boundary in [foundations/serialization.md](foundations/serialization.md).
- Runtime correctness rules live in [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md), [foundations/errors-and-logging.md](foundations/errors-and-logging.md), and [foundations/configuration.md](foundations/configuration.md); typing rules live in [foundations/typing-discipline.md](foundations/typing-discipline.md).
- Proof expectations live in [quality/testing.md](quality/testing.md) and the relevant service or operations docs.
- Ruff, mypy, and import-boundary policy lives in [quality/linting.md](quality/linting.md); time and clock discipline in [foundations/time.md](foundations/time.md); copy-paste scaffolding and exact version pins in [templates/README.md](templates/README.md).
- Architecture decisions and their rationale live in [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).
- Team-process docs — [onboarding-and-handoff.md](onboarding-and-handoff.md), [CONTRIBUTING.md](CONTRIBUTING.md), [checklists/incident-response.md](checklists/incident-response.md), and [glossary.md](glossary.md) — serve humans running the team and handbook; they are not needed to build an app.
- The complete worked example is [reference/exampleservice/](reference/exampleservice/), a `make verify`-green FastAPI+Postgres sidecar; mirror its shape instead of inventing one.

## Fast Path

1. Read this file and identify the project shape from [README.md](README.md). For a brand-new build, run [checklists/spec-intake.md](checklists/spec-intake.md) first.
2. Route the change through the [Change Routing](#change-routing) table below; do not guess where code belongs.
3. Read the relevant foundations doc before editing code in a new area.
4. Implement with the repo defaults unless the repo has already documented an exception.
5. Prove the change with the narrowest meaningful tests first, then the repo-wide baseline.

## Repo-Wide Invariants

- **Runtime floor and pin**: `requires-python = ">=3.11"`; `.python-version` pins a current stable interpreter for development and CI. Code uses no newer language or stdlib feature unless the floor moves deliberately.
- **Single distribution**: one PEP 621 `pyproject.toml` at root, one installable package under `src/<app>/`, and tests under `tests/`. Extra distributions or workspaces require architectural review.
- **Boundary enforcement**: `core` imports nothing from `api`, `db`, `clients`, or `telemetry`; Import Linter makes violations fail `make imports`. Adapters depend inward through core-owned `Protocol` ports.
- **Thin entrypoint**: `[project.scripts]` points to a narrow callable; `__main__.py` delegates. Composition owns config, logging, resources, signals, and shutdown, not business rules.
- **Validation boundary**: Pydantic v2 owns HTTP DTOs, configuration, and message payload parsing. Map once into plain typed classes or dataclasses; `core` never imports FastAPI.
- **Typing discipline**: mypy strict mode gates the repo; public and cross-layer boundaries are typed, IDs use `NewType` where confusion is possible, and libraries ship `py.typed`.
- **Async discipline**: every task has an owner; use `asyncio.TaskGroup`, `asyncio.timeout()`, and bounded concurrency. Re-raise `CancelledError` after cleanup; move blocking work off the event loop.
- **Persistence**: SQLAlchemy 2.0 async + asyncpg + Alembic; parameterized SQL only. Migrations run as an explicit deploy step, never normal startup.
- **Logging**: stdlib `logging` configured once with `dictConfig` at composition; JSON for services, `NullHandler` only for libraries, and log once at the acting boundary.
- **Testing**: every behavior change needs tests. Database and external boundaries need real integration coverage; hand-rolled fakes at `Protocol` seams are the default unit-test double.
- **Observability**: networked behavior includes relevant structured logs, low-cardinality metrics, traces, `/livez`, `/readyz`, and `/metrics` behavior.
- **Dependency posture**: uv owns environments and locking; applications commit `uv.lock`; every non-trivial dependency needs rationale and a reviewed lockfile diff.

## Change Routing

Use this when you know what kind of change you are making but not the file set. Start Here is what you read and touch first; Also Update is the sync surface the change normally drags along; Verify Or Confirm is the proof.

| Change Type | Start Here | Also Update | Verify Or Confirm |
|---|---|---|---|
| Python pin, package metadata, repo layout | `.python-version`, `pyproject.toml`, `src/`, `tests/`, [foundations/project-setup.md](foundations/project-setup.md) | `uv.lock`, CI, Dockerfile, onboarding docs | `uv lock --check`, `uv sync --frozen`, `make verify` |
| HTTP or gRPC contract shape, schema, compatibility, generated stubs | `api/**`, transport package, [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md), [recipes/deprecate-and-remove-contract.md](recipes/deprecate-and-remove-contract.md) | clients, generated-code policy, release notes, tests | schema/generation check, compatibility review |
| Event producers, consumers, payload contracts | `src/<app>/workers/`, `api/**`, [services/eventing-and-messaging.md](services/eventing-and-messaging.md), [recipes/add-event-publisher.md](recipes/add-event-publisher.md), [recipes/add-event-consumer.md](recipes/add-event-consumer.md) | outbox/inbox storage, telemetry, DLQ policy, tests | contract, replay, idempotency, real-broker integration tests |
| HTTP endpoint, middleware, request or response shape | `src/<app>/api/http/`, [services/http-services.md](services/http-services.md), [recipes/add-http-endpoint.md](recipes/add-http-endpoint.md), [recipes/add-http-middleware.md](recipes/add-http-middleware.md) | core use case, Pydantic DTOs, telemetry, router registration, tests | FastAPI client tests, smoke test, readiness behavior |
| Server-rendered pages, templates, static assets, sessions, CSRF | `src/<app>/api/http/`, [services/web-apps.md](services/web-apps.md), [services/http-services.md](services/http-services.md) | Jinja2 templates, session store, security headers, PRG flows, tests | template-render test, CSRF negative test, cookie flags, XSS probe |
| gRPC proto or server method | `api/**`, `src/<app>/api/grpc/`, [services/grpc-services.md](services/grpc-services.md), [recipes/add-grpc-method.md](recipes/add-grpc-method.md) | generated stubs, interceptors, error mapping, docs | generation diff clean, service tests, `grpcurl` |
| Business logic or domain rules | `src/<app>/core/`, [foundations/package-design.md](foundations/package-design.md) | unit tests, transport mappings if contracts changed | targeted unit tests, relevant integration coverage |
| Errors, status mapping, structured logging | [foundations/errors-and-logging.md](foundations/errors-and-logging.md), [foundations/serialization.md](foundations/serialization.md) | exception mapping, error DTOs, log fields at acting boundary | error-mapping tests; log-once review |
| Code review heuristics, comments, naming, API shape | [foundations/style-and-review.md](foundations/style-and-review.md) | docstrings on public contracts, review checklist | `make format-check`, `make lint`, public API reads as a contract |
| DB queries, schema, transaction behavior | `src/<app>/db/`, [services/database.md](services/database.md), [recipes/add-database-feature.md](recipes/add-database-feature.md), [recipes/add-migration.md](recipes/add-migration.md) | models/tables, repositories, Alembic history, callers, telemetry | migration apply/rollback proof, real-Postgres tests |
| Configuration keys or startup defaults | `src/<app>/config.py` or `config/`, [foundations/configuration.md](foundations/configuration.md), [recipes/add-config-key.md](recipes/add-config-key.md) | `.env.example`, startup composition, docs, tests | settings tests; startup fails with required value removed |
| Background workers, schedulers, shutdown paths | `src/<app>/workers/`, [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md), [recipes/add-background-worker.md](recipes/add-background-worker.md), [recipes/add-scheduled-job.md](recipes/add-scheduled-job.md) | root task, telemetry, readiness semantics | cancellation/drain test; no leaked tasks |
| Async correctness: task ownership, cancellation, deadlines, concurrency | [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md) | callers, TaskGroups, semaphores, blocking calls | Ruff ASYNC clean; deterministic async tests |
| External HTTP or gRPC client | `src/<app>/clients/`, [recipes/add-external-client.md](recipes/add-external-client.md), [operations/resilience.md](operations/resilience.md) | core-owned `Protocol`, timeout/retry policy, telemetry, SSRF checks | stub-server tests, timeout/cancellation tests |
| Logging, metrics, traces, health endpoints | `src/<app>/telemetry/`, [operations/observability.md](operations/observability.md), [recipes/add-metric.md](recipes/add-metric.md) | routers, repositories, workers, dashboards/alerts if present | `/livez`, `/readyz`, `/metrics`, cardinality and log-schema review |
| Security-sensitive boundary | boundary module, [operations/security.md](operations/security.md) | validation, auth, secret handling, exposure docs | `make audit`, negative tests, [checklists/security-review.md](checklists/security-review.md) |
| Audit logging of security-relevant actions | [operations/security.md](operations/security.md) | audit sink/retention, actor/resource fields, access controls | event emitted at action boundary; retention/tamper controls honored |
| Data classification, PII, retention, compliance | [operations/data-handling.md](operations/data-handling.md) | field classification, redaction, retention jobs, export/delete paths | no PII in logs/metrics; retention/delete tests |
| Pre-build WHAT decisions | [checklists/spec-intake.md](checklists/spec-intake.md) | acceptance criteria, open/defaulted assumptions, design docs | spec intake resolved before code; criteria testable |
| Idempotent HTTP writes | [recipes/add-idempotent-write.md](recipes/add-idempotent-write.md), [services/http-services.md](services/http-services.md) | key storage, dedupe window, replay response, tests | duplicates collapse to one effect; replay returns stored result |
| CLI commands or options | `[project.scripts]`, `src/<app>/`, [recipes/add-cli-command.md](recipes/add-cli-command.md) | `__main__.py`, config, help, README | command tests, `--help`, package install and smoke run |
| Build, CI, containers, release automation | `.github/workflows/`, Makefile, Dockerfile, [operations/ci-and-release.md](operations/ci-and-release.md) | versioning, changelog, packaging | `make verify`, `uv build`, container smoke test |
| New dependency or framework | [decisions/framework-selection.md](decisions/framework-selection.md), [recipes/bump-dependency.md](recipes/bump-dependency.md) | `pyproject.toml`, `uv.lock`, caller, proof docs | rationale written, lock diff understood, audit clean |
| Ruff, mypy, import-boundary, or format policy | `pyproject.toml` (`[tool.importlinter]`), [quality/linting.md](quality/linting.md) | Makefile, CI, suppression justification | format, lint, imports, and types targets green |
| Time, clocks, timeouts, scheduling | [foundations/time.md](foundations/time.md) | injected clock/sleep seams, callers reading wall time | deterministic tests; no sleeps as synchronization |
| Domain modeling: enums, typed IDs, optional values, collections | [foundations/data-modeling.md](foundations/data-modeling.md) | serialization, DB mapping, validation | construction and round-trip tests; mypy clean |
| Serialization / JSON wire shape | [foundations/serialization.md](foundations/serialization.md) | Pydantic DTOs, compatibility, golden tests | round-trip/golden tests; unknown-field policy exercised |
| Non-obvious or hard-to-reverse architecture decision | [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md) | project `decisions/`, superseded ADR, README/onboarding | ADR has status, alternatives, consequences before merge |
| Project ownership transfer or onboarding | [onboarding-and-handoff.md](onboarding-and-handoff.md), [checklists/handoff.md](checklists/handoff.md) | CODEOWNERS, secret rotation, on-call, deploy access, open decisions | checklist complete; new owner runs `make verify` and deploy dry-run |
| New-repo scaffolding | [reference/exampleservice/](reference/exampleservice/), [templates/README.md](templates/README.md), [checklists/new-project.md](checklists/new-project.md) | copied artifacts and every `<placeholder>` | `make verify` green in fresh repo |
| Outbound resilience: timeouts, retries, backoff, breakers, limits | [operations/resilience.md](operations/resilience.md) | clients, consumers, attempt telemetry | timeout on every call; load test shows bounded shedding |
| Containerization, runtime limits, rollout | [operations/deployment.md](operations/deployment.md), [templates/Dockerfile](templates/Dockerfile) | CI image build, probes, limits, shutdown grace | non-root image; probes wired; `make verify` green |
| SLOs, alerting, runbooks, on-call | [operations/operability.md](operations/operability.md), [templates/runbook.md](templates/runbook.md) | dashboards/alerts, error budgets, on-call docs | SLOs, symptom alerts, current runbook |
| Caching, invalidation, keys, stampede control | [services/caching.md](services/caching.md) | callers, hit/miss telemetry, invalidation, config | correctness under invalidation; bounded memory; no stale regression |
| Feature flags, gating, rollout toggles | [foundations/configuration.md](foundations/configuration.md) | typed settings/accessor, default, cleanup date, tests | default/override tests; stale flag removal tracked |
| Git workflow, branching, commits, CHANGELOG | [foundations/git-workflow.md](foundations/git-workflow.md), [operations/ci-and-release.md](operations/ci-and-release.md) | changelog, release notes, PR template | conventions followed; required changelog entry present |
| Tagging or publishing a distribution version | [recipes/release-library-version.md](recipes/release-library-version.md), [checklists/release.md](checklists/release.md) | CHANGELOG, version/tag, wheel/sdist metadata, consumers | `uv build`; install/inspect artifacts; `make verify` green |
| Reusable shared construct (`core` ports, adapters, repositories, clients, telemetry, test utilities) | [foundations/shared-constructs.md](foundations/shared-constructs.md) | consuming modules and composition wiring | one owner per construct; no `utils.py` junk drawer; imports gate green |

## High-Value Boundaries

- `src/<app>/__main__.py` and the console entry callable delegate to composition and own no business rules.
- `src/<app>/core/` owns domain types, use cases, and consumer-defined `Protocol` ports; it imports nothing from adapters.
- `src/<app>/api/http/` owns FastAPI routers, HTTP DTOs, validation, and transport-to-core mapping; routers never query SQLAlchemy or call HTTPX directly.
- `src/<app>/api/grpc/` owns generated-stub adaptation, status mapping, and interceptors when gRPC exists.
- `src/<app>/db/` owns engine/session construction, SQLAlchemy mappings or Core tables, Alembic migrations, and repository implementations.
- `src/<app>/clients/` owns lifespan-shared outbound HTTPX/gRPC clients and resilience mechanics.
- `src/<app>/config.py` or `config/` owns pydantic-settings loading, defaults, validation, and startup failure behavior.
- `src/<app>/telemetry/` owns logging, OpenTelemetry, Prometheus metrics, and health primitives; `workers/` owns broker settlement and loop lifecycle.

## Proof Hints

- Transport changes need narrow adapter tests and one manual or scripted smoke test.
- Persistence changes need real PostgreSQL integration tests; fakes do not prove SQL or migrations.
- Background work needs cancellation, bounded-drain, and leaked-task proof.
- Config changes are done only when invalid input fails startup and `.env.example` stays synchronized.
- Dependency changes are done only when rationale and the `uv.lock` diff are understood.

## Working Norms

- Prefer small, reviewable changes over broad cleanup.
- Match the repo's current shape unless the task is explicitly architectural.
- Do not bypass boundaries: routers do not query the database, repositories do not contain transport logic, and composition does not absorb business rules.
- When adding a dependency, document why the stdlib or current stack is insufficient.
- Write the failing or proving test before claiming changed behavior complete whenever practical.
- If verification fails, fix it or report it clearly. Do not claim the change is done.

## Baseline Verification

| Goal | Command | Expectation |
|---|---|---|
| lockfile hygiene | `uv lock --check` | committed `uv.lock` matches `pyproject.toml`; no rewrite |
| reproducible environment | `uv sync --frozen` | exact locked environment installs without resolution |
| format | `uv run ruff format --check .` | no diff |
| lint and security rules | `uv run ruff check .` | curated Ruff policy exits zero |
| architecture boundaries | `uv run lint-imports` | declared Import Linter contracts are kept |
| type safety | `uv run mypy .` | strict configuration exits zero |
| functional confidence | `uv run pytest` | unit and configured integration suites pass |
| supply-chain check | `uv run --with pip-audit pip-audit` | no blocking vulnerability findings |
| file-specific correctness | targeted command from the relevant recipe or service doc | expected assertions pass |

The committed [templates/Makefile](templates/Makefile) wraps the ordered gate — lock-check, frozen sync, format-check, lint, imports, types, test, audit — as `make verify`; local and CI run the same target. Each stage remains individually runnable. Use containerized integration, load, compatibility, or package-install checks when the task calls for them.

## Slow Path Docs

- Architecture and package map: [maintainer-reference.md](maintainer-reference.md)
- Startup and package layout: [foundations/project-setup.md](foundations/project-setup.md), [foundations/package-design.md](foundations/package-design.md)
- Runtime correctness: [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md), [foundations/errors-and-logging.md](foundations/errors-and-logging.md), [foundations/configuration.md](foundations/configuration.md)
- Type and boundary correctness: [foundations/typing-discipline.md](foundations/typing-discipline.md), [quality/linting.md](quality/linting.md)
- Proof and verification: [quality/testing.md](quality/testing.md)
