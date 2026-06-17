# Go Project Handbook

This handbook is the default engineering contract for new Go repositories. It is not a Go language tutorial. It exists to make services, workers, CLIs, and libraries converge on the same structure, runtime behavior, dependency posture, and proof of correctness.

## Start Here

- Humans: read this file, then follow the reading path for your project shape.
- Agents: read [AGENTS.md](AGENTS.md) first, then [maintainer-map.md](maintainer-map.md), then the relevant topical docs and recipes.
- Default assumptions unless a repo says otherwise:
  - one module per repo
  - `cmd/` plus `internal/` as the default layout
  - thin `main`
  - `net/http`, `database/sql`, `log/slog`, and the stdlib `testing` package first
  - env-driven configuration with fail-fast validation
  - `golangci-lint`, `go vet ./...`, `go test -race ./...`, and `go tool govulncheck ./...` as baseline proof, all wrapped by `make verify`

## Reading Paths

| If you are building... | Read in this order |
|---|---|
| HTTP service | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/package-design.md](foundations/package-design.md) -> [foundations/configuration.md](foundations/configuration.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [services/http-services.md](services/http-services.md) -> [operations/observability.md](operations/observability.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-http-endpoint.md](recipes/add-http-endpoint.md) |
| gRPC service | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/package-design.md](foundations/package-design.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [services/grpc-services.md](services/grpc-services.md) -> [foundations/errors-and-logging.md](foundations/errors-and-logging.md) -> [services/database.md](services/database.md) -> [operations/observability.md](operations/observability.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-grpc-method.md](recipes/add-grpc-method.md) |
| Background worker | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/context-and-concurrency.md](foundations/context-and-concurrency.md) -> [foundations/configuration.md](foundations/configuration.md) -> [operations/observability.md](operations/observability.md) -> [operations/security.md](operations/security.md) -> [recipes/add-background-worker.md](recipes/add-background-worker.md) |
| Event-driven service or async worker | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [services/eventing-and-messaging.md](services/eventing-and-messaging.md) -> [services/database.md](services/database.md) -> [operations/observability.md](operations/observability.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-event-publisher.md](recipes/add-event-publisher.md) -> [recipes/add-event-consumer.md](recipes/add-event-consumer.md) |
| CLI tool | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/style-and-review.md](foundations/style-and-review.md) -> [foundations/configuration.md](foundations/configuration.md) -> [decisions/framework-selection.md](decisions/framework-selection.md) -> [recipes/add-cli-command.md](recipes/add-cli-command.md) |
| Library | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/package-design.md](foundations/package-design.md) -> [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) -> [foundations/style-and-review.md](foundations/style-and-review.md) -> [quality/testing.md](quality/testing.md) -> [foundations/errors-and-logging.md](foundations/errors-and-logging.md) -> [checklists/release.md](checklists/release.md) -> [recipes/release-library-version.md](recipes/release-library-version.md) |

Every shape also adopts [quality/linting.md](quality/linting.md) and the committed [templates/](templates/) scaffolding, runs `make verify` as its proof gate, and follows [foundations/data-modeling.md](foundations/data-modeling.md), [foundations/serialization.md](foundations/serialization.md), and [foundations/time.md](foundations/time.md) for type, wire-shape, and clock decisions. Networked services additionally follow [operations/resilience.md](operations/resilience.md), [operations/deployment.md](operations/deployment.md), and [operations/operability.md](operations/operability.md).

Cross-cutting concerns apply across shapes: [services/caching.md](services/caching.md) and [foundations/configuration.md](foundations/configuration.md) (feature flags) affect most services, and [foundations/git-workflow.md](foundations/git-workflow.md) governs commits, branches, and changelog discipline everywhere.

## Non-Negotiables

- Default to `internal/`; use `pkg/` only when the repo intentionally exports a library surface for other modules.
- Keep `main` boring. It wires config, logging, dependencies, signals, and shutdown; it does not hold business logic.
- Pass `context.Context` explicitly as the first argument for I/O and long-running work. Never store it in a struct.
- Use `%w`, `errors.Is`, and `errors.As`; do not compare error strings.
- Use `log/slog` for structured logs. Log once at the boundary that can act.
- Use real integration tests for database and external-service boundaries; do not mock everything by default.
- Keep metric labels low-cardinality; request IDs and user IDs never belong in metrics.
- Build pure-Go by default with `CGO_ENABLED=0`; a cgo dependency requires an ADR.
- Do not commit `replace` directives, real secrets, or framework-heavy defaults without explicit justification.

