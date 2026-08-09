# Framework Selection

Rules for deciding when a dependency earns its complexity cost.

## Default Approach

Start with Python 3.11's standard library and the locked stack. Add third-party packages only when they clearly improve correctness, interoperability, or operator experience. Prefer dependencies confined to adapters over frameworks that become the application architecture.

The project uses PEP 621 metadata in `pyproject.toml`; the [Python Packaging User Guide](https://packaging.python.org/en/latest/guides/writing-pyproject-toml/) is authoritative for project scripts and metadata. uv owns environments and locking: [its lock documentation](https://docs.astral.sh/uv/concepts/projects/sync/) defines `uv lock --check` as the drift check and `uv sync --frozen` as installation without re-locking.

### Approval Questions

Before adding a dependency, answer all of these:

1. What concrete problem does the stdlib or locked stack fail to solve well enough?
2. What maintenance, upgrade, typing, licensing, supply-chain, and security cost does this add?
3. Does it introduce runtime discovery, hidden global state, implicit I/O, metaprogramming, event-loop constraints, or framework lock-in?
4. Is it actively maintained, license-compatible, typed enough for strict mypy, auditable, and replaceable behind an existing boundary?

## Default Choices By Concern

| Concern | Default | Acceptable escalation | Avoid by default |
|---|---|---|---|
| runtime | compatibility floor `requires-python = ">=3.11"`; a current stable development/CI interpreter pinned in `.python-version` | raise the floor deliberately through ADR and compatibility proof | coding to the newest local interpreter while metadata still promises 3.11 |
| project metadata and layout | PEP 621 `pyproject.toml`; one distribution and one import package under `src/<app>/`; `tests/` outside `src/` | multiple distributions/workspaces via ADR when independently released packages require them | `setup.py` as primary metadata; flat layouts whose tests import from the checkout accidentally |
| environment and lock workflow | uv; applications commit `uv.lock`; local and CI run the same Make targets | pip-tools for an established organizational workflow that preserves a reviewed lock and identical gate | Poetry for new repos; mixing uv, pip, and Poetry state; ad hoc `pip install` in CI |
| build backend | `uv_build` for pure-Python packages; it validates the single-module structure and integrates with uv ([official guidance](https://docs.astral.sh/uv/concepts/build-backend/)) | Hatchling for build hooks, unusual layouts, or features `uv_build` cannot express; specialized backend for native extensions | legacy implicit setuptools behavior; backend choice hidden from `pyproject.toml` |
| console entry points | `[project.scripts]` calling a thin synchronous function; `__main__.py` delegates | a dedicated runner wrapper for an asyncio process | executable source files as the only interface; business logic in `__main__.py` |
| ASGI framework | FastAPI with thin routers and dependency wiring; Pydantic DTOs mapped into core ([dependencies](https://fastapi.tiangolo.com/tutorial/dependencies/)) | Starlette for a deliberately lower-level ASGI surface; Litestar when mandated or justified by a measured requirement | Django for service sidecars; framework models, requests, or dependency containers in core |
| ASGI server | uvicorn, configured by the process/container entry point | organization-mandated ASGI host through ADR and shutdown proof | development reload in production; invoking server globals from domain code |
| lifespan / resources | FastAPI lifespan async context manager owns app-scoped clients, engines, and exporters ([official lifespan guidance](https://fastapi.tiangolo.com/advanced/events/)) | a separate composition/lifecycle object for workers or multi-host processes | per-request engines/clients; startup-event and lifespan APIs mixed in one app |
| validation | Pydantic v2 at trust boundaries only; explicit DTO-to-domain mapping | stdlib parsing for a small library/CLI boundary | Pydantic models as domain entities; passing raw dictionaries through layers |
| configuration | pydantic-settings, constructed once and validated before startup; env/mounted secrets supplied externally ([official settings docs](https://docs.pydantic.dev/latest/concepts/pydantic_settings/)) | explicit TOML config for an operator workflow via ADR | scattered `os.getenv`; lazy settings reads; secrets committed in dotenv or TOML |
| typing | mypy `strict = true` plus `warn_unreachable` and `disallow_any_unimported`; `Protocol`, `NewType`, `Self`; libraries ship `py.typed` | Pyright as an editor-side second checker | Pyright replacing the gate; broad `Any`, unscoped ignores, stub packages accepted without review |
| boundary enforcement | [Import Linter](https://import-linter.readthedocs.io/en/stable/) forbidden/layer contracts, run by `uv run lint-imports`; Ruff `TID` for import hygiene | a repo-specific architecture test only when Import Linter cannot express a real boundary | relying on folder names and review comments alone; circular adapter imports |
| lint and format | Ruff for both; curated stable families `E`, `W`, `F`, `I`, `UP`, `B`, `SIM`, `C4`, `DTZ`, `T20`, `PT`, `RET`, `RUF`, `ASYNC`, `S`, `PL`, `TRY`, `N`, `A`, `ISC`, `PIE`, `PERF`, `FURB`, `TID` ([rule catalog](https://docs.astral.sh/ruff/rules/)) | narrow per-file ignores with written reason; preview rules only after deliberate review | blanket `noqa`; formatter overlap; selecting all PL/TRY/S rules without curating conflicts and test exceptions |
| outbound HTTP | one lifespan-owned `httpx.AsyncClient` per app with explicit `httpx.Timeout`; bounded pool and response handling ([async client](https://www.python-httpx.org/async/), [timeouts](https://www.python-httpx.org/advanced/timeouts/)) | synchronous `httpx.Client` for a synchronous CLI/library | `requests` in async services; one client per call; disabled or implicit deadlines |
| retries / backoff | small hand-rolled bounded retry loop with exponential backoff, full jitter, explicit retryable outcomes, injected sleep/random seams | Tenacity via ADR when policy composition genuinely outgrows the readable loop | unbounded retries, zero jitter, retrying non-idempotent effects, hidden decorator policy |
| circuit breaker | none; deadlines, concurrency limits, bounded retries, and load shedding first | a breaker via ADR when measured dependency failure demands it and state transitions are observable | a breaker on every client; process-local state treated as globally authoritative |
| persistence | SQLAlchemy 2.0 async engine/session with asyncpg; typed `Mapped[]` ORM models or SQLAlchemy Core tables behind repositories ([asyncio docs](https://docs.sqlalchemy.org/en/20/orm/extensions/asyncio.html)) | SQLAlchemy Core for SQL-first modules; datastore mandated by system constraints | Django ORM in sidecars; ORM entities crossing into core; raw string-interpolated SQL |
| schema migrations | Alembic revisions, explicit deploy/init step, expand/contract for destructive changes | separate migration artifact or job required by the platform | auto-migrate on normal startup; hand-applied SQL without version history |
| transactions | repository/unit-of-work boundary owns one explicit async session scope | SQLAlchemy connection/Core transaction for SQL-heavy paths | ambient sessions; commits hidden inside unrelated helpers; network calls inside open DB transactions |
| caching | no cache until measured; then bounded TTL data with explicit keys, invalidation, and stampede control | redis-py/Valkey for cross-instance sharing or working sets that do not fit a process, via ADR | unbounded dictionaries; caching ORM/Pydantic objects; correctness depending on TTL alone |
| messaging | broker-specific async client behind core-owned publisher/consumer `Protocol`s; outbox/inbox where DB state and events meet | framework or managed bus via ADR after ordering, settlement, retry, and DLQ contracts are written | broker types in core; frameworks hiding acknowledge/offset behavior; publishing inside a DB transaction without outbox |
| job scheduling | platform scheduler (for example, a Kubernetes CronJob) for independent calendar work; owned asyncio loop for fixed intervals in one process | APScheduler or organization scheduler via ADR for real calendar/stateful scheduling | unowned `create_task`; overlapping `sleep` loops; assuming one replica executes a singleton job |
| CLI | stdlib `argparse` | Typer or Click via ADR for substantial nested commands, completion, and shared CLI UX | framework dependency for a few flags; parsing scattered through business logic |
| logging | stdlib `logging`, configured once with `logging.config.dictConfig`; JSON formatter for services; library `NullHandler` only | structlog via ADR when required context binding/processors materially improve the system | `basicConfig` in libraries; multiple logging stacks; secrets or payload dumps; duplicate exception logging |
| metrics | prometheus-client, explicit low-cardinality names/labels, `/metrics` | org-mandated exporter/collector behind a narrow telemetry module | user/request IDs or raw paths as labels; metrics emitted from core through vendor APIs |
| tracing | OpenTelemetry API/SDK and instrumentations configured in composition | org exporter or collector configuration behind OpenTelemetry | vendor tracing SDK in core; spans containing secrets or unrestricted payloads |
| health | `/livez` cheap and local; `/readyz` bounded and dependency-aware; `/metrics` separate | cached readiness aggregation for expensive dependencies | liveness calling dependencies; unbounded fan-out per probe |
| server-rendered templates | Jinja2 with autoescape, templates loaded once, explicit view models | a frontend architecture through ADR when interaction genuinely outgrows server rendering | string-built HTML; marking untrusted content safe; SPA by reflex for form-over-data |
| sessions | opaque high-entropy session ID in Secure, HttpOnly, SameSite cookie; server-side state in an existing store with rotation/expiry ([OWASP session guidance](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)) | framework/session package via ADR after cookie, fixation, revocation, and storage semantics are reviewed | JWT cookie as a session substitute; sensitive state in client-readable signed cookies; hand-rolled cryptography |
| CSRF | synchronizer token tied to the server-side session on every state-changing browser request; Origin/Sec-Fetch validation as defense in depth ([OWASP CSRF guidance](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html)) | vetted middleware integrated at the HTTP adapter | state changes over GET; disabling CSRF globally; SameSite alone treated as complete protection |
| gRPC | `grpcio` runtime with `grpcio-tools` generating Python and type-stub outputs from `.proto` ([official quickstart](https://grpc.io/docs/languages/python/quickstart/)) | Buf generation with pinned Python/protobuf and Python/gRPC plugins when the organization already operates Buf ([Buf guidance](https://buf.build/docs/bsr/remote-plugins/usage/)) | hand-written stubs; generated code as the only schema; unreviewed regeneration diffs |
| testing | pytest; pytest-asyncio with `asyncio_mode = "strict"` and explicit async markers; real PostgreSQL/broker tests; coverage.py/pytest-cov reports | Hypothesis via ADR for parsers, state machines, or algorithms where generated inputs add material proof | `unittest.TestCase` style in new suites; sleep-based synchronization; only mocked boundary tests |
| test doubles | hand-rolled fakes at consumer-owned `Protocol` seams | `unittest.mock` when a focused fake is disproportionately expensive | patching deep implementation paths; autospec-free dynamic mocks; call-order tests that pin internals |
| time | aware UTC `datetime` at boundaries, monotonic clocks for elapsed time, injected clock/sleep for logic | IANA zone handling when the domain requires civil-time rules | naive datetimes crossing boundaries; `datetime.now()` embedded in business decisions |
| serialization | Pydantic v2 JSON at HTTP/message boundaries; protobuf for gRPC; explicit unknown-field and compatibility policy | stdlib `json` for small internal library surfaces | `pickle` across trust boundaries; multiple serializers for the same contract; implicit ORM serialization |
| XML / YAML | `defusedxml` for untrusted XML; `yaml.safe_load` when YAML is required ([OWASP XML guidance](https://cheatsheetseries.owasp.org/cheatsheets/XML_External_Entity_Prevention_Cheat_Sheet.html)) | no parser when the format is unnecessary | stdlib XML parsers on hostile data; `yaml.load` on untrusted input |
| subprocesses | `subprocess.run` with argv list, timeout, checked return; `asyncio.create_subprocess_exec` in async paths | explicitly reviewed shell invocation for a fixed trusted script only | `shell=True` with dynamic data; ignored exit codes; unbounded output |
| secrets manager | env vars or mounted files injected by the platform; settings stores secret values in non-revealing types where practical | org-provided Vault/cloud manager behind `clients/` when runtime fetch is required | secrets in source, `uv.lock`, images, build args, logs, errors, repr, or telemetry |
| audit sink | dedicated structured audit events routed by platform logging to access-controlled retention | SIEM or append-only service behind an adapter when compliance requires tamper evidence | audit mixed indistinguishably with application logs; mutable local file as authoritative audit record |
| supply chain | reviewed `uv.lock`; `uv lock --check`; pip-audit through the verify gate; automated dependency PRs | organization SBOM/signing/provenance controls layered on CI | floating application installs; ignored transitive diff; vulnerability audit disconnected from the lock |
| build entry point | `make verify`: lock-check, frozen sync, format-check, lint, imports, types, test, audit | additional bounded stages for schema generation, packages, containers, or integration suites | CI commands diverging from Make targets; a gate that rewrites files |

## Mandated Frameworks

Sometimes the requester or platform mandates Django, Litestar, Poetry, structlog, Celery, a broker framework, or another stack this table would not choose. Honor the mandate without silently broadening it:

- Record an ADR stating that the requester mandated it, the concrete requirement, allowed package areas, default displaced, lock/runtime consequences, and re-evaluation trigger. The Approval Questions still get written answers; “mandated” answers the first.
- Preserve framework-independent invariants: Python 3.11 floor, `src/`+`tests/`, PEP 621 metadata, core dependency direction, typed boundaries, owned tasks and cancellation, boundary validation, log-once behavior, real integration proof, and full `make verify`.
- Confine framework types to the relevant adapter. HTTP framework types stay in `api/http`; ORM/broker/client types stay in their adapters. Core consumes plain values and core-owned `Protocol`s.
- State which handbook procedures change. Mandated Django replaces FastAPI router/lifespan mechanics and must name equivalent startup, shutdown, validation, migration, and testing proof. Mandated Poetry replaces uv workflow commands only where the mandate explicitly requires it; it does not relax locking, PEP 621 interoperability, or audit proof.
- A mandate covers only the named framework. It is not permission to weaken unrelated defaults.

## Common Mistakes And Forbidden Patterns

- No dependency added because it is familiar, fashionable, or avoids a small amount of explicit code.
- No application with an uncommitted or knowingly stale `uv.lock`; no CI install that silently resolves a different graph.
- No FastAPI, Pydantic DTO, SQLAlchemy model, HTTPX response, protobuf message, broker record, or telemetry SDK type in core.
- No `requests`, blocking file/process calls, or synchronous SQLAlchemy engine on an async service path.
- No unowned task, swallowed `CancelledError`, unbounded gather/fan-out, infinite retry, or external operation without a deadline.
- No auto-migration on normal startup, dynamic SQL interpolation, `shell=True` with external data, unsafe YAML/XML loading, or `pickle` at a trust boundary.
- No structlog, Tenacity, Typer/Click, Hypothesis, distributed cache, scheduler, messaging framework, or alternate ASGI framework without its stated escalation proof and ADR.
- No blanket Ruff/mypy suppression, broad `Any` escape hatch, or package lacking typing review.
- No bare error strings, secret-bearing exceptions/reprs, high-cardinality metric labels, or duplicate logs at multiple layers.
- No exception to this table without an ADR.

## Verification And Proof

A dependency choice is proven before merge:

- Approval Questions are answered in the PR or ADR.
- `pyproject.toml` and `uv.lock` diffs account for every direct and transitive addition; license, typing, maintenance, runtime, and build impact are understood.
- `uv lock --check`, `uv sync --frozen`, `uv run ruff check .`, `uv run lint-imports`, `uv run mypy .`, tests, and `uv run --with pip-audit pip-audit` pass through `make verify`.
- Generated code, wheel/sdist contents, migration behavior, startup/shutdown, or real external boundary proof is added when the dependency affects those surfaces.
- Any departure from the table has an ADR cross-linking [architecture-decision-records.md](architecture-decision-records.md).

### Decision Record

An exception ADR records the package and why the default failed, permitted package areas, operational/security/license/typing tradeoffs, migration and rollback, proof, and removal trigger.
