# Dependency Upgrade Checklist

The periodic maintenance pass that drains the update queue and keeps the toolchain current. This is the recurring sweep; for a single bump use [recipes/bump-dependency.md](../recipes/bump-dependency.md), and for the automation that feeds the queue see [operations/ci-and-release.md](../operations/ci-and-release.md#dependency-updates).

## Cadence & Scope

- [ ] Review the Dependabot/Renovate queue on a fixed schedule (the committed [dependabot.yml](../templates/dependabot.yml) opens grouped PRs weekly); do not let it bank up until an incident forces it.
- [ ] Batch the mechanical patch/minor updates into the grouped PR and merge them together once green.
- [ ] Split every major version bump into its own PR — it earns a changelog read and a rationale, not a grouped checkmark.
- [ ] Split the `go` directive / `toolchain` bump into its own reviewed PR; it never rides along inside a grouped dependency PR.

## Per Update

- [ ] Read the upstream changelog/release notes for any major before touching code; note breaking changes, behavior changes, and any new minimum Go version.
- [ ] Pin the version: `go get <module>@<version>`, then `go mod tidy` to settle the transitive graph and `go.sum`.
- [ ] Inspect the diff with `git diff go.mod go.sum`; understand every transitive line that moved and reject a surprise `go`/`toolchain` bump that snuck in.
- [ ] Update callers and tests for any changed API; cover the changed surface with real tests at any DB/external boundary the dependency touches.

## Toolchain

- [ ] Raise the `go` directive and `toolchain` line to a current supported Go release on the same cadence (Go 1.24+); don't let them rot while only app deps move.
- [ ] Bump the pinned tool deps — `golangci-lint`, `govulncheck`, and the rest declared via `go get -tool` — to current releases with `go get -tool <pkg>@<version>`.
- [ ] Re-run `make verify` after a toolchain or linter bump; new lints or vuln findings surface here, not in production.

## Safety

- [ ] Every PR passes the full `make verify` gate, including `go test -race ./...` and `go tool govulncheck ./...`, before merge — bot PRs included, no fast lane.
- [ ] No `replace` directive is introduced to paper over a break (forbidden for production builds).
- [ ] No auto-merge of a PR that bumped the `go` directive or `toolchain` unreviewed; security-relevant bumps are confirmed against [operations/security.md](../operations/security.md).

## Verification

```bash
# the bot queue is drained on cadence (no stale open dependency PRs)
gh pr list --label dependencies --state open

# the gate is green (wraps tidy, fmt-check, lint, vet, test, race, vuln, build)
make verify
go tool govulncheck ./...

# the toolchain is not stale: go directive + toolchain are a current supported release
go mod edit -json | grep -E '"(Go|Toolchain)"'
```
