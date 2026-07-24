# AGENTS.md - Godot Project Contract

This is the authoritative fast-path contract for autonomous agents working in a new Godot repository.
Read this file first: it carries the repo-wide invariants, the change-routing table, and the verification bar. Use [maintainer-reference.md](maintainer-reference.md) when you need slower-path architecture and rationale. This handbook targets the current Godot 4.x stable line; 3.x is a separate long-term-supported branch and is out of scope (`docs.godotengine.org/en/stable/about/release_policy.html`).

## Purpose

- Use this file for repo-wide invariants, change defaults, change-to-file routing, and the verification bar.
- Use [maintainer-reference.md](maintainer-reference.md) for the scene-tree architecture map, lifecycle guidance, test taxonomy, and troubleshooting.
- For the full catalogs, see the [recipes/README.md](recipes/README.md), [checklists/README.md](checklists/README.md), and [decisions/README.md](decisions/README.md) indexes.

## Source Of Truth

- This file is the fast path. More detailed docs refine it; they do not weaken it.
- Engine version, directory layout, naming, and version-control defaults live in [foundations/project-setup.md](foundations/project-setup.md); the settings conventions themselves in [templates/project-settings-conventions.md](templates/project-settings-conventions.md).
- Imported-asset governance — `.import` sidecars, compression choices, and the import cache — lives in [foundations/asset-pipeline.md](foundations/asset-pipeline.md).
- Scene composition and node-tree boundaries live in [foundations/scene-and-node-architecture.md](foundations/scene-and-node-architecture.md); cross-node communication and autoload policy in [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md).
- Data modeling with custom `Resource` classes lives in [foundations/resources-and-data.md](foundations/resources-and-data.md); the save-file boundary — a different trust surface — in [systems/save-and-load.md](systems/save-and-load.md).
- Script style, static typing, and script structure live in [foundations/gdscript-style-and-typing.md](foundations/gdscript-style-and-typing.md); the language-choice decision in [foundations/gdscript-vs-csharp.md](foundations/gdscript-vs-csharp.md).
- Input, UI, audio, animation, game flow, physics, localization, and multiplayer rules live in [foundations/input-handling.md](foundations/input-handling.md) and the `systems/` docs listed in [Slow Path Docs](#slow-path-docs) below.
- Proof expectations live in [quality/testing.md](quality/testing.md); lint and format policy in [quality/linting-and-formatting.md](quality/linting-and-formatting.md); headless export and release in [operations/ci-and-release.md](operations/ci-and-release.md).
- Architecture decisions and their rationale live in [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).
- Team-process docs — [onboarding-and-handoff.md](onboarding-and-handoff.md), [CONTRIBUTING.md](CONTRIBUTING.md), and the [glossary.md](glossary.md) lookup aid — serve humans running the team and the handbook; they are not needed to build or change a game.
- Complete, verify-green exemplar projects under `reference/` are a planned later phase; until they land, mirror the [recipes/](recipes/README.md) and [templates/](templates/README.md) rather than inventing a structure.

## Fast Path

1. Read this file and identify the project shape from [README.md](README.md). For a brand-new build, run [checklists/new-project.md](checklists/new-project.md) first.
2. Route the change through the [Change Routing](#change-routing) table below; do not guess where code belongs.
3. Read the relevant foundations doc before editing code in a new area.
4. Implement with the repo defaults unless the repo has already documented an exception.
5. Prove the change with the narrowest meaningful tests first, then the repo-wide baseline.

## Repo-Wide Invariants

- **Pinned engine**: pin one exact 4.x stable release per repo and record it in [foundations/project-setup.md](foundations/project-setup.md)'s committed places. Patch releases are bugfix-only; minor releases may break compatibility "in very specific areas," so a minor bump is a reviewed change, not a drive-by (`docs.godotengine.org/en/stable/about/release_policy.html`).
- **`snake_case` filesystem**: folders and files are `snake_case` (C# scripts excepted — PascalCase per class); node names are PascalCase. Exported PCK paths are case-sensitive while dev filesystems usually are not, so inconsistent casing breaks exported builds (`docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html`).
- **Self-contained scenes**: design scenes to have no external dependencies; when a scene needs outside context, the parent injects it (signal connection, method call, node reference, or NodePath) — children never reach up the tree (`docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`).
- **Call down, signal up**: parents call methods on children; children emit signals upward. The phrase is community shorthand (GDQuest and others), but the pattern is the official scene-organization guidance.
- **Autoload restraint**: an autoload must be a wide-scope system that owns its own data; scene-local nodes, `static` funcs on `class_name` scripts, or shared `Resource`s are the defaults it must beat, with the rationale written down (`docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`). Never `free()` or `queue_free()` an autoload at runtime — the engine crashes (`docs.godotengine.org/en/stable/tutorials/scripting/singletons_autoload.html`).
- **Typed GDScript**: every declaration is typed or inferred; enforce with the `UNTYPED_DECLARATION` warning in project settings. Typed code also compiles to optimized opcodes (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/static_typing.html`).
- **Actions, not keycodes**: gameplay code reads InputMap actions, never raw keys or buttons; discrete gameplay events go through `_unhandled_input` so GUI gets first refusal (`docs.godotengine.org/en/stable/tutorials/inputs/inputevent.html`).
- **Physics on the physics tick**: anything touching the physics engine runs in `_physics_process`; all movement is scaled by `delta` (`docs.godotengine.org/en/stable/tutorials/scripting/idle_and_physics_processing.html`).
- **Never load untrusted resource files**: `.tres`/`.tscn`/`.res`/`.scn` can carry embedded scripts that execute on load, so `ResourceLoader.load()` is forbidden on user-writable or user-shared files — saves, mods, UGC (`github.com/godotengine/godot/pull/98168`). Save data uses the formats in [systems/save-and-load.md](systems/save-and-load.md).
- **VCS hygiene**: `.godot/` and `*.translation` are never committed; Git LFS covers binary assets and is set up before their first commit (`docs.godotengine.org/en/stable/tutorials/best_practices/version_control_systems.html`).
- **Third-party code lives in `addons/`**: with an explicit reason for every addon; nothing else earns a place there by reflex.

## Change Routing

Use this when you know what kind of change you are making but not the file set. Start Here is what you read and touch first; Also Update is the sync surface the change normally drags along; Verify Or Confirm is the proof.

| Change Type | Start Here | Also Update | Verify Or Confirm |
|---|---|---|---|
| Engine version, `project.godot` settings, directory layout | `project.godot`, [foundations/project-setup.md](foundations/project-setup.md), [templates/project-settings-conventions.md](templates/project-settings-conventions.md) | CI engine version, export templates version, `.gitignore` | project opens headless without script errors; exported build still launches |
| New-repo scaffolding (settings, ignores, lint config, CI) | [checklists/new-project.md](checklists/new-project.md), [templates/README.md](templates/README.md), [templates/gitignore.txt](templates/gitignore.txt), [templates/gdlintrc.txt](templates/gdlintrc.txt), [templates/github-workflows-ci.yml](templates/github-workflows-ci.yml) | the copied artifacts and their `<PLACEHOLDER>` values | CI green on the fresh repo; `gdlint` clean |
| Version control: ignores, attributes, large binary assets | `.gitignore`, `.gitattributes`, [templates/gitignore.txt](templates/gitignore.txt), [foundations/project-setup.md](foundations/project-setup.md) | Git LFS tracking rules, teammates' checkout notes | `.godot/` and `*.translation` untracked; LFS active before the first binary commit |
| Imported asset, `.import` settings, texture or audio compression | the asset's `.import` sidecar, [foundations/asset-pipeline.md](foundations/asset-pipeline.md) | project-wide import defaults, Git LFS patterns, mobile export presets in CI if per-platform import output changes | source asset and its `.import` committed together; `.godot/imported/` untracked; a cold import in CI reproduces the asset |
| New scene, node-tree shape, scene composition | the owning scene directory, [foundations/scene-and-node-architecture.md](foundations/scene-and-node-architecture.md), [recipes/add-a-scene.md](recipes/add-a-scene.md) | parent injection wiring, signal connections, tests | scene instantiates standalone in a headless test; no external dependency at load |
| Signal contract, cross-node or cross-system communication | [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md), [recipes/add-a-signal-contract.md](recipes/add-a-signal-contract.md) | every connect site, event-bus autoload if one exists, tests | emit and connect covered by a test; no child reaching up the tree |
| Autoload addition or scope change | `project.godot` `[autoload]`, [recipes/add-an-autoload.md](recipes/add-an-autoload.md), [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md) | consumers, [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md) if it owns cross-scene state | written rationale for why a scene-local node, static func, or shared Resource is insufficient |
| Designer-editable data, custom `Resource` classes, `.tres` files | [foundations/resources-and-data.md](foundations/resources-and-data.md) | `@export` surfaces, consuming scenes, save compatibility if the data is persisted | typed loads; resource edits round-trip through the inspector |
| GDScript style, typing, script structure | [foundations/gdscript-style-and-typing.md](foundations/gdscript-style-and-typing.md), [templates/gdlintrc.txt](templates/gdlintrc.txt) | `.gdlintrc`, GDScript warning levels in `project.godot` | `gdformat --check` and `gdlint` clean; zero `UNTYPED_DECLARATION` warnings |
| Language choice, introducing C#, cross-language boundary | [foundations/gdscript-vs-csharp.md](foundations/gdscript-vs-csharp.md), [decisions/README.md](decisions/README.md) | ADR, CI (.NET editor build), export target list — C# cannot export to web | ADR recorded; every shipping platform still exports |
| Input action, binding, device support | `project.godot` `[input]`, [foundations/input-handling.md](foundations/input-handling.md), [recipes/add-an-input-action.md](recipes/add-an-input-action.md) | gameplay scripts polling the action, rebinding UI, player-facing docs | no raw keycodes in gameplay code; action fires in a test or scripted smoke run |
| UI screen, Control layout, theming | [systems/ui-and-theming.md](systems/ui-and-theming.md), [recipes/add-a-ui-screen.md](recipes/add-a-ui-screen.md) | project-wide theme resource, localization keys, focus and navigation order | screen renders at minimum and maximum supported resolutions; theme overrides stay local-only for one-offs |
| Audio playback, buses, mixing | [systems/audio.md](systems/audio.md) | bus layout, scene-local players, persisted volume settings | sound plays through the intended bus; no global players left orphaned across scene changes |
| Animation, tweening, AnimationTree state machines | [systems/animation.md](systems/animation.md) | AnimationTree parameters and `travel()` call sites, method-track targets, [systems/physics-and-movement.md](systems/physics-and-movement.md) if root motion moves a body | one owner per animated property; no `AnimationPlayer.play()` while an AnimationTree is active |
| Scene transition, pause, loading screen, quit path | [systems/game-flow.md](systems/game-flow.md) | transition autoload ([recipes/add-an-autoload.md](recipes/add-an-autoload.md)), `process_mode` on pause-sensitive nodes, tests | `change_scene_*` return values checked; pause reset before every transition; quit routes through `NOTIFICATION_WM_CLOSE_REQUEST` |
| Movement, collision, physics behavior | [systems/physics-and-movement.md](systems/physics-and-movement.md) | collision layers and masks, `_physics_process` bodies, tests | movement is delta-scaled and framerate-independent; physics logic runs on the physics tick |
| Save format, save field, persistence | [systems/save-and-load.md](systems/save-and-load.md), [recipes/add-a-save-field.md](recipes/add-a-save-field.md) | save version, migration path, defaults for old saves, tests | round-trip test passes; an old save still loads; no `ResourceLoader.load()` on user-writable files |
| Translations, locale support, string changes | [systems/localization.md](systems/localization.md) | translation source files, auto-translate modes on Controls, asset remaps | new strings keyed; layout survives pseudolocalized longer strings |
| Multiplayer, RPC, replication | [systems/multiplayer.md](systems/multiplayer.md) | `@rpc` signatures on both peers, spawner/synchronizer config, server-side validation | negative test: server rejects invalid client input; RPC signatures match on both peers |
| Unit or scene test | [quality/testing.md](quality/testing.md), [recipes/add-a-unit-test.md](recipes/add-a-unit-test.md) | CI test stage, fixtures and doubles | test runner passes under `--headless` locally and in CI |
| Lint or format policy | [quality/linting-and-formatting.md](quality/linting-and-formatting.md), [templates/gdlintrc.txt](templates/gdlintrc.txt) | `.gdlintrc`, CI lint stage, pre-commit hooks | `gdlint` and `gdformat --check` clean repo-wide |
| CI, export presets, release | `.github/workflows/`, `export_presets.cfg`, [operations/ci-and-release.md](operations/ci-and-release.md), [recipes/set-up-ci-export.md](recipes/set-up-ci-export.md), [templates/github-workflows-ci.yml](templates/github-workflows-ci.yml) | export templates version, [templates/pull_request_template.md](templates/pull_request_template.md), release notes | headless export succeeds for every preset; the exported artifact launches |
| Release cut, version tag, shipping artifacts | [checklists/release.md](checklists/release.md), [operations/ci-and-release.md](operations/ci-and-release.md) | changelog, player-facing version string, save-migration fixtures if the schema changed | every shipped artifact came from a tag-triggered CI export; the checklist passes top to bottom before anything reaches players |
| PR review of any change above | [checklists/pr-review.md](checklists/pr-review.md) | the files failing boxes point to, routed through the matching row of this table | every box tied to evidence — a diff, a setting, or a passing command |
| Performance regression, profiling, draw calls | [operations/performance-and-profiling.md](operations/performance-and-profiling.md) | material reuse and batching, MultiMesh candidates, captured profiles | profiler evidence before and after; frame budget held on target hardware |
| Non-obvious or hard-to-reverse architecture decision | [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md), [decisions/README.md](decisions/README.md) | any superseded ADR, README/onboarding links | ADR recorded with status, alternatives, and consequences before the change merges |
| Project ownership transfer or onboarding | [onboarding-and-handoff.md](onboarding-and-handoff.md), [checklists/handoff.md](checklists/handoff.md), [README.md](README.md) | store credentials and signing keys, CI secrets, open decisions | new owner runs the full baseline gate and a headless export unaided |
| Handbook doc change (this contract, indexes, glossary) | [CONTRIBUTING.md](CONTRIBUTING.md) | this file's routing table, [README.md](README.md), [maintainer-reference.md](maintainer-reference.md), [glossary.md](glossary.md), [recipes/README.md](recipes/README.md), [checklists/README.md](checklists/README.md), [decisions/README.md](decisions/README.md), [templates/README.md](templates/README.md) | no dead links; fast path, slow path, and recipes agree |

## High-Value Boundaries

- `project.godot` owns registered settings: autoloads, the input map, physics tick rate, GDScript warning levels, and the main scene. It changes through recipes, not ad hoc edits.
- Each scene owns its subtree and exposes a contract of signals, methods, and exported properties; consumers use the contract, never internal node paths.
- Autoloads own cross-scene services and nothing else; per-scene concerns stay in the scene.
- Custom `Resource` classes own shipped, read-only designer data; the save system owns user-writable persistence — the two never share a loader.
- `addons/` owns third-party code; project code does not live there and addon code is not edited in place.
- `export_presets.cfg` owns per-platform output; platform conditionals in gameplay code are a smell to route through it or project settings.

## Proof Hints

- Scene changes usually need a headless instantiation test plus one manual or scripted smoke run of the affected flow.
- Save changes are not done until an old save still loads; write the migration test with the field.
- Multiplayer changes need a negative test — the docs' rule is to treat all client input as untrusted (`docs.godotengine.org/en/stable/tutorials/networking/high_level_multiplayer.html`).
- Performance claims need profiler captures, not intuition; pooling and batching decisions are made from evidence ([operations/performance-and-profiling.md](operations/performance-and-profiling.md)).
- Export-affecting changes (casing, presets, C#, ICU-dependent text) are not done until a headless export succeeds and the artifact launches.

## Working Norms

- Prefer small, reviewable changes over broad cleanup; text-format scenes and resources diff cleanly — keep them that way by saving with the pinned editor version.
- Do not introduce new architecture because it feels cleaner; match the repo's current shape unless the task is explicitly architectural.
- Do not bypass boundaries: gameplay code does not poll raw input, scenes do not grab ancestors with absolute node paths, and autoloads do not accumulate unrelated state.
- When adding an addon or an autoload, document why the built-in or scene-local alternative is insufficient.
- When behavior changes, write the failing or proving test before claiming success whenever practical.
- If verification fails, fix it or report it clearly. Do not claim the change is done.

## Baseline Verification

| Goal | Command | Expectation |
|---|---|---|
| format | `gdformat --check <script dirs>` | no diff (policy in [quality/linting-and-formatting.md](quality/linting-and-formatting.md)) |
| lint | `gdlint <script dirs>` | exit code 0 against the committed `.gdlintrc` |
| tests | the repo's test runner under `godot --headless` | all pass (runner selection owned by [quality/testing.md](quality/testing.md)) |
| export smoke | `godot --headless --export-release "<Preset>" <output>` | export succeeds; artifact launches |
| file-specific correctness | targeted tests from the relevant recipe or systems doc | pass with expected assertions |

The committed [templates/github-workflows-ci.yml](templates/github-workflows-ci.yml) runs this same gate in CI; run the same commands locally so local and CI stay identical. Headless CLI export and the `--headless` flag are the official mechanism (`docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`); export templates for the pinned engine version must be installed wherever the gate runs.

## Slow Path Docs

- Architecture and tree map: [maintainer-reference.md](maintainer-reference.md)
- Setup and layout: [foundations/project-setup.md](foundations/project-setup.md), [foundations/asset-pipeline.md](foundations/asset-pipeline.md)
- Runtime correctness: [foundations/scene-and-node-architecture.md](foundations/scene-and-node-architecture.md), [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md), [foundations/input-handling.md](foundations/input-handling.md)
- Game systems: [systems/ui-and-theming.md](systems/ui-and-theming.md), [systems/audio.md](systems/audio.md), [systems/animation.md](systems/animation.md), [systems/game-flow.md](systems/game-flow.md), [systems/physics-and-movement.md](systems/physics-and-movement.md), [systems/save-and-load.md](systems/save-and-load.md), [systems/localization.md](systems/localization.md), [systems/multiplayer.md](systems/multiplayer.md)
- Proof and verification: [quality/testing.md](quality/testing.md), [operations/ci-and-release.md](operations/ci-and-release.md), [operations/performance-and-profiling.md](operations/performance-and-profiling.md)
