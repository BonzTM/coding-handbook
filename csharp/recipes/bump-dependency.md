# Recipe: Bump A Dependency

Use this when you raise a direct NuGet package's version — routinely, to clear a vulnerability, or as part of a bot PR you are reviewing to merge.

## Files To Touch

- `Directory.Packages.props` — the single version source under Central Package Management
- `packages.lock.json` in every affected project — regenerated, never hand-edited
- the callers (any code in `src/` or `tests/`) if the package's API changed
- tests that exercise the package's surface
- `CHANGELOG.md` if the bump changes caller- or operator-visible behavior

## Steps

1. Scope the bump before touching anything: patch / minor / major, and security-driven vs routine. A patch or minor is usually mechanical; a major can change behavior or contracts and needs the same scrutiny as adopting a new dependency (re-answer the approval questions in [framework-selection.md](../decisions/framework-selection.md)). A security-driven bump follows the supply-chain guidance in [../operations/security.md](../operations/security.md).
2. Read the upstream changelog or release notes for the target version, especially for a major. Note breaking changes, behavior changes, and any new minimum target framework it demands.
3. Edit the `<PackageVersion>` in `Directory.Packages.props`. Under CPM this is the only place a version lives — never add a `Version` attribute to a `<PackageReference>` in a `.csproj`.
4. Regenerate the lock files: `dotnet restore --force-evaluate`. NuGetAudit runs as part of restore and fails the build on high/critical advisories — a bump that introduces one does not merge.
5. Inspect the diff: `git diff Directory.Packages.props '**/packages.lock.json'`. Confirm the direct bump is what you intended and understand every transitive line that moved. Do not paper over a breaking change with a `VersionOverride` or a downgraded transitive pin (forbidden for production builds per [framework-selection.md](../decisions/framework-selection.md)).
6. Update the callers for any changed API and adjust tests. Then run the full gate (`pwsh ./verify.ps1`) plus targeted tests for the package's surface.

## Invariants To Preserve

- every bump passes `pwsh ./verify.ps1` — restore (locked), format-check, build (warnings-as-errors), test, audit — before merge; no exceptions for bot PRs
- a major version bump gets a written rationale; if it shifts behavior or a contract, capture it in an ADR (see [architecture-decision-records.md](../decisions/architecture-decision-records.md))
- no surprise `TargetFramework` or SDK bump rides along unless it is the intended change
- the `Directory.Packages.props` / `packages.lock.json` diff is understood, not blindly accepted; no `VersionOverride` is introduced to dodge a break
- lock-file integrity holds: `dotnet restore --locked-mode` still succeeds after the regen

## Proof

- `pwsh ./verify.ps1` is green (its locked-mode restore proves the lock files are consistent, and NuGetAudit proves no high/critical advisory)
- `dotnet list package --vulnerable --include-transitive` reports no known vulnerabilities (the point of a security bump)
- targeted tests for the changed surface pass, e.g. `dotnet test tests/Orders.UnitTests --filter <DependencyFeature>`
- `git diff Directory.Packages.props '**/packages.lock.json'` reviewed; no TFM or SDK change rides along unless intended
- for a major bump, the ADR or PR description states what changed and why it is safe

For automated bot PRs (Dependabot is the handbook default), the same `pwsh ./verify.ps1` gate runs in CI before the PR is mergeable — see [ci-and-release.md](../operations/ci-and-release.md) and the committed [dependabot.yml](../templates/dependabot.yml).
