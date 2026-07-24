# Release Checklist

Release-cut checklist for a Godot 4.x project: run it on the release commit before pushing the tag, and again before any artifact reaches players or a storefront. The owning doc for every fix is [../operations/ci-and-release.md](../operations/ci-and-release.md).

## Version And Tag

- [ ] Is the player-facing version string bumped in its committed home (`application/config/version` in `project.godot` or the repo's equivalent), so the running build can report exactly what it is?
- [ ] Is the changelog updated for this version with player- and operator-facing changes, including any save-schema bump and its migration behavior?
- [ ] Is the release tag `v<MAJOR>.<MINOR>.<PATCH>` placed on the exact commit CI proved, per [../operations/ci-and-release.md](../operations/ci-and-release.md) (## Versioning And Release Artifacts)?
- [ ] Do the tag, changelog heading, in-game version string, and artifact names all embed the same version?

## Source And CI Gate

- [ ] Is the full baseline gate green on the release commit — format, lint, headless tests, import, and per-preset export ([../AGENTS.md](../AGENTS.md) ## Baseline Verification)?
- [ ] Was every shipped artifact produced by a CI `godot --headless --export-release` job triggered by the tag — none hand-clicked from a dev machine's editor, so the shipped build cannot drift from what CI proves?
- [ ] Do the CI editor install, export templates, and local tooling all still derive from the single pinned engine version — an editor/template version mismatch is the first suspect when export fails only in CI ([../operations/ci-and-release.md](../operations/ci-and-release.md))?
- [ ] Does every preset name referenced in the workflow appear in the committed `export_presets.cfg`, with signing material injected only as CI secrets and no keystore or credential file tracked?
- [ ] Did each export job assert its artifact exists and has plausible size, rather than trusting job status alone?

## Artifact Smoke Test

- [ ] Does each exported binary launch on its target platform — not merely export green — and show the expected version string?
- [ ] Are all shipped artifacts `--export-release` builds, with `--export-debug` reserved for internal QA and crash triage?
- [ ] Are artifacts named `<game>-<version>-<platform>.<ext>` with checksums published alongside, so a storefront upload traces back to an exact commit?
- [ ] If the game ships text needing ICU data (emoji, CJK, Thai, and similar scripts), does the exported build render it — not just the editor ([../systems/localization.md](../systems/localization.md))?

## Save Compatibility

- [ ] Does a save file written by the previous shipped version load through the migration ladder to a validated current-version state in this build ([../systems/save-and-load.md](../systems/save-and-load.md))?
- [ ] If the save schema changed, is the envelope `version` bumped with its stepwise migration, is a fixture for the prior version committed under `test/`, and does a future-version file still refuse to load?

## Store And Platform Metadata

- [ ] Are store listings, icons, screenshots, and content ratings current for anything player-visible this release changed?
- [ ] Are per-platform release requirements met — mobile signing material valid and the platform's monotonically increasing build identifier bumped in the export preset — with platform review lead time in the release plan?
- [ ] Are published release notes consistent with the changelog, mentioning anything an operator or player must act on (save migrations, changed defaults, dropped platforms)?

## Proof

- [ ] `godot --headless --import && godot --headless --export-release "<Preset>" build/<artifact>` reproduces each artifact from a clean checkout of the tagged commit using the pinned engine version.
- [ ] Each uploaded artifact was launched on its target platform and the displayed version matches the tag.
- [ ] The committed previous-version fixture save loads in the release build with no migration or validation error.
- [ ] `git tag --points-at <release-commit>` prints exactly one `v`-prefixed tag, and every published artifact name embeds the same version.
- [ ] `git check-ignore .godot/export_credentials.cfg` succeeds and `git ls-files | grep -iE 'keystore|\.jks|\.p12'` returns nothing.
