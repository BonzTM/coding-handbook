# Recipe: Release A Library Version

Use this when you cut a tagged NuGet release of a reusable library — an SDK, client, or shared package other repos consume by version. (For shipping a deployable service artifact, use the [release checklist](../checklists/release.md) instead; this recipe is the package-tagging path that completes the Library reading path.)

Governing docs: [foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md) decides the version bump; [operations/ci-and-release.md](../operations/ci-and-release.md) owns tagging, packing, and CHANGELOG policy.

## Files To Touch

- `CHANGELOG.md`
- `PublicAPI.Shipped.txt` / `PublicAPI.Unshipped.txt` — the PublicApiAnalyzers surface files in the library project
- the `.csproj` only if package metadata (description, license, readme) changes — never a hand-edited `<Version>`
- the VCS tag (canonical `vX.Y.Z`, `v` prefix required)

## Steps

1. Decide the bump from the public-API diff, not from habit. PublicApiAnalyzers makes the diff explicit: any removed or changed line in the `PublicAPI.*.txt` files — or any breaking change to documented behavior — is a MAJOR bump per [contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md). Additive-only entries in `PublicAPI.Unshipped.txt` are MINOR; none are PATCH. If you cannot point to a breaking surface change, you are not allowed to skip a MAJOR — and you are not allowed to bump MAJOR without one either.
2. Review and promote the surface: read every line in `PublicAPI.Unshipped.txt` as the deliberate act of shipping API, then move those lines into `PublicAPI.Shipped.txt` on the release commit, leaving `Unshipped` empty. Anything you are not ready to support forever gets pulled from the public surface (`internal`, or removed) before the release, not shipped by accident.
3. Update `CHANGELOG.md` in Keep a Changelog form: move the `## [Unreleased]` entries under a new `## [X.Y.Z] - YYYY-MM-DD` heading, grouped by Added / Changed / Deprecated / Removed / Fixed / Security. Every caller-visible change must already have a line (enforced at PR time — see [ci-and-release.md](../operations/ci-and-release.md)).
4. Keep one version source of truth: the tag. The release workflow (copy from [templates/github-workflows-release.yml](../templates/github-workflows-release.yml)) derives the package version from the tag at pack time (`dotnet pack -c Release /p:Version=<tag-without-v>`); do not hand-maintain a `<Version>` in the `.csproj` that can drift from it.
5. Run the gate and land the release commit: `pwsh ./verify.ps1`, then merge the CHANGELOG and PublicAPI edits to the default branch.
6. Tag the release commit with an annotated, canonical tag: `git tag -a v1.2.3 -m "v1.2.3"`. A lightweight tag is forbidden because it carries no tagger or message.
7. Push the tag explicitly: `git push origin v1.2.3`. Pushing the branch does not push tags. The release workflow packs (deterministic CI build, symbols `.snupkg`, SourceLink) and pushes the package to the feed via `dotnet nuget push`.
8. Verify resolution before announcing: in a scratch project against the feed, `dotnet add package Orders.Client --version 1.2.3` restores and compiles. A published NuGet version is immutable — a bad release is fixed by shipping a new PATCH, never by unlisting and reusing the number.

## Invariants To Preserve

- the tag is canonical and `v`-prefixed (`vX.Y.Z`); no bare `1.2.3`, no `release-1.2.3`; the package version is the tag minus the `v`
- the tag is annotated, not lightweight
- any breaking change to the public surface (a `PublicAPI.Shipped.txt` removal/change) or documented behavior forced a MAJOR bump
- `PublicAPI.Unshipped.txt` is empty on the release commit; every shipped API line was reviewed
- `CHANGELOG.md` has a dated entry for this version covering every visible change
- the release commit is on the default branch and passed `pwsh ./verify.ps1`; the package is produced by CI from the tag, never packed and pushed from a laptop
- a published version number is never reused

## Proof

- `pwsh ./verify.ps1` is green on the release commit — restore (locked), format-check, build (warnings-as-errors), test, audit
- `PublicAPI.Unshipped.txt` is empty and the `PublicAPI.Shipped.txt` diff matches the changelog's Added/Changed/Removed entries
- `git cat-file -t v1.2.3` reports `tag` (annotated), not `commit` (lightweight)
- the release workflow run for the tag succeeded and the package is visible on the feed
- a scratch project restores the new version: `dotnet add package Orders.Client --version 1.2.3` followed by `dotnet build` succeeds

Run the cross-cutting gates in the [release checklist](../checklists/release.md) as well — this recipe is the package-tagging specialization of it. To retire public API before a MAJOR, see [deprecate-and-remove-contract.md](deprecate-and-remove-contract.md).
