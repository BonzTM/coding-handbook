# Recipe: Add A UI Screen

Use this when the game gains a menu, HUD panel, dialog, or any screen built from Control nodes. For gameplay scenes, use [add-a-scene.md](add-a-scene.md) instead. Layout, theming, and focus rules are owned by [../systems/ui-and-theming.md](../systems/ui-and-theming.md).

## Files To Touch

- a new `snake_case` folder under the UI area (e.g. `ui/pause_menu/`), holding `<screen_name>.tscn`, `<screen_name>.gd`, and screen-only assets together
- the project theme `.tres` if the screen needs a new type variation
- the translation source file (CSV or PO) for every new string key, per [../systems/localization.md](../systems/localization.md)
- the owner scene (usually the `GUI` branch of the main scene) that instantiates or shows the screen
- a test file per [add-a-unit-test.md](add-a-unit-test.md)

## Steps

1. Create the scene with a Control-derived root anchored Full Rect; build the interior from nested containers (`MarginContainer`, H/VBox, `GridContainer`), never from hand-placed offsets (`docs.godotengine.org/en/stable/tutorials/ui/gui_containers.html`).
2. Attach `<screen_name>.gd`, typed per [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md); keep the screen dependency-free and inject external context from the owner, per [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md).
3. Style through the project-wide theme only: reuse existing theme types, add a type variation for any repeated new look, and reserve local overrides for true one-offs (`docs.godotengine.org/en/stable/tutorials/ui/gui_skinning.html`).
4. Enter every user-visible string as a translation key from the first commit — keys in `text` properties auto-translate on Button/Label-style controls, `tr()` for strings built in code (`docs.godotengine.org/en/stable/tutorials/i18n/internationalizing_games.html`); add the keys to the translation source in the same change.
5. Set focus modes deliberately (interactive controls focusable, decorative ones None) and wire focus neighbors plus Next/Previous explicitly — the engine's automatic guessing "may result in unintended behavior" (`docs.godotengine.org/en/stable/tutorials/ui/gui_navigation.html`).
6. Call `grab_focus()` on the default control when the screen opens; navigation runs on the built-in `ui_*` actions, which stay reserved for UI per [../foundations/input-handling.md](../foundations/input-handling.md).
7. Report outcomes as past-tense signals (`options_applied`, `screen_closed`) per [add-a-signal-contract.md](add-a-signal-contract.md); the owner decides what showing or closing the screen means — the screen never frees itself or swaps scenes.
8. In the owner: instantiate under the `GUI` branch, inject dependencies, connect signals in code, and wire the open/close trigger.

## Invariants To Preserve

- no `position`/`size` set on container children; the container reasserts layout on the next resize and the edit silently vanishes
- one project-wide theme remains the single styling source; a repeated look becomes a type variation, never a copied override
- every string in the `.tscn` and script is a translation key, none a display literal
- the screen opens with visible focus and is fully operable without a mouse
- the screen reaches nothing outside itself — no `get_parent()` calls or `World` access; facts leave via signals
- all file and folder names stay `snake_case`

## Proof

- run the screen standalone (F6) at the smallest and largest supported window sizes; nothing overlaps or clips
- gamepad-only pass: focus visible on open, every interactive control reachable, D-pad and Tab order match the visual order
- switch to a second (or pseudolocalized) locale and confirm every string translates and longer text does not break the layout
- `grep -n "theme_override" <screen_name>.tscn` — every hit is a justified one-off
- a unit test instantiates the screen headless, asserts it reaches `_ready()` without errors, and asserts its signals fire on simulated activation
- `gdformat --check` and `gdlint` pass on the new script
