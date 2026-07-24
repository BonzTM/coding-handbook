# Recipe: Add An Input Action

Use this when gameplay needs a new player-facing input (a jump, an interact, a movement axis) and you must register it as an InputMap action with default bindings instead of matching raw keys or buttons. Rules are owned by [../foundations/input-handling.md](../foundations/input-handling.md).

## Files To Touch

- `project.godot` (`[input]` section — edited through Project Settings > Input Map, diff committed with the change)
- the gameplay script that reads the action (e.g. `characters/player/player.gd`)
- the rebinding settings screen's action list, if the action is player-rebindable
- a test asserting the action exists and driving the behavior via a synthesized event (location per [../quality/testing.md](../quality/testing.md))

## Steps

1. Name the action as a `snake_case` verb phrase (`jump`, `open_inventory`) per [../foundations/input-handling.md](../foundations/input-handling.md) (### InputMap Actions Only). Never reuse or overload an engine `ui_*` action for gameplay; check the existing `[input]` section for a name or binding collision before adding.
2. Register the action under Project Settings > Input Map with BOTH a keyboard/mouse binding and a joypad binding at creation. A device-class exclusion (e.g. mouse-only tool) is documented in the PR, not left implicit. The editor writes the action into `project.godot` `[input]`; review that diff — it is the contract surface [../AGENTS.md](../AGENTS.md) routes through this recipe.
3. Wire gameplay to the action NAME only, choosing one access pattern by the input's shape (never both for one action): discrete events in `_unhandled_input()` via `event.is_action_pressed("jump")` followed by `set_input_as_handled()`; continuous state polled with `Input.is_action_pressed()` inside `_physics_process()`. `_unhandled_input()` runs after the GUI stage, so a focused Control consuming the event silently wins — which is why gameplay never listens in `_input()` (`docs.godotengine.org/en/stable/tutorials/inputs/inputevent.html`).
4. Derive intent once at the owning node and pass it down; do not add `Input` polls to leaf scenes for state a parent already owns (see [../foundations/input-handling.md](../foundations/input-handling.md) (### Polling Versus Events) and [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md)).
5. Keep rebinding intact. If the settings screen enumerates rebindable actions from a list, add the new action to that list; confirm the persistence layer serializes its bindings per action and applies them on boot ([../systems/save-and-load.md](../systems/save-and-load.md) owns the storage). Runtime `InputMap` mutations are in-memory only — defaults are restored with `InputMap.load_from_project_settings()` (`docs.godotengine.org/en/stable/classes/class_inputmap.html`), so the committed `project.godot` bindings ARE the reset-to-defaults state. Do not add the action to the rebinding list with an empty or conflicting default.
6. Add the test: assert `InputMap.has_action("jump")` headlessly so code and `project.godot` cannot drift, and drive the discrete behavior by synthesizing an `InputEventAction` rather than a device event, proving gameplay depends only on the action name.

## Invariants To Preserve

- no raw `keycode`, `physical_keycode`, or `button_index` checks in gameplay scripts — action names are the only input vocabulary gameplay sees
- every action ships with keyboard/mouse and joypad bindings at creation; exclusions are documented in the PR
- one access pattern per action: polled continuous state or discrete event handling, never both
- the node that acts on an event consumes it with `set_input_as_handled()` in the same frame
- `ui_*` actions stay owned by the UI layer — not overloaded, not rebindable from the gameplay rebind screen
- runtime `InputMap` changes are never treated as persistent; the save layer writes player bindings, `project.godot` stays the defaults source

## Proof

- `grep -rn 'keycode\|button_index' <scripts dir>` — no new hits outside the rebinding capture flow
- headless test run passes: `InputMap.has_action()` holds for the new name, and the synthesized-`InputEventAction` behavior test fires the gameplay effect (runner per [../quality/testing.md](../quality/testing.md))
- manual device pass: the keyboard binding and the joypad binding both trigger the action; with the pause menu focused, the action is dead with no ad hoc handling added
- rebinding pass: rebind the action in the settings screen, restart, confirm the new binding survives; reset to defaults and confirm the `project.godot` bindings return
- `project.godot` `[input]` diff reviewed in the same PR as the script that reads the action
