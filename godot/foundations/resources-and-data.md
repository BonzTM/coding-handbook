# Resources And Data

How this handbook models game data: typed custom Resources for designer-editable shipped data, Dictionaries only where speed or schemalessness genuinely wins, and a hard boundary that untrusted resource files are never loaded.

## Default Approach

Split all data by who writes it. Data authored by the team and shipped with the game is a typed custom `Resource` in a `.tres` file. Data written at runtime — saves, settings, downloads, mod content, anything from another machine — is plain serialized data (JSON, `ConfigFile`, `store_var`) validated at the boundary, never a resource file. The split is not a style preference: Godot resource files can carry embedded scripts, so loading one is executing it (see [Untrusted Resource Files](#untrusted-resource-files)).

### Typed Resources For Designer Data

Model every named, designer-tunable data shape — enemy stats, item definitions, wave tables, tuning curves — as a script with `class_name` extending `Resource`, with typed `@export` properties:

```gdscript
class_name EnemyStats
extends Resource

@export var display_name: String = ""
@export var max_health: int = 10
@export var move_speed: float = 80.0
```

Instances are saved as `.tres` files and assigned in the inspector (`@export var stats: EnemyStats` on the consuming node). This buys what the official data-preferences page calls control and clarity over raw collections: a defined API, type safety, and signals for reactive behavior (`docs.godotengine.org/en/stable/tutorials/best_practices/data_preferences.html`). It also composes with the rest of the contract:

- **Inspector-editable.** Designers tune values without touching code, and the exported types constrain what they can enter. Every `@export` is explicitly typed per [gdscript-style-and-typing.md](gdscript-style-and-typing.md).
- **Diffable.** `.tres` files are text-based and mergeable, per the official version-control guidance (`docs.godotengine.org/en/stable/tutorials/best_practices/version_control_systems.html`). Data changes show up in review like code changes.
- **Located near their consumers.** Store `.tres` files in the feature folder of the scene that uses them (`characters/enemies/goblin/goblin_stats.tres`), per the official project-organization guidance and [project-setup.md](project-setup.md).
- **A sanctioned alternative to autoload state.** The official autoloads page lists shared custom Resources among the preferred alternatives to global singletons for shared data (`docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`); see [signals-and-decoupling.md](signals-and-decoupling.md) before reaching for an autoload.

Two handbook conventions on top of the official guidance, adopted here rather than prescribed by the docs: keep Resource scripts logic-light — exported data, derived read-only getters, and signals, with behavior living on the nodes that consume the data; and treat shipped Resources as read-only at runtime — a loaded resource is shareable across every scene that references the file, so runtime mutation belongs on the owning node or in the save system, never on the shared `.tres`.

### Dictionaries Versus Objects

The official tradeoff: Dictionaries win on raw speed — constant-time insert, erase, and lookup — while object property access is slower because the engine "must make so many checks" traversing script and class hierarchies; objects win on control and clarity (`docs.godotengine.org/en/stable/tutorials/best_practices/data_preferences.html`). Pick by role, not habit:

| Model | Use when | Cost |
|---|---|---|
| Typed custom `Resource` | Named, persisted, designer-edited, or cross-module data shapes | Slower property lookup; one class per shape |
| `RefCounted` / inner `class` | Structured in-process data that never needs a `.tres` or the inspector | Not inspector-editable, not serialized by the editor |
| `Dictionary` | Profiler-proven hot paths; decode target for external JSON; genuinely schemaless data | No schema or type safety; a typo is a silent runtime miss, not a compile error |

Two hard rules fall out of that table. First, a Dictionary chosen "for performance" without profiler evidence is a forbidden pattern — measure per [../operations/performance-and-profiling.md](../operations/performance-and-profiling.md) first. Second, a Dictionary decoded from JSON or a network payload is a boundary artifact: convert it to a typed object at the edge and pass the typed value inward; raw Dictionaries do not travel through gameplay layers.

### Resource Loading Boundaries

`load()`, `preload()`, and `ResourceLoader.load()` only ever touch `res://` paths — files committed to the repository, reviewed like code, and exported inside the game binary. That is the entire legal surface for resource loading.

Everything else is data, not a Resource:

- **Save files** are JSON, `ConfigFile`, or `FileAccess.store_var` payloads under `user://`, validated on read. The official saving-games doc covers the three formats and their tradeoffs (`docs.godotengine.org/en/stable/tutorials/io/saving_games.html`); the format decision and validation rules are owned by [../systems/save-and-load.md](../systems/save-and-load.md).
- **Network input** is untrusted by definition — the official multiplayer guidance is to "treat all client input as untrusted" (`docs.godotengine.org/en/stable/tutorials/networking/high_level_multiplayer.html`); see [../systems/multiplayer.md](../systems/multiplayer.md).
- **Mods and user-generated content** may not enter through `ResourceLoader` at all without the ADR-gated exception described below.

### Untrusted Resource Files

Never load a resource file (`.tres`, `.tscn`, `.res`, `.scn`) that was not shipped with the game. These formats can carry embedded GDScript, and loading the file runs it — an `_init` calling `OS.execute` in a crafted "save file" is arbitrary code execution on the player's machine. This is documented engine behavior in Godot 4.x, acknowledged in the merged docs-safety PR `github.com/godotengine/godot/pull/98168` and the open proposals `github.com/godotengine/godot-proposals/issues/4925` and `github.com/godotengine/godot-proposals/issues/10968`; the engine has no built-in "load without scripts" mode.

Consequences of the rule:

- Save files are never resource files, even though `ResourceSaver.save()` on a custom Resource looks like the least code. A player sharing a save, a cloud-sync bug, or a modded save is then remote code execution.
- Mod and UGC support does not mean "load the modder's `.tres`". If a project genuinely needs to load resource-format content from outside `res://`, that requires an ADR ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)) naming the mitigation — the community mitigation is the Safe Resource Loader addon, which scans for embedded scripts before loading (`github.com/derkork/godot-safe-resource-loader`) — and the exception is scoped to the modding surface, not adopted project-wide.
- Code review treats any `load`-family call whose argument is not a literal `res://` path as a security finding until proven otherwise.

