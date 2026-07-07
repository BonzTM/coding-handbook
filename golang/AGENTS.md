# AGENTS.md - Go Project Contract

This is the authoritative fast-path contract for autonomous agents working in a new Go repository.
Read this file first, then use [maintainer-map.md](maintainer-map.md) when you know the change type but not the file set, and [maintainer-reference.md](maintainer-reference.md) when you need slower-path architecture and rationale.

## Purpose

- Use this file for repo-wide invariants, change defaults, and the verification bar.
- Use [maintainer-map.md](maintainer-map.md) to route HTTP, gRPC, DB, config, CLI, worker, and release changes.
- Use [maintainer-reference.md](maintainer-reference.md) for package maps, lifecycle guidance, test taxonomy, and troubleshooting.

## Source Of Truth

- This file is the fast path. More detailed docs refine it; they do not weaken it.
- Layout, module, and toolchain defaults live in [foundations/project-setup.md](foundations/project-setup.md).
- Package and dependency boundaries live in [foundations/package-design.md](foundations/package-design.md) and [decisions/framework-selection.md](decisions/framework-selection.md); reusable shared packages (`internal/runtime`, `internal/telemetry`, `internal/httputil`, `internal/buildinfo`, `internal/testutil`) live in [foundations/shared-constructs.md](foundations/shared-constructs.md).
- Schema, API, and data-boundary rules live in [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md); type modeling in [foundations/data-modeling.md](foundations/data-modeling.md); the JSON/wire boundary in [foundations/serialization.md](foundations/serialization.md).
- Runtime correctness rules live in [foundations/context-and-concurrency.md](foundations/context-and-concurrency.md), [foundations/errors-and-logging.md](foundations/errors-and-logging.md), and [foundations/configuration.md](foundations/configuration.md).
- Proof expectations live in [quality/testing.md](quality/testing.md) and the relevant service or operations docs.
- Lint policy lives in [quality/linting.md](quality/linting.md); time and clock discipline in [foundations/time.md](foundations/time.md); copy-paste scaffolding in [templates/](templates/).
- Architecture decisions and their rationale live in [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md); project ownership transfer in [onboarding-and-handoff.md](onboarding-and-handoff.md).
- Shared vocabulary lives in [glossary.md](glossary.md); how to change this handbook lives in [CONTRIBUTING.md](CONTRIBUTING.md).
- Complete, compiling worked examples live under `reference/` (all `make verify`-green): [reference/exampleservice/](reference/exampleservice/) embodies the HTTP+Postgres shape, [reference/examplegrpc/](reference/examplegrpc/) the gRPC shape, and [reference/exampleworker/](reference/exampleworker/) the event-driven worker shape; mirror the one matching your shape rather than inventing a structure.

## Fast Path

1. Read this file and identify the project shape from [README.md](README.md).
2. Route the change through [maintainer-map.md](maintainer-map.md); do not guess where code belongs.
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

| If you are changing... | Read first |
|---|---|
| project layout, module shape, build flags | [foundations/project-setup.md](foundations/project-setup.md) |
| package boundaries, interfaces, exports | [foundations/package-design.md](foundations/package-design.md) |
| schemas, APIs, compatibility, and migrations | [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) |
| code review heuristics, comments, zero values | [foundations/style-and-review.md](foundations/style-and-review.md) |
| context, cancellation, workers, goroutines | [foundations/context-and-concurrency.md](foundations/context-and-concurrency.md) |
| errors, status mapping, structured logging | [foundations/errors-and-logging.md](foundations/errors-and-logging.md) |
| config loading, env vars, startup validation | [foundations/configuration.md](foundations/configuration.md) |
| event or message producers, consumers, retries, DLQs | [services/eventing-and-messaging.md](services/eventing-and-messaging.md) |
| HTTP handlers or middleware | [services/http-services.md](services/http-services.md) |
| gRPC APIs, protos, interceptors | [services/grpc-services.md](services/grpc-services.md) |
| queries, migrations, transactions | [services/database.md](services/database.md) |
| telemetry, `/metrics`, tracing, readiness | [operations/observability.md](operations/observability.md) |
| auth, secrets, file paths, supply chain | [operations/security.md](operations/security.md) |
| audit logging of security-relevant actions | [operations/security.md](operations/security.md) (### Audit Logging) |
| data classification, PII, retention, compliance | [operations/data-handling.md](operations/data-handling.md) |
| pre-build WHAT decisions (scope, acceptance, spec intake) | [checklists/spec-intake.md](checklists/spec-intake.md) |
| idempotent HTTP writes (Idempotency-Key, safe retries) | [recipes/add-idempotent-write.md](recipes/add-idempotent-write.md) |
| CI, build, release, containers | [operations/ci-and-release.md](operations/ci-and-release.md) |
| new dependency or framework choice | [decisions/framework-selection.md](decisions/framework-selection.md) |
| lint rules, golangci-lint config, formatters | [quality/linting.md](quality/linting.md) |
| time, clocks, timeouts, scheduling correctness | [foundations/time.md](foundations/time.md) |
| enums, typed IDs, optional fields, zero values, slice/map modeling | [foundations/data-modeling.md](foundations/data-modeling.md) |
| JSON tags, wire DTOs, encoding/decoding, `omitzero` | [foundations/serialization.md](foundations/serialization.md) |
| a non-obvious or hard-to-reverse architecture decision | [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md) |
| taking over or handing off a project | [onboarding-and-handoff.md](onboarding-and-handoff.md) |
| bootstrapping scaffolding (Makefile, CI, main, configs) | [templates/](templates/) |
| retries, backoff, circuit breakers, rate limiting, load shedding | [operations/resilience.md](operations/resilience.md) |
| containerization, Dockerfile, runtime limits, deployment | [operations/deployment.md](operations/deployment.md) |
| SLOs, alerting posture, runbooks, on-call | [operations/operability.md](operations/operability.md) |
| caching layers, invalidation, cache keys, stampede control | [services/caching.md](services/caching.md) |
| feature flags, gating, rollout toggles | [foundations/configuration.md](foundations/configuration.md) (### Feature Flags) |
| git workflow, branching, commits, CHANGELOG | [foundations/git-workflow.md](foundations/git-workflow.md) |
| tagging and publishing a library or module version | [recipes/release-library-version.md](recipes/release-library-version.md) |

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
- Detailed routing: [maintainer-map.md](maintainer-map.md)
- Startup and module layout: [foundations/project-setup.md](foundations/project-setup.md)
- Runtime correctness: [foundations/context-and-concurrency.md](foundations/context-and-concurrency.md), [foundations/errors-and-logging.md](foundations/errors-and-logging.md), [foundations/configuration.md](foundations/configuration.md)
- Proof and verification: [quality/testing.md](quality/testing.md)