## Default Stack

| Concern | Default | Reach for something else when |
|---|---|---|
| HTTP routing | `net/http` with Go 1.22+ `ServeMux` | routing, middleware composition, or backwards compatibility needs justify a router like `chi` |
| Database access | `database/sql` | the repo does not persist data, or an approved driver-specific abstraction is required |
| Query authoring | hand-written SQL, then `sqlc` when SQL volume grows | the storage shape is trivial and `database/sql` alone remains clear |
| Logging | `log/slog` | almost never; wrappers are fine, replacing the core logger rarely is |
| Metrics | Prometheus client | the org uses a different required metrics backend |
| Tracing | OpenTelemetry when services are distributed or latency-sensitive | a local CLI or small library does not need trace infrastructure |
| CLI parsing | stdlib `flag` | nested subcommands, completions, or manpage generation justify `cobra` |
| Config loading | explicit env plus flags loader in `internal/config` | a repo-specific operator workflow requires a documented config file |

## Handbook Map

- [AGENTS.md](AGENTS.md) - fast-path contract for autonomous agents and reviewers
- [maintainer-map.md](maintainer-map.md) - change routing and sync surfaces
- [maintainer-reference.md](maintainer-reference.md) - architecture, rationale, and deeper guidance
- [onboarding-and-handoff.md](onboarding-and-handoff.md) - day-one reading path for a new owner and the outgoing-owner handoff
- [glossary.md](glossary.md) - shared vocabulary for handbook terms and conventions
- [CONTRIBUTING.md](CONTRIBUTING.md) - how to propose and make changes to this handbook
- `foundations/` - package layout, contracts, idioms, config, errors, concurrency, time, data modeling, serialization, [shared constructs](foundations/shared-constructs.md), and [git-workflow.md](foundations/git-workflow.md) (commits, branches, changelog)
- `quality/` - test strategy, linting, fuzzing, benchmarks, race detection, and proof commands
- `services/` - transport, eventing, persistence, and [caching.md](services/caching.md) guidance for HTTP, gRPC, messaging, database, and cache work
- `operations/` - telemetry, security, audit logging, resilience, deployment, operability/SLOs, data handling/PII, CI, releases, and runtime expectations
- `decisions/` ([README.md](decisions/README.md)) - architecture decision records (ADRs) plus dependency and framework selection rules
- `checklists/` ([README.md](checklists/README.md)) and `recipes/` ([README.md](recipes/README.md)) - executable startup, review, release, handoff, and implementation guidance
- `templates/` - committed copy-paste scaffolding (Makefile, `.golangci.yml`, `main.go`, project README/AGENTS/CODEOWNERS, ADR template)
- `reference/exampleservice/` - a complete, compiling, `make verify`-green service that composes every handbook pattern (copy it to bootstrap a new repo)

## What This Handbook Optimizes For

- code that still looks obvious six months later
- boundaries that make testing and refactoring cheaper
- runtime behavior that is safe under load and easy to debug
- defaults that keep agents from inventing new architecture every task
- minimal dependency surface unless there is a clear return on complexity

## Where To Go Next

- New repo bootstrap: [checklists/new-project.md](checklists/new-project.md)
- Active agent work: [AGENTS.md](AGENTS.md)
- Routing a change quickly: [maintainer-map.md](maintainer-map.md)
- Choosing third-party libraries: [decisions/framework-selection.md](decisions/framework-selection.md)
- Recording an architecture decision: [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md)
- Copy-paste scaffolding for a new repo: [templates/](templates/)
- Taking over or handing off a project: [onboarding-and-handoff.md](onboarding-and-handoff.md)
- Lint policy and configuration: [quality/linting.md](quality/linting.md)
- Making a networked service production-ready: [operations/resilience.md](operations/resilience.md), [operations/deployment.md](operations/deployment.md), [operations/operability.md](operations/operability.md)
- Implementing a specific change step by step: [recipes/README.md](recipes/README.md)
- Looking up a handbook term: [glossary.md](glossary.md)
- Changing the handbook itself, commits, or the changelog: [CONTRIBUTING.md](CONTRIBUTING.md), [foundations/git-workflow.md](foundations/git-workflow.md)
- A complete worked example to copy: [reference/exampleservice/](reference/exampleservice/) (`make verify`-green)
