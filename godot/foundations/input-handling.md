# Input Handling

Input boundary defaults for the current Godot 4.x stable line: InputMap actions as the only input vocabulary gameplay code sees, a fixed consumption order so UI and gameplay never fight over events, and rebinding support designed in from day one.

## Default Approach

- Gameplay code reads named InputMap actions, never raw keycodes, buttons, or axes.
- UI gets first refusal on every event; gameplay listens in `_unhandled_input()` so a focused Control consuming a click or keypress silently wins.
- Poll `Input` for continuous state each tick; use event callbacks for discrete, edge-triggered gameplay events.
- Treat the action list in `project.godot` as a contract surface: adding, renaming, or removing an action follows [../recipes/add-an-input-action.md](../recipes/add-an-input-action.md).

### InputMap Actions Only

Define every player-facing input as a named action under Project Settings > Input Map, and reference only the action name in scripts. The official docs call this out directly: "it is cleaner and more flexible to use the provided InputMap feature, which allows you to define input actions and assign them different keys" — actions let "the same code to work on different devices with different inputs (e.g., keyboard on PC, Joypad on console)" and allow "Input to be reconfigured at runtime" (`docs.godotengine.org/en/stable/tutorials/inputs/inputevent.html`).

Rules for the action list:

- Action names are `snake_case` verb phrases (`jump`, `move_left`, `open_inventory`) — a convention this handbook adopts to match GDScript identifier casing in [gdscript-style-and-typing.md](gdscript-style-and-typing.md), not an official mandate.
- Every action ships with both a keyboard/mouse binding and a joypad binding unless the design explicitly excludes a device class; document the exclusion in the PR.
- Checking a raw `keycode`, `button_index`, or scancode in gameplay code is forbidden. The only place raw events are inspected is the rebinding capture flow below and engine-level debug tooling.
- Engine defaults like `ui_accept` and `ui_cancel` belong to the UI layer; do not overload them as gameplay actions. Gameplay defines its own action names so [../systems/ui-and-theming.md](../systems/ui-and-theming.md) screens and gameplay can be rebound independently.

### Event Flow And Consumption

The Viewport delivers each event through a fixed pipeline, and a consumer at any stage stops it. Per the stable docs, the order is: `Node._input()` overrides, then the GUI (Control `_gui_input()`), then `Node._shortcut_input()`, then `Node._unhandled_key_input()`, then `Node._unhandled_input()`, then physics object picking. "If any function consumes the event, it can call Viewport.set_input_as_handled(), and the event will not spread any more"; a Control consumes with `Control.accept_event()` instead (`docs.godotengine.org/en/stable/tutorials/inputs/inputevent.html`). Outside the Control hierarchy, propagation is reverse depth-first — deepest nodes see the event first.

Route by layer, not by convenience:

- `_input()` — global capture only: pause toggles, screenshot keys, the rebinding capture flow. Almost nothing else earns this stage, because it runs before the GUI and steals events from focused Controls.
- `_gui_input()` — Control-internal behavior only; owned by [../systems/ui-and-theming.md](../systems/ui-and-theming.md).
- `_shortcut_input()` — editor-style keyboard shortcuts on tool UIs.
- `_unhandled_input()` — the default home for discrete gameplay input. Anything the GUI consumed never arrives here, which is exactly the behavior you want when a menu is open.
- A node that acts on an event calls `set_input_as_handled()` in the same frame. Never let two systems both act on one event; if two systems appear to need the same event, the deeper one owns it and signals upward per [signals-and-decoupling.md](signals-and-decoupling.md).

### Polling Versus Events

Two access patterns, chosen by the shape of the input, never mixed for one action:

- **Continuous state** (movement, aiming, held modifiers): poll `Input.is_action_pressed()` inside `_physics_process()` so the value is sampled once per physics tick, in step with the movement it drives — see [../systems/physics-and-movement.md](../systems/physics-and-movement.md) for the tick model.
- **Discrete events** (jump, interact, confirm): handle the `InputEvent` in `_unhandled_input()` with `event.is_action_pressed("jump")`, then consume it. Event handling fires exactly once per press regardless of frame rate; polling a discrete action in `_process()` can double-fire or drop presses across frame-rate changes.

```gdscript
func _physics_process(delta: float) -> void:
	if Input.is_action_pressed("move_left"):
		_velocity.x = -_speed


func _unhandled_input(event: InputEvent) -> void:
	if event.is_action_pressed("jump"):
		_try_jump()
		get_viewport().set_input_as_handled()
```

Do not poll from deep scene-tree leaves for state a parent already owns; inject input-derived intent downward per [scene-and-node-architecture.md](scene-and-node-architecture.md) so leaf scenes stay testable without a live input device.

### Rebinding Support

Rebinding is a default capability, not a stretch goal — the InputMap-only rule exists precisely so bindings can change without touching gameplay code. The `InputMap` singleton mutates actions at runtime: `action_erase_events()` clears an action's bindings, `action_add_event()` assigns a captured event, `action_get_events()` reads current bindings for the settings UI, and `has_action()` guards against typos (`docs.godotengine.org/en/stable/classes/class_inputmap.html`).

- Capture the new binding in a dedicated settings screen using `_input()`, store the raw `InputEvent`, and call `set_input_as_handled()` so the captured press does not leak into the game.
- Runtime InputMap changes are in-memory only; `InputMap.load_from_project_settings()` "Clears all InputEventAction in the InputMap and load it anew from ProjectSettings" (`docs.godotengine.org/en/stable/classes/class_inputmap.html`). Persist player bindings yourself — serialize per action, apply on boot, and reset to defaults by reloading from project settings. Storage format and location are owned by [../systems/save-and-load.md](../systems/save-and-load.md).
- Never rebind the UI-navigation actions (`ui_*`) from the same screen that gameplay rebinding uses; a broken `ui_accept` binding can lock the player out of the menu that fixes it.
- Reject a capture that collides with an existing action's binding, or surface the conflict for explicit confirmation — silent duplicate bindings are a support-ticket generator.

## Common Mistakes And Forbidden Patterns

- Matching raw keycodes or joypad button indices in gameplay scripts instead of action names — it breaks device portability and makes rebinding impossible.
- Reading gameplay input in `_input()` so it fires even while a menu, text field, or console has focus.
- Polling a discrete action every frame instead of handling its event — double-fires on high frame rates, drops presses on low ones.
- Acting on an event without calling `set_input_as_handled()`, so a deeper node and an ancestor both respond to one press.
- Overloading `ui_accept`/`ui_cancel` as gameplay actions, coupling menu navigation to gameplay rebinds.
- Shipping actions with keyboard bindings only and bolting on joypad support later — every action gets both device bindings at creation.
- Treating runtime `InputMap` mutations as persistent — they vanish on restart unless the save layer writes them.
- Scattering `Input` polls across leaf nodes instead of deriving intent once and passing it down.

## Verification And Proof

- `grep` the scripts directory for `keycode`, `button_index`, and `physical_keycode`: hits outside the rebinding capture flow are violations.
- Every action referenced in code exists in `project.godot` and vice versa — a headless script asserting `InputMap.has_action()` for each name catches drift; wire it into the test suite per [../quality/testing.md](../quality/testing.md).
- A unit test drives discrete behavior by synthesizing `InputEventAction` events rather than real device events, proving gameplay depends only on action names.
- Manual proof for consumption order: open the pause menu and confirm gameplay input is dead while it has focus, with no `set_input_as_handled()` calls added ad hoc to make it so.
- Rebind an action in the settings screen, restart the game, and confirm the binding survives; then reset to defaults and confirm `project.godot` bindings return.
