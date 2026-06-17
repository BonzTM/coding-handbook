# Recipe: Bump A Dependency

Use this when you raise a direct dependency's version — routinely, to clear a vulnerability, or as part of a bot PR you are reviewing to merge.

## Files To Touch

- `go.mod`, `go.sum`
- the caller (any package in `internal/...`, `cmd/...`, or `api/...`) if the dependency's API changed
- tests that exercise the dependency's surface

## Steps

1. Scope the bump before touching anything: patch / minor / major, and security-driven vs routine. A patch or minor is usually mechanical; a major can change behavior or contracts and needs the same scrutiny as adopting a new dependency (re-answer the approval questions in [framework-selection.md](../decisions/framework-selection.md)). A security-driven bump follows the supply-chain guidance in [operations/security.md](../operations/security.md).
2. Read the upstream changelog or release notes for the target version, especially for a major. Note breaking changes, behavior changes, and any new minimum Go version it demands.
3. Pin the version: `go get <module>@<version>` (e.g. `go get example.com/lib@v1.7.2`), then `go mod tidy` to settle the transitive graph and `go.sum`.
4. Inspect the diff: `git diff go.mod go.sum`. Confirm the direct bump is what you intended and understand every transitive line that moved. Watch for an unexpected `go` directive or `toolchain` line bump — only accept it if it is intended.
5. Update the caller for any changed API and adjust tests. Do not paper over a breaking change with a `replace` directive (forbidden for production builds per [framework-selection.md](../decisions/framework-selection.md)).
6. Run the full gate (`make verify`) plus targeted tests for the dependency's surface.

## Invariants To Preserve

- every bump passes `make verify`, including `-race` and `govulncheck`, before merge — no exceptions for bot PRs
- a major version bump gets a written rationale; if it shifts behavior or a contract, capture it in an ADR (see [architecture-decision-records.md](../decisions/architecture-decision-records.md))
- no surprise `go` directive or `toolchain` bump rides along unless it is the intended change
- the `go.mod` / `go.sum` diff is understood, not blindly accepted; no `replace` directive is introduced to dodge a break
- secrets and lockfile integrity are unaffected — `go mod verify` still passes

## Proof

- `make verify` is green (this wraps tidy, fmt-check, lint, vet, test, race, vuln, build)
- `go tool govulncheck ./...` reports no known vulnerabilities (the point of a security bump)
- targeted tests for the changed surface pass, e.g. `go test ./internal/... -run <DependencyFeature>`
- `git diff go.mod go.sum` reviewed; the `go`/`toolchain` directive is unchanged unless the bump intends it
- for a major bump, the ADR or PR description states what changed and why it is safe

For automated bot PRs (Dependabot is the handbook default), the same `make verify` gate runs in CI before the PR is mergeable — see [ci-and-release.md](../operations/ci-and-release.md#dependency-updates) and the committed [dependabot.yml](../templates/dependabot.yml).
