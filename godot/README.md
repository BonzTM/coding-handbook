# Godot Project Handbook

This handbook is the default engineering contract for new Godot 4.x repositories. It is not a Godot tutorial. It exists to make games and interactive tools converge on the same project layout, scene architecture, scripting discipline, data handling, and proof of correctness. Guidance is pinned to the current 4.x stable line (releases at `godotengine.org/download/archive/`); version-sensitive rules say so explicitly.

## Start Here

- Humans: read this file, then follow the reading path for your project shape.
- Agents: read [AGENTS.md](AGENTS.md) first (it includes the change-routing table), then the relevant topical docs and recipes.
- Default assumptions unless a repo says otherwise:
  - Godot 4.x current stable, standard (non-.NET) editor build, version pinned per repo
  - GDScript with static typing enforced via the `UNTYPED_DECLARATION` warning
  - `snake_case` file and folder names; PascalCase node names (`docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html`)
  - self-contained scenes; children signal up, parents call down
  - `.godot/` and `*.translation` gitignored; Git LFS for binary assets, configured before the first asset commit
  - `gdformat` and `gdlint` (gdtoolkit) as the format and lint gate
  - tests run headless in CI, and exports produced by `godot --headless --export-release`

## Reading Paths

| If you are building... | Read in this order |
|---|---|
| 2D or 3D single-player game | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/scene-and-node-architecture.md](foundations/scene-and-node-architecture.md) -> [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md) -> [foundations/gdscript-style-and-typing.md](foundations/gdscript-style-and-typing.md) -> [foundations/resources-and-data.md](foundations/resources-and-data.md) -> [foundations/input-handling.md](foundations/input-handling.md) -> [quality/testing.md](quality/testing.md) -> [recipes/add-a-scene.md](recipes/add-a-scene.md) |
| Physics-driven action game | the single-player path above, inserting [systems/physics-and-movement.md](systems/physics-and-movement.md) after [foundations/input-handling.md](foundations/input-handling.md), then [operations/performance-and-profiling.md](operations/performance-and-profiling.md) and [systems/audio.md](systems/audio.md) |
| UI-heavy game or tool | [foundations/project-setup.md](foundations/project-setup.md) -> [foundations/scene-and-node-architecture.md](foundations/scene-and-node-architecture.md) -> [systems/ui-and-theming.md](systems/ui-and-theming.md) -> [foundations/input-handling.md](foundations/input-handling.md) -> [systems/localization.md](systems/localization.md) -> [recipes/add-a-ui-screen.md](recipes/add-a-ui-screen.md) |
| Game with saves and progression | the single-player path above, then [systems/save-and-load.md](systems/save-and-load.md) -> [recipes/add-a-save-field.md](recipes/add-a-save-field.md) |
| Multiplayer game | the single-player path above, then [systems/multiplayer.md](systems/multiplayer.md) -> [foundations/resources-and-data.md](foundations/resources-and-data.md) -> [operations/ci-and-release.md](operations/ci-and-release.md) |
| Mixed GDScript and C# project | [foundations/gdscript-vs-csharp.md](foundations/gdscript-vs-csharp.md) first — it decides whether C# is justified and owns the interop rules — then the path matching your game shape |

Every shape also adopts [quality/linting-and-formatting.md](quality/linting-and-formatting.md), the committed [templates/](templates/) scaffolding ([templates/project-settings-conventions.md](templates/project-settings-conventions.md) governs project settings), and [operations/ci-and-release.md](operations/ci-and-release.md) for headless test and export gates. Reference exemplar projects that compose these patterns end to end are a planned later phase; until they land, the recipes in [recipes/README.md](recipes/README.md) are the concrete implementation guidance.

## Non-Negotiables

