# PR Review Checklist

Review checklist for Godot changes that affect scenes, scripts, contracts, or the exported build.

## Scene And Resource Diffs

- [ ] Are all touched scenes and resources committed in text formats (`.tscn`/`.tres`) with readable, explainable diffs — no binary `.scn`/`.res`, per [../foundations/project-setup.md](../foundations/project-setup.md)?
- [ ] Is the `.tscn` diff limited to nodes the change actually touched, with no churn from re-saving unrelated scenes?
- [ ] Did renames or moves keep `snake_case` casing, so paths still resolve in the case-sensitive exported PCK?
- [ ] Is the diff free of editor cruft — no `.godot/`, no `*.translation`, nothing the [gitignore template](../templates/gitignore.txt) keeps out of the repo?

## Contracts And Correctness

- [ ] Does every new or changed cross-scene signal keep the contract shape — past-tense name, typed parameters, connections and tests moved together per [../recipes/add-a-signal-contract.md](../recipes/add-a-signal-contract.md)?
- [ ] Do consumers still use each scene's contract of signals, methods, and exported properties, rather than reaching into internal node paths ([../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md))?
- [ ] Is new GDScript fully statically typed, with zero new `UNTYPED_DECLARATION` warnings per [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md)?
- [ ] If the change added third-party content, is it confined to `addons/` with source and version recorded, and the adoption routed through [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)?

## Export And Settings

- [ ] Is `export_presets.cfg` untouched unless the PR is deliberately an export or release change ([../operations/ci-and-release.md](../operations/ci-and-release.md)) — and free of keystores, passwords, and other credentials either way?
- [ ] Are `project.godot` edits deliberate and consistent with [../templates/project-settings-conventions.md](../templates/project-settings-conventions.md), with the engine pin, export templates version, and CI still agreeing?

## Proof

- [ ] Targeted tests prove the actual behavior change under `godot --headless` per [../quality/testing.md](../quality/testing.md), and new scenes instantiate standalone in a headless test ([../recipes/add-a-scene.md](../recipes/add-a-scene.md)).
- [ ] `gdformat --check` and `gdlint` pass per [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md), with no new `gdlint:ignore` that lacks a justification.
- [ ] For export-affecting changes (paths or casing, presets, `project.godot`, C#), a headless export succeeded and the artifact launches.
- [ ] For performance-sensitive changes, a profiler capture identified the bottleneck and a before/after measurement is in the PR, per [../operations/performance-and-profiling.md](../operations/performance-and-profiling.md).
- [ ] The PR body follows [../templates/pull_request_template.md](../templates/pull_request_template.md), with the Scenes And Scripts Touched table filled in.
