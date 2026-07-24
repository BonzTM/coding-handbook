# UI And Theming

Defaults for the UI layer: container-driven Control layout, one project-wide Theme with type variations, and keyboard/controller focus rules that make every screen operable without a mouse.

## Default Approach

A UI screen is a self-contained Control scene under the `GUI` branch of the main scene, per [../foundations/scene-and-node-architecture.md](../foundations/scene-and-node-architecture.md). It talks to the rest of the game by signals up and injected calls down, per [../foundations/signals-and-decoupling.md](../foundations/signals-and-decoupling.md); it never reaches into `World`. New screens follow [../recipes/add-a-ui-screen.md](../recipes/add-a-ui-screen.md). Version-specific guidance below is pinned to the current Godot 4.x stable line.

### Control Layout Discipline

- **Containers own layout inside a screen.** Under any Container-derived node, "all children Control nodes give up their own positioning ability" — manual positions are ignored or invalidated on the parent's next resize (`docs.godotengine.org/en/stable/tutorials/ui/gui_containers.html`). Never set `position` or `size` on a container child in code or in the inspector.
- **Anchors place roots, containers place contents.** Use anchor presets for the screen root (Full Rect) and for free-floating elements like a HUD corner widget; anchors position offsets relative to the parent so the layout survives resolution and aspect-ratio changes (`docs.godotengine.org/en/stable/tutorials/ui/size_and_anchors.html`). Everything inside the root is container-managed.
- **Express sizing intent through size flags**, not pixel constants: Fill/Expand plus shrink modes, with Stretch Ratio for proportional splits. A hard-coded pixel width is a layout bug waiting for the next resolution.
- **Nest containers instead of computing layout.** "The real strength of containers is that they can be nested (as nodes), allowing the creation of very complex layouts that resize effortlessly" (`gui_containers.html`). Layout math in `_process()` or `_ready()` is a forbidden substitute for a container tree.
- **Pick the container that names the intent**: `MarginContainer` for padding, `PanelContainer` for a styled background, `ScrollContainer` for overflow, H/VBox for stacks, `GridContainer` for grids. A generic `Control` with hand-tuned offsets where a container fits is a smell.

### Project-Wide Theme

- **One Theme resource for the whole project**, saved as a `.tres` (text-based and diffable, per [../foundations/project-setup.md](../foundations/project-setup.md)) and registered under Project Settings > GUI > Theme > Custom (`docs.godotengine.org/en/stable/tutorials/ui/gui_skinning.html`).
- Theme lookup runs local overrides first, then themes on the control and its Control ancestors up the tree, then the project-wide theme, then the built-in default: "Whenever a control has a local theme item override, this is the value that it uses" (`gui_skinning.html`). Every level you use above the project theme is a level that can silently shadow a project-theme edit — use as few levels as possible.
- A theme is configuration, not behavior: "A theme only describes the configuration... it is still the job of each individual control to use that configuration" (`gui_skinning.html`). Custom-drawn controls must read their items from the theme, not hard-code colors.
- **Branch-level themes are the exception, not the pattern.** Attach a second Theme resource to a subtree root only for a genuinely distinct UI region (e.g. an in-game terminal styled unlike the menus); it cascades to all children and falls back to the project theme. More than a couple of these requires an ADR per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md).

### Type Variations And Overrides

- **A repeated look is a type variation, never a copied override.** Type variations are named theme types that "extend another, base type" — they inherit the base control's style and replace only what differs (`docs.godotengine.org/en/stable/tutorials/ui/gui_theme_type_variations.html`). Define `DangerButton` or `TitleLabel` once in the project theme; a control opts in via its Theme Type Variation property.
- The official docs call out the alternative as the failure mode: per-control overrides "quickly become hard to manage, if you need to share the same custom look between several controls" (`gui_theme_type_variations.html`).
- **Local theme overrides are for true one-offs only** — a single control needing a granular tweak "while not affecting anything else in the project" (`gui_skinning.html`). The second control that needs the same tweak converts it into a type variation in the same change.
- Styling from code (`add_theme_color_override()` and friends) is reserved for runtime state changes (e.g. flashing a value red on damage). Static appearance lives in the theme, where designers can edit it without touching scripts.

### Focus And Navigation

- **Every screen sets initial focus when it opens.** "For keyboard and controller navigation to work correctly, any node must be focused by using code when the scene starts" — call `grab_focus()` on the default control in `_ready()` (or when the screen becomes visible) (`docs.godotengine.org/en/stable/tutorials/ui/gui_navigation.html`). A screen with no focused control is unusable on gamepad.
- **Set focus modes deliberately.** Buttons default to All, Labels to None; interactive controls that should never take keyboard focus get Click, decorative controls get None. Fewer focusable nodes means fewer wrong stops on the navigation path.
- **Wire focus neighbors explicitly** for any non-trivial layout: "If a node has no focus neighbor configured, the engine will try to guess the next control automatically. This may result in unintended behavior" (`gui_navigation.html`). Directional neighbors cover D-pad/arrows; Next/Previous cover Tab order.
- The built-in `ui_*` actions (`ui_up`, `ui_accept`, `ui_focus_next`, ...) are reserved for UI focus and never reused as gameplay actions; gameplay input lives in project-defined actions handled in `_unhandled_input()`, so the GUI gets first refusal on every event. Ownership of that split is [../foundations/input-handling.md](../foundations/input-handling.md).
- User-visible strings in UI scenes are translation keys, not literals — owned by [localization.md](localization.md).

## Common Mistakes And Forbidden Patterns

- Setting `position`/`size` on a container child, in code or the inspector — the container reasserts layout on the next resize and the edit silently vanishes.
- Pixel-perfect layouts built from hard-coded offsets on a plain `Control`, verified only at the developer's own resolution.
- Layout arithmetic in `_process()`/`_ready()` replicating what a container tree does declaratively.
- Copy-pasted local theme overrides expressing a shared look that belongs in a type variation.
- Scattered Theme resources on individual controls competing with the project-wide theme, so a project-theme edit changes nothing visible.
- Static styling applied from scripts, hiding appearance decisions from the theme editor and from designers.
- A screen that opens with no focused control, or relies on the engine's focus-neighbor guessing across a grid of buttons.
- `ui_accept`/`ui_cancel` doubling as gameplay actions — menu navigation and gameplay now fight over the same event.
- Display strings hard-coded in `.tscn` files instead of translation keys.

## Verification And Proof

- run each screen at the smallest and largest supported window sizes and at a non-default aspect ratio; nothing overlaps, clips, or pins to a wrong corner
- play every screen with only a gamepad: initial focus is visible on open, every interactive control is reachable, Tab/D-pad order matches the visual order
- audit override creep: `grep -rn "theme_override" --include="*.tscn" .` — every hit is either a justified one-off or a conversion candidate to a type variation
- confirm the project theme is wired: temporarily change a base `Button` color in the theme and verify it propagates to every screen
- scene smoke tests instantiate each UI screen headlessly and assert it reaches `_ready()` without errors, per [../quality/testing.md](../quality/testing.md); they run in CI per [../operations/ci-and-release.md](../operations/ci-and-release.md)
