# Recipe: Add A Signal Contract

Use this when a node needs to announce a state change to other nodes or systems — a new signal declaration, its typed payload, and the connections that consume it.

## Files To Touch

- the emitting node's script — the `signal` declaration and every `emit` site
- the script or scene that owns the connection: the shared parent for sibling wiring, the emitter's own `.tscn` for editor connections
- the event bus autoload script only when emitter and receivers share no ancestor that could mediate — gated by [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) (## Event Bus Boundaries)
- every receiver script that adds a handler
- unit tests that watch or await the signal

## Steps

1. Confirm a signal is the right rung on the escalation ladder in [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) (## Default Approach): the emitter is reporting something that already happened. If the caller is starting behavior, this is a method call, not a signal — the official guidance is that signals "should be used only to 'respond' to behavior, not start it" (`docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`).
2. Name it as a past-tense snake_case action phrase and declare it in the style-guide script-order slot, with every parameter type-hinted — naming and payload rules are owned by [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) (## Signal Naming And Contracts).
3. Document the payload contract in a `##` doc comment immediately preceding the declaration: when the signal fires, what each parameter carries, and which system emits it. Doc comments "must immediately precede the member" to attach (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/gdscript_documentation_comments.html`).

```gdscript
## Emitted after damage or healing is applied. Carries the values a
## receiver needs; receivers never reach back into the emitter.
signal health_changed(old_value: int, new_value: int)
```

4. Pick the connection site per [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) (## Connect In Editor Versus Code): editor connection for nodes saved in the same scene, `_ready()` of the owning parent for runtime-instantiated nodes, always the typed `Callable` form.
5. Emit at the point the state change has already been applied, passing values from the emitter's own state. Never emit as a way to ask a receiver to do work, and never pass a reference the receiver uses to reach back into the emitter.
6. For a bus signal only: add the bare declaration to the bus script — signals only, one emitting system — so the bus remains the enumerable registry of cross-system events. Boundaries and the autoload gate live in [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) (## Event Bus Boundaries) and [add-an-autoload.md](add-an-autoload.md).
7. Add a unit test that watches or awaits the signal and asserts both emission and payload values — framework specifics in [../quality/testing.md](../quality/testing.md).

## Invariants To Preserve

- the signal name reports what happened; a command-shaped name means the design is wrong, not the name
- every parameter is type-hinted at the declaration and the doc comment states the full payload contract
- receivers get values, not a handle back into the emitter's internals
- no `get_parent()` or upward NodePaths appear to wire the connection; the owner connects down or injects
- a bus signal has exactly one emitting system, and the bus script still contains only `signal` declarations
- editor connections exist only between nodes saved in the same scene; spawned nodes connect in code

## Proof

- `gdlint` passes on all touched scripts — [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md)
- the unit test fails when the emit is removed and when a payload value is wrong
- `grep -rn "get_parent()" --include="*.gd"` shows no new hits from this change
- editor's Node > Signals panel shows the intended connections and no broken ones after any node rename or move
- for bus signals: the bus script diff adds only the declaration and its doc comment

Governing doc: [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md). Route changes to the wider file set through [../AGENTS.md](../AGENTS.md) (## Change Routing).
