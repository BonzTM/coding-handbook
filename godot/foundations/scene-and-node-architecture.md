# Scene And Node Architecture

Composition rules for Godot scenes and nodes: what a scene may depend on, how dependencies flow into it, and when a node is the wrong tool. Applies to the current Godot 4.x stable line.

## Default Approach

Build the game as a tree of self-contained scenes. Dependencies flow downward — parents inject them into children; information flows upward through signals. Every scene must be instantiable on its own without errors.

### Scenes As Self-Contained Units

- Design every scene to work with "no dependencies" on anything outside itself — the official scene-organization guidance (`docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`). A scene that only functions inside one specific parent is a hidden coupling, not a reusable unit.
- A scene never reaches outside its own root: no `get_parent()` traversal, no `get_node("..")`, no absolute `NodePath`s into another scene's internals, no assumptions about siblings existing.
- When a scene needs external context, it declares the need (an exported property, a signal, a method the parent calls) and the instancing context supplies it. The parent knows the child; the child never knows the parent.
- Proof of self-containment is mechanical: run the scene by itself from the editor. If it errors without its usual parent, it has an undeclared dependency.

### Dependency Injection Techniques

The official docs list five techniques for a parent to hand a child what it needs (`docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`). Prefer them in this order:

| Technique | Use when |
|---|---|
| Connect a signal | The parent must react to something the child did. Signals are "extremely safe, but should be used only to 'respond' to behavior, not start it." |
| Call a method | The parent initiates behavior on the child. |
| Initialize a `Callable` property | The child must invoke parent-supplied behavior without knowing who owns it. |
| Hand over a node or object reference | The child needs an ongoing collaborator the parent owns. |
| Pass a `NodePath` | The child needs to resolve a node the parent locates; weakest option — breaks silently on tree refactors. |

- The direction rule: parents call down, children signal up. "Call down, signal up" is community shorthand (GDQuest and others), not official wording, but it names exactly the pattern the official docs teach. Signal contracts, event buses, and connection style are owned by [signals-and-decoupling.md](signals-and-decoupling.md).
- Injected references arrive before `_ready()` logic depends on them: set exported properties at instantiation, or use an explicit `setup()` method the parent calls — never have the child go looking.

### Tree Shape And Ownership

```text
Main        # entry point, primary controller — orchestrates, holds no gameplay logic
├── World   # gameplay root; swap its children to change levels
└── GUI     # HUD, menus, overlays
```

- The docs recommend a `Main` node as "primary controller" with `World` and `GUI` children, and level changes implemented by swapping `World`'s children (`docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`).
- Choose parent/child relationships by dependency, not by spatial convenience. The docs' test question: "Are the nodes dependent on their parent's existence?" If a node must outlive its parent or be reused elsewhere, move it under a neutral owner and pass a reference instead.
- The main scene is registered in project settings; see [project-setup.md](project-setup.md). Adding a new scene follows [../recipes/add-a-scene.md](../recipes/add-a-scene.md).

### Scenes Versus Scripts

- "Scripts define an engine class extension with imperative code, scenes with declarative code" (`docs.godotengine.org/en/stable/tutorials/best_practices/scenes_versus_scripts.html`). Default pairing: the scene declares structure and initialization; one script on the scene root provides behavior.
- Do not build node hierarchies in `_init()`/`_ready()` code when a scene can declare them. Scene instantiation processes serialized data engine-side and outperforms script-built hierarchies — "Script code like this is much slower than engine-side C++ code."
- The docs' selection rule: scenes for game-specific concepts (easier to track and edit); script classes for reusable tools and abstract node types intended to work across projects.
- A `class_name` script with no scene is correct for non-node logic, shared base classes, and static helpers — see [gdscript-style-and-typing.md](gdscript-style-and-typing.md) for declaration order and typing rules.

### Avoiding Nodes For Everything

