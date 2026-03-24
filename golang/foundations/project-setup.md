# Project Setup

Default repository shape, module policy, and bootstrap expectations for new Go projects.

## Default Approach

Follow the official module layout guidance from `go.dev/doc/modules/layout` and start with a single module.

### Bootstrap Commands

```bash
go mod init github.com/org/repo
mkdir -p cmd/app internal/{api/http,core,db,config,telemetry,runtime,testutil} api scripts
```

### Preferred Tree

```text
repo/
  go.mod
  go.sum
  tools.go
  cmd/
    app/
      main.go
  internal/
    api/
      http/
      grpc/
    config/
    core/
    db/
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
- `internal/config`: config structs, defaults, env and flag loading, validation
- `internal/runtime`: assembly helpers that keep `main` thin
- `internal/telemetry`: logger, metrics, tracing, health helpers
- `internal/testutil`: test-only builders and fixtures that reduce duplication without hiding behavior
- `api/`: `.proto`, OpenAPI, or other wire-contract definitions

If a repo publishes external APIs or schemas, `api/` should hold the authoritative contract sources rather than generated outputs alone.

## Toolchain And Module Policy

- Set the `go` line in `go.mod` to the minimum supported language version for the repo.
- Use the `toolchain` line when you want reproducible local and CI behavior across contributors.
- Stay current on supported patch releases; do not let a new repo start on stale Go versions.
- Default to one module. Split modules only when a package has a genuinely separate release cadence or import surface.

## Tool Dependencies

Track tool-only dependencies in `tools.go` so CI and local environments resolve the same versions.

```go
//go:build tools

package tools

import (
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
	_ "golang.org/x/vuln/cmd/govulncheck"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
```

Use `tools.go` for CLI tools only. Do not smuggle runtime dependencies into it.

## Build Defaults

- Use `-trimpath` for normal builds.
- Keep `-buildvcs` enabled for release artifacts unless reproducibility requirements force a different mode.
- Consider `default.pgo` only after collecting representative profiles; do not cargo-cult PGO into day-one repos.
- `CGO_ENABLED=0` is a good default when the repo does not need cgo-backed libraries.

## Common Mistakes And Forbidden Patterns

- Nested modules without a release-boundary reason.
- A catch-all `pkg/` that quietly becomes the real application.
- A root `main.go` that grows into the service instead of delegating to `internal/`.
- Committed `replace` directives used as a substitute for proper dependency versioning.
- Build scripts that require undeclared local tools or hidden shell state.

## Verification And Proof

```bash
go mod tidy
go mod verify
go build -trimpath ./...
go test ./...
```

Proof is complete when the module graph is clean, all binaries build, and the package tree matches the intended boundaries rather than a temporary prototype layout.
