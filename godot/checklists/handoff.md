# Handoff Checklist

Ownership-transfer checklist for a Godot repo built from this handbook. Walk it with both the outgoing and incoming owner present; [../onboarding-and-handoff.md](../onboarding-and-handoff.md) owns the day-one reading path and the questions the new owner must answer unaided.

## Ownership And Access

- [ ] Is `CODEOWNERS` (or the platform equivalent) updated so reviews and notifications route to the new owner, with the outgoing owner removed where no longer warranted?
- [ ] Are store dashboards, web hosting, and any crash-reporting or analytics accounts the project uses transferred to the new owner?
- [ ] Can the new owner authenticate to every system needed to build, export, and publish — verified live, not assumed?
- [ ] Is the outgoing owner's access revoked everywhere it should not outlive their ownership?

## Engine And Toolchain

- [ ] Are the pinned Godot version (standard or .NET build), the matching export template version, and the lint/test toolchain pins recorded in the repo, and do they match what CI and both owners' machines actually run ([../operations/ci-and-release.md](../operations/ci-and-release.md))?
- [ ] Does the new owner know where the pinned editor and export templates are downloaded from, and has the engine-upgrade watch — who tracks new 4.x releases, who decides when the project moves — been handed off ([../onboarding-and-handoff.md](../onboarding-and-handoff.md))?
- [ ] Does opening the project in the pinned editor on the new owner's machine regenerate the gitignored `.godot/` cleanly, with the main scene running and no missing dependencies?

## Signing Keys And Secrets

- [ ] Is the location of every credential documented — signing keystores, store upload keys, CI secrets — with store, path, scope, and rotation procedure per credential, and who is now responsible for each?
- [ ] Are signing keystores and their recovery paths in the new owner's custody before the outgoing owner loses the ability to re-issue them — the highest-stakes item in the transfer ([../onboarding-and-handoff.md](../onboarding-and-handoff.md))?
- [ ] Are credentials the outgoing owner personally held rotated, and CI secrets re-issued under the new owner's control?
- [ ] Do sensitive export options still live only in the gitignored `.godot/export_credentials.cfg` locally and in CI secrets at job time, with nothing in the repo or its logs ([../operations/ci-and-release.md](../operations/ci-and-release.md))?

## Addons And Asset Licenses

- [ ] Does every addon in `addons/` have its source and vendored version recorded, so the next upgrade is a diff against a known base rather than archaeology ([../foundations/project-setup.md](../foundations/project-setup.md))?
- [ ] Is every local patch to a vendored addon recorded in an ADR, so the next upgrade re-applies or drops it deliberately ([../foundations/project-setup.md](../foundations/project-setup.md))?
- [ ] Is the license inventory complete enough that the new owner can answer "are we allowed to ship this?" for every addon and binary asset in the repo?
- [ ] Is Git LFS working on the new owner's clone, with `.gitattributes` covering every binary asset type present?

## Decisions And Knowledge

- [ ] Are all open and proposed ADRs in `decisions/` surfaced to the new owner with their current status ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md))?
- [ ] Is tribal knowledge — undocumented decisions, known failure modes, the hotfix path and how long it takes end to end — converted into an ADR or repo doc before transfer, not kept in the outgoing owner's head?
- [ ] Has the outgoing owner walked through the load-bearing choices: scene architecture, the autoload set, the save format and its compatibility story ([../systems/save-and-load.md](../systems/save-and-load.md)), and the GDScript/C# stance ([../foundations/gdscript-vs-csharp.md](../foundations/gdscript-vs-csharp.md))?
- [ ] Does the new owner know what triggers a release, how it is tagged, and how builds reach each shipping channel ([../operations/ci-and-release.md](../operations/ci-and-release.md))?

## Proof

The handoff is complete only when the new owner can do all of the following unaided, from a clean clone, without the outgoing owner present:

- [ ] The full Baseline Verification gate in [../AGENTS.md](../AGENTS.md) (## Baseline Verification) passed on the new owner's machine from a clean clone.
- [ ] `godot --headless --import && godot --headless --export-release "<Preset>" build/<artifact>` succeeded for every shipping preset, with each artifact present and plausibly sized.
- [ ] `git check-ignore .godot` succeeds and no keystore or credential file is tracked in the repo.
- [ ] The new owner answered every day-one question in [../onboarding-and-handoff.md](../onboarding-and-handoff.md) without help.
