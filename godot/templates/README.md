# Templates

Committed, copy-paste-ready starting artifacts for a new Godot repository, so every handbook-following repo converges on the same scaffolding instead of re-deriving it per project. Each template is a fill-in skeleton with explicit `<PLACEHOLDER>`s, not finished prose.

This tree is the artifact home that the prose docs keep implying. When [foundations/project-setup.md](../foundations/project-setup.md) says `.godot/` is never tracked or [operations/ci-and-release.md](../operations/ci-and-release.md) says CI runs the same gate as a local checkout, the concrete file lives here. For routing a change rather than bootstrapping a repo, start at the Change Routing table in [../AGENTS.md](../AGENTS.md) (## Change Routing). Verify-green exemplar projects under `reference/` are a planned later phase; until they land, these templates plus [../recipes/README.md](../recipes/README.md) are the canonical scaffolding.

## How To Use

1. Create the repo skeleton per [checklists/new-project.md](../checklists/new-project.md), letting the Godot editor generate `project.godot` — never hand-write it.
2. Copy each template to the destination path in the table below, decoding the filename per the conventions in the next section.
3. Replace every `<PLACEHOLDER>` token — the pinned Godot 4.x engine version, export preset names, and script directory paths.
4. Run the baseline gate from [../AGENTS.md](../AGENTS.md) (## Baseline Verification) — `gdformat --check`, `gdlint`, headless tests, headless export — before the first commit.

## Filename And Destination Conventions

A template's filename encodes its destination in a fresh repo. The tree stays flat and greppable:

- Dotfile targets drop the leading dot and carry a trailing `.txt` so this docs repo does not apply them to itself: `gitignore.txt` -> `.gitignore`, `gdlintrc.txt` -> `.gdlintrc`. Drop the `.txt` and restore the dot when you copy.
- Slashes in a destination become `-` in the filename: `github-workflows-ci.yml` -> `.github/workflows/ci.yml`.
- `pull_request_template.md` keeps the literal name GitHub requires: `.github/pull_request_template.md`.
- `project-settings-conventions.md` is the one non-copy template: `project.godot` and `export_presets.cfg` are editor-managed, so this doc lists the settings to apply through the Project Settings dialog and the values a reviewer checks in the diff — it is consulted, not copied.

## Template Index

| Template | Destination in a new repo | Governing handbook doc |
|---|---|---|
| [gitignore.txt](gitignore.txt) | `.gitignore` | [foundations/project-setup.md](../foundations/project-setup.md) |
| [gdlintrc.txt](gdlintrc.txt) | `.gdlintrc` | [quality/linting-and-formatting.md](../quality/linting-and-formatting.md) |
| [github-workflows-ci.yml](github-workflows-ci.yml) | `.github/workflows/ci.yml` | [operations/ci-and-release.md](../operations/ci-and-release.md) |
| [project-settings-conventions.md](project-settings-conventions.md) | applied to `project.godot` via the editor, not copied | [foundations/project-setup.md](../foundations/project-setup.md) |
| [pull_request_template.md](pull_request_template.md) | `.github/pull_request_template.md` | [../AGENTS.md](../AGENTS.md) (## Working Norms) |

The CI workflow is the single verification entrypoint mirrored in both directions: it runs the same format, lint, test, and headless-export stages as the local baseline gate, so local and CI can never disagree about green. Headless CLI export is the official mechanism (`docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`).

## Governing Docs

Every template is downstream of exactly one owning doc; when a rule changes there, the template changes in the same PR (the sync surface in [../CONTRIBUTING.md](../CONTRIBUTING.md)):

- `gitignore.txt` encodes the official version-control guidance — `.godot/` and `*.translation` are never tracked, and Git LFS covers binary assets before their first commit (`docs.godotengine.org/en/stable/tutorials/best_practices/version_control_systems.html`). Policy owner: [foundations/project-setup.md](../foundations/project-setup.md).
- `gdlintrc.txt` is the committed configuration for `gdlint` from gdtoolkit (`github.com/Scony/godot-gdscript-toolkit`); the rule choices and the typed-GDScript warning policy live in [quality/linting-and-formatting.md](../quality/linting-and-formatting.md) and [foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md).
- `github-workflows-ci.yml` pins the engine version and export-templates version together — they must match or headless export fails. Owner: [operations/ci-and-release.md](../operations/ci-and-release.md); first-time setup walkthrough in [recipes/set-up-ci-export.md](../recipes/set-up-ci-export.md).
- `project-settings-conventions.md` records the non-default `project.godot` values the handbook mandates (GDScript warning levels, physics tick, internationalization fallback). Owner: [foundations/project-setup.md](../foundations/project-setup.md).
- `pull_request_template.md` encodes the proof expectations from [../AGENTS.md](../AGENTS.md) (## Proof Hints) so every PR states which gate stages ran.

## Where To Go Next

- Bootstrapping a repo: [checklists/new-project.md](../checklists/new-project.md)
- The layout these templates land in: [foundations/project-setup.md](../foundations/project-setup.md)
- Wiring the copied workflow to a first green run: [recipes/set-up-ci-export.md](../recipes/set-up-ci-export.md)
