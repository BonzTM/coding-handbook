# AGENTS.md - TypeScript Project Contract

This is the authoritative fast-path contract for autonomous agents working in a new TypeScript repository.
Read this file first; use [maintainer-reference.md](maintainer-reference.md) for slower-path architecture and rationale.

## Purpose

- Use this file for repo-wide invariants, change defaults, change-to-file routing, and the verification bar.
- Use [maintainer-reference.md](maintainer-reference.md) for lifecycle, architecture, taxonomy, and troubleshooting.
- Use [recipes/README.md](recipes/README.md), [checklists/README.md](checklists/README.md), and [decisions/README.md](decisions/README.md) for procedures and gates.

## Source Of Truth

- This file is the fast path. Detailed docs refine it; they do not weaken it.
- Runtime, package, ESM, compiler, and layout defaults live in [foundations/project-setup.md](foundations/project-setup.md).
- Import direction lives in [foundations/module-design.md](foundations/module-design.md); type policy in [foundations/type-system.md](foundations/type-system.md); shared seams in [foundations/shared-constructs.md](foundations/shared-constructs.md).
- Domain, wire, and compatibility rules live in [foundations/data-modeling.md](foundations/data-modeling.md), [foundations/serialization.md](foundations/serialization.md), and [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md).
- Runtime correctness lives in [foundations/async-and-cancellation.md](foundations/async-and-cancellation.md), [foundations/errors-and-logging.md](foundations/errors-and-logging.md), [foundations/configuration.md](foundations/configuration.md), and [foundations/time.md](foundations/time.md).
- Proof and lint policy live in [quality/testing.md](quality/testing.md) and [quality/linting.md](quality/linting.md).
- Binding dependency decisions live in [decisions/framework-selection.md](decisions/framework-selection.md); hard-to-reverse exceptions require [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).
- Worked examples live at [reference/exampleservice/](reference/exampleservice/), [reference/exampleworker/](reference/exampleworker/), and [reference/examplefrontend/](reference/examplefrontend/).

## Fast Path