## Common Mistakes And Forbidden Patterns

- Loading a `.tres` or `.res` file from `user://`, a download, or another player with `ResourceLoader.load()` — this executes any embedded script; it is the security boundary of this doc, not a style nit.
- Saving game state with `ResourceSaver.save()` because it round-trips custom Resources "for free", thereby forcing the forbidden load on the way back in.
- Shipping designer-tunable data as untyped Dictionaries or JSON blobs, losing inspector editing, type checks, and reviewable diffs.
- Passing a decoded JSON Dictionary through gameplay layers instead of converting it to a typed object at the boundary.
- Choosing Dictionary over a typed Resource "for performance" with no profiler evidence.
- Putting node behavior (movement, combat logic, scene manipulation) inside a Resource script instead of on the consuming node.
- Mutating a shipped Resource at runtime as instance state — the loaded resource is shareable across every scene referencing the file; runtime state lives on the node or in the save system.
- Building an autoload to hold shared data that a shared custom Resource or a `static` function already covers — see [signals-and-decoupling.md](signals-and-decoupling.md).

## Verification And Proof

```bash
# Every load-family call site; each hit must be a literal res:// path or reviewed as a finding.
grep -rn --include="*.gd" -E "(ResourceLoader\.load|preload|[^_a-z]load)\(" . | grep -v "res://"

# Untyped exports and declarations are caught by the typing gate.
gdlint .
```

Data modeling is done when:

- every designer-tunable data shape is a `class_name` Resource with typed `@export` properties, and its `.tres` instances live beside the scenes that consume them.
- no `load`, `preload`, or `ResourceLoader.load` call site takes a `user://` path, a network-derived path, or any non-literal path that untrusted input can influence — the grep above returns only reviewed hits.
- save and settings data round-trips through the format owned by [../systems/save-and-load.md](../systems/save-and-load.md), with a test rejecting a malformed payload rather than crashing or silently accepting it (test harness per [../quality/testing.md](../quality/testing.md)).
- any Dictionary on a non-boundary code path has a profiler measurement or a schemaless justification attached in review.
- any exception to the untrusted-resource rule exists only as an ADR with a named mitigation.

## Where To Go Next

- [../systems/save-and-load.md](../systems/save-and-load.md) — save formats, validation, and versioning for user-writable data.
- [gdscript-style-and-typing.md](gdscript-style-and-typing.md) — the typing rules every `@export` in a Resource follows.
- [signals-and-decoupling.md](signals-and-decoupling.md) — shared Resources versus autoloads for cross-scene state.
- [scene-and-node-architecture.md](scene-and-node-architecture.md) — where behavior lives while Resources carry the data.
- [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md) — the escape hatch for any exception to the untrusted-resource rule.
