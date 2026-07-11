# Dependency Upgrade Checklist

The periodic maintenance pass that drains the update queue and keeps the toolchain current. This is the recurring sweep; for a single bump use [recipes/bump-dependency.md](../recipes/bump-dependency.md), and for the automation that feeds the queue see [operations/ci-and-release.md](../operations/ci-and-release.md).

## Cadence & Scope

- [ ] Review the Dependabot queue on a fixed schedule (the committed [dependabot.yml](../templates/dependabot.yml) opens grouped PRs weekly); do not let it bank up until an incident forces it.
- [ ] Batch the mechanical patch/minor updates into the grouped PR and merge them together once green.
- [ ] Split every major version bump into its own PR — it earns a changelog read and a rationale, not a grouped checkmark.
- [ ] Split the `global.json` SDK bump into its own reviewed PR; it never rides along inside a grouped dependency PR.

## Per Update

- [ ] Read the upstream changelog/release notes for any major before touching code; note breaking changes, behavior changes, and any new minimum target framework.
- [ ] Bump the version in `Directory.Packages.props` (the single pin point under Central Package Management), then regenerate lock files with `dotnet restore --force-evaluate` so `packages.lock.json` settles the transitive graph.
- [ ] Inspect the diff with `git diff Directory.Packages.props '**/packages.lock.json'`; understand every transitive line that moved and reject a surprise target-framework or SDK requirement that snuck in.
- [ ] Update callers and tests for any changed API; cover the changed surface with real tests at any DB/external boundary the dependency touches.

## Toolchain

- [ ] Raise the `global.json` SDK pin to a current supported line on the same cadence (`rollForward: latestFeature` absorbs feature bands, but the pinned baseline still needs a periodic reviewed bump); don't let it rot while only app packages move.
- [ ] Keep the `TargetFramework` on the current LTS; moving to a new major is a planned migration with its own ADR per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md), never a rider on a dependency PR.
- [ ] Re-run `pwsh ./verify.ps1` after an SDK or analyzer bump; new analyzer warnings surface here as build errors (`AnalysisLevel=latest-all` + warnings-as-errors), not in production.

## Safety

- [ ] Every PR passes the full `pwsh ./verify.ps1` gate — restore (locked), format-check, build (warnings-as-errors), test, audit — before merge; bot PRs included, no fast lane.
- [ ] No floating version (`*` or open range), local package source, or `nuget.config` edit is introduced to paper over a break (forbidden for production builds).
- [ ] No auto-merge of a PR that bumped the `global.json` SDK pin or `TargetFramework` unreviewed; security-relevant bumps are confirmed against [operations/security.md](../operations/security.md).

## Verification

```powershell
# the bot queue is drained on cadence (no stale open dependency PRs)
gh pr list --label dependencies --state open

# the gate is green (restore (locked), format-check, build (warnings-as-errors), test, audit)
pwsh ./verify.ps1
dotnet restore --locked-mode
dotnet list package --vulnerable --include-transitive

# the toolchain is not stale: the SDK pin is a current supported release line
Get-Content global.json
```
