# AGENTS.md - Go Project Contract

This is the authoritative fast-path contract for autonomous agents working in a new Go repository.
Read this file first: it carries the repo-wide invariants, the change-routing table, and the verification bar. Use [maintainer-reference.md](maintainer-reference.md) when you need slower-path architecture and rationale.

## Purpose

- Use this file for repo-wide invariants, change defaults, change-to-file routing, and the verification bar.
- Use [maintainer-reference.md](maintainer-reference.md) for package maps, lifecycle guidance, test taxonomy, and troubleshooting.
- For the full catalogs, see the [recipes/README.md](recipes/README.md), [checklists/README.md](checklists/README.md), and [decisions/README.md](decisions/README.md) indexes.

## Source Of Truth

- This file is the fast path. More detailed docs refine it; they do not weaken it.
- Layout, module, and toolchain defaults live in [foundations/project-setup.md](foundations/project-setup.md).
- Package and dependency boundaries live in [foundations/package-design.md](foundations/package-design.md) and [decisions/framework-selection.md](decisions/framework-selection.md); reusable shared packages (`internal/runtime`, `internal/telemetry`, `internal/httputil`, `internal/buildinfo`, `internal/testutil`) live in [foundations/shared-constructs.md](foundations/shared-constructs.md).
- Schema, API, and data-boundary rules live in [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md); type modeling in [foundations/data-modeling.md](foundations/data-modeling.md); the JSON/wire boundary in [foundations/serialization.md](foundations/serialization.md).
- Runtime correctness rules live in [foundations/context-and-concurrency.md](foundations/context-and-concurrency.md), [foundations/errors-and-logging.md](foundations/errors-and-logging.md), and [foundations/configuration.md](foundations/configuration.md).
- Proof expectations live in [quality/testing.md](quality/testing.md) and the relevant service or operations docs.
- Lint policy lives in [quality/linting.md](quality/linting.md); time and clock discipline in [foundations/time.md](foundations/time.md); copy-paste scaffolding in [templates/](templates/).
- Architecture decisions and their rationale live in [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).
- Team-process docs — [onboarding-and-handoff.md](onboarding-and-handoff.md), [CONTRIBUTING.md](CONTRIBUTING.md), [checklists/incident-response.md](checklists/incident-response.md), and the [glossary.md](glossary.md) lookup aid — serve humans running the team and the handbook; they are not needed to build or change an app.
- Complete, compiling worked examples live under `reference/` (all `make verify`-green): [reference/exampleservice/](reference/exampleservice/) embodies the HTTP+Postgres shape, [reference/examplegrpc/](reference/examplegrpc/) the gRPC shape, and [reference/exampleworker/](reference/exampleworker/) the event-driven worker shape; mirror the one matching your shape rather than inventing a structure.

## Fast Path

