# Localization

Translation keys, source formats, auto-translate boundaries, per-locale assets and fonts, and the pseudolocalization gate that proves layout survives real languages. Guidance targets the current Godot 4.x stable line (`docs.godotengine.org/en/stable`).

## Default Approach

Every player-facing string is a translation key resolved through Godot's translation system — no display text is hardcoded in scripts or scenes outside the translation sources. Localize from day one even for a single-language project: retrofitting keys into a shipped UI is a full rewrite of every scene, while starting with keys costs nothing. The official reference is `docs.godotengine.org/en/stable/tutorials/i18n/internationalizing_games.html`.

### Translation Keys And Sources

- Resolve strings in code with `Object.tr()` (`label.text = tr("LEVEL_5_NAME")`), which "will just look up the text in the translations and convert it if found."
- Plurals go through `tr_n(singular, plural, count)`, never an `if count == 1` branch — "most languages require different strings depending on whether an object is in singular or plural form," and several have more than two plural forms your branch cannot express.
- Use named `String.format()` placeholders, not positional `%s` chains: the docs state named placeholders "should be used whenever possible, as they also allow translators to choose the order in which placeholders appear." Translate whole sentences; never concatenate translated fragments, because word order is not portable across languages.
- Disambiguate identical source strings with a translation context: `tr("Close", "Actions")` versus `tr("Close", "Distance")`.
- Source of truth is a committed CSV (or gettext PO for larger projects — `docs.godotengine.org/en/stable/tutorials/assets_pipeline/importing_translations.html`). CSV shape per `docs.godotengine.org/en/stable/tutorials/i18n/localization_using_spreadsheets.html`: first column is the key, remaining column headers are engine-valid locale codes, and a header starting with `_` "is served as comment and won't be imported." Files "must be saved with UTF-8 encoding without a byte order mark" — Excel defaults to ANSI, so edit in LibreOffice or Google Sheets.
- Import generates compressed `*.translation` resources next to the CSV. These are build artifacts: the source CSV/PO is committed, `*.translation` is gitignored — the rule is owned by [../foundations/project-setup.md](../foundations/project-setup.md) and the committed pattern by [../templates/gitignore.txt](../templates/gitignore.txt).
- An imported CSV "is **not** automatically registered as a translation source" — add the generated `*.translation` files under Project Settings > Localization > Translations, and review that `project.godot` diff like code.
- Set the fallback locale under Project Settings > Internationalization > Locale (empty falls back to `en`). Switch at runtime with `TranslationServer.set_locale()`; the player's chosen locale is a user setting, persisted per [save-and-load.md](save-and-load.md).

### Auto-Translate Boundaries

- Controls such as `Button` and `Label` "will automatically fetch a translation if their text matches a translation key." This is the default path for static UI text: put the key in the scene, let the control translate itself, and keep scripts free of presentation strings.
- Auto-translate is a trust boundary. Any control that ever displays user-generated or external text — player names, chat, save-slot titles, server messages — must have Auto Translate > Mode set to `Disabled` in the inspector, per the official doc. Otherwise a player who names themselves after a translation key gets their name silently rewritten, and attacker-controlled text gains a lookup into your translation table.
- The same boundary applies in code: never pass runtime-assembled or user-supplied strings through `tr()`. `tr()` takes known keys from the committed source files, nothing else.

### Asset Remaps And Fonts

- Locale-specific assets — voice-over, textures or videos with baked-in text — are swapped via Project Settings > Localization > Remaps: register the base resource and one alternative per locale. Code keeps loading the base path; the engine substitutes per locale.
- Fonts do not remap: "The resource remapping system isn't supported for DynamicFonts. Use the DynamicFont fallback system instead." Configure fallback chains on the project theme's fonts so every shipped locale's script has glyph coverage; the theme and its font resources are owned by [ui-and-theming.md](ui-and-theming.md).
- Right-to-left locales (Arabic, Hebrew) get automatic mirroring of anchors, text alignment, and control order; override per control with `text_direction`, `language`, and `layout_direction` only when the automatic result is wrong, and record why.
- Complex-script and emoji rendering in exported builds needs ICU data in the export templates (`docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`); export coverage per locale is proven in the release pipeline, owned by [../operations/ci-and-release.md](../operations/ci-and-release.md).

