# CI and Release

Delivery guidance for Go repos that should build the same way locally, in CI, and at release time.

## Default Approach

Every PR should run a predictable baseline pipeline.

| Stage | Commands | Purpose |
|---|---|---|
| module hygiene | `go mod tidy`, `go mod verify` | clean module graph |
| formatting | `gofmt -s -l .` | consistent source shape |
| static analysis | `go vet ./...`, optionally `staticcheck ./...` | catch correctness issues early |
| tests | `go test ./...`, `go test -race ./...` | functional and concurrency confidence |
| security | `govulncheck ./...` | supply-chain and known-vulnerability check |
| build | `go build -trimpath ./...` | compile and link proof |

Tailor the matrix to the repo, but do not quietly drop the safety stages because they are slow.

## Release Defaults

- tag or label releases with semantic versions in plain `1.2.3` form if the repo publishes stable artifacts
- inject build metadata through linker flags or a small `internal/buildinfo` package
- use multi-stage container builds for service artifacts
- add `default.pgo` only after collecting representative production profiles and deciding it materially helps
- introduce GoReleaser only when release complexity justifies it, not on day one by reflex

## Compatibility And Rollback

- API, schema, and event changes need an explicit backward-compatibility story.
- Migrations should be deploy-safe: avoid patterns that require every instance to stop at once unless the rollout plan says so.
- Release notes should call out operator-visible changes such as new env vars, port changes, migrations, or default timeout changes.
- Message contract changes should also call out replay expectations, DLQ handling, and whether mixed-version consumers or producers are safe during rollout.

## Common Mistakes And Forbidden Patterns

- skipping `-race` in CI for code that uses concurrency
- builds that require undeclared local tools or shell state
- manual release steps that drift from CI reality
- opaque version output that prevents operators from identifying a running binary
- cargo-cult PGO or container hardening that nobody has actually tested

## Verification And Proof

- latest CI run on the release commit is green
- `go version -m <binary>` or the repo's equivalent shows expected module metadata
- release artifact starts successfully and reports its version/build info
- deploy-safe changes have a documented rollback or compatibility plan
