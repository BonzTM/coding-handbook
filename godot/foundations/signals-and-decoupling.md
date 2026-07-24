# Signals And Decoupling

This doc owns inter-node communication: how nodes talk without referencing each other, when to use a direct call versus a signal versus a group versus an event bus, and where each mechanism stops being appropriate. Scene structure and ownership rules live in [scene-and-node-architecture.md](scene-and-node-architecture.md); this doc governs the wires between the boxes.

## Default Approach

Godot 4.x. Signals are the engine's built-in observer mechanism: "a delegation mechanism built into Godot that allows one game object to react to a change in another without them referencing one another" (`docs.godotengine.org/en/stable/getting_started/step_by_step/signals.html`).

Pick the communication mechanism with this escalation ladder, and stop at the first rung that works:

1. **Direct method call** — the caller initiates behavior on a node it owns (parent to child). Calls start behavior.
2. **Signal** — a node announces something that already happened, and interested parties react (child to parent, or sibling via a mediating parent). The official scene-organization doc is explicit: signals "should be used only to 'respond' to behavior, not start it" (`docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`).
3. **Group** — one sender must reach many receivers that share a role, not a parent (`docs.godotengine.org/en/stable/tutorials/scripting/groups.html`).
4. **Event bus autoload** — the emitter and receivers live in unrelated scene branches and no reasonable ancestor can mediate. This rung requires the boundaries in [Event Bus Boundaries](#event-bus-boundaries).

Skipping rungs is the failure mode. Every rung you climb trades away debuggability, so each climb needs the lower rung to be demonstrably insufficient, not merely more typing.

Scenes stay dependency-free: a scene never assumes anything about its environment beyond its own tree. When a scene needs outside context, the parent injects it — the official docs list connecting signals, calling methods, initializing `Callable` properties, handing over node references, and passing NodePaths as the injection techniques (`docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`).

## Signal Naming And Contracts

- Name signals as **past-tense snake_case action phrases**: `door_opened`, `item_collected`, `health_changed`. This is the official naming rule in both the signals tutorial and the GDScript style guide (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/gdscript_styleguide.html`).
- A signal name states what happened, never what should happen. `open_door` is a method; `door_opened` is a signal. If the name reads as a command, the design is a call wearing a signal costume — fix the design, not the name.
- Declare every parameter in the `signal` declaration with a type hint, so the payload contract is visible at the declaration site and handler signatures can be checked against it:

```gdscript
signal health_changed(old_value: int, new_value: int)
```

- Signal declarations sit in the script-order slot the style guide prescribes (after `extends` and doc comments, before constants and exports) — see [gdscript-style-and-typing.md](gdscript-style-and-typing.md).
- Payloads carry data about the event, not references the receiver uses to reach back into the emitter's internals. A receiver that needs to call methods on the emitter after every signal is coupled; pass the values it needs instead.
- Every signal that crosses a scene boundary is a contract change. Route it through [../recipes/add-a-signal-contract.md](../recipes/add-a-signal-contract.md) so the declaration, connections, and tests move together.

## Call Down, Signal Up

"Call down, signal up" is community shorthand — GDQuest and other practitioner sources name it (`gdquest.com/tutorial/godot/best-practices/signals/`, `gamedevartisan.com/tips/call-down-signal-up`) — but the pattern itself is what the official scene-organization doc teaches: parents call methods on the children they own; children emit signals upward and never reach for their ancestors.

- **Parents call down.** A parent instantiated its children, knows their types, and may call their methods directly. Direct calls between a parent and its own children are correct and need no signal.
- **Children signal up.** A child announces state changes by emitting signals. It does not know, and must not care, whether anyone is connected.
- **Siblings never talk directly.** The shared parent mediates: it connects sibling A's signal to a method that calls down into sibling B. The parent owns the wiring because the parent owns both nodes.
- **`get_parent()` and upward NodePaths (`"../..."`) are forbidden in gameplay scripts.** A child that climbs the tree has a hidden dependency on an ancestor arrangement it does not own, and breaks the moment the scene is reused elsewhere — the exact dependency the official docs' "no dependencies" principle exists to prevent. The only exception is engine-mandated plumbing (e.g. a `Control` querying its container), documented at the call site.

## Connect In Editor Versus Code

The official signals tutorial defines the split: connect in the editor for relationships between nodes that are both saved in the same scene; connect in code for nodes instantiated at runtime (`docs.godotengine.org/en/stable/getting_started/step_by_step/signals.html`).

- **Editor connections** are for static, scene-internal wiring — a `Button` in a menu scene connected to the menu script. Keep the editor's generated handler naming (`_on_<node>_<signal>`) so the connection is discoverable from the method name alone.
- **Code connections** happen in `_ready()` of the node that owns the relationship — normally the parent that instantiated the child:

```gdscript
func _spawn_enemy() -> void:
	var enemy: Enemy = ENEMY_SCENE.instantiate()
	enemy.died.connect(_on_enemy_died)
	add_child(enemy)
```

- Use the Godot 4 `Callable`-based form (`node.signal_name.connect(handler)`), not string-based `connect("signal_name", ...)` — the typed form fails loudly at parse time when the signal name is wrong instead of silently at runtime.
- The connecting node is the node that owns the relationship. A child never connects itself to an ancestor's signal by climbing the tree; the ancestor connects down, or injection provides the reference.
- Before code-review sign-off on any scene refactor, re-check editor connections: renaming or moving a node can silently break wiring that only exists in the `.tscn`.

## Groups For Broadcast

Groups are tag-like membership for one-to-many broadcast where per-receiver signal wiring would be unwieldy. The official docs' example is calling `enter_alert_mode` on every node in a `guards` group via `get_tree().call_group()` (`docs.godotengine.org/en/stable/tutorials/scripting/groups.html`).

- Group names are `snake_case`, per the same doc. Define them as constants in one script (the system that owns the broadcast), never as scattered string literals.
- A group plus the method it broadcasts is a contract: every member must implement the method with the same signature. Document the pair where the group constant is defined.
- Membership is assigned in the editor for scene-static members, or with `add_to_group()` in `_ready()` for spawned nodes — the same static/dynamic split as signal connections.
- `call_group()` is fire-and-forget: no return values, no delivery guarantee to nodes not yet in the tree. Use it for notifications, never for queries. To query, use `get_tree().get_nodes_in_group()` and iterate.
- Groups do not replace signals for single-receiver relationships. One known receiver means a signal or a call, not a broadcast.

## Event Bus Boundaries

An event bus is an autoload whose script contains only signal declarations — a named rendezvous point for events that must cross unrelated scene branches. GDQuest's guidance matches this handbook's ladder: call-down/signal-up for close-proximity nodes, an event bus only for cross-system communication (`gdquest.com/tutorial/godot/best-practices/signals/`).

The official autoload best-practices page documents why this rung is expensive: global access destroys debuggability — "now that any object can call Sound.play() from anywhere, there's no longer an easy way to find the source of a bug" — and global state is "a centralized failure point" (`docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`). The same trade applies to a global signal hub, so the bus is bounded hard:

- **Signals only.** The bus script declares signals and nothing else — no state, no methods with logic, no node references. The moment a bus accumulates state it becomes the centralized failure point the docs warn about; that system belongs in its own autoload with an owned API, routed through [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md).
- **Cross-system events only.** A bus signal is justified only when emitter and receivers share no ancestor that could reasonably mediate (e.g. `player_died` consumed by UI, audio, and save systems in separate branches). Parent-child and sibling communication never routes through the bus.
- **Same naming contract.** Bus signals follow the past-tense rule and carry typed payloads; each one is registered via [../recipes/add-a-signal-contract.md](../recipes/add-a-signal-contract.md) so the full cross-system event surface is enumerable in one place.
- **Emit from one owner.** Each bus signal has exactly one emitting system. Many receivers, one sender — a signal any system may emit is untraceable.
- Never `free()` or `queue_free()` an autoload at runtime — the official docs state the engine will crash (`docs.godotengine.org/en/stable/tutorials/scripting/singletons_autoload.html`).

## Common Mistakes And Forbidden Patterns

- Command-named signals (`open_door`, `play_sound`) — a signal that starts behavior instead of reporting it inverts the pattern; use a method call.
- `get_parent()`, `owner`, or upward NodePaths in gameplay scripts to reach ancestors or siblings.
- Editor-connecting signals for nodes that are instantiated at runtime — the connection exists only in the editing scene and silently never fires in the composed game.
- String-based `connect("signal_name", ...)` where the Godot 4 typed form is available.
- Routing local parent-child events through the event bus because the bus was already imported.
- Event bus autoloads that grow state, helper methods, or node references.
- Two systems emitting the same bus signal.
- Group broadcasts used as queries, or a group created for a single receiver.
- Signal handlers that immediately call back into the emitter — the payload was wrong; pass the data the receiver actually needs.
- Untyped signal payloads (`signal changed(data)`) that force every receiver to re-validate shape.

## Verification And Proof

- `gdlint` passes on all scripts, enforcing signal and group naming conventions — see [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md)
- `grep -rn "get_parent()" --include="*.gd"` returns only call sites with a documented engine-plumbing justification
- every cross-scene and bus signal appears in the signal contract registry per [../recipes/add-a-signal-contract.md](../recipes/add-a-signal-contract.md)
- unit tests watch or await the signal and assert both emission and payload for every contract signal — framework specifics in [../quality/testing.md](../quality/testing.md)
- scene refactors re-verified in the editor's Node > Signals panel: no broken connections after node renames or moves
- the event bus script diff shows only `signal` declarations; any other member fails review

## Related

- [scene-and-node-architecture.md](scene-and-node-architecture.md) — scene ownership and dependency-injection rules the wiring depends on
- [gdscript-style-and-typing.md](gdscript-style-and-typing.md) — script order and typing rules for declarations
- [../recipes/add-a-signal-contract.md](../recipes/add-a-signal-contract.md) — the fixed-shape procedure for new signal contracts
- [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md) — gate for any new autoload, including a bus
- [../systems/multiplayer.md](../systems/multiplayer.md) — RPCs, not signals, cross the network boundary
