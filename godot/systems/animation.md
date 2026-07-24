# Animation

Who owns motion in Godot 4.x projects: a fixed AnimationPlayer / AnimationTree / Tween decision table, state-machine and root-motion rules, method-track boundaries, and which process callback animation runs on.

## Default Approach

Every animated value has exactly one owner. Pick the owner from the table below by whether the target values are authored ahead of time, blended from gameplay state, or computed at runtime — then do not let a second system write the same property.

### Choosing The Animation Owner

| Tool | Owns motion when | Never |
|---|---|---|
| `AnimationPlayer` | values are authored and editor-tweaked keyframes — it can animate "anything available in the Inspector, such as Node transforms, sprites, UI elements, particles, visibility and color of materials" and call functions (`docs.godotengine.org/en/stable/tutorials/animation/introduction.html`) | blending or state logic — its blend support is limited to a fixed cross-fade time (`docs.godotengine.org/en/stable/tutorials/animation/animation_tree.html`) |
| `AnimationTree` | authored clips must be blended or sequenced from gameplay state — it is "a node used for advanced animation transitions in an AnimationPlayer" (`docs.godotengine.org/en/stable/classes/class_animationtree.html`) | containing animations; "it uses animations contained in an AnimationPlayer node" (`docs.godotengine.org/en/stable/tutorials/animation/animation_tree.html`) |
| `Tween` | the target value is computed at runtime — "Tween is more suited than AnimationPlayer for animations where you don't know the final values in advance", e.g. "interpolating a dynamically-chosen camera zoom value" (`docs.godotengine.org/en/stable/classes/class_tween.html`) | replacing authored animation; Tweens suit "simple animations or general tasks that don't require visual tweaking provided by the editor" |
| gameplay code (`_physics_process`) | the motion is gameplay-relevant displacement of a physics body — see [physics-and-movement.md](physics-and-movement.md) | being keyframed around; animation decorates gameplay motion, it does not replace it |

- One owner per property. The Tween docs state the failure mode for their corner of it — "If two or more tweens animate one property at the same time, the last one created will take priority" (`docs.godotengine.org/en/stable/classes/class_tween.html`) — and the same last-writer-wins chaos applies when a script, an AnimationPlayer, and a Tween share a property.
- Tween discipline: create via `create_tween()` — "Tweens created manually (i.e. by using `Tween.new()`) are invalid and can't be used for tweening values"; "Tweens are not designed to be reused"; and "Tweens start immediately, so only create a Tween when you want to start animating" (`docs.godotengine.org/en/stable/classes/class_tween.html`). Kill the previous tween on the same property before starting a replacement. A node-created Tween is auto-killed when its node is freed, so scene-owned tweens clean up with their scene.

### AnimationTree State Machines

- **Once an AnimationTree is active, it owns playback.** The class docs are explicit: "Playback and transitions should be handled using only the AnimationTree and its constituent AnimationNode(s). The AnimationPlayer node should be used solely for adding, deleting, and editing animations" (`docs.godotengine.org/en/stable/classes/class_animationtree.html`). A `play()` call on the underlying AnimationPlayer while the tree is active is a defect.
- Drive an `AnimationNodeStateMachine` from code through its playback object — `animation_tree["parameters/playback"]` — and change state with `travel()`, which "go[es] from the current state to another one, while visiting all the intermediate ones... via the A* algorithm" (`docs.godotengine.org/en/stable/tutorials/animation/animation_tree.html`). Because `travel()` pathfinds, audit the graph's transition edges: an unintended edge becomes an unintended intermediate state.
- Pick transition switch modes deliberately — *Immediate* switches now, *Sync* switches now but seeks to the matching playback position, *At End* waits for the current state to finish (`docs.godotengine.org/en/stable/tutorials/animation/animation_tree.html`). Synced walk/run clips use *Sync*; attack-to-idle uses *At End*.
- Keep gameplay decisions in gameplay code. Prefer `travel()` calls or boolean advance conditions set from the owning script over advance *expressions*; an expression evaluated against `advance_expression_base_node` is game logic hidden inside a resource where no test or grep will find it.
- Use `BlendSpace1D`/`BlendSpace2D` for parameterized locomotion: "Points representing animations are added to a 2D space and then a position between them is controlled to determine the blending" (`docs.godotengine.org/en/stable/tutorials/animation/animation_tree.html`). Feed the blend position from the same velocity the physics code computes — one source of truth for how fast the character moves.

### Animation-Driven Versus Code-Driven Movement

