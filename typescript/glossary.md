# Glossary

> **Lookup aid, not required reading.** Terms are alphabetical and link to the governing rule.

## Adapter
A module translating an external framework or representation into domain-facing values and ports. Owned by [module design](foundations/module-design.md).

## ADR
An immutable record of a consequential decision, alternatives, and consequences. Owned by [architecture decision records](decisions/architecture-decision-records.md).

## AbortSignal
The cancellation contract passed through I/O, long-running work, React-owned requests, and process shutdown. Owned by [async and cancellation](foundations/async-and-cancellation.md).

## Boundary Parsing
Conversion of untrusted `unknown` input into a validated typed value, normally through Zod. Owned by [type system](foundations/type-system.md) and [serialization](foundations/serialization.md).

## Composition Root
`src/index.ts`, the only module that knows concrete adapters on every side and constructs the runnable application. Owned by [module design](foundations/module-design.md).

## Contract
An externally observable HTTP, event, persistence, configuration, browser, or package behavior that independent code may rely on. Owned by [contracts and compatibility](foundations/contracts-and-compatibility.md).

## Core
`src/core/`, which owns domain values, business decisions, use cases, and consumer-owned ports without framework or adapter dependencies. Owned by [module design](foundations/module-design.md).

## DTO
A boundary-specific data-transfer value parsed and mapped explicitly rather than reused as a domain model. Owned by [serialization](foundations/serialization.md).

## Envelope
Stable event metadata containing ID, type, version, occurred-at instant, producer, trace context, and payload. Owned by [eventing](services/eventing-and-messaging.md).

## Error Budget
The allowed unsuccessful portion of an SLI population over an SLO window. Owned by [operability](operations/operability.md).

## Expand/Migrate/Contract
A staged schema or contract change: add compatible shape, move/backfill use, then remove old shape only after old code is gone. Owned by [database](services/database.md).

## Idempotency
The property that repeating a request or message does not repeat its durable effect. Owned by [HTTP services](services/http-services.md) and [eventing](services/eventing-and-messaging.md).

## Inbox
A durable consumer dedupe record keyed by event ID plus consumer identity and committed with the effect. Owned by [eventing](services/eventing-and-messaging.md).

## NodeNext
The backend TypeScript module and resolution mode that models Node's ESM behavior, including runtime-correct relative extensions. Owned by [project setup](foundations/project-setup.md).

## Outbox
A publication-intent row committed in the same PostgreSQL transaction as the domain change, then relayed asynchronously. Owned by [eventing](services/eventing-and-messaging.md).

## Port
A narrow behavior contract defined by the consumer, implemented by an adapter, and wired in composition. Owned by [shared constructs](foundations/shared-constructs.md).

## Problem Details
The RFC 9457-shaped HTTP failure representation produced by transport-owned error mapping. Owned by [errors and logging](foundations/errors-and-logging.md).

## Readiness And Liveness
Readiness answers whether the process can accept work; liveness answers whether it should be restarted. Owned by [observability](operations/observability.md).

## Server State
Remote data whose fetch, cache, invalidation, and mutation lifecycle is owned by TanStack Query in a React application. Owned by [frontend data and state](services/frontend-data-and-state.md).

## Thin Entrypoint
`src/main.ts`, which loads configuration, creates logging/lifetime control, handles fatal startup and signals, and contains no business rules. Owned by [project setup](foundations/project-setup.md).

## Trust Boundary
A point where HTTP, environment, database, message, file, storage, or third-party data enters owned code and must be bounded and parsed. Owned by [security](operations/security.md).

## Verify Gate
`npm run verify`: format check, typed lint, typecheck, deterministic tests, audit policy, and build in stable fail-fast order. Owned by [CI and release](operations/ci-and-release.md).

## Zod Schema
The runtime parser that establishes a typed value from untrusted input; it is a boundary artifact, not a domain dependency requirement. Owned by [serialization](foundations/serialization.md).

## Related

- [README.md](README.md)
- [AGENTS.md](AGENTS.md)
- [maintainer-reference.md](maintainer-reference.md)
