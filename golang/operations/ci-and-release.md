# CI and Release

Delivery guidance for Go repos that should build the same way locally, in CI, and at release time.

## Default Approach

Every PR should run a predictable baseline pipeline.

| Stage | Commands | Purpose |
|---|---|---|
| module hygiene | `go mod tidy`, `go mod verify` | clean module graph |
| formatting | `go tool golangci-lint fmt --diff` (gofumpt + gci, via `make fmt-check`) | consistent source shape |
| static analysis | `go vet ./...`, `go tool golangci-lint run` (includes `staticcheck`) | catch correctness issues early |
| tests | `go test ./...`, `go test -race ./...` | functional and concurrency confidence |
| security | `govulncheck ./...` | supply-chain and known-vulnerability check |
| build | `go build ./...` | compile and link proof (release artifacts add `-trimpath`; see Release Defaults) |

Tailor the matrix to the repo, but do not quietly drop the safety stages because they are slow. The committed [CI workflow](../templates/github-workflows-ci.yml) runs `make verify`, which wraps this table; keep the Makefile and this table in sync rather than hand-editing CI.

The `verify` job is the *offline* gate: it runs with no network and no database, so the table above proves the in-memory path. The workflow adds a second `integration` job that stands up a `postgres:16` service container and runs `go test -tags=integration ./...` with `TEST_DATABASE_DSN` pointing at it, so the real SQL path (and the embedded migrations) is proven in CI rather than only at deploy time. The same suite is gated behind `//go:build integration` and skips when `TEST_DATABASE_DSN` is unset, so it never runs in the offline inner loop — see [quality/testing.md](../quality/testing.md).

## Release Defaults

- tag releases with canonical Go module versions in `v1.2.3` form (the `v` prefix is required by the module system and the module proxy); changelog or display strings may render them without the `v`, but the VCS tag must include it
- inject build metadata through linker flags or a small `internal/buildinfo` package
- use multi-stage container builds for service artifacts; the hardened base-image stance, runtime limits, and probe wiring live in [deployment.md](deployment.md), with a committed [Dockerfile](../templates/Dockerfile) and [.dockerignore](../templates/.dockerignore)
- add `default.pgo` only after collecting representative production profiles and deciding it materially helps
- introduce GoReleaser only when release complexity justifies it, not on day one by reflex

### Release Pipeline

Pushing a `v*` tag triggers the release workflow (committed template: [github-workflows-release.yml](../templates/github-workflows-release.yml)): it builds the committed [Dockerfile](../templates/Dockerfile) with the `VERSION`/`COMMIT`/`CREATED` build-args — so the stamped binary, the OCI labels, and the VCS tag all agree per [deployment.md](deployment.md) — and pushes the image to the container registry chosen at spec-intake ([checklists/spec-intake.md](../checklists/spec-intake.md)); the template uses `ghcr.io` as a placeholder.

### Dependency Updates

Dependencies drift whether or not you touch them; let automation surface the drift on a schedule instead of discovering it during an incident.

- Automate the PRs. The handbook default is **Dependabot** (GitHub-native, pairs with the committed CI workflow); the committed [dependabot.yml](../templates/dependabot.yml) runs weekly for both the `gomod` and `github-actions` ecosystems, groups minor/patch updates into one PR, and caps open PRs. **Renovate** is the supported alternative for non-GitHub repos or advanced cross-repo grouping.
- Keep a regular cadence: bump patch-level dependencies on the weekly schedule, and on the same rhythm raise the `go` directive and `toolchain` line to a current supported release rather than letting them rot.
- Every bot PR is a normal PR: it must pass the same `make verify` gate — including `-race` and `govulncheck` — before it can merge. There is no fast lane that skips the safety stages.
- Review the lockfile diff and follow [recipes/bump-dependency.md](../recipes/bump-dependency.md): grouped minor/patch updates are usually mechanical, but a major arrives as its own PR and earns a changelog read and a rationale.

### Changelog Policy

Every release maintains a `CHANGELOG.md` in [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format — newest version first, an `Unreleased` section at the top, and only the standard sections (`Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`). Start from the committed [changelog template](../templates/changelog.md).

- **Operator-visible changes MUST have an entry.** New or changed env vars, migrations, ports, default timeouts, and API/schema/message contract changes go in the changelog so an operator reading it can predict what a deploy will require. This is the same operator-visible set the release checklist enforces; the changelog is where it is written down.
- **Curate the `Unreleased` section per-PR, or generate it from Conventional Commits.** The default is to add the entry in the same PR that makes the change, while the rationale is fresh. Repos that prefer automation derive `Unreleased` from [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) subjects (`feat` -> Added, `fix` -> Fixed, `BREAKING CHANGE`/`!` -> a Changed/Removed entry and a MAJOR bump) per [foundations/git-workflow.md](../foundations/git-workflow.md). Either way the section is never empty-by-neglect at release time.
- **Release moves `Unreleased` to a version heading.** On tag, rename `Unreleased` to the new version with a `YYYY-MM-DD` date and open a fresh empty `Unreleased`. Headings drop the `v` per Keep a Changelog; the git tag keeps it (`v1.2.3` <-> `[1.2.3]`), matching the canonical tag rule in [Release Defaults](#release-defaults). This is part of the release flow in [checklists/release.md](../checklists/release.md).

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
- merging a dependency-bot PR on a green checkmark alone without reviewing the `go.mod`/`go.sum` diff, or auto-merging one that quietly bumped the `go` directive or toolchain
- letting the `go` directive and toolchain go stale because only application dependencies are on an update schedule
- shipping an operator-visible change (env var, migration, port, contract change) with no `CHANGELOG.md` entry, so operators discover the new requirement at deploy time
- inventing changelog section names instead of the six Keep a Changelog categories, or releasing with a stale `Unreleased` section that was never moved to a version heading
- a changelog version heading whose date or version drifts from the actual VCS tag (`[1.2.3]` must correspond to tag `v1.2.3`)

## Verification And Proof

- latest CI run on the release commit is green
- `go version -m <binary>` or the repo's equivalent shows expected module metadata
- release artifact starts successfully and reports its version/build info
- deploy-safe changes have a documented rollback or compatibility plan
- dependency-update automation is committed (`.github/dependabot.yml` or Renovate config) and its PRs run the same `make verify` gate as any other PR
- `CHANGELOG.md` exists in Keep a Changelog format; the released version has a heading with a date matching its `v`-prefixed tag, and the `Unreleased` section was moved on release
- every operator-visible change in the release range has a changelog entry; a diff of env vars, migrations, ports, and contracts against the changelog shows no gaps
