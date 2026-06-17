# Maintainer Map

Purpose: route maintainers and agents to the right files and related obligations for common Go changes.
Read [AGENTS.md](AGENTS.md) first. Use this file when you know what kind of change you are making but do not want to rediscover the normal file set from scratch.

## Documentation Policy

- Keep [AGENTS.md](AGENTS.md) as the fast path only.
- Keep cross-cutting routing information here instead of scattering it through package docs.
- Document sync surfaces and proof steps that code search does not reveal cheaply.
- If a fact is obvious from the package tree alone, prefer search over prose.
- For the full catalogs, see the [recipes/README.md](recipes/README.md), [checklists/README.md](checklists/README.md), and [decisions/README.md](decisions/README.md) indexes.

## Change Map

| Change Type | Start Here | Also Update | Verify Or Confirm |
|---|---|---|---|
| Module path, repo layout, `go` or `toolchain` lines | `go.mod`, `cmd/`, `internal/`, [foundations/project-setup.md](foundations/project-setup.md) | CI build scripts, `go.mod` `tool` directives, onboarding docs | `go mod tidy`, `go build ./...` |
| HTTP or gRPC contract shape, schema, compatibility, generated stubs | `api/**`, transport package, [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md), [recipes/deprecate-and-remove-contract.md](recipes/deprecate-and-remove-contract.md) | clients, generated code policy, release notes, tests | schema or generation check, compatibility review |
| Event or message producers, consumers, payload contracts | eventing package, `api/**`, [services/eventing-and-messaging.md](services/eventing-and-messaging.md), [reference/exampleworker/](reference/exampleworker/) | outbox or inbox storage, telemetry, release notes, tests, DLQ policy | contract tests, replay/idempotency tests, integration against real broker path |
| HTTP endpoint, middleware, request or response shape | `internal/api/http/`, [services/http-services.md](services/http-services.md), [recipes/add-http-endpoint.md](recipes/add-http-endpoint.md), [recipes/add-http-middleware.md](recipes/add-http-middleware.md) | `internal/core/`, request validation, telemetry, route registration, tests | handler tests, smoke test, readiness behavior |
| gRPC proto or server method | `api/**`, `internal/api/grpc/`, [services/grpc-services.md](services/grpc-services.md), [recipes/add-grpc-method.md](recipes/add-grpc-method.md), [reference/examplegrpc/](reference/examplegrpc/) | generated code policy, interceptors, error mapping, docs | proto lint or generate check, service tests, `grpcurl` |
| Business logic or domain rules | `internal/core/`, [foundations/package-design.md](foundations/package-design.md) | package-local tests, transport adapters if contracts changed | targeted unit tests, relevant integration coverage |
| DB queries, schema, transaction behavior | `internal/db/`, migrations, [services/database.md](services/database.md), [recipes/add-database-feature.md](recipes/add-database-feature.md), [recipes/add-migration.md](recipes/add-migration.md) | `sqlc` inputs and outputs, repositories, callers, observability | migration apply test, real DB integration tests |
| Configuration keys or startup defaults | `internal/config/`, [foundations/configuration.md](foundations/configuration.md), [recipes/add-config-key.md](recipes/add-config-key.md) | `.env.example`, startup wiring, docs, tests | config loader tests, startup smoke test |
| Background workers, schedulers, shutdown paths | worker package, [foundations/context-and-concurrency.md](foundations/context-and-concurrency.md), [recipes/add-background-worker.md](recipes/add-background-worker.md), [recipes/add-scheduled-job.md](recipes/add-scheduled-job.md), [reference/exampleworker/](reference/exampleworker/) | `main`, supervisor wiring, telemetry, readiness semantics | `go test -race`, shutdown smoke test, leak check |
| External HTTP or gRPC client | `internal/client/`, [recipes/add-external-client.md](recipes/add-external-client.md) | interfaces in consuming package, retries/timeouts, telemetry, security checks | `httptest.Server` or test server integration, timeout tests |
| Logging, metrics, traces, health endpoints | `internal/telemetry/`, [operations/observability.md](operations/observability.md), [recipes/add-metric.md](recipes/add-metric.md) | handlers, repositories, workers, dashboards or alerts if repo has them | `/metrics`, `/livez`, `/readyz`, log schema review |
| Security-sensitive boundary | boundary package, [operations/security.md](operations/security.md) | validation, auth seams, secret handling, release docs if exposure changed | `govulncheck`, targeted negative tests, path or SSRF checks |
| Audit logging of security-relevant actions | [operations/security.md](operations/security.md) (### Audit Logging) | audit sink/retention, actor and resource fields, telemetry, access controls on the log | audit events emitted at the action boundary; tamper-evidence and retention honored |
| Data classification, PII, retention, compliance | [operations/data-handling.md](operations/data-handling.md) | field-level classification, log/metric redaction, retention jobs, export/delete paths | PII never in logs/metrics; retention and deletion paths exercised by tests |
| Pre-build WHAT decisions (scope, acceptance, spec intake) | [checklists/spec-intake.md](checklists/spec-intake.md) | the task's acceptance criteria, open questions, downstream design docs | spec intake complete before code; acceptance criteria agreed and testable |
| Idempotent HTTP writes (Idempotency-Key, safe retries) | [recipes/add-idempotent-write.md](recipes/add-idempotent-write.md), [services/http-services.md](services/http-services.md) | idempotency-key storage, dedupe window, replay response shape, tests | duplicate requests collapse to one effect; replays return the stored result |
| CLI commands or flags | `cmd/<app>/`, [recipes/add-cli-command.md](recipes/add-cli-command.md) | config loading, help output, README if user-visible | command tests, `--help`, build and smoke run |
| Build, CI, containers, release automation | `.github/workflows/`, build scripts, Dockerfile, [operations/ci-and-release.md](operations/ci-and-release.md) | version injection, changelog, release packaging | CI dry run, `go version -m`, container smoke test |
| New dependency or framework | [decisions/framework-selection.md](decisions/framework-selection.md), [recipes/bump-dependency.md](recipes/bump-dependency.md) | the caller package, proof docs, a `go.mod` `tool` directive if tool-only | written rationale, dependency diff, proof it earns its cost |
| Lint config, golangci-lint policy, formatters | `.golangci.yml`, [quality/linting.md](quality/linting.md) | golangci-lint `go.mod` tool directive, CI lint stage, Makefile | `make lint` clean, `make verify` green |
| Time, clocks, timeouts, scheduling | [foundations/time.md](foundations/time.md) | injectable Clock seam, `internal/testutil` fake clock, callers reaching for `time.Now()` | deterministic tests with a fake clock, `make race` clean |
| Domain type modeling: enums, typed IDs, optional/zero values, slices/maps | [foundations/data-modeling.md](foundations/data-modeling.md) | the wire shape (serialization), DB null mapping, validation | round-trip + zero-value tests; `make race` for shared-slice aliasing |
| Serialization / JSON wire shape: tags, `omitzero`, DTOs, decode | [foundations/serialization.md](foundations/serialization.md) | transport DTOs, contract compatibility, golden tests | round-trip + golden tests; decoder handles unknown fields per policy |
| Non-obvious or hard-to-reverse architecture decision | [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md) | the project's `decisions/` directory, any superseded ADR, README/onboarding links | ADR recorded with status, alternatives, and consequences before the change merges |
| Project ownership transfer or onboarding | [onboarding-and-handoff.md](onboarding-and-handoff.md), [checklists/handoff.md](checklists/handoff.md) | CODEOWNERS, secrets/rotation docs, on-call, deploy access, open decisions | handoff checklist complete; new owner runs `make verify` and a deploy dry-run unaided |
| New-repo scaffolding (Makefile, CI, main, configs, project docs) | the [reference/](reference/) tree — [exampleservice](reference/exampleservice/) (HTTP+Postgres), [examplegrpc](reference/examplegrpc/) (gRPC), [exampleworker](reference/exampleworker/) (event-driven) — plus [templates/](templates/), [checklists/new-project.md](checklists/new-project.md) | the copied artifacts and their `<placeholder>` values | `make verify` green on the fresh repo |
| Outbound resilience: timeouts, retries, backoff, breakers, rate limits | [operations/resilience.md](operations/resilience.md) | external clients, eventing consumers, telemetry for retries/breakers | load test shows graceful shedding; a timeout on every external call |
| Containerization, Dockerfile, runtime limits, rollout | [operations/deployment.md](operations/deployment.md), [templates/Dockerfile](templates/Dockerfile) | CI image build, `GOMEMLIMIT`/`GOMAXPROCS`, probes, shutdown grace period | image builds static + nonroot; probes wired; `make verify` green |
| SLOs, alerting, runbooks, on-call | [operations/operability.md](operations/operability.md), [templates/runbook.md](templates/runbook.md) | dashboards/alerts (repo-specific), error budgets, on-call docs | each service has SLOs, symptom alerts, and a current runbook |
| Caching layer, invalidation, cache keys, stampede control | [services/caching.md](services/caching.md) | callers, telemetry for hit/miss, invalidation paths, config keys | cache hit/miss metrics; correctness under invalidation; no stale-read regressions |
| Feature flags, gating, rollout toggles | [foundations/configuration.md](foundations/configuration.md) (### Feature Flags) | config loader, default state, cleanup of stale flags, tests | flag default and override tested; stale flags scheduled for removal |
| Git workflow, branching, commits, CHANGELOG | [foundations/git-workflow.md](foundations/git-workflow.md), [operations/ci-and-release.md](operations/ci-and-release.md) (### Changelog Policy) | CHANGELOG, release notes, PR template | commit and branch conventions followed; changelog entry present where required |
| Tagging or publishing a library/module version | [recipes/release-library-version.md](recipes/release-library-version.md), [checklists/release.md](checklists/release.md) | CHANGELOG, semver tag, compatibility review, consumers | `v1.2.3` tag is canonical; compatibility honored; `make verify` green |
| Reusable shared package (`internal/runtime`, `internal/telemetry`, `internal/httputil`, `internal/buildinfo`, `internal/testutil`) | [foundations/shared-constructs.md](foundations/shared-constructs.md) | the consuming packages and `main` wiring | each helper states exactly what it owns; `main` reads top-to-bottom; no junk-drawer growth |

## High-Value Boundaries

- `cmd/` owns process startup and shutdown; it should not hold business rules.
- `internal/core/` owns domain behavior and interfaces consumed from the outside.
- `internal/api/http` and `internal/api/grpc` translate transport concerns into core calls.
- `internal/db` owns SQL, migrations, transaction helpers, and storage-specific mapping.
- `internal/config` owns loading, defaults, validation, and startup failure behavior.
- `internal/telemetry` owns reusable logging, metrics, trace, and health primitives.
- `api/` owns published wire-contract definitions such as `.proto` or OpenAPI sources when the repo uses them.
- eventing packages own broker interaction, settlement, retry policy, and payload mapping rather than scattering those concerns through handlers or `main`.

## Proof Hints

- Transport changes usually need both narrow handler tests and one manual or scripted smoke test.
- Persistence changes usually need real database integration tests; mocks are not enough.
- New background work usually needs race detection plus shutdown verification.
- Config changes are not done until invalid input fails fast and documented keys stay in sync.
- Dependency changes are not done until the rationale is explicit and the lockfile diff is understood.
