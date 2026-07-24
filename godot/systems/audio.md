# Audio

Audio defaults for Godot 4.x projects: a small named bus layout committed to the repo, players owned by the scenes that emit sound, explicit spatial settings, and volume persisted as linear values through the settings file.

## Default Approach

All audio routes through named buses, and every sound is played by an `AudioStreamPlayer` (or its 2D/3D variant) that lives inside the scene emitting it. There is **no global sound singleton**. The official best-practices page names the failure mode directly: with a global `Sound.play()` autoload, "there's no longer an easy way to find the source of a bug," and a pooled set of global players is sized wrong for individual scenes — "either you have unnecessary constraints... or the audio arbitrarily cuts out" (`docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`). A scene that owns its players stays self-contained and reusable, which is the dependency rule in [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md).

The one legitimate autoload in the audio system is a music/ambience controller — cross-scene playback state is a wide-scope concern that owns its own data. Add it via [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md); it plays music and nothing else.

### Bus Layout

Godot channels all playback through buses, routed right-to-left into the leftmost **Master** bus, and "the default bus layout is automatically saved to the `res://default_bus_layout.tres` file" (`docs.godotengine.org/en/stable/tutorials/audio/audio_buses.html`). Rules:

- **Commit `default_bus_layout.tres`.** It is text-based project state like any scene; the bus layout is part of the contract, not a local editor preference.
- **Standard layout — `Master`, `Music`, `SFX`, `UI`** — a convention this handbook adopts, not official doctrine. Add sub-buses (e.g. `Ambience` routed into `SFX`) only when a settings slider or a shared effect needs them; buses exist to be mixed or processed as a group, not to mirror the scene tree.
- **Reference buses by name, never by raw index.** Buses are identified by name and can be reordered without breaking connections (`docs.godotengine.org/en/stable/tutorials/audio/audio_buses.html`). `AudioServer.get_bus_index()` "returns -1 if no bus with the specified name exist" (`docs.godotengine.org/en/stable/classes/class_audioserver.html`) — assert the result is not `-1` at startup so a renamed bus fails loudly, not silently on Master.
- **Headroom on Master.** Volume is logarithmic decibels — "for every 6 dB, sound amplitude doubles or halves" — and the docs warn to keep the final mix under 0 dB on Master to avoid clipping (`docs.godotengine.org/en/stable/tutorials/audio/audio_buses.html`). Balance individual sounds with the player's `volume_db`; reserve bus volume for user-facing sliders and mix groups.
- **Effects live on buses**, applied in order top-to-bottom. Put shared processing (reverb, compression, ducking) on a bus, not duplicated per player.

### Scene-Owned Players

- **The emitting scene owns the player.** A door scene owns the `AudioStreamPlayer3D` for its creak; a button owns the click player on the `UI` bus. Parents trigger playback by calling down; completion flows up via the `finished` signal ("Emitted when the audio stops playing," `docs.godotengine.org/en/stable/classes/class_audiostreamplayer3d.html`) — the same call-down/signal-up split as [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md).
- **Set `bus` explicitly on every player.** The property defaults to `"Master"` (`docs.godotengine.org/en/stable/classes/class_audiostreamplayer3d.html`); a player left on Master bypasses the SFX/Music/UI sliders and is a settings-screen bug waiting to be filed.
- **Overlapping instances of the same sound use `max_polyphony`**, not extra players or a global pool. Default is 1; "playing additional sounds after this value is reached will cut off the oldest sounds" (`docs.godotengine.org/en/stable/classes/class_audiostreamplayer3d.html`). Raise it on rapid-fire emitters (gunshots, footsteps) instead of hand-rolling player pools.
- **Plain `AudioStreamPlayer` is for non-positional sound only** — music and UI. It "can play to any bus" without spatialization (`docs.godotengine.org/en/stable/tutorials/audio/audio_streams.html`); anything that exists in the world uses a positional variant.

### Spatial Audio Defaults

Positional sound uses `AudioStreamPlayer2D` in 2D and `AudioStreamPlayer3D` in 3D, which "can position sound in stereo, 5.1 or 7.1 depending on the chosen audio setup" (`docs.godotengine.org/en/stable/tutorials/audio/audio_streams.html`). Defaults this handbook mandates on every 3D player, with engine defaults per `docs.godotengine.org/en/stable/classes/class_audiostreamplayer3d.html`:

