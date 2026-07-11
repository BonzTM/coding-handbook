# CI and Release

Delivery guidance for .NET repos that should build the same way locally, in CI, and at release time — on every OS a developer touches.

## Default Approach

Every PR runs the same gate a developer runs locally: `pwsh ./verify.ps1`. CI adds nothing to the gate; it just runs it on more machines.

| Stage | Commands | Purpose |
|---|---|---|
| restore | `dotnet restore --locked-mode` | committed `packages.lock.json` matches the dependency graph exactly; drift fails, it is not silently rewritten |
| format-check | `dotnet format --verify-no-changes` | consistent source shape from `.editorconfig`; read-only |
| build | `dotnet build -c Release` with warnings-as-errors | compile proof AND static analysis — the built-in analyzers (`AnalysisLevel=latest-all`, `EnforceCodeStyleInBuild=true`) run inside the build, so a warning-free build is the analysis gate |
| test | `dotnet test` (unit projects; xUnit v3 on Microsoft.Testing.Platform) | functional confidence, offline |
| audit | NuGetAudit on restore (`NuGetAuditMode=all`), failing on high/critical advisories | supply-chain and known-vulnerability check |

The committed [CI workflow](../templates/github-workflows-ci.yml) runs `pwsh ./verify.ps1` on an **ubuntu/windows/macos matrix** — the cross-platform mandate from [../foundations/cross-platform.md](../foundations/cross-platform.md) is enforced here, not trusted. `setup-dotnet` installs the SDK from the committed [global.json](../templates/global.json), so local, CI, and container builds agree on the toolchain. Keep `verify.ps1` and this table in sync rather than hand-editing CI; the [`Makefile`](../templates/Makefile) is a one-line shim to the same script.

The matrix jobs are the *offline* gate: no network beyond package restore, no database, so the table above proves the in-memory path on all three OSes. The workflow adds a second **integration job — ubuntu only** — that runs `pwsh ./verify.ps1 -Integration`, standing up real dependencies (a version-pinned Postgres via Testcontainers; the pin lives in the test code) so the real SQL path and the committed EF Core migrations are proven in CI rather than only at deploy time. It is ubuntu-only for an honest reason: GitHub's ubuntu runners ship Docker; the macOS runners have none and the windows runners run Windows containers — Linux containers are not available there without setup this handbook does not assume. The same suite sits behind the explicit `-Integration` switch locally and skips when Docker is absent, so it never blocks the offline inner loop — see [quality/testing.md](../quality/testing.md).

## Release Defaults

- Tag releases as `v1.2.3` — SemVer 2.0 with the `v` prefix on the VCS tag; changelog headings render without it (see Changelog Policy).
- **Version injection default: MinVer.** The package reference (pinned in [Directory.Packages.props](../templates/Directory.Packages.props)) derives the assembly, package, and informational versions from the latest `v*` tag in git history at build time — no version committed in any csproj, no manual bump PRs. Two operational consequences: checkouts that build a release must be **full-depth** (`fetch-depth: 0`; a shallow clone leaves MinVer blind to tags and it emits a default `0.0.0-alpha` height version), and Docker builds have no `.git` in context, so CI computes the version once and passes it into the image build as `-p:Version=...` (see [deployment.md](deployment.md)). The escape hatch — a static `<Version>` in `Directory.Build.props` — is for repos whose release process cannot key off tags, and takes an ADR.
- Use multi-stage container builds for service artifacts; the hardened base-image stance, runtime limits, and probe wiring live in [deployment.md](deployment.md), with a committed [Dockerfile](../templates/Dockerfile) and [.dockerignore](../templates/.dockerignore).
- Adopt ReadyToRun, Native AOT, or profile-guided optimization only after measuring on representative load and recording the decision in an [ADR](../decisions/architecture-decision-records.md) — never by reflex.
- Introduce a build framework (Nuke, Cake) only when release complexity justifies it, not on day one; `verify.ps1` is the entrypoint until then.

### Release Pipeline

Pushing a `v*` tag triggers the release workflow (committed template: [github-workflows-release.yml](../templates/github-workflows-release.yml)): it checks out full history, resolves the version from the tag, and builds the committed [Dockerfile](../templates/Dockerfile) with the `VERSION`/`COMMIT`/`CREATED` build args — so the stamped assembly metadata, the OCI labels, and the VCS tag all agree per [deployment.md](deployment.md) — then pushes the image to the container registry chosen at spec-intake ([checklists/spec-intake.md](../checklists/spec-intake.md)); the template uses `ghcr.io` as a placeholder. The tag is only pushed from a commit whose CI run is green; the release checklist ([checklists/release.md](../checklists/release.md)) walks the full sequence.

### NuGet Package Publishing

Library repos release packages instead of images, on the same tag trigger:

- `dotnet pack -c Release` after the full verify gate; MinVer stamps the version from the tag, so the `.nupkg` version and the VCS tag cannot drift.
- Pack with `ContinuousIntegrationBuild=true` (deterministic build), SourceLink metadata (on by default in current SDKs), and a symbols package (`.snupkg`) — a consumer stepping into your stack trace is part of the contract.
- Package metadata is not optional: license expression, repository URL, a package README. NuGet.org surfaces all three.
- Turn on package validation (`EnablePackageValidation` with `PackageValidationBaselineVersion` set to the last released version) so an accidental breaking API change fails the pack, not the consumer — the mechanical enforcement of [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md).
- Push with `dotnet nuget push --skip-duplicate` using an API key held in CI secrets, never committed. The step-by-step flow is [../recipes/release-library-version.md](../recipes/release-library-version.md).

### Dependency Updates

