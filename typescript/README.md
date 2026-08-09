# TypeScript Project Handbook

This handbook is the default engineering contract for new TypeScript repositories targeting Node.js services and React applications.

## Start Here

- Humans: read this file, then follow the reading path for the project shape.
- Agents: read [AGENTS.md](AGENTS.md), route the change, then read the named topical docs and recipes.
- Defaults: one npm package, Node.js 24 LTS, ESM, the newest TypeScript compiler supported by the pinned type-aware lint toolchain, strict type checking, Zod at trust boundaries, and `npm run verify` as the canonical gate.

## Reading Paths

| If you are building... | Read in this order |
|---|---|
| HTTP service | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/module-design.md](foundations/module-design.md) -> [foundations/configuration.md](foundations/configuration.md) -> [services/http-services.md](services/http-services.md) -> [services/database.md](services/database.md) -> [operations/observability.md](operations/observability.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-http-endpoint.md](recipes/add-http-endpoint.md); exemplar: [reference/exampleservice/](reference/exampleservice/) |
| React frontend app | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/type-system.md](foundations/type-system.md) -> [services/react-applications.md](services/react-applications.md) -> [services/frontend-data-and-state.md](services/frontend-data-and-state.md) -> [operations/security.md](operations/security.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-react-component.md](recipes/add-react-component.md); exemplar: [reference/examplefrontend/](reference/examplefrontend/) |
| Full-stack feature | the HTTP service and React paths above, plus [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md), [foundations/serialization.md](foundations/serialization.md), and [recipes/add-frontend-route.md](recipes/add-frontend-route.md) |
| Background worker | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/async-and-cancellation.md](foundations/async-and-cancellation.md) -> [foundations/configuration.md](foundations/configuration.md) -> [operations/observability.md](operations/observability.md) -> [operations/resilience.md](operations/resilience.md) -> [recipes/add-background-worker.md](recipes/add-background-worker.md) |
| Event-driven service | [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [services/eventing-and-messaging.md](services/eventing-and-messaging.md) -> [services/database.md](services/database.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-event-publisher.md](recipes/add-event-publisher.md) -> [recipes/add-event-consumer.md](recipes/add-event-consumer.md); exemplar: [reference/exampleworker/](reference/exampleworker/) |
| CLI tool | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/style-and-review.md](foundations/style-and-review.md) -> [foundations/configuration.md](foundations/configuration.md) -> [decisions/framework-selection.md](decisions/framework-selection.md) -> [recipes/add-cli-command.md](recipes/add-cli-command.md) |
| Library | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/module-design.md](foundations/module-design.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [foundations/style-and-review.md](foundations/style-and-review.md) -> [quality/testing.md](quality/testing.md) -> [checklists/release.md](checklists/release.md) -> [recipes/release-library-version.md](recipes/release-library-version.md) |

Every shape adopts [quality/linting.md](quality/linting.md), [foundations/data-modeling.md](foundations/data-modeling.md), [foundations/serialization.md](foundations/serialization.md), [foundations/time.md](foundations/time.md), and the committed [templates/](templates/) scaffolding. Networked processes additionally adopt [operations/resilience.md](operations/resilience.md), [operations/deployment.md](operations/deployment.md), and [operations/operability.md](operations/operability.md).

## Non-Negotiables

- Pin Node.js 24 in `engines` and `.nvmrc`; commit `package-lock.json`; use `npm ci` in CI.
- Ship ESM only. Backend production builds are JavaScript emitted by `tsc`; native type stripping is a bounded development convenience, never the type checker or deployment artifact.
- Keep `src/api` and `src/db` pointed inward at `src/core`; core code does not import transport, persistence, framework, or process-global configuration.
- Parse every trust boundary with Zod. Raw HTTP, environment, database, queue, and third-party values do not cross into domain logic.
- Pass `AbortSignal` through I/O and long-running work. Every external operation has a timeout; concurrency and retries are bounded.
- Handle every promise. Floating promises, swallowed rejections, and fire-and-forget work without ownership are forbidden.
- Use typed errors and RFC 9457 problem details. Log once at the boundary that can act, with Pino redaction configured.
- Use real PostgreSQL integration tests with Testcontainers. Components are tested through user-visible behavior with React Testing Library.
- Keep telemetry dimensions low-cardinality. Request IDs and user IDs belong in logs and traces, not metric attributes.
- Run `npm run verify`; unfixable audit findings require a time-bounded, owned exception rather than a silent ignore.

## Default Stack

| Concern | Default | Reach for something else when |
|---|---|---|
| Runtime | Node.js 24 LTS, ESM | another runtime is an organizational constraint, via ADR |
| Language/build | current stable TypeScript; `tsc` for backend; Vite for frontend | a library needs a multi-format publishing build, via ADR |
| Package manager | npm, committed lockfile, `npm ci` | workspace scale and measured install pressure justify pnpm, via ADR |
| HTTP | Fastify v5 with Zod type provider | Express 5 compatibility is required by middleware or an inherited app |
| Validation | Zod 4 at every trust boundary | a published JSON Schema is the controlling contract and generation is proven |
| Persistence | `pg`, handwritten parameterized SQL, Zod row parsing | SQL volume justifies Kysely; an ORM requires an ADR |
| Migrations | `node-pg-migrate`, run as an explicit deployment step | the platform already mandates another PostgreSQL migrator |
| Logging | Pino JSON with child loggers and redaction | the hosting platform mandates an adapter; reusable code still accepts an injected logger |
| Testing | Jest 30, React Testing Library, Testcontainers, MSW | Vite-native transform speed materially dominates and Vitest compatibility is proven, via ADR |
| Frontend | React 19, Vite, React Router, TanStack Query | a framework requirement controls rendering or routing, via ADR |
| Client state | local state, then lifted state/context; TanStack Query for server state | complex client-only event transitions justify a dedicated state machine; Redux is not default |
| Observability | OpenTelemetry JS traces/metrics, Pino logs, OTLP | the platform requires a compatible exporter or collector topology |
| Time | injected `Clock`; `Date` at the edge | Temporal is available across every supported runtime or an approved polyfill is adopted |

See [decisions/framework-selection.md](decisions/framework-selection.md) for rationale and escalation rules. Current platform claims are anchored in the [TypeScript download page](https://www.typescriptlang.org/download/), [Node.js 24 documentation](https://nodejs.org/download/release/latest-v24.x/docs/api/), [Fastify reference](https://fastify.dev/docs/latest/Reference/), [Jest 30 ESM guide](https://jestjs.io/docs/30.0/ecmascript-modules), and [React 19 release](https://react.dev/blog/2024/12/05/react-19).

## Handbook Map

- [AGENTS.md](AGENTS.md) — fast-path contract, routing, and verification.
- [maintainer-reference.md](maintainer-reference.md) — slow-path architecture and rationale.
- `foundations/` — project, module, type, boundary, async, error, config, time, shared-seam, style, and Git contracts.
- `quality/` — testing and linting policy.
- `services/` — Fastify, React, server state, PostgreSQL, messaging, and caching.
- `operations/` — observability, resilience, deployment, operability, security, data handling, CI, and release.
- `decisions/` — ADR and framework-selection rules.
- [checklists/README.md](checklists/README.md), [recipes/README.md](recipes/README.md), and [templates/README.md](templates/README.md) — executable guidance and scaffolding.
- `reference/` — verify-green [service](reference/exampleservice/), [worker](reference/exampleworker/), and [frontend](reference/examplefrontend/) exemplars.
- Team process: [onboarding-and-handoff.md](onboarding-and-handoff.md), [glossary.md](glossary.md), and [CONTRIBUTING.md](CONTRIBUTING.md).

## What This Handbook Optimizes For

- obvious dependency direction and small, testable units
- explicit trust boundaries and failure behavior
- bounded work under load and predictable shutdown
- one backend/frontend contract story without unsafe type sharing
- reproducible builds and reviewable dependency changes
- operable systems with useful telemetry and rollback paths

## Where To Go Next

- New repository: [checklists/spec-intake.md](checklists/spec-intake.md), then [checklists/new-project.md](checklists/new-project.md).
- Active work: [AGENTS.md](AGENTS.md) and the matching [recipes/README.md](recipes/README.md) entry.
- Dependency decision: [decisions/framework-selection.md](decisions/framework-selection.md).
- Architecture decision: [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).
- Production readiness: [operations/resilience.md](operations/resilience.md), [operations/deployment.md](operations/deployment.md), and [operations/operability.md](operations/operability.md).
- Worked examples: [reference/exampleservice/](reference/exampleservice/), [reference/exampleworker/](reference/exampleworker/), and [reference/examplefrontend/](reference/examplefrontend/).
