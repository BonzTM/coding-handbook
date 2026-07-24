# Project Setup

Default repository shape, naming rules, addon policy, baseline settings, and version-control hygiene for new Godot projects. Guidance targets the current Godot 4.x stable line (`docs.godotengine.org/en/stable`).

## Default Approach

Organize by feature, not by file type. Official guidance is to group assets as close as possible to the scenes that use them — a scene, its script, and its art live in the same directory (`docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html`). A global `scenes/` + `scripts/` + `assets/` split by file type is the forbidden inverse of this rule. Start every new repo from [../checklists/new-project.md](../checklists/new-project.md) and the files in [../templates/README.md](../templates/README.md).

## Directory Layout

The official docs prescribe feature grouping and a top-level `addons/`; the specific `autoload/`, `ui/`, and `test/` directories below are this handbook's convention layered on that guidance, not part of the official doc.

```text
project/
  project.godot
  .gitignore
  .gitattributes
  LICENSE
  addons/
  autoload/
  characters/
    player/
      player.tscn
      player.gd
      player.png
    enemies/
  levels/
  ui/
    theme/
  test/
```

### What Goes Where

- `project.godot`: the project settings file, always committed; settings conventions live in [../templates/project-settings-conventions.md](../templates/project-settings-conventions.md)
- `addons/`: third-party assets and editor plugins only — never first-party code (see [Addons And Third-Party Assets](#addons-and-third-party-assets))
- `autoload/`: scripts registered as autoload singletons; adding one goes through [../recipes/add-an-autoload.md](../recipes/add-an-autoload.md), and the (narrow) cases that justify one are owned by [signals-and-decoupling.md](signals-and-decoupling.md)
- `characters/`, `levels/`, `ui/`: feature directories — each scene co-located with its script and the assets only it uses
- `ui/theme/`: the project-wide Theme resource and its styleboxes; theming rules are owned by [../systems/ui-and-theming.md](../systems/ui-and-theming.md)
- `test/`: unit and integration tests; framework choice and layout are owned by [../quality/testing.md](../quality/testing.md)

Assets shared by genuinely many features may live in a common directory, but reach for that only after co-location fails — a shared directory that accumulates everything is the file-type split reappearing under a new name.

## Naming And Case Rules

- **`snake_case` for all folder and file names**, with the single exception of C# scripts, which are PascalCase to match their class name (`docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html`).
- The rationale is not style: Godot's exported PCK filesystem is case-sensitive while Windows and macOS filesystems typically are not, so a path that resolves in the editor with the wrong case breaks only in the exported build.
- Node names are PascalCase, matching built-in node casing.
- In-code naming — classes, functions, signals, constants, enums — is owned by [gdscript-style-and-typing.md](gdscript-style-and-typing.md); do not restate it here.

## Addons And Third-Party Assets

- All third-party content goes in top-level `addons/`, per the official project-organization doc. Nothing first-party lives there.
- Treat `addons/` as vendored code: commit it, never edit it in place. A local patch to an addon is lost on upgrade; if a patch is unavoidable, record it in an ADR ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)) so the next upgrade knows to re-apply or drop it.
- Record where each addon came from and which version is vendored, so upgrades are a diff against a known base rather than archaeology.
- Adding an addon is a dependency decision, not a download: check maintenance health and Godot-4 compatibility before it lands, and prefer engine features over addons when both exist.
- Use an empty `.gdignore` file to exclude a folder from Godot's importer (`docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html`) — useful for tooling directories that should never generate `.import` metadata.

## Baseline Project Settings

Set these in `project.godot` on day one; the full annotated list is owned by [../templates/project-settings-conventions.md](../templates/project-settings-conventions.md).

- `application/config/name` and the main scene — a project that only runs from the editor's current tab is not runnable in CI.
- The GDScript `UNTYPED_DECLARATION` warning, which enforces static typing project-wide; the typing policy itself is owned by [gdscript-style-and-typing.md](gdscript-style-and-typing.md).
- Input actions in the InputMap rather than raw event checks in code; owned by [input-handling.md](input-handling.md).
- The physics tick rate stays at the default 60 unless [../systems/physics-and-movement.md](../systems/physics-and-movement.md) gives a reason to move it.

Settings changes are diffs to a committed text file — review them like code, because they are code.

## Version Control Defaults

The official VCS doc (`docs.godotengine.org/en/stable/tutorials/best_practices/version_control_systems.html`) defines the baseline; the committed form lives in [../templates/gitignore.txt](../templates/gitignore.txt).

- **Ignore `.godot/`** — it is the import and project cache, fully regenerable, and enormous in diffs.
- **Ignore `*.translation`** — binary artifacts of CSV translation imports; the source CSV/PO files are what gets committed (see [../systems/localization.md](../systems/localization.md)).
- **Keep scenes and resources in text formats** (`.tscn`, `.tres`), which Godot generates as "mostly readable and mergeable files"; binary `.scn`/`.res` in version control forfeits diff and merge for no gain.
- **Set up Git LFS for binary assets before the first commit** of models, images, audio, and fonts, with a `.gitattributes` covering those types — LFS added after the fact leaves the binaries in history.
- On Windows, set `core.autocrlf` to `input` to avoid line-ending churn in text scenes.

## Common Mistakes And Forbidden Patterns

- A `scenes/` + `scripts/` + `assets/` layout split by file type instead of by feature.
- Mixed-case file paths that load in the editor on Windows or macOS and 404 in the exported PCK.
- `.godot/` or `*.translation` committed to the repository.
- First-party code placed in `addons/`, or vendored addon code edited in place without an ADR.
- Git LFS configured after binary assets are already in history.
- Binary `.scn`/`.res` scene and resource formats where text formats would do.
- An addon vendored with no record of its source or version.
- No `LICENSE` file, or one copied in by reflex without deciding whether it fits how the project ships.

## Verification And Proof

```bash
ls project.godot LICENSE .gitignore .gitattributes
git check-ignore .godot
git lfs track
git ls-files | grep -E '\.(scn|res|translation)$'   # expect no output
git ls-files | grep -vE '\.cs$' | grep -E '[A-Z]'   # review any hit against the naming rules
```

Proof is complete when the ignore rules actually match (`git check-ignore` succeeds on `.godot`), LFS patterns cover every binary asset type in the repo, no binary scene/resource or `*.translation` artifacts are tracked, filenames outside C# scripts are `snake_case`, and the directory tree groups by feature rather than by file type.
