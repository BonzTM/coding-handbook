# Recipe: Add A Scene

Use this when a feature adds one self-contained scene (a character, prop, level chunk, or reusable gameplay unit). For screens built from Control nodes, use [add-a-ui-screen.md](add-a-ui-screen.md) instead.

## Files To Touch

- a new `snake_case` folder next to its closest relatives (e.g. `characters/goblin/`), holding the `.tscn`, its script, and its assets together per `docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html`
- `<scene_name>.tscn` and `<scene_name>.gd` inside that folder
- the owner scene (or spawner script) that instantiates it
- a test file per [add-a-unit-test.md](add-a-unit-test.md)

## Steps

1. Create the folder and scene; name files `snake_case`, the root node PascalCase, and pick the root type by what the scene is (`CharacterBody2D`/`3D`, `Node2D`, `Control`), per [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md).
2. Attach `<scene_name>.gd` to the root, typed per [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md); add `class_name` only if other code refers to the type.
3. Keep the scene dependency-free: it must load and run with no assumptions about its parent or siblings (`docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`).
4. Take external context by injection from the owner — method calls, handed-over node references, `Callable` properties, or NodePaths — never by the scene reaching up or out itself.
5. Declare outbound facts as past-tense signals following [add-a-signal-contract.md](add-a-signal-contract.md); signals report what happened, they do not command the parent.
6. In the owner: `preload` the `PackedScene`, `instantiate()`, inject dependencies, `add_child()`, and connect signals in code — editor-wired connections are only for nodes that exist in the same saved scene (`docs.godotengine.org/en/stable/getting_started/step_by_step/signals.html`).
7. Add a routing row to [../AGENTS.md](../AGENTS.md) only if the scene introduces a new change type; otherwise the existing scene row already covers it.

## Invariants To Preserve

- the scene runs standalone (F6 / Run Current Scene) without errors
- no `get_parent()` calls, absolute `get_node("/root/...")` paths, or autoload reads inside the scene's core logic — cross-scene state goes through [add-an-autoload.md](add-an-autoload.md) deliberately, not by default
- all file and folder names stay `snake_case`; exported PCK filesystems are case-sensitive even when the dev OS is not
- assets used only by this scene live in its folder, not a global `assets/` pile
- signal connections for dynamically instantiated scenes are made in code, not left dangling from a deleted editor wiring

## Proof

- standalone launch of the scene shows no errors in the debugger output
- a unit test instantiates the scene headless, injects fakes for its dependencies, and asserts on emitted signals
- `gdformat --check` and `gdlint` pass on the new script
- the owner-scene test (or a smoke run) shows the instance spawning and its signals firing
