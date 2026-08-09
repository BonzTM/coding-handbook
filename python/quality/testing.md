# Testing

Testing strategy for Python repos that need trustworthy behavior, not just green checkmarks.

## Default Approach

Use pytest, pytest-asyncio, pytest-cov, and the smallest real boundary that proves the behavior. Keep tests under `tests/`, mirroring `src/<app>/` by responsibility.

### Test Taxonomy

| Test Type | Use for | Default proof |
|---|---|---|
| unit | domain decisions, value objects, parsers, retry policy, small adapters | plain pytest plus hand-rolled fakes at core-owned `Protocol` seams |
| integration | SQLAlchemy repositories, Alembic migrations, outbound clients, broker settlement | real PostgreSQL, protocol test server, or real containerized dependency |
| contract | exact HTTP/message/protobuf/public-library shapes and compatibility | golden examples, generated schema/stub diffs, provider/consumer fixtures |
| smoke | assembled process, probes, startup/shutdown, one critical path | built artifact with real local dependencies and bounded process control |

Test the narrowest deterministic boundary first. A few smoke tests prove wiring; they never replace unit, integration, or contract proof.

### Organization And Naming

Mirror production ownership:

```text
src/<app>/core/orders.py        -> tests/core/test_orders.py
src/<app>/api/http/orders.py    -> tests/api/http/test_orders.py
src/<app>/db/orders.py          -> tests/db/test_orders.py
```

Name files `test_<subject>.py` and functions `test_<behavior>_<condition>`. Test names state observable behavior, not implementation method names. Keep one primary act per test and use plain `assert` so pytest renders the failing values.

Use test classes only to group related cases; do not use inheritance or xUnit setup methods in new suites. Fixtures own setup and deterministic teardown. A `yield` fixture closes resources in `finally` semantics supplied by pytest; use the narrowest scope that matches the resource lifetime.

### Fixtures And conftest.py

Prefer explicit fixtures over setup methods and autouse state. Function scope is the default. Module/session scope is reserved for expensive immutable infrastructure such as one Postgres container, while each test still gets isolated schema/data state.

