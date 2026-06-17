# Recipe: Release A Library Version

Use this when you cut a tagged release of a reusable module — a library, SDK, or shared internal package other repos import by version. (For shipping a deployable service artifact, use the [release checklist](../checklists/release.md) instead; this recipe is the module-tagging path that completes the Library reading path.)

Governing docs: [foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md#versioning-and-deprecation-policy) decides the version bump; [operations/ci-and-release.md](../operations/ci-and-release.md) owns tagging and CHANGELOG policy.

## Files To Touch

- `CHANGELOG.md`
- `go.mod` — only for a v2+ major: the module path gets a `/vN` suffix
- the `version`/`buildinfo` package, if the library exposes one
- the VCS tag (canonical `vX.Y.Z`, `v` prefix required)

## Steps

1. Decide the bump from the diff, not from habit. Any breaking change to an exported identifier, signature, or documented behavior is a MAJOR bump per [contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md#versioning-and-deprecation-policy). Additive-only changes are MINOR; fixes are PATCH. If you cannot point to a breaking export, you are not allowed to skip a MAJOR — and you are not allowed to bump MAJOR without one either.
2. For a v2+ major, change the module path to carry the `/vN` suffix: `module example.com/lib/v2` in `go.mod`, and update every internal import that referenced the old path. SemVer import compatibility (the "import compatibility rule") makes `/v2` a distinct module — skipping the suffix ships a silently-broken release. v0 and v1 take no suffix.
3. Update `CHANGELOG.md` in Keep a Changelog form: move the `## [Unreleased]` entries under a new `## [X.Y.Z] - YYYY-MM-DD` heading, grouped by Added / Changed / Deprecated / Removed / Fixed / Security. Every operator- or caller-visible change must already have a line (enforced at PR time — see [ci-and-release.md](../operations/ci-and-release.md)).
4. Bump any exposed `version`/`buildinfo` constant to match the tag, if the library carries one. Most libraries rely on the VCS tag alone; do not invent a parallel source of truth.
5. Run the gate and commit: `make verify`, then commit the CHANGELOG and any path/version edits on the release branch and merge to the default branch.
6. Tag the release commit with an annotated, canonical tag: `git tag -a v1.4.0 -m "v1.4.0"`. The `v` prefix is required for Go module resolution; a lightweight tag is forbidden because it carries no tagger or message. For a v2+ major in a subdirectory-less repo the tag is still `v2.0.0` (the proxy maps it to the `/v2` module path).
7. Push the tag explicitly: `git push origin v1.4.0`. Pushing the branch does not push tags.
8. Verify resolution from a clean cache before announcing: `GOFLAGS= go list -m example.com/lib@v1.4.0` (or a throwaway `go get`) must resolve through the module proxy.

## Invariants To Preserve

- the tag is canonical and `v`-prefixed (`vX.Y.Z`); no bare `1.4.0`, no `release-1.4.0`
- the tag is annotated, not lightweight
- any breaking change to an exported symbol or documented behavior forced a MAJOR bump
- a v2+ release carries the `/vN` suffix in both the `go.mod` module path and import paths
- `CHANGELOG.md` has a dated entry for this version covering every visible change
- the release commit is on the default branch and passed `make verify`

## Proof

- `go mod tidy` leaves `go.mod`/`go.sum` clean (no diff) on the tagged commit
- `make verify` is green on the release commit (tidy, fmt-check, lint, vet, test, race, vuln, build)
- `go list -m example.com/lib@v1.4.0` resolves the new version (or a fresh `go get example.com/lib@v1.4.0` in a scratch module succeeds) — proves the proxy picked it up
- `git cat-file -t v1.4.0` reports `tag` (annotated), not `commit` (lightweight)
- for a v2+ release, `go list -m example.com/lib/v2@v2.0.0` resolves and importers compile against the `/v2` path

Run the cross-cutting gates in the [release checklist](../checklists/release.md) as well — this recipe is the module-tagging specialization of it.