Nodes are cheap to create, but "the more complex their behavior though, the larger the strain each one adds to a project's performance" (`docs.godotengine.org/en/stable/tutorials/best_practices/node_alternatives.html`). Reach for a node only when the object needs to live in the scene tree — otherwise use the lightest class that fits:

| Class | Use for | Memory model |
|---|---|---|
| `Object` | Custom data structures at scale (the docs' example: `TreeItem`s behind a `Tree` node) | Manual — you `free()` it |
| `RefCounted` | Plain logic and data classes that need no serialization | Automatic reference counting |
| `Resource` | Serializable, Inspector-editable data — stats, items, config | Reference-counted, plus save/load |
| `Node` | Anything that needs tree membership: processing callbacks, transforms, tree-scoped lifetime | Owned by its parent in the tree |

- Default for a new non-node class is `RefCounted`; escalate to `Resource` only when the data must be saved, shared, or designer-edited in the Inspector. Resource design and the untrusted-file loading rules are owned by [resources-and-data.md](resources-and-data.md).
- Thousands of lightweight objects beat thousands of nodes; if profiling shows tree overhead, see [../operations/performance-and-profiling.md](../operations/performance-and-profiling.md).

### Autoloads Versus Regular Nodes

- Default to no autoload. The official position: global state makes "one object ... responsible for all objects' data" — a centralized failure point — and once "any object can call Sound.play() from anywhere, there's no longer an easy way to find the source of a bug" (`docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`).
- Prefer, in order: keeping the capability inside the scene that uses it (e.g. per-scene `AudioStreamPlayer`s), `static` functions on a `class_name` script, or a shared custom `Resource` — all named by the same page as the alternatives to reach for first.
- An autoload is legitimate for a wide-scope system that owns its own data and no other object's — quest systems, dialogue systems: "autoloaded nodes can simplify your code for systems with a wide scope."
- An autoload is not a true singleton — it can still be instanced again — and it "must not be removed using `free()` or `queue_free()` at runtime, or the engine will crash" (`docs.godotengine.org/en/stable/tutorials/scripting/singletons_autoload.html`).
- Every new autoload goes through [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md), which requires stating why the scene-local, static-function, and shared-Resource alternatives do not fit.

## Common Mistakes And Forbidden Patterns

- A child reaching up: `get_parent()`, `get_node("..")`, or any assumption about who instanced it.
- Absolute `NodePath`s (`get_node("/root/...")`) crossing scene boundaries instead of injected references.
- A `Globals` autoload grab-bag accreting unrelated state because adding a field there is easier than injecting it.
- Signals used to command behavior downward — signals respond to what happened; parents start behavior with method calls.
- Building in code what should be a scene: assembling node hierarchies in `_ready()` instead of instancing a `PackedScene`.
- Nodes as plain data records where a `Resource` or `RefCounted` class suffices.
- Sibling coupling: a scene that errors unless a specific sibling happens to exist in the parent.
- Freeing an autoload at runtime.

## Verification And Proof

- Run each changed scene by itself from the editor (or instantiate it alone in a test — see [../quality/testing.md](../quality/testing.md)); it must load and reach `_ready()` without errors.
- `grep -rn 'get_parent()\|get_node("/root\|get_node("\.\.' --include='*.gd' .` — every hit needs a justification or a refactor to injection.
- Review the `[autoload]` block in `project.godot`: each entry names a wide-scope system that owns its own data, and each has a recipe-backed rationale.
- For each new class, name why it needs its base: a `Node` must need the tree, a `Resource` must need serialization or the Inspector; otherwise downgrade it.
- Review a proposed scene by asking: who instances it, what does the parent inject, and what does it signal back?

Related: [signals-and-decoupling.md](signals-and-decoupling.md), [resources-and-data.md](resources-and-data.md), [project-setup.md](project-setup.md), [../recipes/add-a-scene.md](../recipes/add-a-scene.md), [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md), [../AGENTS.md](../AGENTS.md) (## Change Routing).
