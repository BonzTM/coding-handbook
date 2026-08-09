# Maintainer Reference

Purpose: hold slower-path architecture, package-map, lifecycle, and rationale guidance that is useful but not worth loading for every task.
Audience: maintainers and agents working in Python repositories that use this handbook.
Read [AGENTS.md](AGENTS.md) first. Use this file when you need the fuller background behind the fast-path rules.

## Architecture Snapshot

This handbook assumes one PEP 621 distribution package with explicit inward dependencies and a minimal public surface. The dominant shape is:

```text
repo/
  pyproject.toml
  uv.lock
  .python-version
  Makefile
  api/
  src/
    orders/
      __init__.py
      __main__.py
      main.py
      config.py
      core/
      api/
        http/
        grpc/
      db/
        migrations/
      clients/
      telemetry/
      workers/
  tests/
    unit/
    integration/
```

The [Python Packaging User Guide](https://packaging.python.org/en/latest/discussions/src-layout-vs-flat-layout/) explains why `src/` prevents the checkout root from shadowing the installed package. Folder conventions alone do not enforce architecture, so the repository also commits [Import Linter](https://import-linter.readthedocs.io/en/stable/) contracts: `core` may not import `api`, `db`, `clients`, or `telemetry`. [Ruff TID rules](https://docs.astral.sh/ruff/rules/#flake8-tidy-imports-tid) govern import hygiene; `uv run lint-imports` governs dependency direction.

A complete `make verify`-green instance of this architecture lives at [reference/exampleservice/](reference/exampleservice/). Read it alongside this map to see the boundaries embodied in code.

## Two-Speed Documentation Model

- Fast path: [AGENTS.md](AGENTS.md) for invariants, the task loop, change-type-to-file-set routing, and baseline proof.
- Slow path: this file for architecture, package map, test taxonomy, lifecycle, and rationale.

Use the fast path for most tasks. Use this file when a change crosses layers, introduces new runtime behavior, or challenges an existing default.

## Package Map

| Package Area | Owns | Must Not Own |
|---|---|---|
| `src/<app>/__main__.py` | delegation to the console entry callable | composition details, business rules, resource construction |
| `src/<app>/main.py` | composition, validated config, logging setup, lifespan/resource wiring, root process lifecycle | domain rules, SQL, request validation details |
| `src/<app>/core/` | domain types, use cases, orchestration, consumer-defined `Protocol` ports | FastAPI/Pydantic DTOs, SQLAlchemy, HTTPX, telemetry implementation |
| `src/<app>/api/http/` | FastAPI routers, dependencies, Pydantic HTTP DTOs, status/error mapping | business rules, engine/session construction, direct external calls |
| `src/<app>/api/grpc/` | generated-stub adaptation, proto-to-core mapping, status codes, interceptors | SQL or business state mutation outside core |
| `src/<app>/db/` | async engine/session factory, SQLAlchemy mappings or SQLAlchemy Core tables, repositories, transaction policy, Alembic migrations | HTTP/gRPC types; leaking ORM entities into core |
| `src/<app>/clients/` | lifespan-shared HTTPX/gRPC clients, wire DTOs, timeout/retry mechanics, external error mapping | domain decisions; creating a client per call |
| `src/<app>/config.py` or `config/` | pydantic-settings sources, typed settings, defaults, startup validation | lazy env reads spread through business logic |
| `src/<app>/telemetry/` | `dictConfig`, JSON formatter, Prometheus instruments, OpenTelemetry, health primitives | business decisions about request meaning |
| `src/<app>/workers/` | consumer loops, broker adapters, settlement, retry/DLQ mechanics, scheduled-loop ownership | domain rules that belong in core; unowned tasks |
| `tests/unit/` | fast deterministic tests and hand-rolled fakes | Docker, real network, or real database dependencies |
| `tests/integration/` | real PostgreSQL, broker, migration, and external-protocol proof | unit tests that belong in the inner loop |

Libraries keep the same `src/` layout but omit unused service adapters. They declare their typed marker (`py.typed`), never configure application logging, and treat exported names as compatibility contracts.

## Dependency Direction

The dependency arrows point inward:

```text
api/http ─┐
api/grpc ─┼──> core <── db
clients ──┤       <── workers
main ─────┘
```

`telemetry` and `config` are composed at the process boundary. Core may define narrow telemetry-neutral result types or ports but never import concrete logging, metric, tracing, settings, transport, client, or persistence packages. Python's structural typing lets the consumer own the seam: [typing.Protocol](https://docs.python.org/3.11/library/typing.html#typing.Protocol) is preferred over adapter-owned abstract base classes.

## Lifecycle Model

For services and workers, the normal process lifecycle is:

1. The console entry callable starts one root asyncio runner; `__main__.py` only delegates.
2. Pydantic Settings reads injected environment or mounted secret files and validates every required value before listeners or loops start.
3. Composition configures stdlib logging once, then constructs metrics/tracing, the async SQLAlchemy engine/session factory, and one shared `httpx.AsyncClient` with explicit timeouts.
4. Core use cases are wired to adapter implementations. FastAPI `Depends` exposes those ports without becoming a service locator.
5. The FastAPI [lifespan context manager](https://fastapi.tiangolo.com/advanced/events/) owns startup and deterministic resource closure; workers run under a root `TaskGroup`.
6. SIGTERM or Ctrl+C cancels the root task. The process stops accepting work, drains within explicit deadlines, settles only durable work, and closes clients, engines, exporters, and executors.

Python 3.11 provides `TaskGroup` and `asyncio.timeout()` for structured concurrency and deadlines. Its documentation requires caught `CancelledError` to be propagated after cleanup because swallowing it can break those primitives: [Coroutines and Tasks](https://docs.python.org/3.11/library/asyncio-task.html).

If a repository shape does not fit this lifecycle, document the exception explicitly. Full rules live in [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md) and [foundations/shared-constructs.md](foundations/shared-constructs.md).

## Test Taxonomy

| Test Type | Default Location | What It Proves |
|---|---|---|
| unit tests | `tests/unit/`, mirroring `src/<app>/` | core rules and edge cases, offline and deterministic |
| transport tests | `tests/unit/api/` with FastAPI's test client or HTTPX ASGI transport and fake core ports | routing, request validation, mapping, status/error shape, middleware |
| repository integration tests | `tests/integration/db/` with real PostgreSQL | SQL, transactions, SQLAlchemy mappings, and committed Alembic migrations |
| external client tests | unit tests with a bounded stub server or transport fake | request construction, deadlines, retry bounds, response/error mapping |
| worker integration tests | `tests/integration/workers/` with the real broker path when practical | settlement, replay, idempotency, retry, and DLQ semantics |
| contract tests | golden payloads plus Pydantic/protobuf round trips | compatibility and unknown-field policy |
| property tests | parser/algorithm tests, ADR-gated | broad generated-input invariants where examples are insufficient |
| benchmarks | dedicated measured hot-path suite, ADR-gated | allocation, latency, or throughput under repeatable conditions |

The principle is not "more tests". It is "the right tests at the right boundary". A fake repository does not replace a real migration or transaction test. Pytest-asyncio is configured in strict mode and async tests carry explicit markers so ownership is visible; see its [configuration reference](https://pytest-asyncio.readthedocs.io/en/stable/reference/configuration.html).

## Runtime Contracts Worth Remembering

- Every asyncio task has an owner, stop condition, failure-observation path, and shutdown proof.
- Every external operation has a deadline; retry loops, pagination, streams, queues, and concurrency are bounded.
- Blocking I/O and CPU-heavy work never run on the event loop; use `asyncio.to_thread()` or an explicitly owned executor.
- Every network-facing component has readiness (`/readyz`) distinct from cheap liveness (`/livez`).
- Every non-trivial feature adds telemetry where operators will act on it; metric labels remain finite.
- Every dependency becomes part of the patch, typing, build, and vulnerability surface represented by `uv.lock`.

## Contract Surfaces

- HTTP and gRPC boundaries have an obvious source of truth for payloads, validation, status mapping, and errors.
- Pydantic models are boundary DTOs, not domain entities. Explicit mapping prevents transport refactors from rewriting core behavior.
- Database schema, migration order, transaction behavior, and compatibility windows are data contracts.
- Event payloads have explicit envelopes, versioning, idempotency, settlement, and redelivery rules.
- Generated code is never the only source: `.proto`, OpenAPI, schema, or migration history remains authoritative.
- A library's public names, signatures, exceptions, typing metadata, wheel, and sdist are compatibility surfaces.

## Dependency Rationale

- Python 3.11 is the floor because its stdlib provides the structured concurrency and deadline primitives this lifecycle standardizes on.
- uv owns the environment and lock workflow; its official docs define `uv lock --check` as the no-write drift test and `uv sync --frozen` as locked installation without re-locking: [Locking and syncing](https://docs.astral.sh/uv/concepts/projects/sync/).
- `uv_build` is the pure-Python default because it integrates with uv, validates metadata and structure, and supports the single-module `src/` shape; Hatchling is the escalation for build hooks or layouts it cannot express: [uv build backend](https://docs.astral.sh/uv/concepts/build-backend/).
- SQLAlchemy keeps SQL and transaction behavior explicit while providing an async engine/session boundary; Alembic preserves ordered migration history.
- FastAPI, Pydantic, HTTPX, Prometheus, and OpenTelemetry stay at adapters or composition so core behavior remains replaceable and directly testable.
- Import Linter earns one development dependency because it turns the highest-value architectural rule into a repeatable gate; Ruff's import rules complement it but do not encode layer-specific direction.

## Common Failure Modes

| Symptom | Likely Cause | First Fix |
|---|---|---|
| routers know too much about storage | use cases leaked out of `core` | move orchestration behind a core-owned `Protocol` |
| `main.py` grows with every feature | composition and domain behavior are mixed | extract focused constructors and adapter registration helpers |
| tests pass but deploys fail | no real PostgreSQL, migration, or startup proof | add real-boundary integration and container smoke tests |
| workers hang or lose work on shutdown | tasks have no owner, cancellation is swallowed, or settlement is early | root `TaskGroup`, re-raised cancellation, bounded drain, settle last |
| requests stall under load | blocking calls or unbounded fan-out on the event loop | Ruff ASYNC cleanup, `to_thread`, semaphores, explicit limits |
| HTTP connection count churns | `AsyncClient` created per request | create one lifespan-owned client and close it once |
| metrics explode | request/user identifiers used as labels | reduce labels to stable finite dimensions |
| config fails only in production paths | lazy env reads or unvalidated defaults | construct one Settings object and fail startup |
| mypy passes while boundaries rot | broad `Any`, ignored imports, or missing architecture contracts | remove escape hatches; run strict mypy and `lint-imports` |

## Primary Sources Behind These Defaults

- Python 3.11 asyncio tasks and cancellation: `https://docs.python.org/3.11/library/asyncio-task.html`
- Python typing protocols and `NewType`: `https://docs.python.org/3.11/library/typing.html`
- packaging metadata and scripts: `https://packaging.python.org/en/latest/guides/writing-pyproject-toml/`
- `src/` layout: `https://packaging.python.org/en/latest/discussions/src-layout-vs-flat-layout/`
- uv projects, locking, and build backend: `https://docs.astral.sh/uv/concepts/projects/`, `https://docs.astral.sh/uv/concepts/projects/sync/`, `https://docs.astral.sh/uv/concepts/build-backend/`
- FastAPI lifespan and dependencies: `https://fastapi.tiangolo.com/advanced/events/`, `https://fastapi.tiangolo.com/tutorial/dependencies/`
- SQLAlchemy asyncio: `https://docs.sqlalchemy.org/en/20/orm/extensions/asyncio.html`
- HTTPX async clients and timeouts: `https://www.python-httpx.org/async/`, `https://www.python-httpx.org/advanced/timeouts/`
- Ruff rules and mypy strict flags: `https://docs.astral.sh/ruff/rules/`, `https://mypy.readthedocs.io/en/stable/command_line.html#cmdoption-mypy-strict`

## Related Docs

- Fast path and change routing: [AGENTS.md](AGENTS.md)
- Project layout: [foundations/project-setup.md](foundations/project-setup.md)
- Package boundaries: [foundations/package-design.md](foundations/package-design.md)
- Async lifecycle: [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md)
- Contracts and compatibility: [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md)
- Event delivery: [services/eventing-and-messaging.md](services/eventing-and-messaging.md)
- Proof and testing: [quality/testing.md](quality/testing.md)