1. Identify the project shape in [README.md](README.md); for new work, resolve [checklists/spec-intake.md](checklists/spec-intake.md).
2. Route the change through [Change Routing](#change-routing).
3. Read the named topical docs before editing a new boundary.
4. Follow repository-established exceptions; otherwise use these defaults.
5. Prove the narrow behavior first, then run `npm run verify`.

## Repo-Wide Invariants

- **One package first**: one root `package.json`; npm workspaces require an ADR.
- **Pinned runtime**: Node.js 24 LTS in `engines` and `.nvmrc`; CI uses the pin and `npm ci`.
- **ESM only**: `type: module`; backend imports and output obey NodeNext. No new CommonJS.
- **Strict types**: TypeScript strict baseline is mandatory; `any`, unchecked assertions, and floating promises are forbidden.
- **Boundary parsing**: Zod parses HTTP, env, database rows, messages, files, and external responses before domain use.
- **Dependency direction**: `api -> core <- db`; core imports neither adapter. Composition owns concrete wiring.
- **Cancellation**: `AbortSignal` reaches every I/O and long-running operation; timeouts, retries, and fan-out are bounded.
- **Failures**: typed errors preserve `cause`; expected outcomes may use discriminated results; process-level uncaught failures terminate.
- **Logging**: Pino JSON, injected into reusable code, redacted at construction, and emitted once at the acting boundary.
- **Persistence**: `pg` plus parameterized SQL and parsed rows; schema migration is an explicit deployment step.
- **Frontend**: React function components and hooks; TanStack Query owns server state; accessibility is a merge gate.
- **Verification**: format, lint, types, tests, audit, and build run through `npm run verify`.

## Change Routing

| Change Type | Start Here | Also Update | Verify Or Confirm |
|---|---|---|---|
| Node, npm, TypeScript, ESM, or repository layout | `package.json`, `.nvmrc`, tsconfig, [foundations/project-setup.md](foundations/project-setup.md) | lockfile, CI, Dockerfile, onboarding | `npm ci`; typecheck; build |
| Module boundary, import direction, alias, or export surface | [foundations/module-design.md](foundations/module-design.md) | package exports, lint cycle rules, consumers | lint; build; import smoke test |
| Type strictness, assertion, generic, or augmentation | [foundations/type-system.md](foundations/type-system.md) | tsconfig, lint rules, public types, tests | typecheck and negative type proof |
| Domain types, branded IDs, optionality, or collections | [foundations/data-modeling.md](foundations/data-modeling.md) | serialization, DB mapping, contracts | construction and round-trip tests |
| JSON or other wire serialization | [foundations/serialization.md](foundations/serialization.md) | schemas, DTO mappers, golden fixtures | parse, round-trip, golden, rejection tests |
| OpenAPI, shared schema, or compatibility change | [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md), [recipes/deprecate-and-remove-contract.md](recipes/deprecate-and-remove-contract.md) | clients, release notes, schema artifacts | compatibility review and contract tests |
| Error taxonomy, mapping, or structured logging | [foundations/errors-and-logging.md](foundations/errors-and-logging.md) | problem details, logger fields, redaction | mapping tests; log-once review |
| Async flow, timeout, cancellation, or concurrency | [foundations/async-and-cancellation.md](foundations/async-and-cancellation.md) | callers, shutdown, telemetry | timeout, abort, drain, and bound tests |
| Configuration key or startup default | [foundations/configuration.md](foundations/configuration.md), [recipes/add-config-key.md](recipes/add-config-key.md) | `.env.example`, deploy config, tests | invalid config fails startup |
| Feature flag | [foundations/configuration.md](foundations/configuration.md) (### Feature Flags) | default, owner, expiry, rollout tests | both branches; removal issue |
| Clock, timezone, timer, or schedule | [foundations/time.md](foundations/time.md) | Clock seam, fake timers, storage/wire mapping | deterministic tests; no sleeps |
| Shared helper or cross-cutting construct | [foundations/shared-constructs.md](foundations/shared-constructs.md) | consumers and composition | ownership statement; no cycles |
| Naming, comments, JSDoc, or review heuristics | [foundations/style-and-review.md](foundations/style-and-review.md) | public docs and review checklist | format, lint, API readability |
| Git, commit, branch, or changelog | [foundations/git-workflow.md](foundations/git-workflow.md), [operations/ci-and-release.md](operations/ci-and-release.md) | changelog, PR template | convention and release-note review |
| HTTP endpoint or response | `src/api/`, [services/http-services.md](services/http-services.md), [recipes/add-http-endpoint.md](recipes/add-http-endpoint.md) | core use case, schemas, telemetry | `app.inject()` tests; smoke test |
| HTTP hook, plugin, or middleware | [services/http-services.md](services/http-services.md), [recipes/add-http-middleware.md](recipes/add-http-middleware.md) | registration scope/order, auth, tests | encapsulation and negative tests |
| Idempotent HTTP write | [services/http-services.md](services/http-services.md), [recipes/add-idempotent-write.md](recipes/add-idempotent-write.md) | dedupe store, replay contract, metrics | duplicates yield one effect |
| React component | [services/react-applications.md](services/react-applications.md), [recipes/add-react-component.md](recipes/add-react-component.md) | styles, accessibility, component tests | RTL user behavior and a11y review |
| React hook | [services/react-applications.md](services/react-applications.md), [recipes/add-react-hook.md](recipes/add-react-hook.md) | call sites, effect cleanup, tests | hook behavior under rerender/unmount |
| Frontend route | [services/react-applications.md](services/react-applications.md), [recipes/add-frontend-route.md](recipes/add-frontend-route.md) | navigation, auth, errors, lazy boundary | navigation and deep-link tests |
| Frontend server-state query or mutation | [services/frontend-data-and-state.md](services/frontend-data-and-state.md) | query keys, API client, invalidation | loading/error/success/cancel tests |
| Frontend form | [services/react-applications.md](services/react-applications.md), [recipes/add-form.md](recipes/add-form.md) | Zod schema, mutation, accessible errors | keyboard, validation, submit tests |
| Frontend bundle, Vite, assets, or code splitting | [services/react-applications.md](services/react-applications.md), [operations/deployment.md](operations/deployment.md) | browser targets, cache headers, CI | production build and bundle review |
| Database query, transaction, or row mapping | `src/db/`, [services/database.md](services/database.md), [recipes/add-database-feature.md](recipes/add-database-feature.md) | core port, schema, telemetry | real PostgreSQL integration test |
| Database migration | migrations, [services/database.md](services/database.md), [recipes/add-migration.md](recipes/add-migration.md) | deploy order, rollback, compatibility | apply on empty and prior schema |
| External HTTP client | `src/lib/http/`, [recipes/add-external-client.md](recipes/add-external-client.md) | timeout, SSRF controls, schemas, telemetry | local-server timeout/parse tests |
| Background worker or scheduler | [foundations/async-and-cancellation.md](foundations/async-and-cancellation.md), [recipes/add-background-worker.md](recipes/add-background-worker.md), [recipes/add-scheduled-job.md](recipes/add-scheduled-job.md) | composition, readiness, telemetry | abort and drain smoke tests |
| Event producer or consumer | [services/eventing-and-messaging.md](services/eventing-and-messaging.md), [recipes/add-event-publisher.md](recipes/add-event-publisher.md), [recipes/add-event-consumer.md](recipes/add-event-consumer.md) | schemas, outbox/inbox, DLQ, telemetry | replay, idempotency, broker integration |
| Cache, key, TTL, or invalidation | [services/caching.md](services/caching.md) | callers, config, metrics | hit/miss/invalidation/stampede tests |
| Logs, metrics, traces, or health endpoints | [operations/observability.md](operations/observability.md), [recipes/add-metric.md](recipes/add-metric.md) | dashboards, alerts, route hooks | `/livez`, `/readyz`, export smoke test |
| Timeout, retry, breaker, rate limit, or load shedding | [operations/resilience.md](operations/resilience.md) | clients/consumers, telemetry, idempotency | fault and load tests |
| Container, runtime limits, signals, or rollout | [operations/deployment.md](operations/deployment.md), [templates/Dockerfile](templates/Dockerfile) | CI image, probes, config | non-root image and shutdown smoke |
| SLO, alert, runbook, or on-call | [operations/operability.md](operations/operability.md), [templates/runbook.md](templates/runbook.md) | dashboards, ownership, error budget | symptom alert and runbook exercise |
| Security-sensitive boundary | [operations/security.md](operations/security.md), [checklists/security-review.md](checklists/security-review.md) | validation, authz, secrets, release note | negative tests; audit; SSRF/injection review |
| Audit logging | [operations/security.md](operations/security.md) (### Audit Logging) | event schema, sink access, retention | action emits tamper-resistant audit record |
| PII, classification, retention, export, or deletion | [operations/data-handling.md](operations/data-handling.md) | schemas, redaction, jobs, access controls | deletion/retention proof; no telemetry leak |
| CI, dependency automation, publishing, or release | [operations/ci-and-release.md](operations/ci-and-release.md) | lockfile, changelog, artifacts, rollback | clean checkout `npm run verify` |
| New dependency or framework | [decisions/framework-selection.md](decisions/framework-selection.md), [recipes/bump-dependency.md](recipes/bump-dependency.md) | adapter, lockfile, security review | rationale and dependency diff |
| ESLint, Prettier, or type-check policy | [quality/linting.md](quality/linting.md) | scripts, CI, suppressions | lint, format check, typecheck |
| Test taxonomy, fixtures, mocks, or coverage | [quality/testing.md](quality/testing.md) | changed boundary and CI stage | deterministic relevant suite |
| Architecture decision | [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md) | new/superseded ADR and affected docs | accepted record before implementation |
| CLI command or flag | [recipes/add-cli-command.md](recipes/add-cli-command.md) | config, help, exit codes | command tests, `--help`, smoke run |
| Package/library version release | [recipes/release-library-version.md](recipes/release-library-version.md), [checklists/release.md](checklists/release.md) | exports, types, changelog, consumers | pack/install smoke and compatibility |
| New repository scaffold | [checklists/new-project.md](checklists/new-project.md), [templates/README.md](templates/README.md), matching `reference/` exemplar | placeholders, ownership, CI | fresh clone `npm run verify` |
| Ownership transfer | [onboarding-and-handoff.md](onboarding-and-handoff.md), [checklists/handoff.md](checklists/handoff.md) | CODEOWNERS, access, runbooks, decisions | new owner verifies and deploys unaided |

## High-Value Boundaries

- `src/index.ts` owns composition and process lifetime; it holds no business rules.
- `src/core/` owns domain behavior and ports; it imports no Fastify, React, Pino, `pg`, or environment state.
- `src/api/` translates validated transport DTOs into core calls and maps outcomes to transport responses.
- `src/db/` owns SQL, transaction helpers, row schemas, and persistence mapping.
- `src/config/` owns environment parsing and exports a typed object created once at startup.
- `src/telemetry/` owns logger and OpenTelemetry construction, not domain decisions.
- `src/lib/http/` owns outbound fetch policy, timeouts, safe URL resolution, response parsing, and telemetry.
- `src/testutil/` owns test-only builders and controlled fakes; production code never imports it.

## Proof Hints

- Transport changes need narrow handler/component tests and a boundary smoke test.
- Persistence changes need Testcontainers-backed PostgreSQL proof; mocked `pg` calls are insufficient.
- Async changes need cancellation, timeout, ownership, and shutdown tests.
- Config changes need valid, defaulted, malformed, and missing-input tests.
- Dependency changes need a reviewed lockfile, compatibility proof, and audit outcome.
- UI changes need keyboard, accessible-name, loading, empty, error, and success coverage where applicable.

## Working Norms

- Prefer small, reviewable changes; preserve established architecture unless architecture is the task.
- Do not bypass layers because TypeScript structural typing makes it easy.
- Keep functions focused, use guard clauses, and extract parsing and decisions from I/O orchestration.
- Add no global singleton when explicit composition can pass the dependency.
- Write the proving test before claiming changed behavior works whenever practical.
- Fix verification failures or report the exact failure; never claim a red gate is green.

## Baseline Verification

| Goal | Command or stage | Expectation |
|---|---|---|
| lockfile honesty | clean-checkout `npm ci` | `package.json` and committed lockfile agree; install succeeds without rewriting |
| formatting | `prettier --check .` | no diff |
| lint correctness | `eslint .` | flat config and type-aware rules pass with zero warnings |
| type safety | `tsc --noEmit` or project references/build type gate | zero diagnostics |
| tests | `jest` | deterministic unit/component suites pass offline |
| integration | explicit integration flag/script | real dependencies pass where available; CI runs the gated suite |
| supply chain | `npm audit` with repository threshold | no unaccepted blocking advisories |
| build | backend `tsc` emit or frontend `vite build` | production artifact builds from a clean install |
| canonical gate | `npm run verify` | all stages above run in stable order and fail fast |

The [templates/Makefile](templates/Makefile) keeps `verify` as a one-line shim: `verify: ; npm run verify`. The npm script is canonical; Make does not duplicate policy.

## Slow Path Docs

- Architecture and lifecycle: [maintainer-reference.md](maintainer-reference.md)
- Project and module shape: [foundations/project-setup.md](foundations/project-setup.md), [foundations/module-design.md](foundations/module-design.md)
- Runtime correctness: [foundations/async-and-cancellation.md](foundations/async-and-cancellation.md), [foundations/errors-and-logging.md](foundations/errors-and-logging.md), [foundations/configuration.md](foundations/configuration.md)
- Backend and frontend services: [services/http-services.md](services/http-services.md), [services/react-applications.md](services/react-applications.md), [services/frontend-data-and-state.md](services/frontend-data-and-state.md)
- Proof and operations: [quality/testing.md](quality/testing.md), [operations/observability.md](operations/observability.md), [operations/security.md](operations/security.md)
