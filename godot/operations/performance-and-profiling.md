# Performance And Profiling

Performance discipline for Godot projects on the current 4.x stable line: measure first, spend the frame budget deliberately, and treat every optimization as a change that needs proof.

## Default Approach

| Concern | Default | Notes |
|---|---|---|
| measurement | built-in editor profiler plus the debugger's pipeline-compilation monitor | manually started; never optimize from intuition |
| frame budget | 16.7 ms per frame at a 60 FPS target | game logic, physics, and rendering share it |
| physics tick | fixed 60 Hz default | change only via Project Settings, with an ADR |
| draw calls | shared materials, batching, `MultiMesh` for massive instance counts | fewest distinct materials wins |
| shader stutter | engine pipeline precompilation (4.4+) plus Shader Baker on export (4.5+) | plus the scene-design rules below |
| object pooling | none | allowed only with profiler evidence and an ADR |

The engine's own stance frames every rule here: "In the performance world, there are always tradeoffs" — Godot favors balanced algorithms over ones that are fast in narrow cases (`docs.godotengine.org/en/stable/tutorials/performance/index.html`). Do not import optimization habits from other engines without re-measuring in Godot.

### Profile Before Optimizing

- No optimization lands without a bottleneck identified in a profiler capture and a before/after measurement in the PR. The official CPU optimization doc is blunt: "We have to know where the 'bottlenecks' are to know how to speed up our program" (`docs.godotengine.org/en/stable/tutorials/performance/cpu_optimization.html`).
- Use the built-in profiler first. It "must be manually started and stopped" because recording measurements itself slows the project — start it around the scenario you are diagnosing, not globally.
- Reach for an external C++ profiler (e.g. Callgrind) only when the bottleneck is inside the engine itself, per the same doc.
- Typed GDScript is a free baseline win: the compiler emits optimized opcodes when types are known. Typing rules are owned by [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md); do not leave hot-path scripts untyped.
- Profile on the minimum-spec target device, not only the development machine. Mobile tile-based GPUs have different failure modes (dense vertex clusters, missing LOD) than desktop GPUs (`docs.godotengine.org/en/stable/tutorials/performance/gpu_optimization.html`).

### Draw Calls And Batching

- The GPU optimization goal is to "reduce these instructions to a bare minimum and group together similar objects" (`docs.godotengine.org/en/stable/tutorials/performance/gpu_optimization.html`).
- Reuse materials and shaders aggressively: "the fewer different materials in the scene, the faster the rendering." Share one material across visually identical objects instead of duplicating it per scene instance.
- 2D rendering batches similar items into single draw calls automatically; keep sprites on shared textures/materials so batches stay unbroken.
- 3D: merge static meshes that always render together, or use `MultiMesh` — "a single draw primitive that can draw up to millions of objects in one go" (`docs.godotengine.org/en/stable/tutorials/performance/using_multimesh.html`).
- Know the batching tradeoff before applying it: a `MultiMesh` has no per-instance frustum culling — "millions of objects will be _always_ or _never_ drawn, depending on the visibility of the whole MultiMesh." Split large worlds into several `MultiMesh`es per area so culling still works.
- VRAM-compress 3D textures; uncompressed textures burn memory bandwidth on every frame.

### Process Versus Physics Budgets

- `_process(delta)` runs once per rendered frame at a variable rate; `_physics_process(delta)` runs on the fixed physics tick, "at a fixed rate, 60 times per second by default" (`docs.godotengine.org/en/stable/tutorials/scripting/idle_and_physics_processing.html`).
- Anything touching the physics engine — moving a body that collides with the environment, raycasts against physics state — belongs in `_physics_process`. The docs warn plainly: "_process() is not synchronized with physics." Movement and collision rules are owned by [../systems/physics-and-movement.md](../systems/physics-and-movement.md).
- Rendering, UI, and cosmetic updates belong in `_process`. Always multiply movement by `delta` in both callbacks for framerate independence.
- The physics tick rate is a Project Settings value; raising it multiplies fixed-cost physics work per second across the whole project. Changing it is a project-wide contract change and requires an ADR ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)).
- Budget per frame, not per system in isolation: at 60 FPS everything — scripts, physics, rendering, audio — shares 16.7 ms. A system that is "fast" but eats 8 ms owns half the frame.

### Shader Compilation Stutter

First-use pipeline compilation is the classic Godot hitch: the GPU driver converts the shader's intermediate format into pipelines on demand, and before 4.4 there was no mitigation — objects hitching as they first entered view was the infamous "shader stutter" (`docs.godotengine.org/en/stable/tutorials/performance/pipeline_compilations.html`).

- Godot 4.4+ mitigates automatically with ubershaders (specialization constants swapped in the background) and load-time pipeline precompilation on mesh load and scene-node addition. Do not disable these; design so they can work:
- Enable every rendering feature you will use in a scene loaded as early as possible, before the majority of assets load, so precompilation covers the real pipeline set.
- Instantiate gameplay-spawned effects (projectiles, explosions, pickups) once at load time as hidden nodes so their pipelines precompile before first spawn.
- Never toggle rendering features mid-gameplay; that invalidates the precompiled pipeline set and reintroduces stutter.
- Watch the pipeline-compilation monitor in the Godot debugger to find stutter sources; a full playthrough after load screens should show zero mid-gameplay compilations.
- On Godot 4.5+, enable the Shader Baker export option so shipped builds skip shader compilation (final pipelines still compile on device). Export presets are owned by [ci-and-release.md](ci-and-release.md).

### Object Pooling Policy

- Default: no object pools. Godot is reference-counted, not garbage-collected, so the GC-pressure rationale imported from Unity does not apply; pooling pays off only at extreme spawn rates — bullet-hell volumes, thousands of short-lived objects — or when the profiler shows instantiation spikes (`forum.godotengine.org/t/what-is-the-best-way-to-do-object-pooling/28960`, `popcar.bearblog.dev/unity-to-godot-what-to-expect/`).
- A pool may be introduced only when a profiler capture attributes a frame spike to instantiation, and the pool ships with that capture plus an ADR recording the threshold that justified it.
- A pool is shared mutable state with a lifecycle; it trades the scene-tree's ownership model for manual reset bugs. Scene ownership rules are owned by [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md).

## Common Mistakes And Forbidden Patterns

- Optimizing without a profiler capture, or shipping an "optimization" with no before/after numbers.
- Porting Unity-style object pools by reflex into a reference-counted engine.
- Duplicating materials per instance so every object breaks batching.
- One world-spanning `MultiMesh` that defeats frustum culling for every instance at once.
- Physics-engine work in `_process`, or cosmetic per-frame work in `_physics_process`.
- Raising the physics tick rate to paper over a rendering or logic problem.
- Dismissing first-playthrough hitching as "loading" instead of checking the pipeline-compilation monitor.
- Enabling a new rendering feature mid-gameplay and reintroducing shader stutter.
- Untyped GDScript in hot loops, forfeiting the typed-opcode fast path.

## Verification And Proof

- every optimization PR attaches profiler captures showing the bottleneck before and the improvement after
- a full playthrough with the debugger's pipeline-compilation monitor open shows zero mid-gameplay compilations
- the target scene holds frame time under 16.7 ms on the minimum-spec device, not just the dev machine
- any pool in the tree links to its ADR and the profiler evidence that justified it
- exported release builds (4.5+) have Shader Baker enabled in the preset checked by [ci-and-release.md](ci-and-release.md)

Related: [../systems/physics-and-movement.md](../systems/physics-and-movement.md), [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md), [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md), [ci-and-release.md](ci-and-release.md), [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md).
