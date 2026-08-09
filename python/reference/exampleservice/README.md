# exampleservice

A small, complete FastAPI sidecar that is the executable reference for the Python engineering handbook. It manages an orders feature end to end and proves the handbook's boundaries, lifecycle, tooling, and verification gate on real code.

## What It Demonstrates

- PEP 621 metadata, a `src/` package, a thin console entrypoint, a committed uv lock, and one canonical `make verify` gate.
- A framework-free `core/` with frozen dataclasses, `NewType` identifiers, typed exceptions, consumer-owned `Protocol` ports, and an injected UTC `Clock`.
- Thin FastAPI routers, strict Pydantic v2 request DTOs, explicit response mapping, RFC 9457-style problem responses, bounded request IDs, and request deadlines.
- One lifespan-owned HTTPX client with an explicit four-phase timeout, pool bound, per-dependency semaphore, bounded retry attempts, full jitter, and injectable monotonic clock, sleep, and randomness seams.
- A private Prometheus registry, `/livez`, drain-aware `/readyz`, `/metrics`, JSON logging, and offline-safe OpenTelemetry server/client instrumentation.
- One bounded queue and fixed-size `asyncio.TaskGroup` worker set, owned and drained by application lifespan.
- Deterministic pytest proof through hand-written Protocol fakes, `httpx.AsyncClient` with `ASGITransport`, HTTPX `MockTransport`, a fake clock/sleeper, a lifespan test, and an exact serialization golden.

## Deliberate Persistence Scope

The exemplar uses a concurrency-safe in-memory repository behind the core-owned `OrderRepository` Protocol. It proves dependency direction, adapter substitution, and process-local behavior without Docker or a database.

It does not claim durability, cross-replica idempotency, SQL correctness, or migration proof. Production persistence follows [database.md](../../services/database.md) and [add-migration.md](../../recipes/add-migration.md): SQLAlchemy 2.0 async with asyncpg, explicit Alembic deploy migrations, and real-PostgreSQL integration tests. An in-process dictionary is not a valid implementation of the handbook's durable idempotent-write contract, so this exemplar intentionally does not advertise idempotent writes.

## Run It

Copy the safe local example, replace its non-production catalog token if needed, then start the service:

```bash
cp .env.example .env
make run
```

The process starts without contacting the catalog. `GET` requests and operational endpoints work offline; `POST /orders` calls the configured catalog dependency and therefore needs a compatible upstream at `OUTBOUND_BASE_URL`.

```bash
curl --fail http://127.0.0.1:8080/livez
curl --fail http://127.0.0.1:8080/readyz
curl --fail http://127.0.0.1:8080/metrics
curl --fail http://127.0.0.1:8080/orders/order-1
curl --fail --request POST http://127.0.0.1:8080/orders \
  --header 'Content-Type: application/json' \
  --header 'X-Request-ID: local-example-1' \
  --data '{"orderId":"order-1","sku":"sku-1"}'
```

## Verify

```bash
make verify
```

The local gate and CI contract are identical:

| Stage | Command | Proof |
|---|---|---|
| lock | `uv lock --check` | metadata and committed resolution agree |
| sync | `uv sync --frozen` | the exact graph installs without resolution |
| format | `uv run ruff format --check .` | Ruff owns formatting and reports no diff |
| lint | `uv run ruff check .` | curated correctness, async, and security rules pass |
| imports | `uv run lint-imports` | core and adapter dependency contracts hold |
| types | `uv run mypy .` | strict typing passes for source and tests |
| test | `uv run pytest` | deterministic unit, transport, lifecycle, and golden proof passes |
| audit | `uv run --with pip-audit pip-audit` | the resolved environment has no blocking known vulnerability |

Each stage is also an individually runnable Make target.

## Package Map

| Package / file | Responsibility | Governing handbook doc |
|---|---|---|
| `src/exampleservice/__main__.py` | validated settings, one logging configuration, Uvicorn handoff | [project setup](../../foundations/project-setup.md) |
| `src/exampleservice/app.py` | app factory, resource wiring, TaskGroup ownership, ordered bounded shutdown | [concurrency and asyncio](../../foundations/concurrency-and-asyncio.md), [HTTP services](../../services/http-services.md) |
| `src/exampleservice/config.py` | immutable, fail-fast pydantic-settings graph | [configuration](../../foundations/configuration.md) |
| `src/exampleservice/core/` | domain values, typed failures, use cases, consumer-owned ports, Clock seam | [package design](../../foundations/package-design.md), [time](../../foundations/time.md) |
| `src/exampleservice/api/http/` | DTO mapping, `Depends` wiring, problem boundary, request-ID middleware, probes | [HTTP services](../../services/http-services.md), [serialization](../../foundations/serialization.md) |
| `src/exampleservice/db/memory.py` | replaceable in-memory repository adapter | [database](../../services/database.md) |
| `src/exampleservice/clients/catalog.py` | typed response parsing, timeouts, retries, jitter, bulkhead | [resilience](../../operations/resilience.md) |
| `src/exampleservice/telemetry/` | `dictConfig`, JSON logs, private Prometheus registry, health, OpenTelemetry | [observability](../../operations/observability.md) |
| `src/exampleservice/workers/` | bounded queue, fixed concurrency, owned TaskGroup, drain | [concurrency and asyncio](../../foundations/concurrency-and-asyncio.md) |
| `tests/` | fakes, ASGI transport, fake HTTP transport/time, lifecycle, golden contract | [testing](../../quality/testing.md) |

## Production Extensions

The handbook carries the intentionally omitted concerns: PostgreSQL and Alembic in [database.md](../../services/database.md), durable idempotency in [add-idempotent-write.md](../../recipes/add-idempotent-write.md), authentication and authorization in [security.md](../../operations/security.md), and an exporting OpenTelemetry pipeline in [observability.md](../../operations/observability.md). Add them through the documented adapters and proof lanes; do not move them into `core/`.
