# Glossary

> **Lookup aid, not required reading.** Consult one entry when a handbook term is unclear; nothing here needs to be read front-to-back to build.

Canonical vocabulary for this handbook. Each term has one meaning across Python repos. Each entry points to the doc that owns the full rule. Terms are alphabetical.

## Adapter (transport or infrastructure adapter)
A module under `api/`, `db/`, `clients/`, or `workers/` that translates a framework, wire, persistence, or broker concern into a core-owned `Protocol`. It holds mapping and mechanism, never business rules. Owned by [foundations/package-design.md](foundations/package-design.md).

## ADR (Architecture Decision Record)
A short immutable record of one decision, its forcing context, alternatives, and accepted consequences. Accepted records are superseded, never rewritten. Owned by [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).

## ASGI
The asynchronous server/application interface between uvicorn and FastAPI/Starlette. ASGI types stop at the HTTP adapter; core does not depend on them. Owned by [services/http-services.md](services/http-services.md).

## Composition root
The narrow process boundary that constructs validated settings, logging, telemetry, engines, clients, repositories, use cases, routers, and lifetime ownership. It coordinates dependencies and owns no business behavior. Owned by [foundations/project-setup.md](foundations/project-setup.md).

## Contract (wire or data contract)
The externally observable shape another component depends on: HTTP/protobuf/message payload and errors for wire contracts; schema and migration compatibility for data contracts. Owned by [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md).

## Core / domain (`src/<app>/core`)
The framework-free package holding domain types, use cases, and consumer-defined ports. Import Linter forbids it from importing `api`, `db`, `clients`, or `telemetry`. Owned by [foundations/package-design.md](foundations/package-design.md).

## Distribution package
The installable and publishable project identified by PEP 621 `[project].name`, producing wheel and sdist artifacts. Its name may contain hyphens and need not equal the import package spelling. Owned by [foundations/project-setup.md](foundations/project-setup.md).

## DTO (data transfer object)
A Pydantic or generated wire type defined at a trust boundary and explicitly mapped to or from a plain domain type. It is not a domain entity. Owned by [foundations/serialization.md](foundations/serialization.md).

## Event loop
The thread-confined asyncio scheduler that runs tasks cooperatively. Blocking I/O or CPU work on it stalls every task sharing that loop. Owned by [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md).

## Exception group
The Python 3.11 aggregate raised when multiple child tasks fail, commonly from `TaskGroup`; selective handling uses `except*` without discarding unrelated failures. Owned by [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md).

## Fail-fast config
One pydantic-settings object loaded and fully validated before any listener, worker loop, engine, or external client starts. Invalid config terminates startup. Owned by [foundations/configuration.md](foundations/configuration.md).

## Idempotency
The property that replaying the same request or message produces no additional effect. Dedupe state is durable and keyed by an explicit request or event identity. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Import package
The Python module hierarchy imported by code, such as `orders`, located under `src/orders/`. It is distinct from the distribution package metadata name. Owned by [foundations/package-design.md](foundations/package-design.md).

## Inbox
A durable consumer-side dedupe record keyed by event ID, committed with the business effect before settlement so redelivery is harmless. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Lifespan
The FastAPI async context manager that constructs app-scoped resources before serving and closes them after draining. It owns shared clients, engines, exporters, and similar resources. Owned by [services/http-services.md](services/http-services.md).

## Liveness vs readiness
`/livez` proves the process can answer cheaply; `/readyz` proves it can accept work and may include bounded critical-dependency state. A process may be live but not ready. Owned by [operations/observability.md](operations/observability.md).

## Lockfile (`uv.lock`)
The resolved dependency graph uv uses for reproducible application environments. Applications commit it; `uv lock --check` proves metadata has not drifted. Owned by [foundations/project-setup.md](foundations/project-setup.md).

## Outbox
A durable publish-intent row written in the same transaction as business state, then relayed asynchronously to avoid a dual-write gap. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Port (`Protocol` seam)
A narrow structurally typed interface owned by the consumer in core and implemented by an outer adapter. It names what the use case needs, not the technology providing it. Owned by [foundations/typing-discipline.md](foundations/typing-discipline.md).

## Pydantic boundary
The trust boundary where untrusted HTTP, environment, file, message, or third-party data is parsed into a typed DTO. The DTO is then mapped into plain domain values; Pydantic does not spread through core. Owned by [foundations/data-modeling.md](foundations/data-modeling.md).

## `py.typed`
The marker included in a library distribution to declare that inline type information is part of the installed package. Libraries ship and test it in the built wheel. Owned by [foundations/typing-discipline.md](foundations/typing-discipline.md).

## Relay
The owned worker that publishes committed outbox rows and records success or retry state. It exposes delivery and failure semantics instead of hiding them behind a repository. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Settlement
The consumer action that acknowledges a message or commits its offset, performed only after durable effects and inbox state commit. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Sdist (source distribution)
The standardized source archive used to build a distribution. Library releases build and inspect both sdist and wheel because each is a consumer contract. Owned by [operations/ci-and-release.md](operations/ci-and-release.md).

## SLI (Service Level Indicator)
A ratio of good events to valid events computed from emitted telemetry; one SLI measures one user-visible promise. Owned by [operations/operability.md](operations/operability.md).

## SLO (Service Level Objective)
A target for an SLI over an explicit rolling window. Both target and window are required. Owned by [operations/operability.md](operations/operability.md).

## Source (`src/`) layout
The layout that keeps import packages under `src/` so tests and tools exercise the installed project instead of accidentally importing from the checkout root. Owned by [foundations/project-setup.md](foundations/project-setup.md).

## Structured concurrency
Concurrency in which child tasks have a lexical owner, propagate failure, and are awaited before the owning scope exits. `asyncio.TaskGroup` is the default primitive. Owned by [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md).

## Task group
Python 3.11's async context manager for owning related tasks. Exiting waits for children; a child failure cancels siblings and raises the resulting exception group. Owned by [foundations/concurrency-and-asyncio.md](foundations/concurrency-and-asyncio.md).

## Thin entrypoint
The contract that `[project.scripts]` and `__main__.py` only delegate to composition and process lifetime; they contain no business logic. Owned by [foundations/project-setup.md](foundations/project-setup.md).

## Trust boundary
Any point where data enters from HTTP, environment, files, queues, databases, caches, subprocesses, or third parties. Parse and normalize it before business logic. Owned by [operations/security.md](operations/security.md).

## Verify gate (`make verify`)
The single committed proof gate: lock-check, frozen sync, format-check, lint, imports, types, test, audit. Local and CI run the same ordered targets. Owned by [quality/testing.md](quality/testing.md) and [operations/ci-and-release.md](operations/ci-and-release.md); summarized in [AGENTS.md](AGENTS.md).

## Wheel
The built distribution installed without executing a source build. Library proof installs and inspects the wheel to catch missing modules, data, and `py.typed`. Owned by [operations/ci-and-release.md](operations/ci-and-release.md).

## Related

- [README.md](README.md) — how layers, defaults, and proof fit together.
- [AGENTS.md](AGENTS.md) — the fast-path contract using this vocabulary.
- [AGENTS.md](AGENTS.md) (## Change Routing) — which doc owns each change surface.
