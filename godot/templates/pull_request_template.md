<!--
This file is the handbook template for a repo's PR template.
Install it at: .github/pull_request_template.md
The Proof section mirrors the Baseline Verification table in AGENTS.md — that table
is the source of truth for gate depth. Keep the two in sync.
Delete this comment block when installing.
-->

## Summary

<!-- What changed and why, in two or three sentences. Link the issue/ticket.
     Link the ADR if this change is architectural or shifts a contract
     (per ../decisions/architecture-decision-records.md). -->

-

## Scenes And Scripts Touched

<!-- List every scene (.tscn), resource (.tres), and script this PR adds, edits,
     moves, or renames. Renames matter more than in most stacks: the exported PCK
     filesystem is case-sensitive while dev filesystems usually are not, so a
     casing-only rename can break the exported build while passing locally
     (../foundations/project-setup.md). -->

| Path | Change | Notes |
|---|---|---|
| | | |

## Proof

<!-- Check what ran and paste outcomes. Commands and expectations are defined in
     ../AGENTS.md (## Baseline Verification); CI runs the same gate via the
     committed workflow (../templates/github-workflows-ci.yml). -->

- [ ] `gdformat --check` clean on all touched script directories.
- [ ] `gdlint` exit 0 against the committed `.gdlintrc`.
- [ ] Test runner green under `godot --headless`, locally and in CI.
- [ ] `godot --headless --export-debug "<Preset>" <output>` succeeds and the artifact launches — required when the change touches file paths or casing, `project.godot`, `export_presets.cfg`, or C# (../operations/ci-and-release.md).

## Checklist

- [ ] Tests added or updated that prove the actual behavior change (../quality/testing.md); new scenes instantiate standalone in a headless test (../recipes/add-a-scene.md).
- [ ] New signals are past-tense and their contract is documented (../recipes/add-a-signal-contract.md).
- [ ] New gameplay input goes through InputMap actions, not raw keycodes (../foundations/input-handling.md).
- [ ] Save-schema changes bump the save version and ship a migration (../systems/save-and-load.md); `@rpc` signature changes are flagged as breaking — peers checksum-match signatures (../systems/multiplayer.md).
- [ ] Engine version, export templates version, and CI pin still agree if `project.godot` or presets moved (../operations/ci-and-release.md).
