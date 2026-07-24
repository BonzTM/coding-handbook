# Recipe: Add An Autoload

Use this only after the decision in [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md) (### Autoloads Versus Regular Nodes) says an autoload is justified: a wide-scope system that owns its own data and no other object's (`docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`). If the need is only carrying events across unrelated scene branches, add an event bus under the stricter rules in [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) (### Event Bus Boundaries) instead.

## Files To Touch

- `autoload/<system_name>.gd` — `snake_case` filename; the `autoload/` directory is the handbook layout convention in [../foundations/project-setup.md](../foundations/project-setup.md)
- `project.godot` `[autoload]` block, edited through the Project Settings Autoload UI, not by hand
- every consumer that calls the new singleton or connects to its signals
- [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md) — an ADR when the autoload owns cross-scene state, per the routing row in [../AGENTS.md](../AGENTS.md) (## Change Routing)
- tests for the autoload's public API, per [../quality/testing.md](../quality/testing.md)

## Steps

1. Write down why each official alternative fails for this system — scene-local nodes (e.g. per-scene `AudioStreamPlayer`s), `static` funcs on a `class_name` script, a shared custom `Resource` (`docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`). This rationale goes in the ADR or the PR description; no rationale, no autoload.
2. Create `autoload/<system_name>.gd` following [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md): typed public methods for the data the system owns, past-tense signals for the state changes it announces.
3. Register it in Project Settings under Autoload with a PascalCase name (autoloads are nodes; node names are PascalCase). Confirm the diff touches only the `[autoload]` block of `project.godot`.
4. Wire consumers through the public surface only: they call the autoload's methods and connect to its signals. The autoload never reaches into scene internals or holds references to scene nodes it does not own.
5. Add headless-runnable tests exercising the public API through a scene-free seam where possible — see [add-a-unit-test.md](add-a-unit-test.md).

## Invariants To Preserve

- the autoload owns its own data and no other object's — no `Globals` grab-bag accreting unrelated state
- never `free()` or `queue_free()` an autoload at runtime; the official docs state the engine will crash (`docs.godotengine.org/en/stable/tutorials/scripting/singletons_autoload.html`); lifetime is the engine's, teardown logic belongs in `_exit_tree()` or an explicit `reset()` the game calls
- an autoload is not a true singleton — it can be instanced again (`docs.godotengine.org/en/stable/tutorials/scripting/singletons_autoload.html`); never `new()` or instance its script or scene manually
- state that must survive quit lives in the save system, not the autoload — [../systems/save-and-load.md](../systems/save-and-load.md)
- an event-bus autoload stays signals-only; the moment it grows state or logic it re-enters this recipe as a system autoload

## Proof

- `gdformat --check autoload/` and `gdlint autoload/` exit clean
- headless test run covering the new public API passes
- `git diff project.godot` shows exactly one new `[autoload]` entry and nothing else
- written rationale exists naming why scene-local, `static` func, and shared `Resource` do not fit
- project search shows no `free()`/`queue_free()` call targeting the autoload and no manual instancing of its script

If consumers need to react to the autoload rather than call it, define the signals via [add-a-signal-contract.md](add-a-signal-contract.md).
