# Physics And Movement

Frame/tick split, delta discipline, body type selection, and collision layer conventions for movement code that behaves identically at any framerate. Guidance targets the current Godot 4.x stable line.

## Default Approach

- All physics work — moving bodies, applying forces, raycasts, overlap queries — runs in `_physics_process`. Rendering, UI, and cosmetic updates run in `_process`.
- Every per-frame change of a continuous value is multiplied by `delta`, with the single documented exception of `move_and_slide()`.
- Pick the body type from the table below by who controls the motion; do not fight a body type's contract with per-frame position writes.
- Every collision layer in use is named in Project Settings, and the layer/mask assignment for the project is recorded once (see [Collision Layers And Masks](#collision-layers-and-masks)).

### Process Versus Physics Process

- `_process(delta)` runs once per rendered frame; its frequency "depends on your application's framerate, which varies over time and across devices." `_physics_process(delta)` runs on the physics tick: "at a fixed rate, 60 times per second by default," configurable in Project Settings (`docs.godotengine.org/en/stable/tutorials/scripting/idle_and_physics_processing.html`).
- The official rule for choosing: use `_physics_process` "for anything that involves the physics engine, like moving a body that collides with the environment." The docs warn that "`_process()` is not synchronized with physics" — a body moved from `_process` moves at a rate the physics engine never sees consistently, producing jitter and missed collisions.
- Continuous input polling (`Input.is_action_pressed()`) belongs next to the movement it drives, so it lives in `_physics_process` for physics-driven characters. Discrete input events stay in the event callbacks owned by [../foundations/input-handling.md](../foundations/input-handling.md).
- Do not duplicate state mutation across both callbacks. A value either advances on the render frame or on the physics tick — never both.

### Delta Discipline

- Multiply rates by `delta` to get per-frame amounts: "Use this parameter to make calculations independent of the framerate" (`docs.godotengine.org/en/stable/tutorials/scripting/idle_and_physics_processing.html`). This applies in both callbacks — `_physics_process` has a fixed tick by default, but the tick rate is a project setting, not a constant your code may assume.
- The exception is `move_and_slide()`, which "automatically includes the timestep in its calculation, so you should not multiply the velocity vector by delta" (`docs.godotengine.org/en/stable/tutorials/physics/physics_introduction.html`). `velocity` is a speed in units per second; hand it over un-scaled.
- `move_and_collide()` takes a motion vector, not a velocity — pass `velocity * delta`.
- Acceleration, gravity applied to `velocity`, lerp-toward-target rates, and timers still need `delta` even on a `CharacterBody2D`/`CharacterBody3D`; the exemption covers only the final `move_and_slide()` call.
- Never encode the tick rate as a literal (`velocity / 60.0`). Raising the physics tick in Project Settings must not change gameplay speed.

### Body Type Selection

Choose by who controls the motion. 2D names shown; the 3D classes split identically (`docs.godotengine.org/en/stable/tutorials/physics/physics_introduction.html`).

| Body | Motion controlled by | Use for | Never |
|---|---|---|---|
| `StaticBody2D` | nobody — "not moved by the physics engine" | walls, floors, level geometry, obstacles | moving it per-frame from script |
| `AnimatableBody2D` | your code or an `AnimationPlayer`; "when moved manually, it affects other bodies in its path" | moving platforms, doors, elevators | combining `sync_to_physics` with `move_and_collide()` — the class docs forbid it (`docs.godotengine.org/en/stable/classes/class_animatablebody2d.html`) |
| `CharacterBody2D` | your code; bodies "detect collisions with other bodies, but are not affected by physics properties like gravity or friction" | players, NPCs, anything needing precise game-feel movement | expecting free gravity or bounce — you write it |
| `RigidBody2D` | the physics engine; you apply forces and "the physics engine calculates the resulting movement" | crates, debris, projectiles with real physics | setting `position` or `linear_velocity` directly per-frame — the docs warn this "can result in unexpected behavior"; use forces or `_integrate_forces()` |
| `Area2D` | n/a — "detection and influence" only | hitboxes, pickups, triggers, gravity/audio zones | collision response; it detects, it does not block |

- Do not use `RigidBody2D` for a player character to "get physics for free" and then fight the simulation with position writes; that is the `CharacterBody2D` contract.
- Area overlap reactions arrive as signals (`body_entered`/`area_entered`); wire them per [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md), not by polling overlap lists in `_process`.

### Collision Layers And Masks

- The split, per the official docs: `collision_layer` describes "the layers that the object appears in"; `collision_mask` describes "what layers the body will scan for collisions" (`docs.godotengine.org/en/stable/tutorials/physics/physics_introduction.html`).
- Name every layer you use under Project Settings > Layer Names — the docs recommend this because "keeping track of what you're using each layer for can be difficult." A bare bit index in a scene or script is a review-blocking offense; the name is the contract.
- Record the project's layer table (number, name, what sits on it, what masks it) in the settings conventions doc seeded from [../templates/project-settings-conventions.md](../templates/project-settings-conventions.md), and keep it in sync when a layer is added — this recorded table is a convention this handbook adopts, not an engine requirement.
- Make detection one-directional and minimal: the detector masks the detected. A coin masks the player layer; the player does not mask coins. Masking everything against everything makes every collision pair intentional-looking and none of them auditable.
- Set layers and masks in the scene (inspector), not in `_ready()` code, unless the value genuinely changes at runtime (e.g. a phasing enemy). Scene-declared physics config is diffable and visible; script-assigned bits are neither.

## Common Mistakes And Forbidden Patterns

- Moving a collision body from `_process` — "`_process()` is not synchronized with physics."
- Multiplying the `move_and_slide()` velocity by `delta`, applying the timestep twice and making speed tick-rate-dependent.
- Omitting `delta` on movement, acceleration, or timers, so gameplay speed tracks framerate.
- Hardcoding the physics tick (`/ 60.0`, frame-count timers) instead of using `delta`.
- Writing `position` or `linear_velocity` on a `RigidBody2D` every frame instead of applying forces or overriding `_integrate_forces()`.
- Choosing `RigidBody2D` for a player character, then suppressing the simulation it exists to provide.
- Unnamed collision layers, or magic bit indexes in scripts.
- All-ones collision masks — every body scanning every layer.
- Assigning layers/masks in `_ready()` when the scene inspector could declare them.
- Duplicating movement state updates across `_process` and `_physics_process`.

## Verification And Proof

- Grep for physics APIs in the wrong callback and review every hit:

```bash
grep -rn --include="*.gd" -A 20 "func _process(" . | grep -nE "move_and_slide|move_and_collide|apply_force|apply_impulse"
```

- Prove framerate independence by running the game at a forced frame rate and confirming identical traversal times: `godot --fixed-fps 30` and `--fixed-fps 120` ("Force a fixed number of frames per second," `docs.godotengine.org/en/stable/tutorials/editor/command_line_tutorial.html`). Distance covered per wall-clock second must not change.
- Movement and collision behavior that can be expressed as pure functions (speed curves, damage from overlap sets) is unit-tested per [../quality/testing.md](../quality/testing.md); scene-level movement gets headless scene-runner tests where the framework supports them.
- On any layer/mask change: open Project Settings > Layer Names and confirm every used bit is named, and confirm the recorded layer table matches the scenes touched.
- Route the change through [../AGENTS.md](../AGENTS.md) (## Change Routing) so the sync surface (scenes, layer table, tests) moves together.

If you cannot state which callback owns a moving value and why its rate is framerate-independent, the movement code is not done.
