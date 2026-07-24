# GDScript Style And Typing

The GDScript conventions every repo adopts: the official style guide as the formatting and naming baseline, the prescribed script member order, and static typing enforced project-wide by escalating the `UNTYPED_DECLARATION` warning to an error.

## Default Approach

Adopt the official GDScript style guide wholesale (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/gdscript_styleguide.html`) and make every declaration typed. The style guide itself says "Keeping your code consistent in your projects and within your team is more important than following this guide to a tee" — this handbook removes that latitude by making the guide the repo standard, so consistency and guide-compliance are the same thing. Deviations are ADR material, not per-file preferences.

### Style Guide Baseline

Formatting is enforced by `gdformat`, not by review; the tool configuration and gate wiring live in [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md). The rules worth knowing because they show up in review anyway:

- Tabs for indentation. Lines under 100 characters, ideally under 80. One statement per line. Double quotes unless single quotes avoid escapes.
- Naming, per the style guide:

| Element | Convention | Example |
|---|---|---|
| File | `snake_case.gd` | `health_component.gd` |
| Class (`class_name`) | PascalCase | `HealthComponent` |
| Functions and variables | snake_case | `take_damage`, `max_health` |
| Private or virtual members | leading underscore | `_recalculate_path()`, `_counter` |
| Constants | CONSTANT_CASE | `MAX_SPEED` |
| Enum name / members | PascalCase / CONSTANT_CASE | `Element.FIRE` |
| Signals | past-tense snake_case | `door_opened`, `item_collected` |

- File and folder casing is not cosmetic: exported PCK filesystems are case-sensitive while desktop filesystems usually are not, so inconsistent casing breaks exported builds. That rule and the directory layout it protects are owned by [project-setup.md](project-setup.md).
- Signal naming is the surface of a larger contract — signals respond to what happened, they do not command what happens next. The design rules are owned by [signals-and-decoupling.md](signals-and-decoupling.md).

### Script Member Order

Every script declares its members in the order the style guide prescribes (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/gdscript_styleguide.html`):

1. `@tool`, `@icon` annotations
2. `class_name`
3. `extends`
4. doc comment
5. signals
6. enums
7. constants
8. `@export` variables
9. public variables, then private (`_`-prefixed) variables
10. `@onready` variables
11. `_init()`, `_ready()`, and other built-in virtual callbacks
12. public methods
13. private methods
14. inner classes

`gdlint` enforces this mechanically via its `class-definitions-order` check (`github.com/Scony/godot-gdscript-toolkit/wiki/3.-Linter`), configured through the committed [../templates/gdlintrc.txt](../templates/gdlintrc.txt) and routed via [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md). Do not reorder members to group "related" code; a reader locates any member by kind, not by feature.

### Static Typing Enforcement

Every declaration carries a type: variables, `@export` vars, function parameters, and return values (`-> void` included). Typed GDScript is not a taste decision — "typed GDScript improves performance by using optimized opcodes when operand/argument types are known at compile time," and it is what makes autocompletion and refactoring reliable (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/static_typing.html`).

The docs support two typed styles — explicit (`var damage: float = 10.5`) and inferred (`var damage := 10.5`) — and advise sticking to one. This handbook's convention (adopted here, not prescribed by the docs): explicit types on every API surface — `@export` vars, signal parameters, function parameters, and return types — and `:=` inference allowed for locals only when the right-hand side makes the type obvious (a literal, a constructor, or a typed call). When the right-hand side is a `Variant`-returning call, write the type explicitly.

Treat `as` casts as fallible: the docs warn that `as` "silently casts the variable to `null` in case of a type mismatch at runtime, without an error/warning" (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/static_typing.html`). Check or assert the result before use; never chain a call off a bare `as` expression.

### Warnings As Gate

Enforcement is a project setting, not a review habit. Per the GDScript warning system, warnings are configured under Project Settings > Debug > GDScript (Advanced Settings enabled), each escalatable individually: "you can turn them into errors if you'd like. This way your game won't compile unless you fix all warnings" (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/warning_system.html`).

- Set `debug/gdscript/warnings/untyped_declaration` to **Error**. An untyped declaration now fails compilation instead of scrolling past in the output panel. This setting is part of the committed `project.godot` conventions in [../templates/project-settings-conventions.md](../templates/project-settings-conventions.md).
- Leave `INFERRED_DECLARATION` at its default. It exists for teams that forbid `:=` entirely (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/static_typing.html`); this handbook permits inference for obvious locals, so enabling it would fight the convention above.
- Per-line suppression exists — `@warning_ignore("untyped_declaration")` (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/warning_system.html`) — but for typing warnings it is forbidden by default. An untyped declaration that "cannot" be typed is a design smell (usually a `Variant` leaking across a boundary); fix the boundary. A genuinely unavoidable suppression carries a comment stating why and is a review flag.
- Never downgrade a warning project-wide to silence noise. Warning levels changed from the template require the same justification as any other invariant change: an ADR per [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md).

## Common Mistakes And Forbidden Patterns

- Untyped declarations anywhere — `var health = 100` instead of `var health: int = 100` — or functions without return types.
- Suppressing `untyped_declaration` with `@warning_ignore` instead of fixing the type, or downgrading the warning in `project.godot`.
- Chaining calls off a bare `as` cast; a runtime type mismatch yields `null` silently and the failure surfaces far from its cause.
- Passing `Variant` through layers because the original value came from an untyped API (`get_node`, JSON parsing) — convert to a typed value at the boundary, once.
- Spaces for indentation, or mixed tabs and spaces.
- camelCase or PascalCase function and variable names; PascalCase `.gd` filenames (breaks exported builds — see [project-setup.md](project-setup.md)).
- Signals named as imperative commands (`open_door`) rather than past-tense facts (`door_opened`) — the contract side is owned by [signals-and-decoupling.md](signals-and-decoupling.md).
- Grouping script members by feature instead of the prescribed member order, or scattering `@onready` vars among methods.
- Hand-formatting to personal taste and arguing with `gdformat` output — the formatter's output is the style, per [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md).

## Verification And Proof

```bash
gdformat --check .
gdlint .
godot --headless --check-only -s res://scripts/health_component.gd
```

- `gdformat --check` and `gdlint` are the batch gates and run in CI on every push; the exact invocation and configuration are owned by [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md).
- `--check-only` will "Only parse for errors and quit (use with `--script`)" (`docs.godotengine.org/en/stable/tutorials/editor/command_line_tutorial.html`). With `untyped_declaration` escalated to Error, an untyped declaration fails this parse — use it for a fast single-file check without opening the editor.
- The headless test run in [../quality/testing.md](../quality/testing.md) loads every script under test, so a typing regression that slips past a spot check still fails CI per [../operations/ci-and-release.md](../operations/ci-and-release.md).
- In review, confirm `project.godot` diffs never touch `debug/gdscript/warnings/*` without an ADR link in the PR.