- Pin the Godot minor version per repo and upgrade deliberately: the official release policy allows "minor compatibility breakage in very specific areas" between minor versions (`docs.godotengine.org/en/stable/about/release_policy.html`).
- `snake_case` for all file and folder names. Exported PCK filesystems are case-sensitive while desktop filesystems usually are not; inconsistent casing breaks exported builds (`docs.godotengine.org/en/stable/tutorials/best_practices/project_organization.html`).
- Static typing everywhere in GDScript, enforced with the `UNTYPED_DECLARATION` warning. Typed GDScript gets optimized opcodes and real autocompletion (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/static_typing.html`).
- Scenes are self-contained with no external dependencies; a parent injects what a child needs. Children signal up, parents call down — signals respond to behavior, they do not start it (`docs.godotengine.org/en/stable/tutorials/best_practices/scene_organization.html`).
- A new autoload requires an ADR. Global state is a centralized failure point and destroys call-site debuggability; prefer self-contained scenes, `static` functions, or shared Resources first (`docs.godotengine.org/en/stable/tutorials/best_practices/autoloads_versus_regular_nodes.html`).
- Never `ResourceLoader.load()` a file a user can modify. `.tres`/`.tscn`/`.res` files can embed scripts that execute on load; user-writable saves use JSON, `ConfigFile`, or `store_var` with validation (see [systems/save-and-load.md](systems/save-and-load.md)).
- Gameplay reads InputMap actions, never raw keycodes; discrete events go through `_unhandled_input()` so GUI gets first refusal (`docs.godotengine.org/en/stable/tutorials/inputs/inputevent.html`).
- Physics-interacting movement lives in `_physics_process()` and scales by `delta`; `_process()` is not synchronized with physics (`docs.godotengine.org/en/stable/tutorials/scripting/idle_and_physics_processing.html`).
- Never commit `.godot/` or `*.translation`; binary assets go through Git LFS from the first commit (`docs.godotengine.org/en/stable/tutorials/best_practices/version_control_systems.html`).
- Never `free()` or `queue_free()` an autoload at runtime — the engine crashes (`docs.godotengine.org/en/stable/tutorials/scripting/singletons_autoload.html`).

## Default Stack

| Concern | Default | Reach for something else when |
|---|---|---|
| Scripting language | typed GDScript | compute-heavy systems or .NET library needs justify C# per [foundations/gdscript-vs-csharp.md](foundations/gdscript-vs-csharp.md); C# rules out web export |
| Cross-node communication | direct calls down, signals up | one-to-many broadcast justifies groups; cross-system decoupling justifies a signal-contract event bus per [foundations/signals-and-decoupling.md](foundations/signals-and-decoupling.md) |
| Designer-editable data | custom `Resource` classes with `class_name` and `@export` | hot-path or schemaless data where Dictionary speed measurably matters |
| Save data | JSON or `ConfigFile` (user settings) with validation on load | binary `store_var` when file size or non-JSON types dominate; never resource files for user-writable data |
| Testing | GUT, run headless in CI | C# tests or IDE test-adapter integration justify gdUnit4 |
| Lint and format | `gdlint` + `gdformat` from gdtoolkit, configured via [templates/gdlintrc.txt](templates/gdlintrc.txt) | almost never; editor-side plugins supplement, not replace, the CI gate |
| UI consistency | one project-wide `Theme` resource plus type variations | branch-level themes for distinct sections; local overrides only for one-off tweaks |
| Multiplayer transport | high-level `MultiplayerAPI` over `ENetMultiplayerPeer` | web targets need `WebSocketMultiplayerPeer`; P2P needs WebRTC |
| CI export | `godot --headless --export-release` in GitHub Actions, per [templates/github-workflows-ci.yml](templates/github-workflows-ci.yml) | the org mandates a different CI system; keep the same headless stages |

## Handbook Map

- [AGENTS.md](AGENTS.md) - fast-path contract and change routing for autonomous agents and reviewers
- [maintainer-reference.md](maintainer-reference.md) - architecture, rationale, and deeper guidance
- `foundations/` - [project setup](foundations/project-setup.md), [asset pipeline](foundations/asset-pipeline.md), [scene and node architecture](foundations/scene-and-node-architecture.md), [signals and decoupling](foundations/signals-and-decoupling.md), [resources and data](foundations/resources-and-data.md), [GDScript style and typing](foundations/gdscript-style-and-typing.md), [GDScript vs C#](foundations/gdscript-vs-csharp.md), and [input handling](foundations/input-handling.md)
- `quality/` - [testing.md](quality/testing.md) and [linting-and-formatting.md](quality/linting-and-formatting.md), the proof commands every repo runs
- `systems/` - gameplay-facing subsystems: [UI and theming](systems/ui-and-theming.md), [audio](systems/audio.md), [animation](systems/animation.md), [game flow](systems/game-flow.md), [physics and movement](systems/physics-and-movement.md), [save and load](systems/save-and-load.md), [localization](systems/localization.md), and [multiplayer](systems/multiplayer.md)
- `operations/` - [ci-and-release.md](operations/ci-and-release.md) (headless test and export pipeline) and [performance-and-profiling.md](operations/performance-and-profiling.md)
- `decisions/` ([README.md](decisions/README.md)) - [architecture decision records](decisions/architecture-decision-records.md); ADRs gate autoloads, C# adoption, and new addons
- `checklists/` ([README.md](checklists/README.md)) and `recipes/` ([README.md](recipes/README.md)) - executable startup and implementation guidance, including [new-project.md](checklists/new-project.md), [pr-review.md](checklists/pr-review.md), [release.md](checklists/release.md), and [handoff.md](checklists/handoff.md)
- `templates/` ([README.md](templates/README.md)) - committed copy-paste scaffolding: [gitignore](templates/gitignore.txt), [gdlintrc](templates/gdlintrc.txt), [CI workflow](templates/github-workflows-ci.yml), [project settings conventions](templates/project-settings-conventions.md), and [PR template](templates/pull_request_template.md)
- `reference/` - complete exemplar projects composing these patterns end to end; planned for a later phase and not yet available
- Team process (human-facing; not read during app builds): [onboarding-and-handoff.md](onboarding-and-handoff.md) for ownership transfer, [glossary.md](glossary.md) as a term lookup, and [CONTRIBUTING.md](CONTRIBUTING.md) for changing the handbook itself

## What This Handbook Optimizes For

- code that still looks obvious six months later
- boundaries that make testing and refactoring cheaper
- runtime behavior that is safe under load and easy to debug
- defaults that keep agents from inventing new architecture every task
- minimal dependency surface unless there is a clear return on complexity

## Where To Go Next

- New repo bootstrap: [checklists/new-project.md](checklists/new-project.md)
- Active agent work: [AGENTS.md](AGENTS.md)
- Routing a change quickly: [AGENTS.md](AGENTS.md) (## Change Routing)
- Deciding GDScript vs C#: [foundations/gdscript-vs-csharp.md](foundations/gdscript-vs-csharp.md)
- Recording an architecture decision: [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md)
- Copy-paste scaffolding for a new repo: [templates/README.md](templates/README.md)
- Taking over or handing off a project: [onboarding-and-handoff.md](onboarding-and-handoff.md)
- Lint and format policy: [quality/linting-and-formatting.md](quality/linting-and-formatting.md)
- Setting up headless CI export: [recipes/set-up-ci-export.md](recipes/set-up-ci-export.md)
- Implementing a specific change step by step: [recipes/README.md](recipes/README.md)
- Looking up a handbook term: [glossary.md](glossary.md)
- Changing the handbook itself: [CONTRIBUTING.md](CONTRIBUTING.md)
