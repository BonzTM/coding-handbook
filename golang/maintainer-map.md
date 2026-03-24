# Maintainer Map

Purpose: route maintainers and agents to the right files and related obligations for common Go changes.
Read [AGENTS.md](AGENTS.md) first. Use this file when you know what kind of change you are making but do not want to rediscover the normal file set from scratch.

## Documentation Policy

- Keep [AGENTS.md](AGENTS.md) as the fast path only.
- Keep cross-cutting routing information here instead of scattering it through package docs.
- Document sync surfaces and proof steps that code search does not reveal cheaply.
- If a fact is obvious from the package tree alone, prefer search over prose.

## Change Map

| Change Type | Start Here | Also Update | Verify Or Confirm |
|---|---|---|---|
| Module path, repo layout, `go` or `toolchain` lines | `go.mod`, `cmd/`, `internal/`, [foundations/project-setup.md](foundations/project-setup.md) | CI build scripts, `tools.go`, onboarding docs | `go mod tidy`, `go build -trimpath ./...` |
| HTTP or gRPC contract shape, schema, compatibility, generated stubs | `api/**`, transport package, [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md) | clients, generated code policy, release notes, tests | schema or generation check, compatibility review |
| Event or message producers, consumers, payload contracts | eventing package, `api/**`, [services/eventing-and-messaging.md](services/eventing-and-messaging.md) | outbox or inbox storage, telemetry, release notes, tests, DLQ policy | contract tests, replay/idempotency tests, integration against real broker path |
| HTTP endpoint, middleware, request or response shape | `internal/api/http/`, [services/http-services.md](services/http-services.md) | `internal/core/`, request validation, telemetry, route registration, tests | handler tests, smoke test, readiness behavior |
| gRPC proto or server method | `api/**`, `internal/api/grpc/`, [services/grpc-services.md](services/grpc-services.md) | generated code policy, interceptors, error mapping, docs | proto lint or generate check, service tests, `grpcurl` |
| Business logic or domain rules | `internal/core/`, [foundations/package-design.md](foundations/package-design.md) | package-local tests, transport adapters if contracts changed | targeted unit tests, relevant integration coverage |
| DB queries, schema, transaction behavior | `internal/db/`, migrations, [services/database.md](services/database.md) | `sqlc` inputs and outputs, repositories, callers, observability | migration apply test, real DB integration tests |
| Configuration keys or startup defaults | `internal/config/`, [foundations/configuration.md](foundations/configuration.md) | `.env.example`, startup wiring, docs, tests | config loader tests, startup smoke test |
| Background workers, schedulers, shutdown paths | worker package, [foundations/context-and-concurrency.md](foundations/context-and-concurrency.md) | `main`, supervisor wiring, telemetry, readiness semantics | `go test -race`, shutdown smoke test, leak check |
| External HTTP or gRPC client | `internal/client/`, [recipes/add-external-client.md](recipes/add-external-client.md) | interfaces in consuming package, retries/timeouts, telemetry, security checks | `httptest.Server` or test server integration, timeout tests |
| Logging, metrics, traces, health endpoints | `internal/telemetry/`, [operations/observability.md](operations/observability.md) | handlers, repositories, workers, dashboards or alerts if repo has them | `/metrics`, `/livez`, `/readyz`, log schema review |
| Security-sensitive boundary | boundary package, [operations/security.md](operations/security.md) | validation, auth seams, secret handling, release docs if exposure changed | `govulncheck`, targeted negative tests, path or SSRF checks |
| CLI commands or flags | `cmd/<app>/`, [recipes/add-cli-command.md](recipes/add-cli-command.md) | config loading, help output, README if user-visible | command tests, `--help`, build and smoke run |
| Build, CI, containers, release automation | `.github/workflows/`, build scripts, Dockerfile, [operations/ci-and-release.md](operations/ci-and-release.md) | version injection, changelog, release packaging | CI dry run, `go version -m`, container smoke test |
| New dependency or framework | [decisions/framework-selection.md](decisions/framework-selection.md) | the caller package, proof docs, `tools.go` if tool-only | written rationale, dependency diff, proof it earns its cost |

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
