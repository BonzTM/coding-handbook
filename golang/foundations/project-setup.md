# Project Setup

Default repository shape, module policy, and bootstrap expectations for new Go projects.

## Default Approach

Start with a single module and the official `cmd/` plus `internal/` baseline from `go.dev/doc/modules/layout`. The fuller tree below — the `internal/api`, `internal/core`, `internal/db`, `api/`, and `scripts/` split — is this handbook's own convention layered on that baseline, not prescribed by the official layout doc.

### Bootstrap Commands

```bash
go mod init github.com/org/repo
mkdir -p cmd/app internal/{api/http,core,db,httputil,config,telemetry,runtime,buildinfo,testutil} api scripts
```

### Preferred Tree

```text
repo/
  go.mod
  go.sum
  LICENSE
  cmd/
    app/
      main.go
  internal/
    api/
      http/
      grpc/
    buildinfo/
    config/
    core/
    db/
    httputil/
    runtime/
    telemetry/
    testutil/
  api/
  scripts/
```

### What Goes Where

- `cmd/<app>/main.go`: config loading, dependency injection, signal handling, process exit
- `internal/core`: domain logic plus interfaces consumed from the outside
- `internal/api/http` and `internal/api/grpc`: transport adapters only
- `internal/db`: repositories, queries, migrations, transaction helpers
- `internal/httputil`: JSON/error-response/request-size/timeout helpers shared across transport adapters; not a dumping ground
- `internal/config`: config structs, defaults, env and flag loading, validation
- `internal/runtime`: assembly helpers that keep `main` thin
- `internal/buildinfo`: version, commit, and build-time metadata stamped via `-ldflags` and surfaced in logs and `version` output
- `internal/telemetry`: logger, metrics, tracing, health helpers
- `internal/testutil`: test-only builders and fixtures that reduce duplication without hiding behavior
- `api/`: `.proto`, OpenAPI, or other wire-contract definitions

If a repo publishes external APIs or schemas, `api/` should hold the authoritative contract sources rather than generated outputs alone.

`internal/buildinfo` and `internal/httputil` are optional shared packages: add them when they earn their place (real build-metadata stamping, or transport helpers actually shared across adapters), not by reflex on day one. They are shown in the tree so the boundary is clear when you do reach for them, not to imply every repo must carry them.

A complete, compiling instance of this layout lives at [../reference/exampleservice/](../reference/exampleservice/). It is `make verify`-green (build, vet, lint, race tests, govulncheck) and is the fastest way to bootstrap a service: copy it, rename the module path, and replace the example domain.

## Toolchain And Module Policy

- Set the `go` line in `go.mod` to the minimum supported language version for the repo.
- Use the `toolchain` line when you want reproducible local and CI behavior across contributors.
- Stay current on supported patch releases; do not let a new repo start on stale Go versions.
- Default to one module. Split modules only when a package has a genuinely separate release cadence or import surface.

### License And Headers

- Every repo has a top-level `LICENSE` from day one. The license is a deliberate choice — the org default for internal/proprietary code, or a per-project OSI license for anything published — and a non-default pick is ADR-worthy (see [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).
- Default header policy: rely on the `LICENSE` file plus the module path for provenance; do not add per-file copyright headers. Add SPDX headers (`// SPDX-License-Identifier: <id>`) only when the org policy or a regulatory/compliance context requires them — then apply them uniformly and enforce them in `make verify`, not by hand.
- Review third-party license obligations when adding a dependency: confirm the license is compatible with how you ship, and capture the check in the dependency decision (see [../decisions/framework-selection.md](../decisions/framework-selection.md) and [../operations/security.md](../operations/security.md)). A copyleft or attribution-required dependency is a decision, not an accident discovered at release.

## Tool Dependencies

Track tool-only dependencies with `go.mod` `tool` directives (Go 1.24+) so CI and local environments resolve the same versions. Add a tool with `go get -tool`, run it with `go tool`, and upgrade the whole set with `go get tool`.

```bash
go get -tool golang.org/x/vuln/cmd/govulncheck
go get -tool honnef.co/go/tools/cmd/staticcheck
go get -tool github.com/sqlc-dev/sqlc/cmd/sqlc
go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint
```

Each command records a `tool` directive in `go.mod`; run them as `go tool govulncheck ./...`, `go tool staticcheck ./...`, and so on. The pre-1.24 `tools.go` blank-import pattern is obsolete — reach for it only when a module is pinned below `go 1.24`. Use tool directives for CLI tools only; do not smuggle runtime dependencies into them.

Pinning current tools as `tool` directives (golangci-lint, sqlc) can raise the module's `go` directive to match the highest `go` floor those tools require — the reference modules sit at `go 1.26.x` for this reason. That is expected and fine; the handbook's language baseline stays "1.24+", and the `go` line simply tracks the tool floor when tools are pinned.

## Build Defaults

- Use `-trimpath` for release and distribution builds so local filesystem paths are not embedded in the binary; it is unnecessary for the routine `go build ./...` compile check and forks the build cache.
- Keep `-buildvcs` enabled for release artifacts unless reproducibility requirements force a different mode.
- Consider `default.pgo` only after collecting representative profiles; do not cargo-cult PGO into day-one repos.
- Build pure-Go static binaries by default: `CGO_ENABLED=0`. Cgo is an exception that requires an ADR — it breaks easy cross-compilation, blocks `static`/distroless deployment, and enlarges the build and supply-chain surface.
- Prefer a pure-Go library over a cgo one when both exist (e.g. `modernc.org/sqlite` over `mattn/go-sqlite3`). The default Postgres driver `pgx` is already pure-Go.

## Common Mistakes And Forbidden Patterns

- Nested modules without a release-boundary reason.
- A catch-all `pkg/` that quietly becomes the real application.
- A root `main.go` that grows into the service instead of delegating to `internal/`.
- Committed `replace` directives used as a substitute for proper dependency versioning.
- Build scripts that require undeclared local tools or hidden shell state.
- No `LICENSE` file, or a license copied in by reflex without deciding whether it fits how the code ships.
- Per-file copyright headers added ad hoc when no policy requires them, or SPDX headers applied to some files and not others.
- Pulling in a dependency without checking its license obligations against how the binary is distributed.

## Verification And Proof

```bash
go mod tidy
go mod verify
go build ./...
go test ./...
ls LICENSE
```

Proof is complete when the module graph is clean, all binaries build, the package tree matches the intended boundaries rather than a temporary prototype layout, a deliberate `LICENSE` is present, and the header policy (none, or SPDX everywhere) is applied uniformly.
