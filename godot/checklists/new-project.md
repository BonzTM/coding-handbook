# New Project Checklist

Bootstrap checklist for a brand-new Godot repo using this handbook, targeting the current 4.x stable line.

## Project Settings

- [ ] Is exactly one Godot 4.x stable version pinned, with the editor, export templates, and CI tooling all derived from that single pin ([../operations/ci-and-release.md](../operations/ci-and-release.md))?
- [ ] Are `application/config/name` and the main scene set, so the project runs headless and not only from the editor's current tab?
- [ ] Is the GDScript `UNTYPED_DECLARATION` warning enabled, per the typing policy in [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md)?
- [ ] Are the remaining baseline settings applied from [../templates/project-settings-conventions.md](../templates/project-settings-conventions.md), and is the physics tick rate still the default 60 unless [../systems/physics-and-movement.md](../systems/physics-and-movement.md) justified a change?
- [ ] Do input bindings live in the InputMap as named actions rather than raw event checks in code ([../foundations/input-handling.md](../foundations/input-handling.md))?
- [ ] Is the scripting-language decision (GDScript vs C#) made against the target platforms per [../foundations/gdscript-vs-csharp.md](../foundations/gdscript-vs-csharp.md) — remembering C# rules out web export?

## Layout And Version Control

- [ ] Does the directory tree group by feature per [../foundations/project-setup.md](../foundations/project-setup.md), with no `scenes/` + `scripts/` + `assets/` file-type split?
- [ ] Are all folder and file names `snake_case` (C# scripts excepted), so paths resolve identically in the case-sensitive exported PCK?
- [ ] Is the [gitignore template](../templates/gitignore.txt) copied in, with `.godot/` and `*.translation` ignored and `project.godot` committed?
- [ ] Is Git LFS tracking every binary asset type via `.gitattributes` before the first commit of any binary?
- [ ] Are scenes and resources committed in text formats (`.tscn`/`.tres`), with no binary `.scn`/`.res` tracked?
- [ ] Is third-party content confined to `addons/`, vendored with its source and version recorded?
- [ ] Is a `LICENSE` committed as a deliberate choice, not a reflex copy?

## Quality Gates

- [ ] Is gdtoolkit pinned to the engine major (`pip3 install "gdtoolkit==4.*"`) and the [gdlintrc template](../templates/gdlintrc.txt) copied in, per [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md)?
- [ ] Do `gdformat --check <first-party dirs>` and `gdlint <first-party dirs>` pass on the initial commit, scoped to exclude `addons/` per [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md)?
- [ ] Is the test framework chosen per [../quality/testing.md](../quality/testing.md) (GUT for GDScript-only, gdUnit4 when C# or IDE test adapters matter), with `test/` scaffolded and one real test passing under `--headless`?

## Export And CI

- [ ] Is an export preset created for each shipping platform, with `export_presets.cfg` committed and no keystore, password, or other credential anywhere in the repo?
- [ ] Is the [CI workflow template](../templates/github-workflows-ci.yml) copied in and wired per [../recipes/set-up-ci-export.md](../recipes/set-up-ci-export.md), running format, lint, headless tests, `--headless --import`, and per-preset export on every push and pull request?
- [ ] Does each export job assert the artifact exists and has plausible size, rather than trusting job status alone?
- [ ] Does the repo README link to [../AGENTS.md](../AGENTS.md) or its own equivalent fast-path contract?

## Proof

- [ ] `godot --headless --import && godot --headless --export-release "<Preset>" build/<artifact>` succeeds from a clean checkout.
- [ ] `git check-ignore .godot` succeeds and `git lfs track` lists every binary asset type present.
- [ ] `git ls-files | grep -E '\.(scn|res|translation)$'` returns nothing.
- [ ] CI is green on the first PR, including an uploaded export artifact per shipping preset.