- Gameplay displacement is code-driven: bodies move in `_physics_process` under the contracts in [physics-and-movement.md](physics-and-movement.md). Animation-driven movement of a colliding object is legal only through `AnimatableBody2D/3D` (platforms, doors) — that body type exists for AnimationPlayer-moved geometry.
- **Root motion** is the bridge for animation-authored locomotion. Set `root_motion_track` on the tree; for position/rotation/scale tracks "the transformation will be canceled visually, and the animation will appear to stay in place" (`docs.godotengine.org/en/stable/classes/class_animationmixer.html`). Read the per-frame delta with `get_root_motion_position()` / `get_root_motion_rotation()` and feed it into movement code — "this can be fed to functions such as `CharacterBody3D.move_and_slide` to control the character movement" (`docs.godotengine.org/en/stable/tutorials/animation/animation_tree.html`). Divide the position delta by `delta` before assigning it to `velocity`, since `move_and_slide()` re-applies the timestep (see [physics-and-movement.md](physics-and-movement.md) ### Delta Discipline).
- When the character can rotate, multiply through the rotation accumulator as the tutorial shows — the accumulator "is necessary to apply the root motion position correctly, taking rotation into account" (`docs.godotengine.org/en/stable/classes/class_animationmixer.html`).

### Method Call Tracks And Signals

- A call method track exists to "call a function at a precise time from within an animation" — footstep sounds, hitbox toggles, spawn moments (`docs.godotengine.org/en/stable/tutorials/animation/animation_track_types.html`).
- Method tracks may only call methods on nodes inside the owning scene. A track that reaches outside the scene breaks the self-containment rule in [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md); if the moment matters to anyone else, the called method emits a signal and the parent connects it per [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md).
- Know the two execution gaps: "The events placed on the call method track are not executed when the animation is previewed in the editor for safety" (`docs.godotengine.org/en/stable/tutorials/animation/animation_track_types.html`), and calls are deferred by default — `callback_mode_method` batches calls "then do[es] the calls after events are processed" (`docs.godotengine.org/en/stable/classes/class_animationmixer.html`). Switch to immediate mode only when a same-frame ordering dependency is documented at the site that needs it.

### Animation And The Physics Tick

- `callback_mode_process` on any `AnimationMixer` (AnimationPlayer and AnimationTree both inherit it) defaults to idle processing. Set it to physics — "especially useful when animating physics bodies" (`docs.godotengine.org/en/stable/classes/class_animationmixer.html`) — whenever the animation moves an `AnimatableBody`, drives root motion consumed in `_physics_process`, or keys any collision-relevant transform. This is the animation-system face of the repo invariant in [../AGENTS.md](../AGENTS.md): physics work happens on the physics tick.
- The same split applies to Tweens: a Tween touching a physics-visible value runs with `TWEEN_PROCESS_PHYSICS`, which "updates after each physics frame" (`docs.godotengine.org/en/stable/classes/class_tween.html`); cosmetic tweens stay on the idle default.
- `ANIMATION_CALLBACK_MODE_PROCESS_MANUAL` plus `advance()` — "Do not process animation. Use advance() to process the animation manually" (`docs.godotengine.org/en/stable/classes/class_animationmixer.html`) — is the deterministic stepping hook for headless tests.

## Common Mistakes And Forbidden Patterns

- Two owners writing one property — a script fighting an AnimationPlayer, or stacked Tweens where "the last one created will take priority".
- Calling `AnimationPlayer.play()` while an active AnimationTree drives the same player; the tree owns playback.
- `Tween.new()` (invalid by contract), reusing a finished Tween (undefined behavior), or creating a Tween per frame in `_process` instead of once when the animation should start.
- Keyframing a colliding object's transform on anything but an `AnimatableBody`, or applying root motion by letting the animation move the collision body directly instead of feeding the deltas to `move_and_slide()`.
- Animating physics-relevant transforms with the mixer or Tween left in idle process mode, so collision state updates off the physics tick.
- Game logic in advance expressions, or method tracks calling across scene boundaries — untestable, ungreppable control flow hidden in resources.
- Relying on method-track side effects while scrubbing in the editor; they are suppressed in preview by design.
- State-machine graphs with convenience transitions everywhere, turning every `travel()` into an unpredictable A* tour.

## Verification And Proof

- Ownership sweep: for each animated property, name its single owner. Grep for tween churn and playback bypass, and review every hit:

```bash
grep -rn --include="*.gd" -B 2 "create_tween" . | grep "func _process\|func _physics_process"
grep -rn --include="*.gd" -E '\$AnimationPlayer.*\.play\(|animation_player\.play\(' .
```

- Root motion and `AnimatableBody` animation: prove framerate independence the same way as [physics-and-movement.md](physics-and-movement.md) — traversal distance per wall-clock second is identical under `godot --fixed-fps 30` and `--fixed-fps 120`, and the mixer's `callback_mode_process` is physics in the scene file.
- Method tracks: a headless scene test (see [../quality/testing.md](../quality/testing.md)) plays the animation with manual processing and `advance()`, asserting the tracked method's observable effect fires — editor preview cannot prove this, because preview suppresses method tracks.
- State machines: a test drives `parameters/playback` through the expected `travel()` sequence and asserts `get_current_node()` at each step, so an added convenience transition that reroutes A* fails loudly.

## Related

- [physics-and-movement.md](physics-and-movement.md) — body-type contracts, delta discipline, and the physics-tick rule this doc's motion rules plug into.
- [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md) — scene self-containment, the boundary method tracks must respect.
- [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) — how animation-triggered moments leave the scene.
- [../quality/testing.md](../quality/testing.md) — the headless test surface for animation proofs.
- [audio.md](audio.md) — scene-owned players that method tracks and audio tracks trigger.
