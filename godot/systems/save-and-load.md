# Save And Load

The persistence contract: which format a save file uses, how its schema is versioned and migrated, and how writes stay atomic and loads survive corruption. Guidance targets the current Godot 4.x stable line (`docs.godotengine.org/en/stable`).

## Default Approach

A save file is a wire contract with your own past releases, and it lives on a disk you do not control. Treat it accordingly: declare the shape explicitly, version it from the first release, write it atomically, and treat every byte read back as untrusted input.

- All saves and settings live under `user://` — the per-user, per-project writable directory (`docs.godotengine.org/en/stable/tutorials/io/data_paths.html`). Never write into `res://`; it is read-only in exported builds.
- One save service owns all persistence I/O. It is an autoload registered via [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md) — a wide-scope system that owns its data, the legitimate autoload case per [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md). Gameplay code hands it typed state and receives typed state; no scene script opens a `FileAccess` to a save path directly.
- The default game-state format is JSON with a versioned envelope; user settings go in a `ConfigFile`; binary `store_var` is the measured exception. See [Save Format Selection](#save-format-selection).
- Loading is parse, then migrate, then validate — in that order, with a backup fallback at every step. See [Atomic Writes And Corruption Recovery](#atomic-writes-and-corruption-recovery).
- Adding or changing a persisted field goes through [../recipes/add-a-save-field.md](../recipes/add-a-save-field.md); it is a schema change, not a code tweak.

## Save Format Selection

The official saving-games doc offers three mechanisms (`docs.godotengine.org/en/stable/tutorials/io/saving_games.html`). Each has one job:

| Format | Use for | Never for |
|---|---|---|
| JSON (`JSON.stringify` / `JSON.parse_string`) | Game-state saves: progress, inventory, world state | Raw engine types without explicit encoding |
| `ConfigFile` | User settings: volume, keybinds, window mode | Game-state graphs; object-valued entries |
| `FileAccess.store_var` / `get_var` | Large binary payloads where profiling shows JSON cost matters | Anything with `full_objects` / `allow_objects` enabled |
| Custom `Resource` (`.tres`) via `ResourceLoader` | Shipped, read-only design data inside `res://` | Any user-writable file — see [Never Load Untrusted Resources](#never-load-untrusted-resources) |

- **JSON is the default** because it is human-readable, diffable in a bug report, and forces an explicit schema. Its known limit: JSON cannot round-trip Godot types — the docs call out that saving `Vector2`, `Color`, and friends requires you to encode and decode them yourself (`docs.godotengine.org/en/stable/tutorials/io/saving_games.html`). Encode them as explicit fields (`{"x": ..., "y": ...}`) in the save service; never rely on `var_to_str` stringification leaking into the schema.
- **`ConfigFile` is for configuration only.** The docs are direct: "If you're looking to save user configuration, you can use the ConfigFile class for this purpose" (`docs.godotengine.org/en/stable/tutorials/io/saving_games.html`). Store primitives — numbers, strings, bools, arrays of those. `ConfigFile` values are Variants and can carry serialized objects, which makes a settings file on disk a code-execution surface under the same rule as resources below (`github.com/godotengine/godot/pull/98168`).
- **Binary `store_var` is an optimization, not a default.** Reach for it only when a profiled save is too large or too slow as JSON, and record the decision in an ADR ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)). Always leave `full_objects` (write) and `allow_objects` (read) at their default `false` (`docs.godotengine.org/en/stable/classes/class_fileaccess.html`).
- Settings and game state never share a file. A corrupted save must not take the player's keybinds with it, and a settings reset must not touch progress.

## Schema Versioning And Migration

- Every save file carries a top-level integer `version`, written on every save and read before anything else on load. A file without a version field is version-unknown and is treated as corrupt, not guessed at.
- Evolution is additive by default: a new field gets a default value applied when the key is absent, and does not bump the version. Renames, removals, type changes, and semantic changes bump the version and ship with a migration.
- Migrations are a ladder of pure, stepwise functions — `_migrate_1_to_2(data)`, `_migrate_2_to_3(data)` — each taking and returning a `Dictionary`. Load runs them in order from the file's version up to the current one; the loop is bounded by the version distance. No migration reads files, touches the scene tree, or skips a step.
- A version **newer** than the running build refuses to load and leaves the file untouched. An older build silently rewriting a newer save destroys data the newer build needed; tell the player instead.
- After migration, validate: required keys present, types correct, values in range, enum-like strings in the allowed set. Validation failure is a load failure and falls through to the backup, exactly like a parse failure.
- The migrated result is written back through the atomic path below only after validation succeeds, and the pre-migration file is kept as the backup generation until the next successful save.
- Keep one committed fixture save file per historical version under `test/`; they are the proof the ladder still works (see [Verification And Proof](#verification-and-proof)).

## Never Load Untrusted Resources

The rule is owned by [../foundations/resources-and-data.md](../foundations/resources-and-data.md): custom `Resource`s are for shipped, read-only data. This section is its application to persistence, where the file being loaded is user-writable by definition.

- `.tres`, `.tscn`, `.res`, and `.scn` files can carry embedded GDScript, and loading such a file runs that script — including `_init()` — so loading a tampered resource file is arbitrary code execution (`github.com/godotengine/godot/pull/98168`, `github.com/godotengine/godot-proposals/issues/4925`). There is no built-in "load without scripts" mode as of 4.x (`github.com/godotengine/godot-proposals/issues/10968`).
- Therefore: never call `ResourceLoader.load()` (or `load()` / `preload()`) on any path under `user://`, any cloud-synced save, any mod or UGC file, or anything another player produced. A "save my custom Resource with `ResourceSaver`" tutorial pattern is forbidden here regardless of how convenient the inspector round-trip is.
- The same rule covers Variant deserialization. The `FileAccess.get_var` docs warn: "Deserialized objects can contain code which gets executed. Do not use this option if the serialized object comes from untrusted sources to avoid potential security threats such as remote code execution" (`docs.godotengine.org/en/stable/classes/class_fileaccess.html`). `get_var(true)`, `bytes_to_var_with_objects`, and `str_to_var` on object-bearing strings are all the same door.
- The community mitigation — a scanning loader addon (`github.com/derkork/godot-safe-resource-loader`) — is a defense-in-depth option for a project that genuinely must load resource-format UGC, not an alternative to this rule. Adopting it is an addon decision under [../foundations/project-setup.md](../foundations/project-setup.md) and requires an ADR.

## Atomic Writes And Corruption Recovery

A save that dies mid-write (power loss, crash, full disk) must not destroy the previous good save. The write path is: serialize to a string first, write to a temporary file, flush, then rename over the real path.

```gdscript
func _write_atomically(path: String, text: String) -> Error:
	var tmp_path := path + ".tmp"
	var file := FileAccess.open(tmp_path, FileAccess.WRITE)
	if file == null:
		return FileAccess.get_open_error()
	file.store_string(text)
	file.flush()
	file.close()
	return DirAccess.rename_absolute(tmp_path, path)
```

- `FileAccess.flush()` "writes the file's buffer to disk" (`docs.godotengine.org/en/stable/classes/class_fileaccess.html`); the rename is the commit point, since `DirAccess.rename_absolute` moves the file and overwrites an existing destination (`docs.godotengine.org/en/stable/classes/class_diraccess.html`). A reader never observes a half-written save — it sees the old file or the new one.
- Check every return value on this path: `FileAccess.open()` returns `null` on failure and the cause comes from `FileAccess.get_open_error()`; `store_var` returns a success bool; `rename_absolute` returns an `Error` (`docs.godotengine.org/en/stable/classes/class_fileaccess.html`). A save that fails must report failure to the caller — a silently ignored `Error` here is data loss deferred.
- Keep one rolling backup generation: before the rename, move the current good save to `save.json.bak` (or copy the new file over the backup after a verified write). The load order is main file, then backup, then a fresh default state — and reaching the backup or the default is surfaced to the player, never silent.
- The load path treats failure as an expected input, not an exception: unreadable file, JSON parse error, missing or future `version`, and validation failure all take the same fallback ladder without crashing.
- Autosave frequency is bounded and off the critical path — saving every frame or in `_process()` is an I/O storm; save on meaningful boundaries (checkpoint, scene exit, quit) through the one save service.

## Common Mistakes And Forbidden Patterns

- Calling `ResourceLoader.load()` on anything under `user://`, a mod folder, or downloaded content — that file can carry an embedded script and loading it runs the script.
- Using `store_var(value, true)` or `get_var(true)` on user-writable data; `full_objects` deserialization is code execution per the `FileAccess` docs.
- Writing directly over the live save file instead of temp-write, flush, rename — a crash mid-write corrupts the only copy.
- Shipping a save schema with no `version` field, then shape-sniffing old files forever.
- Passing `Vector2`/`Color`/`Transform` values through `JSON.stringify` and expecting them back — JSON cannot round-trip them; encode explicitly.
- Ignoring the `null` return of `FileAccess.open()` or the `Error` from `rename_absolute`, so a failed save reports success.
- Silently resetting progress on a parse failure instead of falling back to the backup and telling the player.
- Storing user settings inside the game-state save, or game state inside the `ConfigFile`.
- An older build overwriting a save whose `version` is newer than it understands.
- Migrating in place with no surviving pre-migration copy, so a migration bug is unrecoverable.
- Scene scripts doing their own file I/O instead of going through the save service.

## Verification And Proof

Tests run under the framework and layout owned by [../quality/testing.md](../quality/testing.md), headless in CI per [../operations/ci-and-release.md](../operations/ci-and-release.md).

- Every persisted field has a round-trip test: save, load, and compare typed state — added alongside the field per [../recipes/add-a-save-field.md](../recipes/add-a-save-field.md).
- A golden fixture pins the current envelope — exact keys and the `version` value — so an accidental schema change shows up as a diff.
- One committed fixture per historical schema version loads through the migration ladder to a validated current-version state.
- Corruption tests feed the loader a truncated file, invalid JSON, a missing `version`, and a future `version`, and assert the backup fallback runs and nothing crashes or clobbers the input file.
- An interrupted-write test (temp file present, rename never happened) proves the previous save loads intact.
- Static proof that the forbidden loaders are absent from the save path:

```bash
grep -rn --include="*.gd" -E "get_var\(true\)|bytes_to_var_with_objects|store_var\(.+,\s*true\)" .
grep -rn --include="*.gd" -E "(ResourceLoader\.load|[^_a-z]load)\(.*user://" .
```

Both greps return nothing, or every hit is justified in an ADR.

## Where To Go Next

- [../foundations/resources-and-data.md](../foundations/resources-and-data.md) — owner of the untrusted-resource rule this doc applies.
- [../recipes/add-a-save-field.md](../recipes/add-a-save-field.md) — the end-to-end recipe for a schema change.
- [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md) — registering the save service.
- [../quality/testing.md](../quality/testing.md) — where the round-trip, migration, and corruption tests live.
- [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md) — the escape hatch for binary formats and resource-loading exceptions.
