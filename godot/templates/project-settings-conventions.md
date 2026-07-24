<!--
Destination: docs/project-settings.md

Fill-in conventions sheet for <PROJECT_NAME>'s `project.godot`. Contract:
- These are the settings every new repo sets DELIBERATELY on day one; everything else
  stays at the engine default until a handbook doc gives a reason to move it.
- Apply values through the editor (Project Settings), commit the `project.godot` diff,
  and record the chosen value plus rationale here IN THE SAME CHANGE. `project.godot`
  is code; this sheet is why the code says what it says.
- Replace every <PLACEHOLDER>. "Engine default, on purpose" is a valid entry — keep the
  row and say so. A missing row means nobody decided.
- The day-one list is owned by godot/foundations/project-setup.md; each section below
  names the doc that owns its policy. Setting paths verified against
  docs.godotengine.org/en/stable/classes/class_projectsettings.html (4.x stable line).
-->

# Project Settings Conventions

Deliberate `project.godot` settings for <PROJECT_NAME>: what each is set to, why, and which handbook doc owns the rule. Review every `project.godot` diff against this sheet.

## Application Settings

| Setting | Value | Why |
|---|---|---|
| `application/config/name` | `<PROJECT_NAME>` | Used by the Project Manager and exporters, and it determines the `user://` data folder path — renaming later strands existing player data under the old folder name. Set it once, correctly. |
| `application/run/main_scene` | `res://<PATH_TO_MAIN>.tscn` | A project that only runs from the editor's current tab is not runnable headless or in CI. |
| `application/config/version` | `<0.1.0>` | The human-readable version identifier exporters use unless overridden per preset; left empty, exports silently ship as `1.0.0`. Bumping it is part of the release process (godot/operations/ci-and-release.md). |

## Display And Rendering

Stretch and base-size guidance is from the official multiple-resolutions doc (`docs.godotengine.org/en/stable/tutorials/rendering/multiple_resolutions.html`); renderer descriptions from the `ProjectSettings` class reference.

| Setting | Value | Why |
|---|---|---|
| `rendering/renderer/rendering_method` | `<forward_plus / mobile / gl_compatibility>` | Chosen per target platform: Forward+ is the "high-end renderer designed for desktop devices"; Mobile has "lower base overhead"; Compatibility targets old/low-end hardware. Decide on day one — it shapes every material and light that follows. |
| `display/window/size/viewport_width` x `viewport_height` | `<1920x1080>` | Base design resolution. Docs baselines: 1920x1080 desktop non-pixel-art, 1280x720 mobile landscape, 720x1280 mobile portrait, 640x360 pixel art. The engine default 1152x648 is a decision by omission — replace it. |
| `display/window/stretch/mode` | `<canvas_items / viewport / disabled>` | `canvas_items` is "recommended for most games that don't use a pixel art aesthetic" and is the new-project default starting in Godot 4.7; `viewport` is "recommended for games that use a pixel art aesthetic"; `disabled` is "recommended for non-game applications". |
| `display/window/stretch/aspect` | `<keep / expand / keep_width / keep_height>` | `keep` letterboxes; `expand` fills the screen at any ratio; `keep_height` is "usually the best option for 2D games that scroll horizontally", `keep_width` for scaling GUIs. State which edges of the world players may see. |
| `display/window/stretch/scale_mode` | `<fractional / integer>` | `integer` floors the scale factor for "a crisp pixel art appearance" — pair it with `viewport` mode for pixel art; otherwise leave `fractional`. |

Window and stretch settings are exercised by real UI: container behavior under resize is owned by godot/systems/ui-and-theming.md.

## Input Map Baseline

Policy is owned by godot/foundations/input-handling.md: gameplay code reads named InputMap actions only, never raw keycodes. Adding an action goes through godot/recipes/add-an-input-action.md; this sheet records the shipped baseline. The committed `[input]` section IS the reset-to-defaults state — `InputMap.load_from_project_settings()` restores exactly what is written here.

- Action names are `snake_case` verb phrases. Never reuse or overload an engine `ui_*` action for gameplay.
- Every action ships with both a keyboard/mouse and a joypad binding, or documents the exclusion.

| Action | Keyboard/Mouse default | Joypad default | Notes |
|---|---|---|---|
| `<move_left>` | `<A / Left>` | `<left stick -X>` | <axis pair with move_right> |
| `<jump>` | `<Space>` | `<bottom face button>` | |
| `<pause>` | `<Escape>` | `<start>` | <consumed by UI layer first> |
| `<ACTION>` | `<BINDING>` | `<BINDING or "none — reason">` | |

## Physics Tick

| Setting | Value | Why |
|---|---|---|
| `physics/common/physics_ticks_per_second` | `60` (engine default) | Stays at 60 unless godot/systems/physics-and-movement.md gives a reason: per the class reference, "CPU usage scales approximately with the physics tick rate", below ~30 "physics behavior can break down", and higher rates buy accuracy for fast-moving objects (e.g. racing games). |

If this row is not 60, record the reason here: <RATIONALE OR "n/a">. Movement code must never encode the tick rate as a literal — raising this setting must not change gameplay speed.

## GDScript Warnings

Typing policy is owned by godot/foundations/gdscript-style-and-typing.md; this sheet records the committed enforcement. Warning levels changed from these values require an ADR.

| Setting | Value | Why |
|---|---|---|
| `debug/gdscript/warnings/untyped_declaration` | Error | The engine default is Ignore; escalating to Error makes an untyped declaration fail compilation project-wide instead of scrolling past in the output panel. |
| `debug/gdscript/warnings/inferred_declaration` | `<Ignore / Warn / Error>` | Optional companion: forbids `:=` inference if the project standardizes on explicit types. The static-typing doc's rule is "stick to one style for consistency" — record the choice. |

## Internationalization

Localization policy is owned by godot/systems/localization.md.

| Setting | Value | Why |
|---|---|---|
| `internationalization/locale/fallback` | `<en>` | The locale used when a translation is missing. The engine default is `en`; set it explicitly so the fallback is a decision, not an accident, and so the fallback locale's CSV/PO column is the one kept complete. |
| Localization > Translations list | `<res://... .translation entries>` | Imported translation sources must be registered here or `tr()` silently returns keys; the registration is a `project.godot` diff — review it like code. |

Pseudolocalization (`internationalization/pseudolocalization/*`) is a development aid for layout testing; it must be off in every committed `project.godot`.

## Proof

- Project opens and runs headless from a clean clone: `godot --headless --quit` exits without script errors, proving the main scene and warning escalations hold outside the editor.
- A headless test asserts `InputMap.has_action()` for every action in the table above, so code, `project.godot`, and this sheet cannot drift (runner per godot/quality/testing.md).
- `git diff project.godot` in any PR touches only settings with a row here, or the PR adds the row.

## Related

- Day-one settings list and repo shape: godot/foundations/project-setup.md
- Input action rules and rebinding: godot/foundations/input-handling.md
- Typing enforcement rationale: godot/foundations/gdscript-style-and-typing.md
- Tick-rate and delta rules: godot/systems/physics-and-movement.md
- Translation workflow: godot/systems/localization.md
- Escalation path for changing an invariant here: godot/decisions/architecture-decision-records.md