Place a fixture in `conftest.py` only when every test beneath that directory has a legitimate use for it. Feature-specific fixtures stay beside the tests that consume them. Root `conftest.py` owns only suite-wide infrastructure; it is not a hidden service locator. The [pytest fixture guide](https://docs.pytest.org/en/stable/explanation/fixtures.html) defines fixture scope and teardown behavior.

Builders return valid domain values and expose focused overrides. Fixtures contain no assertions about the behavior under test and do not hide network calls.

### Parametrization

Use `@pytest.mark.parametrize` when the behavior and setup are the same shape across inputs. Give non-obvious cases explicit `id=` values. Split cases when setup branches, expected failures differ materially, or a table obscures the contract. Pytest's [parametrization guide](https://docs.pytest.org/en/stable/how-to/parametrize.html) is the supported mechanism; do not build loops that stop at the first failure.

### Asyncio Configuration

Set `asyncio_mode = "strict"` in `[tool.pytest.ini_options]`. Every async test uses `@pytest.mark.asyncio`; every async fixture uses `@pytest_asyncio.fixture`. In strict mode pytest-asyncio handles only those explicit tests and fixtures, preventing accidental plugin ownership; this is the behavior documented in its [concepts guide](https://pytest-asyncio.readthedocs.io/en/stable/concepts.html#test-discovery-modes).

Use function-scoped event loops by default. A wider `loop_scope` requires a resource whose lifecycle genuinely spans tests and consistent neighboring scope. Never create loop-bound clients, engines, or tasks at import time.

### Deterministic Async Tests

- Coordinate with `asyncio.Event`, bounded queues, barriers, or observable fake calls; never sleep to “let the task run.”
- Use the injected `Clock` and sleep seam from [time](../foundations/time.md) for deadlines, expiry, retry, and schedules.
- Own every task in the test. Cancel and await it during teardown, then assert its terminal result.
- Force sibling failure, cancellation, timeout, queue saturation, and shutdown ordering explicitly.
- Sort unordered results and fix random inputs, timestamps, IDs, and jitter sources.
- Run high-risk suites with `PYTHONASYNCIODEBUG=1` and warnings as errors; pending tasks and unclosed transports fail the test.

Do not assert one exact interleaving unless ordering is the contract. Assert invariants before and after controlled synchronization points.

### Test Doubles

The default double is a hand-rolled fake implementing a consumer-defined `Protocol`:

```python
class FakeWidgetStore:
    def __init__(self) -> None:
        self.items: dict[WidgetId, Widget] = {}

    async def get(self, widget_id: WidgetId) -> Widget | None:
        return self.items.get(widget_id)
```

Fakes retain behavior and state needed for assertions; they do not reproduce the real dependency. Keep them in `tests/testutil/` only after multiple tests share the same coherent seam.

Use `unittest.mock` only when a focused fake is disproportionately expensive: a context manager callback, a one-off failure injection, or a hard-to-construct library object. Patch the name looked up by the consumer, not the library's defining module. Use `autospec=True` or `create_autospec()` so nonexistent attributes and wrong signatures fail. Prefer outcome assertions; call assertions are appropriate only when the call itself is the contract, such as audit emission or settlement.

No global monkeypatch of time, environment, clients, or database sessions when constructor/dependency injection can expose the seam honestly.

### FastAPI And HTTPX Tests

Use `httpx.AsyncClient` with `httpx.ASGITransport(app=app)` for async HTTP adapter tests. HTTPX documents this transport as the direct ASGI test path in its [transport guide](https://www.python-httpx.org/advanced/transports/). Override FastAPI dependencies with fakes to isolate the router, then assert status, headers, exact body, and the DTO-to-domain mapping.

`ASGITransport` does not trigger ASGI lifespan. For tests that must prove lifespan-owned resources, use `TestClient` as a context manager; FastAPI documents that the context runs startup and shutdown in [testing events](https://fastapi.tiangolo.com/advanced/testing-events/). Use `TestClient` for synchronous transport tests and lifespan proof. Use `AsyncClient` when the test must await application dependencies or coordinate async work on the same loop.

Do not mix `TestClient` resources with loop-bound objects created on the pytest loop. Build the app through `create_app()` for every test or fixture and close the client before asserting teardown.

### Outbound HTTP Tests

Default to a hand-written `httpx.AsyncBaseTransport`/`MockTransport` fake when a client needs a small fixed response matrix. It keeps the seam at HTTPX transport, avoids global patching, and adds no dependency. Use a local ASGI app through `ASGITransport` when request/response protocol behavior matters.

RESPX is an ADR-level escalation for a large HTTPX surface where route matching, repeated calls, or complex request inspection materially reduces test code. A real local stub server remains necessary when DNS, TLS, connection pooling, streaming, or wire behavior is the subject.

Test success, malformed payloads, non-success statuses, timeout, cancellation, connection failure, bounded retries, response closure, and redaction. Do not mock the domain-specific client in every test and leave its actual parsing unproved.

### Real PostgreSQL Integration

Repository and migration proof uses PostgreSQL, not SQLite. SQLite does not prove asyncpg behavior, PostgreSQL types, locking, constraints, isolation, or query plans.

Use testcontainers-python to start a disposable PostgreSQL instance for the suite; the official [PostgreSQL module guide](https://testcontainers.com/modules/postgresql/) documents the Python container. Pin the image in project configuration, start it in a session-scoped fixture, run Alembic from an empty database, and dispose the engine before stopping the container. Isolate tests with a fresh database/schema or a transaction strategy whose rollback semantics are understood.

Docker Compose is the fallback when CI or developer infrastructure already owns a named test stack. The test still receives its DSN explicitly, waits with a bounded health check, and cleans its data. Never silently point integration tests at a shared or production-like database.

### Marker Policy

Register markers and enable `strict_markers = true`:

```toml
[tool.pytest.ini_options]
asyncio_mode = "strict"
strict_markers = true
markers = [
  "integration: requires a real containerized dependency",
  "contract: validates a published boundary",
  "smoke: exercises the assembled artifact",
]
```

All real external-boundary tests carry `@pytest.mark.integration`; contract and smoke may be combined with integration when both properties apply. The ordinary full gate runs the suite configured by the project. Developers without Docker may run:

```bash
uv run pytest -m "not integration"
```

A separate CI integration lane must run `uv run pytest -m integration` with Docker available. A skip is allowed only when the dependency capability is absent and reports the reason; CI's integration lane treats unexpected skips as failure. Pytest documents registration and strict-marker validation in its [marker guide](https://docs.pytest.org/en/stable/how-to/mark.html).

### Contract And Golden Tests

Golden tests pin observable wire shapes: JSON keys and aliases, null versus omission, timestamps, enums, problem responses, event envelopes, protobuf fixtures, and generated OpenAPI. Keep small expected objects inline; use focused files under `tests/testdata/` for larger stable payloads.

Regenerate golden files only through an explicit command, then review the diff as a contract change. Normalize only fields declared nondeterministic by the contract. Never delete a field from expected output merely to make a refactor pass.

Provider tests prove produced shapes. Consumer tests validate only fields they use, tolerate additive unknown fields where policy requires it, and exercise unknown enum/version behavior. See [contracts and compatibility](../foundations/contracts-and-compatibility.md).

### Coverage Policy

Use pytest-cov over coverage.py with branch coverage enabled. The template records a meaningful initial floor based on the exemplar, then the repo ratchets it without lowering the threshold to merge a change.

```bash
uv run pytest --cov=src/<app> --cov-branch --cov-report=term-missing
```

Coverage is a map, not the deliverable. Core decisions, parsing, authorization, error/status mapping, retries, cancellation, and security-sensitive negative paths require coverage. Generated code, trivial entrypoint wiring, and framework glue do not earn contrived tests. A high percentage with untested failure branches is coverage theater.

Review missing lines and branches with every behavior change. Combine unit and integration results when reporting the product's real exercised surface.

### Property-Based Tests

Hypothesis is an ADR-level escalation for parsers, normalization, serialization round trips, ordering laws, and state machines whose input space defeats a readable example table. State the invariant, bound generation and execution, and retain every discovered failure as a focused regression example. Do not introduce it for business logic that a short parametrized table proves completely.

### Flake Discipline

A flaky test is a failing test. Quarantine only to restore signal during an active incident: link an owner and issue, set a removal deadline, and keep the failure visible. Do not add reruns until green, widen timing windows, or mark intermittent behavior `xfail`.

Reproduce with fixed seeds, asyncio debug, repeated targeted runs, and constrained parallelism. Repair hidden clocks, shared globals, leaked tasks, unordered data, port collisions, or resource lifetime. Record the root cause in the fixing change.

### What Not To Test

- Framework behavior already owned by FastAPI, Pydantic, SQLAlchemy, or pytest.
- Private helper call order when public behavior proves the same contract.
- Dataclass field storage, trivial getters, or type annotations at runtime.
- Generated protobuf implementation details.
- Third-party client behavior through mocks of that same client.
- `main.py` line coverage beyond one startup/smoke proof.

Test owned decisions, mappings, configuration, and failure behavior. Integration tests prove the assumptions made about dependencies.

### Load, Soak, And End-To-End Proof

Load tests justify concurrency, pool, timeout, cache, and SLO settings; soak tests expose slow resource growth. They run against a dedicated deployed environment, not inside pytest or `make verify`. Record request mix, duration, artifact, dependency capacity, latency/error results, and resource curves.

Keep end-to-end tests few. Start the built artifact against real local dependencies, prove probes plus one critical path, send SIGTERM during in-flight work, and assert bounded drain. Every detailed behavior remains proven lower in the pyramid.

### Eventing-Specific Proof

- Contract tests pin envelope and payload compatibility.
- Duplicate delivery proves idempotency and inbox behavior.
- Replay and out-of-order delivery prove the actual ordering contract.
- Transient/permanent failures prove retry classification, settlement, and DLQ exhaustion.
- Real-broker integration proves prefetch, redelivery, and acknowledgement semantics when those are broker-specific.

## Common Mistakes And Forbidden Patterns

- Tests under `src/`, or tests importing an uninstalled checkout accidentally.
- Broad root `conftest.py`, autouse state, or session fixtures for mutable business data.
- Unmarked async tests or `@pytest.fixture` used for async fixtures under strict mode.
- `time.sleep()`/`asyncio.sleep()` used as synchronization, or ambient current time in expected output.
- Tasks, clients, engines, responses, containers, or event loops leaked by teardown.
- Dynamic mocks without autospec, patching the definition instead of the consumer, or call scripts that pin internals.
- `ASGITransport` tests assumed to have run lifespan.
- SQLite presented as proof of PostgreSQL behavior.
- Integration markers never exercised in CI, or skips hiding a broken integration lane.
- Golden output updated blindly, coverage floor lowered, or easy lines added for percentage only.
- Flakes hidden with retries, timing padding, permissive `xfail`, or permanent quarantine.

## Verification And Proof

```bash
uv run pytest -m "not integration"
uv run pytest -m integration
PYTHONASYNCIODEBUG=1 uv run pytest -W error
uv run pytest --cov=src/<app> --cov-branch --cov-report=term-missing
make verify
```

Testing is done when the selected proof matches the change risk; async work is deterministic and fully owned; real boundaries are exercised where their semantics matter; contract changes have reviewed exact-shape evidence; the integration lane runs rather than merely collecting; coverage does not regress; and no warning, leak, skip, rerun, or quarantine hides a defect.

Related: [time](../foundations/time.md), [concurrency and asyncio](../foundations/concurrency-and-asyncio.md), [serialization](../foundations/serialization.md), and [framework selection](../decisions/framework-selection.md).