### Pseudolocalization Checks

Pseudolocalization "simulates changes that might take place during localization" so i18n breakage is caught "early on during development" (`docs.godotengine.org/en/stable/tutorials/i18n/pseudolocalization.html`) — before a single translation exists. Enable it under Project Settings > General > Internationalization > Pseudolocalization (Advanced settings toggle), or at runtime via `TranslationServer.pseudolocalization_enabled`.

- `replace_with_accents` exposes untranslated strings and missing glyphs; `double_vowels` and `expansion_ratio` (the docs suggest 0.3 for ~30% growth) simulate the length expansion real translations bring; `fake_bidi` smoke-tests right-to-left layout; `override` replaces every localized character with `*`, so anything still readable is a hardcoded string that bypassed the key system.
- Keep `skip_placeholders` on so `%s`/`%d` format markers survive the transform.
- Run the pseudolocalization pass as a UI review gate before release: walk every screen with expansion enabled and fix clipping by making containers size to content, not by shortening English — container sizing rules are owned by [ui-and-theming.md](ui-and-theming.md).
- For spot checks of real locales, use the editor's Preview Translation menu, launch with `godot --language <locale>`, or set the Test property under Internationalization > Locale.

## Common Mistakes And Forbidden Patterns

- Player-facing literal strings in scripts or scenes instead of keys in the translation source.
- Concatenating translated fragments, or positional placeholders where `String.format()` named placeholders belong — both hardcode English word order.
- Hand-rolled plural logic (`if count == 1`) instead of `tr_n()`.
- Auto-translate left enabled on a control that displays user-generated or external text.
- Passing runtime or user-supplied strings through `tr()`.
- `*.translation` artifacts committed, or the source CSV saved from Excel with ANSI encoding or a BOM.
- Importing a CSV and assuming it is registered — translations that work in the editor preview but were never added under Localization > Translations.
- Attempting to remap a font through the resource-remap system instead of configuring font fallbacks.
- Containers and labels sized to the English string, discovered clipped only after translations arrive — pseudolocalization exists to catch this before they do.
- Shipping a locale whose script has no glyph coverage in the theme's font fallback chain.

## Verification And Proof

```bash
git ls-files | grep -E '\.translation$'       # expect no output; sources (.csv/.po) are what is committed
git ls-files | grep -E '\.(csv|po)$'          # translation sources present and committed
grep -n 'translations=' project.godot         # every generated .translation registered
godot --headless --check-only path/to/script  # scripts still parse after key refactors
```

Localization is done when:

- every shipped locale launches via `godot --language <locale>` with no missing-key text visible,
- a full-screen walk with pseudolocalization `override` enabled shows no readable English (every readable string is a bypassed key),
- the same walk with `expansion_ratio` 0.3 and `double_vowels` shows no clipped or overflowing controls,
- `fake_bidi` produces mirrored layout with no manually-anchored breakage,
- every control bound to user-generated text has auto-translate disabled, asserted in a scene test per [../quality/testing.md](../quality/testing.md),
- and no `*.translation` artifact is tracked while every source CSV/PO is.

## Where To Go Next

- [ui-and-theming.md](ui-and-theming.md) — theme fonts, fallback chains, and container sizing that survives expansion.
- [save-and-load.md](save-and-load.md) — persisting the player's chosen locale as a user setting.
- [../foundations/project-setup.md](../foundations/project-setup.md) — the VCS rule that ignores `*.translation` and commits sources.
- [../operations/ci-and-release.md](../operations/ci-and-release.md) — proving export builds render every shipped locale.
- [../recipes/add-a-ui-screen.md](../recipes/add-a-ui-screen.md) — the recipe where new screens pick up keys instead of literals.
