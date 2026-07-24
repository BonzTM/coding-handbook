# Recipe: Add A Save Field

Use this when a feature must persist new state in the player's save file and existing saves in the wild must keep loading correctly.

Format and loader rules are owned by [../systems/save-and-load.md](../systems/save-and-load.md): saves are JSON, `ConfigFile`, or `FileAccess.store_var` data — the official formats for user-writable persistence (`docs.godotengine.org/en/stable/tutorials/io/saving_games.html`). They are never `.tres`/`.res` resources, because resource files can carry embedded scripts that execute on load (`github.com/godotengine/godot/pull/98168`). Shipped read-only data belongs in custom `Resource` classes per [../foundations/resources-and-data.md](../foundations/resources-and-data.md); do not use this recipe for it.

## Files To Touch

- the save service script (the autoload or `class_name` service the repo registered per [add-an-autoload.md](add-an-autoload.md)) — the `SAVE_VERSION` constant, the serialize/deserialize functions, and the migration chain
- the game-state object the field lives on, with a typed declaration and a sane default
- a committed legacy save fixture at the **pre-change** version, e.g. `test/fixtures/saves/save_v<N-1>.json`
- the save-system test file under the layout defined by [../quality/testing.md](../quality/testing.md)

## Steps

1. Bump `SAVE_VERSION` by exactly one. Every save file embeds its version at write time; a file without a readable version is treated as corrupt, not as version 1.
2. Before touching serialization, capture a save produced by the **current** build and commit it as the `v<N-1>` fixture. Fixtures are immutable once committed — they are the only proof that real old saves still load.
3. Add the field to the game-state object with an explicit type and default, per [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md).
4. Extend serialize **and** deserialize together. If the field is a `Vector2`, `Color`, or other engine type and the save is JSON, encode/decode it explicitly — JSON cannot round-trip those types on its own (`docs.godotengine.org/en/stable/tutorials/io/saving_games.html`).
5. Write the migration step `_migrate_v<N-1>_to_v<N>(data)` that fills the new field's default into old data. Migrations chain sequentially — load runs every step from the file's version up to `SAVE_VERSION`, so a v1 save reaches v4 through v2 and v3. Never edit a shipped migration step; append the next one.
6. Validate on load before the data reaches game state: type-check the field, clamp numeric ranges, and fall back to the default on a missing or malformed value. A save file is external input from an untrusted trust boundary, not a memory snapshot.
7. If loading now fails for a case it previously accepted, that is a schema decision, not a bug fix — route it through [../systems/save-and-load.md](../systems/save-and-load.md) and record it if it strands players' saves.

## Invariants To Preserve

- every save file carries its schema version; `SAVE_VERSION` only increases and a value is never reused
- the migration chain is append-only and sequential — one step per version bump, shipped steps never edited
- serialize and deserialize change in the same commit; a field one side knows about and the other does not is a silent data loss
- load never trusts file contents: missing, extra, or wrong-typed fields produce validated defaults or an explicit load error, never a crash or script injection
- `ResourceLoader.load()` never touches a user-writable file — the loader boundary in [../AGENTS.md](../AGENTS.md) (## Repo-Wide Invariants) holds
- committed legacy fixtures are never regenerated to make a failing test pass

## Proof

- round-trip test: build a state with the new field set to a non-default value, save, load, assert equality — green under the repo's headless runner from [../quality/testing.md](../quality/testing.md)
- legacy-load test: the committed `v<N-1>` fixture loads, the new field equals its migration default, and every pre-existing field survives unchanged
- corrupt-input test: a truncated file and a wrong-typed field both produce the documented error path, not a crash
- `gdformat --check` and `gdlint` clean on the touched scripts
- the baseline gate in [../AGENTS.md](../AGENTS.md) (## Baseline Verification) is green