1. Read this file and identify the project shape from [README.md](README.md). For a brand-new build, run [checklists/spec-intake.md](checklists/spec-intake.md) first — it says which WHAT decisions to take from the spec, which to ask about, and which defaults apply when the spec is silent.
2. Route the change through the [Change Routing](#change-routing) table below; do not guess where code belongs.
3. Read the relevant foundations doc before editing code in a new area.
4. Implement with the repo defaults unless the repo has already documented an exception.
5. Prove the change with the narrowest meaningful tests first, then the repo-wide baseline.

## Repo-Wide Invariants

- **Single module**: start with one `go.mod` in the repo root. Nested modules require explicit architectural review.
- **Official layout**: follow `cmd/` plus `internal/` per `go.dev/doc/modules/layout` (which prescribes that baseline and `internal/`, but does not define `pkg/`). Use `pkg/` only for an intentionally public library surface — a convention this handbook adopts, not part of the official layout doc.
- **Thin main**: `main.go` wires config, logger, dependencies, signal handling, and shutdown. Business logic belongs in `internal/`.
- **Context discipline**: `ctx context.Context` is the first parameter for I/O and long-running work. Never store it in a struct or use it as an optional bag of dependencies.
- **Errors**: wrap with `%w`, inspect with `errors.Is` and `errors.As`, and log once at the boundary that can act.
- **Logging**: use `log/slog`; do not introduce global loggers into reusable packages.
- **Persistence**: default to `database/sql`; use `sqlc` when query count or complexity makes manual scanning noisy.
- **Testing**: every behavior change needs tests. DB and external boundaries need real integration coverage, not only mocks.
- **Observability**: new networked behavior adds logs, metrics, and health/readiness behavior where appropriate.
- **Pure-Go default**: build with `CGO_ENABLED=0`; a cgo dependency requires an ADR.
- **Dependency posture**: stdlib first, explicit rationale for every non-trivial dependency, no committed `replace` directives.

## Change Routing

Use this when you know what kind of change you are making but not the file set. Start Here is what you read and touch first; Also Update is the sync surface the change normally drags along; Verify Or Confirm is the proof.

| Change Type | Start Here | Also Update | Verify Or Confirm |
|---|---|---|---|
| Module path, repo layout, `go` or `toolchain` lines | `go.mod`, `cmd/`, `internal/`, [foundations/project-setup.md](foundations/project-setup.md) | CI build scripts, `go.mod` `tool` directives, onboarding docs | `go mod tidy`, `go build ./...` |
| HTTP or gRPC contract shape, schema, compatibility, generated stubs | `api/**`, transport package, [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md), [recipes/deprecate-and-remove-contract.md](recipes/deprecate-and-remove-contract.md) | clients, generated code policy, release notes, tests | schema or generation check, compatibility review |
| Event or message producers, consumers, payload contracts | eventing package, `api/**`, [services/eventing-and-messaging.md](services/eventing-and-messaging.md), [reference/exampleworker/](reference/exampleworker/) | outbox or inbox storage, telemetry, release notes, tests, DLQ policy | contract tests, replay/idempotency tests, integration against real broker path |
| HTTP endpoint, middleware, request or response shape | `internal/api/http/`, [services/http-services.md](services/http-services.md), [recipes/add-http-endpoint.md](recipes/add-http-endpoint.md), [recipes/add-http-middleware.md](recipes/add-http-middleware.md) | `internal/core/`, request validation, telemetry, route registration, tests | handler tests, smoke test, readiness behavior |
| Server-rendered pages, templates, static assets, sessions, CSRF | `internal/api/web/`, [services/web-apps.md](services/web-apps.md), [services/http-services.md](services/http-services.md) | template set and `embed.FS`, session store, security headers, flash/PRG flows, tests | template-parse test, CSRF negative test, session cookie-flag assertions, XSS probe |
| gRPC proto or server method | `api/**`, `internal/api/grpc/`, [services/grpc-services.md](services/grpc-services.md), [recipes/add-grpc-method.md](recipes/add-grpc-method.md), [reference/examplegrpc/](reference/examplegrpc/) | generated code policy, interceptors, error mapping, docs | proto lint or generate check, service tests, `grpcurl` |
| Business logic or domain rules | `internal/core/`, [foundations/package-design.md](foundations/package-design.md) | package-local tests, transport adapters if contracts changed | targeted unit tests, relevant integration coverage |
| Errors, status mapping, structured logging | [foundations/errors-and-logging.md](foundations/errors-and-logging.md) | error envelope, transport error mapping, log fields at the acting boundary | error-mapping tests; log-once review at the boundary |
| Code review heuristics, comments, naming, zero values | [foundations/style-and-review.md](foundations/style-and-review.md) | godoc comments, exported-contract docs, review checklist | `make fmt-check`, `go doc ./...` reads as a contract |
| DB queries, schema, transaction behavior | `internal/db/`, migrations, [services/database.md](services/database.md), [recipes/add-database-feature.md](recipes/add-database-feature.md), [recipes/add-migration.md](recipes/add-migration.md) | `sqlc` inputs and outputs, repositories, callers, observability | migration apply test, real DB integration tests |
| Configuration keys or startup defaults | `internal/config/`, [foundations/configuration.md](foundations/configuration.md), [recipes/add-config-key.md](recipes/add-config-key.md) | `.env.example`, startup wiring, docs, tests | config loader tests, startup smoke test |
| Background workers, schedulers, shutdown paths | worker package, [foundations/context-and-concurrency.md](foundations/context-and-concurrency.md), [recipes/add-background-worker.md](recipes/add-background-worker.md), [recipes/add-scheduled-job.md](recipes/add-scheduled-job.md), [reference/exampleworker/](reference/exampleworker/) | `main`, supervisor wiring, telemetry, readiness semantics | `go test -race`, shutdown smoke test, leak check |
| External HTTP or gRPC client | `internal/client/`, [recipes/add-external-client.md](recipes/add-external-client.md) | interfaces in consuming package, retries/timeouts, telemetry, security checks | `httptest.Server` or test server integration, timeout tests |
| Logging, metrics, traces, health endpoints | `internal/telemetry/`, [operations/observability.md](operations/observability.md), [recipes/add-metric.md](recipes/add-metric.md) | handlers, repositories, workers, dashboards or alerts if repo has them | `/metrics`, `/livez`, `/readyz`, log schema review |
| Security-sensitive boundary | boundary package, [operations/security.md](operations/security.md) | validation, auth seams, secret handling, release docs if exposure changed | `govulncheck`, targeted negative tests, path or SSRF checks |
| Audit logging of security-relevant actions | [operations/security.md](operations/security.md) (### Audit Logging) | audit sink/retention, actor and resource fields, telemetry, access controls on the log | audit events emitted at the action boundary; tamper-evidence and retention honored |
| Data classification, PII, retention, compliance | [operations/data-handling.md](operations/data-handling.md) | field-level classification, log/metric redaction, retention jobs, export/delete paths | PII never in logs/metrics; retention and deletion paths exercised by tests |
| Pre-build WHAT decisions (scope, acceptance, spec intake) | [checklists/spec-intake.md](checklists/spec-intake.md) | the task's acceptance criteria, open questions or defaulted assumptions, downstream design docs | spec intake resolved (answered, asked, or defaulted-and-disclosed) before code; acceptance criteria agreed and testable |
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
- `internal/api/http`, `internal/api/web`, and `internal/api/grpc` translate transport concerns into core calls.
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

## Working Norms

- Prefer small, reviewable changes over broad cleanup.
- Do not introduce new architecture because it feels cleaner; match the repo's current shape unless the task is explicitly architectural.
- Do not bypass boundaries: handlers do not query the database directly, repositories do not contain transport logic, and `main` does not absorb business rules.
- When adding a dependency, document why the stdlib or an existing dependency is insufficient.
- When behavior changes, write the failing or proving test before claiming success whenever practical.
- If verification fails, fix it or report it clearly. Do not claim the change is done.

## Baseline Verification

| Goal | Command | Expectation |
|---|---|---|
| module hygiene | `go mod tidy -diff` + `go mod verify` (via `make tidy-check`) | no diff; committed module files are tidy |
| format | `go tool golangci-lint fmt --diff` (via `make fmt-check`; gofumpt + gci) | no diff output |
| static analysis | `go vet ./...` | exit code 0 |
| lint | `go tool golangci-lint run` | exit code 0 (policy in [quality/linting.md](quality/linting.md)) |
| type and compile safety | `go build ./...` | exit code 0 |
| concurrency safety | `go test -race ./...` | all relevant packages pass |
| supply-chain check | `go tool govulncheck ./...` | no blocking findings |
| file-specific correctness | targeted tests from the relevant recipe or service doc | pass with expected assertions |

The committed [templates/Makefile](templates/Makefile) wraps this full gate as `make verify`; run it so local and CI stay identical. `staticcheck` already runs inside `golangci-lint`. Use broader tools such as containerized integration suites, fuzzing, or benchmark checks when the repo or task calls for them.

## Slow Path Docs

- Architecture and package map: [maintainer-reference.md](maintainer-reference.md)
- Startup and module layout: [foundations/project-setup.md](foundations/project-setup.md)
- Runtime correctness: [foundations/context-and-concurrency.md](foundations/context-and-concurrency.md), [foundations/errors-and-logging.md](foundations/errors-and-logging.md), [foundations/configuration.md](foundations/configuration.md)
- Proof and verification: [quality/testing.md](quality/testing.md)