- **Set `max_distance` to a real gameplay radius.** The engine default is `0.0`, which means no cutoff — "only has an effect if set to a value greater than 0.0" — so every emitter in the level is mixed and audible everywhere. Distant silent emitters are wasted mixing work; see [../operations/performance-and-profiling.md](../operations/performance-and-profiling.md).
- **Tune loudness with `unit_size`** ("higher values make the sound audible over a larger distance," default `10.0`) and leave `attenuation_model` on the default inverse-distance model unless a specific sound justifies otherwise.
- **Leave `panning_strength` at `1.0`** project-wide; adjust the project-level panning setting, not per-node values, if the whole mix needs softer panning.
- **Doppler is opt-in** — `doppler_tracking` is disabled by default and must be enabled per player; enable it only on fast movers where the effect is audible.
- **Area3D reverb routing** sends dry and wet audio to separate buses for room acoustics. Caveat for web exports: reverb buses are not supported when the stream's playback mode is Sample — set those streams to Stream mode (`docs.godotengine.org/en/stable/tutorials/audio/audio_streams.html`).

### Volume And Settings Persistence

Settings sliders control **bus** volume, and the values persist as linear floats:

- **Sliders are linear 0.0–1.0, applied via `AudioServer.set_bus_volume_linear()`** (equivalent to `set_bus_volume_db()` with `@GlobalScope.linear_to_db()` applied — `docs.godotengine.org/en/stable/classes/class_audioserver.html`). Mapping a slider straight onto decibels makes the top of the range do almost nothing and the bottom fall off a cliff, because dB is logarithmic.
- **Store linear values, one key per named bus** (`audio/music_volume = 0.8`), in the user-settings `ConfigFile` owned by [save-and-load.md](save-and-load.md). Never store raw dB; `-80.0` in a hand-edited config file is indistinguishable from a corrupt value.
- **Clamp on load.** Settings files are external input: clamp each value to `[0.0, 1.0]` and fall back to the default on parse failure before it touches `AudioServer`.
- **Apply saved volumes at startup before the first sound plays** — in the settings autoload's `_ready()` — or the title-screen music plays one frame at full volume.
- **Mute is `AudioServer.set_bus_mute()`** (`docs.godotengine.org/en/stable/classes/class_audioserver.html`), stored as its own boolean. Do not fake mute by writing `-80 dB` over the user's slider value — unmuting then has nothing to restore.

## Common Mistakes And Forbidden Patterns

- **A global `Sound.play()` autoload for effects.** The exact anti-pattern the official autoload page warns about: untraceable playback sources and a pool sized wrong for every scene. Scene-owned players; music controller is the only audio autoload.
- **Players left on the `Master` bus.** Every player names its bus, or the volume sliders silently do not cover it.
- **Hardcoded bus indices.** `set_bus_volume_db(1, ...)` breaks the moment buses are reordered; resolve by name via `get_bus_index()` and assert it is not `-1`.
- **Linear sliders mapped to dB ranges.** Interpolating `-80..0` dB from a slider; use `set_bus_volume_linear()` or `linear_to_db()`.
- **`max_distance` left at `0.0`** on 3D players, mixing every emitter in the level at all times.
- **Positional sounds on plain `AudioStreamPlayer`**, so the sound ignores the listener entirely.
- **Hand-rolled player pools for overlap** instead of `max_polyphony`.
- **Uncommitted `default_bus_layout.tres`**, so bus names exist on one machine and `get_bus_index()` returns `-1` on every other.
- **Mix state hidden in code.** Setting bus volumes/effects imperatively at startup instead of in the committed bus layout, so the editor's Audio panel lies about the real mix.

## Verification And Proof

- **Bus contract test.** A unit test (see [../quality/testing.md](../quality/testing.md)) asserts `AudioServer.get_bus_index()` returns a valid index for every bus name the project references, and fails on `-1`.
- **Player-routing sweep.** A test (or CI script) walks project scenes and asserts no `AudioStreamPlayer*` node has `bus == "Master"` outside an allowlist. Grep is the cheap version: `grep -rn 'AudioStreamPlayer' --include='*.tscn' | grep -v 'bus ='` surfaces players on the default bus.
- **Settings round-trip test.** Write a volume, reload the `ConfigFile`, assert the bus's linear volume matches; corrupt the value and assert the clamp/fallback path holds.
- **Mix check in the editor.** Play the loudest plausible scene with the Audio bottom panel open and confirm the Master meter stays below 0 dB.
- **Layout is in version control.** `git ls-files default_bus_layout.tres` returns the file; a fresh clone plays with the same mix.

## Related

- [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md) — scene self-containment and dependency injection, the rule that puts players inside emitters.
- [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md) — `finished` and other completion signals flowing up.
- [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md) — adding the music controller the approved way.
- [save-and-load.md](save-and-load.md) — the `ConfigFile` settings surface that owns persisted volumes.
- [../operations/performance-and-profiling.md](../operations/performance-and-profiling.md) — measuring mixing cost before adding buses or emitters.
