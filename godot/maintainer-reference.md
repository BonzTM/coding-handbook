# Maintainer Reference

Purpose: hold slower-path architecture, scene-map, lifecycle, engine-version, and rationale guidance that is useful but not worth loading for every task.
Audience: maintainers and agents working in Godot repositories that use this handbook.
Read [AGENTS.md](AGENTS.md) first. Use this file when you need the fuller background behind the fast-path rules.

## Architecture Snapshot

This handbook assumes a single Godot 4.x project pinned to one stable minor release, organized feature-first: assets live next to the scenes that use them, per the official project-organization guidance at `docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html` ("group assets as close to scenes as possible"). The dominant shape is:

```text
repo/
  project.godot
  export_presets.cfg
  .gitignore            # ignores .godot/ and *.translation
  .gitattributes        # Git LFS patterns for binary assets
  addons/               # third-party only; never hand-edited
  autoload/             # scripts and scenes registered as autoloads
  characters/
    player/             # player.tscn, player.gd, sprites, sfx together
    enemies/
  levels/
  ui/
    theme/              # the single project-wide Theme resource
  data/                 # custom Resource scripts and .tres instances
  test/
    unit/
    integration/
```

All folder and file names are `snake_case` (C# scripts excepted); the exported PCK filesystem is case-sensitive while Windows and macOS filesystems typically are not, so inconsistent casing breaks exported builds, not editor runs. Compiling reference projects embodying this layout are a planned later phase; until they land, [foundations/project-setup.md](foundations/project-setup.md) is the authority on layout details.

## Two-Speed Documentation Model

- Fast path: [AGENTS.md](AGENTS.md) for invariants, the task loop, change-type-to-file-set routing, and baseline proof.
- Slow path: this file for architecture, scene map, lifecycle, test taxonomy, version policy, and rationale.

Use the fast path for most tasks. Use this file when a change crosses scenes, introduces new runtime behavior, or challenges an existing default.

## Scene And Directory Map

| Area | Owns | Must Not Own |
|---|---|---|
| `Main` scene | entry point, primary controller, swapping `World` children to change levels, top-level `World`/`GUI` split | gameplay rules, per-entity state |
| `autoload/` | cross-scene systems that own their data (scene transitions, save orchestration, event bus if adopted) | per-scene state, anything a single scene could own, pooled resources sized for one scene |
| `characters/`, `levels/` | self-contained feature scenes with their scripts and assets co-located | reaching out of their own subtree; external context arrives by dependency injection from the parent |
| `ui/` | Control scenes, the project-wide Theme, type variations | game logic beyond presentation and input forwarding |
| `data/` | `class_name` Resource scripts and shipped `.tres` instances (stats, items, config) | runtime save files, user-writable data |
| `addons/` | third-party plugins as imported | local patches; fork upstream instead |
| `test/` | GUT or gdUnit4 suites and test doubles | production scripts or assets |

The `Main`-as-primary-controller shape and the design rule that scenes "have no dependencies" — parents inject what children need via signals, method calls, `Callable` properties, references, or NodePaths — come from `docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`. The autoload boundary comes from `docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`: global state centralizes failure and hides bug sources ("there's no longer an easy way to find the source of a bug"), but "autoloaded nodes can simplify your code for systems with a wide scope." Ownership of the decoupling rules lives in [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md).

## Lifecycle Model

For any node entering the running game, the order is:

1. `_init` runs on object construction, before the node is in the tree.
2. `_enter_tree` runs when the node joins the SceneTree; these calls "cascade down the tree" — parents before children.
3. `_ready` runs once per node after all its children have finished theirs — "a reverse cascade going up back to the tree's root". Wire dynamic signal connections here.
4. Per frame: `_process(delta)` at the variable render rate; `_physics_process(delta)` on the fixed physics tick, 60 Hz by default. Physics-interacting movement belongs in `_physics_process`, always scaled by `delta`; "`_process()` is not synchronized with physics."
5. Input flows `_input()` → Control `_gui_input()` → `_shortcut_input()` → `_unhandled_key_input()` → `_unhandled_input()` → physics picking; a consumer calls `Viewport.set_input_as_handled()` to stop propagation. Gameplay reads actions, not raw events, in `_unhandled_input` so GUI gets first refusal.
6. `_exit_tree` on removal; `NOTIFICATION_PREDELETE` fires "before the engine deletes an Object, i.e. a 'destructor'". Autoloads must never be freed at runtime "or the engine will crash".

Ordering and notification semantics: `docs.godotengine.org/en/stable/tutorials/best_practices/godot_notifications.html`. Process split: `docs.godotengine.org/en/stable/tutorials/scripting/idle_and_physics_processing.html`. Input flow: `docs.godotengine.org/en/stable/tutorials/inputs/inputevent.html`. Autoload crash warning: `docs.godotengine.org/en/stable/tutorials/scripting/singletons_autoload.html`.

## Test Taxonomy

| Test Type | Default Location | What It Proves |
|---|---|---|
| unit tests | `test/unit/` | pure logic in scripts and custom Resources, no SceneTree required |
| scene tests | `test/unit/` | a scene instantiates cleanly, `_ready` wiring holds, signals fire with the declared payloads |
| scene-runner / simulated-input tests | `test/unit/`; `test/integration/` when simulating physics ticks | input actions drive the expected state transitions frame by frame |
| integration tests | `test/integration/` | multi-scene interaction, physics-tick behavior, save/load round-trips, autoload-backed systems |
| C# tests | `test/` via gdUnit4Net | C# gameplay logic with IDE test-adapter integration |
| headless smoke run | CI job | the project boots and exports under `--headless` with the pinned engine version |

Default framework: GUT (9.x targets Godot 4.x) for GDScript-only projects; gdUnit4 when C# tests or IDE integration matter — both run headless in CI (`github.com/bitwes/Gut`, `github.com/godot-gdunit-labs/gdUnit4`). The important principle is not "more tests"; it is the right tests at the right boundary — a unit test on a stats Resource does not replace a scene test proving the HUD actually reacts to `health_changed`. Selection rationale and structure are owned by [quality/testing.md](quality/testing.md).

## Runtime Contracts Worth Remembering

- Every signal connection must have an owner and a declared payload shape; signals respond to what happened, they do not command what happens next.
- Every autoload must justify its wide scope and must never be freed at runtime.
- Every physics-interacting movement must run in `_physics_process` and scale by `delta`.
- Every file path must be `snake_case`, because the exported PCK is case-sensitive and the editor filesystem may not be.
- Loading a `.tres`/`.tscn`/`.res`/`.scn` file can execute embedded GDScript; never `ResourceLoader.load()` a file a user can modify — user-writable saves use JSON, `ConfigFile`, or `store_var` with validation (`github.com/godotengine/godot/pull/98168`).
- GDScript and C# cannot inherit across the language boundary, and C# projects cannot export to the web platform (`docs.godotengine.org/en/stable/tutorials/scripting/c_sharp/index.html`).

## Contract Surfaces

- Signal names and payloads are API contracts between scenes; changing one drags every connection site and its tests along.
- InputMap action names are the contract between devices and gameplay; raw keycodes in scripts bypass it and break rebinding.
- Save-file schemas are data contracts with every past player; they need explicit versioning and migration rules, owned by [systems/save-and-load.md](systems/save-and-load.md).
- Translation keys are contracts with every locale CSV or PO file; renaming a key orphans its translations.
- `@rpc` methods are cross-peer contracts — "Both RPCs must have the same signature which is evaluated with a checksum" on both peers, and all client input is untrusted (`docs.godotengine.org/en/stable/tutorials/networking/high_level_multiplayer.html`).
- Export presets in `export_presets.cfg` are the contract between the repo and CI; the CLI exports whatever preset name CI passes to `--export-release`.

## Engine Version Policy

- Pin one 4.x minor per project and record it; the current stable release is listed at `godotengine.org/download/archive/`. Editor, export templates, and CI must use the identical version — templates are version-matched.
- Minor releases (one 4.x minor to the next) add features, and "minor compatibility breakage in very specific areas *may* happen"; treat a minor upgrade as a change with its own verification pass, not a drive-by. Patch releases fix bugs without breaking compatibility and are safe to take promptly.
- A stable branch is supported "at least until the next stable branch is released and has received its first patch update" — do not sit more than one minor behind stable.
- Godot 3.x is the long-term-supported branch; this handbook targets 4.x only.

Source: `docs.godotengine.org/en/stable/about/release_policy.html`. Upgrade mechanics and CI pinning are owned by [operations/ci-and-release.md](operations/ci-and-release.md).

## Common Failure Modes

| Symptom | Likely Cause | First Fix |
|---|---|---|
| works in editor, exported build cannot find a file | path casing mismatch; exported PCK is case-sensitive | rename to `snake_case`, fix references, re-export |
| engine crash on scene change or quit | `free()`/`queue_free()` called on an autoload | remove the free; autoloads live for the process lifetime |
| bug source untraceable, everything touches one global | autoload overuse for state single scenes could own | push state down to scenes; use signals or `static` funcs on `class_name` scripts |
| scene breaks when instanced in a new context | hidden external dependency (`get_node("/root/...")` or hard-coded sibling paths) | invert it: parent injects the dependency at instantiation |
| jittery or framerate-dependent movement | physics movement in `_process`, or `delta` not applied | move to `_physics_process(delta)` and scale by `delta` |
| noisy diffs, unmergeable repo | `.godot/` committed, binary assets not in LFS | gitignore `.godot/` and `*.translation`; add LFS patterns before committing binaries |
| arbitrary code execution via shared saves or mods | `ResourceLoader.load()` on user-writable files | switch to JSON/`ConfigFile`/`store_var` with validation |
| GUI swallows gameplay input or vice versa | gameplay listening in `_input` ahead of Controls | move gameplay to `_unhandled_input`; consume with `set_input_as_handled()` |

## Primary Sources Behind These Defaults

- release and support policy: `https://docs.godotengine.org/en/stable/about/release_policy.html`
- best-practices index: `https://docs.godotengine.org/en/stable/tutorials/best_practices/index.html`
- project organization: `https://docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html`
- scene organization and dependency injection: `https://docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`
- autoload boundaries: `https://docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`
- notification and callback ordering: `https://docs.godotengine.org/en/stable/tutorials/best_practices/godot_notifications.html`
- process vs physics tick: `https://docs.godotengine.org/en/stable/tutorials/scripting/idle_and_physics_processing.html`
- input event flow: `https://docs.godotengine.org/en/stable/tutorials/inputs/inputevent.html`
- GDScript style and static typing: `https://docs.godotengine.org/en/stable/tutorials/scripting/gdscript/gdscript_styleguide.html`, `https://docs.godotengine.org/en/stable/tutorials/scripting/gdscript/static_typing.html`
- saving games and the untrusted-resource caveat: `https://docs.godotengine.org/en/stable/tutorials/io/saving_games.html`, `https://github.com/godotengine/godot/pull/98168`
- headless export: `https://docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`
- version control: `https://docs.godotengine.org/en/stable/tutorials/best_practices/version_control_systems.html`
- multiplayer contracts: `https://docs.godotengine.org/en/stable/tutorials/networking/high_level_multiplayer.html`
- test frameworks: `https://github.com/bitwes/Gut`, `https://github.com/godot-gdunit-labs/gdUnit4`

## Related Docs

- Fast path and change routing: [AGENTS.md](AGENTS.md)
- Project layout: [foundations/project-setup.md](foundations/project-setup.md)
- Asset import pipeline: [foundations/asset-pipeline.md](foundations/asset-pipeline.md)
- Scene composition rules: [foundations/scene-and-node-architecture.md](foundations/scene-and-node-architecture.md)
- Signal contracts: [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md)
- Data and Resources: [foundations/resources-and-data.md](foundations/resources-and-data.md)
- Proof and testing: [quality/testing.md](quality/testing.md)
- Engine pinning in CI: [operations/ci-and-release.md](operations/ci-and-release.md)
- Save schemas: [systems/save-and-load.md](systems/save-and-load.md)
- Animation ownership: [systems/animation.md](systems/animation.md)
- Scene transitions and game flow: [systems/game-flow.md](systems/game-flow.md)
