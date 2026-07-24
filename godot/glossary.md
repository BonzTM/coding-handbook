# Glossary

> **Lookup aid, not required reading.** Consult a single entry when a handbook term is unclear; nothing here needs to be read front-to-back to build.

Canonical vocabulary for this handbook. Every term below has exactly one meaning across all repos; use the word for nothing else, and define it nowhere else. Each entry is one or two sentences plus the doc that owns the full rule — read that doc before relying on the term. Start from [README.md](README.md) for how these pieces fit together.

Terms are alphabetical.

## Autoload
A script or scene registered in Project Settings that loads before the running scene and persists across scene changes. It is not a true singleton — it can still be instanced again — and it must never be freed at runtime; reserve it for wide-scope systems that own their own data, never as a dumping ground for shared state. Owned by [foundations/scene-and-node-architecture.md](foundations/scene-and-node-architecture.md); added via [recipes/add-an-autoload.md](recipes/add-an-autoload.md).

## Call down, signal up
The dependency direction for node communication: parents call methods on children they own; children report what happened by emitting signals, and never reach up or sideways for references. The name is community convention; the official scene-organization guidance prescribes the same direction of dependency without naming it. Owned by [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md).

## Export preset
A per-platform export configuration, stored in `export_presets.cfg` and addressed by name from the CLI (`--export-release "Preset Name"`). A release is only reproducible when the preset is committed, not hand-built in the editor per machine. Owned by [operations/ci-and-release.md](operations/ci-and-release.md); created via [recipes/set-up-ci-export.md](recipes/set-up-ci-export.md).

## Export template
The precompiled engine binary for a target platform that an export wraps around the project's packed resources. Templates are version-matched to the editor and must be installed on any machine — including the CI runner — before a headless export can run. Owned by [operations/ci-and-release.md](operations/ci-and-release.md).

## Group
Tag-like membership on nodes (snake_case names) used for one-to-many broadcast via `get_tree().call_group()`. It covers the "notify many peers" case where wiring individual signals would be unwieldy; it is not a substitute for a signal contract between two nodes. Owned by [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md).

## Import cache (`.godot/`)
The per-machine folder the editor regenerates from source assets and project state. It is never committed; ignoring it (plus `*.translation` artifacts) is the first line of the version-control contract. Owned by [foundations/project-setup.md](foundations/project-setup.md); ignore rules in [templates/gitignore.txt](templates/gitignore.txt).

## Input action
A named InputMap abstraction over raw device events (`jump`, not a key code), defined in Project Settings, polled with `Input.is_action_pressed()` or matched in `_unhandled_input()`. Gameplay code references actions only, so devices can be remapped without touching logic. Owned by [foundations/input-handling.md](foundations/input-handling.md); added via [recipes/add-an-input-action.md](recipes/add-an-input-action.md).

## Node
The engine's base building block: one behavior, a PascalCase name, a place in exactly one tree. Composition of small nodes — not inheritance depth — is the default unit of design. Owned by [foundations/scene-and-node-architecture.md](foundations/scene-and-node-architecture.md).

## PCK
The default packed archive (`.pck`) an export produces alongside the platform binary, holding the project's resources. Its virtual filesystem is case-sensitive even when the host OS is not — the reason all file and folder names are snake_case. ZIP is the alternative when players should be able to read or mod the contents. Owned by [operations/ci-and-release.md](operations/ci-and-release.md).

## Physics tick
The fixed-rate update where `_physics_process(delta)` runs — 60 times per second by default — and the only place code that touches the physics engine belongs. `_process(delta)` runs once per rendered frame at a variable rate and is not synchronized with physics. Owned by [systems/physics-and-movement.md](systems/physics-and-movement.md).

## Resource (custom Resource)
A `class_name` script extending `Resource` with `@export` fields — the typed, inspector-editable container for designer-owned data, saved as `.tres`. Shipped read-only data only: loading a resource file a user can modify can execute embedded scripts, so saves never go through `ResourceLoader`. Owned by [foundations/resources-and-data.md](foundations/resources-and-data.md); the untrusted-load rule lives in [systems/save-and-load.md](systems/save-and-load.md).

## Scene
A saved, self-contained node tree (`.tscn`) designed to have no dependencies on where it is instanced; anything it needs from outside arrives by injection from its parent. The scene is this handbook's unit of reuse and of review. Owned by [foundations/scene-and-node-architecture.md](foundations/scene-and-node-architecture.md); added via [recipes/add-a-scene.md](recipes/add-a-scene.md).

## Signal
Godot's built-in observer mechanism: a node announces something that already happened — named past-tense snake_case (`door_opened`) — and connected listeners react without either side referencing the other. Signals respond to behavior; they never start it — starting behavior is a downward method call. Owned by [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md).

## Signal contract
A handbook term, not an engine concept: the declared surface of a signal — its name, typed parameters, emitting scene, and intended listeners — defined once on the owning scene's root script and treated as a compatibility surface like any API. Owned by [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md); added via [recipes/add-a-signal-contract.md](recipes/add-a-signal-contract.md).

## Theme
The Control-skinning resource holding colors, fonts, font sizes, icons, constants, and StyleBoxes. Default posture: one project-wide theme, type variations for named presets, local overrides only for one-off layout tweaks — a theme describes configuration; each control still applies it. Owned by [systems/ui-and-theming.md](systems/ui-and-theming.md).

## Typed GDScript
GDScript with a static type on every declaration (`var health: int`, `func heal(amount: int) -> void:`), enforced project-wide by treating the `UNTYPED_DECLARATION` warning as an error. Typed code compiles to optimized opcodes and gives real autocompletion; untyped scripts do not pass review. Owned by [foundations/gdscript-style-and-typing.md](foundations/gdscript-style-and-typing.md).

## Related

- [README.md](README.md) — how the scenes, signals, data, and proof gate fit together.
- [AGENTS.md](AGENTS.md) — the fast-path contract that uses this vocabulary.
- [AGENTS.md](AGENTS.md) (## Change Routing) — which doc owns each change surface.
