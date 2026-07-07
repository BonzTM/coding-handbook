# Glossary

> **Lookup aid, not required reading.** Consult a single entry when a handbook term is unclear; nothing here needs to be read front-to-back to build.

Canonical vocabulary for this handbook. Every term below has exactly one meaning across all repos; use the word for nothing else, and define it nowhere else. Each entry is one or two sentences plus the doc that owns the full rule — read that doc before relying on the term. Start from [README.md](README.md) for how these pieces fit together.

Terms are alphabetical.

## Adapter (transport adapter)
A package under `internal/api/http` or `internal/api/grpc` that translates a wire request into a core call and a core result back into a wire response. Adapters hold decode/validate/map logic and the error-to-status mapping; they hold no business rules and never touch SQL or migrations. Owned by [foundations/package-design.md](foundations/package-design.md).

## ADR (Architecture Decision Record)
A short, immutable document stating one decision, the context that forced it, and the consequences accepted; once accepted it is frozen and only superseded, never edited. Owned by [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).

## Contract (wire / data contract)
The externally observable shape another system depends on: a wire contract is the request/response or message payload; a data contract is the persisted schema. Both are first-class engineering surfaces, versioned and compatibility-checked like code. Owned by [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md).

## Core / domain (`internal/core`)
The package layer that holds business rules and domain types, depending only on the stdlib, domain packages, and narrowly scoped contracts. It must not import transport or database-specific packages; everything else points inward at it. Owned by [foundations/package-design.md](foundations/package-design.md).

## DTO (data transfer object)
A wire struct defined in the transport package and mapped explicitly to and from domain types at the boundary; it is never the domain or `internal/core` type. The DTO decouples the public wire contract from internal refactors. Owned by [foundations/serialization.md](foundations/serialization.md).

## Envelope
A stable metadata wrapper around a message payload (event ID, type, source, time, correlation/trace context), CloudEvents-style even where CloudEvents is not adopted literally. It carries the fields delivery and dedupe logic key on, separate from the business payload. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Error budget
The agreed allowance of failure: `(1 - target) × valid events` over an SLO window. It is spent deliberately on releases and risk, and its named owner decides what happens when it is exhausted. Owned by [operations/operability.md](operations/operability.md).

## Expand/contract migration
The staged pattern for destructive schema changes across releases: add the new shape and dual-write, backfill, switch reads, then drop the old shape only in a later release once no running version references it. Never drop, rename, or narrow in the same release as code that still reads the old shape. Owned by [services/database.md](services/database.md).

## Fail-fast config
Configuration loaded and fully validated at startup, before any listener opens, worker starts, or external client is created; an invalid config aborts the process rather than failing later under load. Owned by [foundations/configuration.md](foundations/configuration.md); shown in `reference/exampleservice/internal/config`.

## Idempotency
The property that processing the same message or request more than once yields the same effect as processing it once. Consumers are made idempotent (typically via an inbox/dedupe table keyed by event ID) before retries are added. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Inbox
A durable dedupe table on the consumer side, keyed by event ID, that records which messages have already been handled so duplicate delivery is harmless. The consumer-side counterpart to the outbox. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Liveness vs readiness
Liveness (`/livez`) answers "is the process up and not deadlocked?"; readiness (`/readyz`) answers "can it serve traffic right now?" and is dependency-aware. They are distinct endpoints: a service can be live but not ready when a critical dependency is down. Owned by [operations/observability.md](operations/observability.md).

## Outbox
A transactional outbox table written in the same DB transaction as the business state, so committing state and recording the intent to publish is one atomic action; the row is relayed to the broker asynchronously. Used by default when a service both commits state and emits an event. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Relay
The asynchronous worker that reads committed outbox rows and publishes them to the broker, decoupling the publish from the original transaction. Its failure semantics are explicit, not hidden behind a repository abstraction. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## Runtime / assembly (`internal/runtime`)
The package of assembly helpers that build and compose dependencies so `main` stays thin: it constructs adapters, stores, and telemetry into a runnable whole and manages lifecycle, without owning business rules. Owned by [foundations/project-setup.md](foundations/project-setup.md).

## Settlement
The point at which a consumer durably finalizes a message — acknowledge or commit offset — performed only after durable side effects, never after mere parsing. Owned by [services/eventing-and-messaging.md](services/eventing-and-messaging.md).

## SLI (Service Level Indicator)
A ratio of good events to valid events, computed entirely from telemetry the service already emits; one SLI measures one user-visible promise. Owned by [operations/operability.md](operations/operability.md).

## SLO (Service Level Objective)
A target for an SLI over a rolling window, always stated as `target over window` (e.g. "99.9% over 30 rolling days"); both halves are load-bearing. Owned by [operations/operability.md](operations/operability.md).

## Store / repository seam
The narrow persistence interface defined in the consumer's package and named for what the consumer needs (`Store`, not `PostgresUserRepository`), with the database implementation in `internal/db`. It is the substitution point for fakes and the boundary keeping SQL out of core and adapters. Owned by [foundations/package-design.md](foundations/package-design.md); persistence side in [services/database.md](services/database.md).

## Telemetry (`internal/telemetry`)
The package that owns the logging, metrics, and tracing seam — the wiring that turns runtime behavior into logs (`log/slog`), low-cardinality metrics, and traces. It is the construction point observability and SLIs build on. Owned by [operations/observability.md](operations/observability.md); seam shown in `reference/exampleservice/internal/telemetry`.

## Thin main
The contract that `main` only wires config, logging, dependencies, signals, and shutdown — it holds no business logic. Assembly that would bloat `main` moves into `internal/runtime`. Owned by [foundations/project-setup.md](foundations/project-setup.md) and [foundations/package-design.md](foundations/package-design.md).

## Verify gate (`make verify`)
The single committed proof gate: tidy, fmt-check, lint, vet, test, race, vuln, and build, run identically locally and in CI. Coverage is deliberately *not* part of it — coverage runs separately via `make cover`. A change is not done until `make verify` is green. Owned by [quality/testing.md](quality/testing.md) and [operations/ci-and-release.md](operations/ci-and-release.md); the fast-path summary is in [AGENTS.md](AGENTS.md).

## Related

- [README.md](README.md) — how the layers, defaults, and proof gate fit together.
- [AGENTS.md](AGENTS.md) — the fast-path contract that uses this vocabulary.
- [AGENTS.md](AGENTS.md) (## Change Routing) — which doc owns each change surface.
