<!--
TEMPLATE -> repo root AGENTS.md
Fill every <PLACEHOLDER>. This is the per-PROJECT fast path. It mirrors the Go
handbook's AGENTS.md two-speed model but is scoped to one repo and points back to
the handbook for the full rules. Keep it short; do not restate handbook prose here.
-->

# AGENTS.md - <project-name> Contract

Fast-path contract for autonomous agents and reviewers working in this repository.
Read this file first. It is scoped to this project; the full engineering rules live
in the Go handbook (`<link or path to the Go handbook AGENTS.md>`). Where this file
is silent, the handbook governs. This file may add constraints but never weakens it.

## Purpose

- This project is a <service | worker | CLI | library>: <one-line purpose>.
- Use this file for project-specific invariants, change routing, and the verification bar.
- Use [README.md](README.md) for project shape, config keys, and how to run it.
- Use the Go handbook for the complete, cross-project rules and rationale.

## Repo-Wide Invariants

- **Single module**: one `go.mod` at the root, module `github.com/<org>/<repo>`, Go 1.24+.
- **Layout**: `cmd/` + `internal/` per the handbook; `pkg/` only for an intentional public surface.
- **Thin main**: `cmd/<app>/main.go` wires config, logger, dependencies, signals, shutdown. No business logic.
- **Context discipline**: `ctx context.Context` is the first parameter for I/O and long-running work; never stored in a struct.
- **Errors**: wrap with `%w`, inspect with `errors.Is`/`errors.As`, log once at the boundary that can act.
- **Logging**: `log/slog`; no global loggers in reusable packages.
- **Config**: loaded and validated in `internal/config`, fail-fast at startup; every key documented in [README.md](README.md) and `.env.example`.
- **Persistence**: `database/sql` (sqlc when scanning gets noisy); real integration tests at the DB boundary.
- **Testing**: every behavior change ships with a test.
- **Dependencies**: stdlib first, explicit rationale per non-trivial dependency, no committed `replace` directives.
- **Project-specific**: <invariants unique to this repo — domain rules, external system, SLAs — or "none beyond the above".>

## Change Routing

Route the change; do not guess where code belongs. Owners per area: [CODEOWNERS](.github/CODEOWNERS).

| If you are changing... | Start in | Read first |
|---|---|---|
| process startup, wiring, shutdown | `cmd/<app>/` | handbook `foundations/project-setup.md` |
| domain logic or rules | `internal/core/` | handbook `foundations/package-design.md` |
| HTTP handlers, middleware, routes | `internal/api/http/` | handbook `services/http-services.md` |
| gRPC methods, protos, interceptors | `internal/api/grpc/`, `api/` | handbook `services/grpc-services.md` |
| queries, schema, transactions | `internal/db/` | handbook `services/database.md` |
| config keys or startup defaults | `internal/config/`, `.env.example`, [README.md](README.md) | handbook `foundations/configuration.md` |
| logging, metrics, traces, health | `internal/telemetry/` | handbook `operations/observability.md` |
| auth, secrets, input boundaries | the boundary package | handbook `operations/security.md` |
| CI, build, release, containers | `.github/workflows/`, build scripts | handbook `operations/ci-and-release.md` |
| a new dependency or framework | the consuming package + `go.mod` | handbook `decisions/framework-selection.md` |
| <project-specific area> | `<path>` | `<doc>` |

## Working Norms

- Prefer small, reviewable changes; match the repo's current shape, do not invent new architecture.
- Do not bypass boundaries: handlers do not query the DB, repositories carry no transport logic, `main` holds no business rules.
- When adding a dependency, document why stdlib or an existing dependency is insufficient.
- Write the proving test before claiming success whenever practical.
- If verification fails, fix it or report it clearly. Do not claim the change is done.

## Baseline Verification

```bash
make verify
```

`make verify` is the single ordered gate (tidy, format check, lint, `go vet`,
tests, race tests, `govulncheck`, build). Run the narrowest meaningful tests first, then this gate.
A change is not done until `make verify` passes locally and in CI.