Dependencies drift whether or not you touch them; let automation surface the drift on a schedule instead of discovering it during an incident.

- Automate the PRs. The handbook default is **Dependabot** (GitHub-native, pairs with the committed CI workflow); the committed [dependabot.yml](../templates/dependabot.yml) runs weekly for both the NuGet and GitHub Actions ecosystems, groups minor/patch updates into one PR, and caps open PRs. **Renovate** is the supported alternative for non-GitHub repos or advanced cross-repo grouping.
- Keep a regular cadence: bump patch-level dependencies on the weekly schedule, and on the same rhythm raise the [global.json](../templates/global.json) SDK pin to the current patch line rather than letting it rot — Dependabot's `dotnet-sdk` ecosystem or a manual check, but on the calendar either way.
- Every bot PR is a normal PR: it must pass the same `pwsh ./verify.ps1` gate — restore (locked), format-check, build (warnings-as-errors), test, audit — before it can merge. There is no fast lane that skips the safety stages.
- Review the lockfile diff and follow [recipes/bump-dependency.md](../recipes/bump-dependency.md): a version bump rewrites `packages.lock.json`, and the transitive changes in that diff are the actual change under review. Grouped minor/patch updates are usually mechanical, but a major arrives as its own PR and earns a changelog read and a rationale.

### Changelog Policy

Every release maintains a `CHANGELOG.md` in [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format — newest version first, an `Unreleased` section at the top, and only the standard sections (`Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, `Security`). Start from the committed [changelog template](../templates/changelog.md).

- **Operator-visible changes MUST have an entry.** New or changed env vars, migrations, ports, default timeouts, and API/schema/message contract changes go in the changelog so an operator reading it can predict what a deploy will require. This is the same operator-visible set the release checklist enforces; the changelog is where it is written down.
- **Curate the `Unreleased` section per-PR, or generate it from Conventional Commits.** The default is to add the entry in the same PR that makes the change, while the rationale is fresh. Repos that prefer automation derive `Unreleased` from [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) subjects (`feat` -> Added, `fix` -> Fixed, `BREAKING CHANGE`/`!` -> a Changed/Removed entry and a MAJOR bump) per [foundations/git-workflow.md](../foundations/git-workflow.md). Either way the section is never empty-by-neglect at release time.
- **Release moves `Unreleased` to a version heading.** On tag, rename `Unreleased` to the new version with a `YYYY-MM-DD` date and open a fresh empty `Unreleased`. Headings drop the `v` per Keep a Changelog; the git tag keeps it (`v1.2.3` <-> `[1.2.3]`), matching the canonical tag rule in [Release Defaults](#release-defaults). This is part of the release flow in [checklists/release.md](../checklists/release.md).

## Compatibility And Rollback

- API, schema, and event changes need an explicit backward-compatibility story ([../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md)).
- Migrations should be deploy-safe: expand/contract over destructive-in-place, per [services/database.md](../services/database.md) — during a rolling deploy, old pods run against the new schema.
- Release notes should call out operator-visible changes such as new env vars, port changes, migrations, or default timeout changes.
- Message contract changes should also call out replay expectations, DLQ handling, and whether mixed-version consumers or producers are safe during rollout.
- Library releases: a breaking public-API change is a MAJOR version, and package validation (above) is the tripwire that keeps "breaking" from being discovered by a consumer.

## Common Mistakes And Forbidden Patterns

- Restoring in CI without `--locked-mode`, so the lockfile silently regenerates and CI proves a dependency graph nobody committed.
- Dropping matrix legs because "production is Linux anyway" — the windows/macos legs are what keep developer machines working; the cross-platform mandate dies quietly the day they go.
- Builds that require undeclared local tools or shell state; bash-only steps in the verify path (PowerShell 7 is the one blessed script runtime).
- Manual release steps that drift from CI reality.
- Opaque version output that prevents operators from identifying a running binary, or a shallow release checkout that leaves MinVer computing a default `0.0.0-alpha` version instead of the tag's.
- Cargo-cult AOT/ReadyToRun/PGO or container hardening that nobody has actually tested.
- Merging a dependency-bot PR on a green checkmark alone without reviewing the `packages.lock.json` diff, or auto-merging one that quietly bumped a major or the SDK pin.
- Letting the `global.json` SDK pin go stale because only application packages are on an update schedule.
- Shipping an operator-visible change (env var, migration, port, contract change) with no `CHANGELOG.md` entry, so operators discover the new requirement at deploy time.
- Inventing changelog section names instead of the six Keep a Changelog categories, or releasing with a stale `Unreleased` section that was never moved to a version heading.
- A changelog version heading whose date or version drifts from the actual VCS tag (`[1.2.3]` must correspond to tag `v1.2.3`).

## Verification And Proof

- Latest CI run on the release commit is green on all three matrix legs, and the integration job passed on ubuntu.
- `dotnet restore --locked-mode` succeeds on a clean checkout — the committed lockfile is authoritative.
- The release artifact starts successfully and reports its version/commit at startup, matching the VCS tag and the image's OCI labels.
- Deploy-safe changes have a documented rollback or compatibility plan.
- Dependency-update automation is committed (`.github/dependabot.yml` or Renovate config) and its PRs run the same `pwsh ./verify.ps1` gate as any other PR.
- `CHANGELOG.md` exists in Keep a Changelog format; the released version has a heading with a date matching its `v`-prefixed tag, and the `Unreleased` section was moved on release.
- Every operator-visible change in the release range has a changelog entry; a diff of env vars, migrations, ports, and contracts against the changelog shows no gaps.
- Library releases: the packed version equals the tag, symbols and SourceLink resolve, and package validation passed against the previous release.
