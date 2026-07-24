# CI And Release

Delivery guidance for Godot repos: the same pinned editor version, headless commands, and committed presets must produce the same artifact locally and in CI.

## Default Approach

Every PR runs a predictable baseline pipeline, entirely headless.

| Stage | Commands | Purpose |
|---|---|---|
| formatting | `gdformat --check <first-party dirs>` | consistent source shape; scope and rules owned by [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md) — never vendored `addons/` |
| lint | `gdlint <first-party dirs>` | static GDScript checks against the committed [gdlintrc](../templates/gdlintrc.txt), same first-party scope |
| tests | GUT CLI or gdUnit4 runner under `--headless` | functional confidence; framework choice owned by [../quality/testing.md](../quality/testing.md) |
| import | `godot --headless --import` | rebuild the gitignored `.godot/` import cache on a fresh runner |
| export | `godot --headless --export-release "<Preset>" <output>` | artifact proof, one job per platform preset |

Pin exactly one engine version — a current 4.x stable release (release archive at `godotengine.org/download/archive/`) — as a single variable in the committed [CI workflow](../templates/github-workflows-ci.yml), and derive the editor install, the export templates, and any local tooling from that one pin. Setting up a new repo's pipeline is [../recipes/set-up-ci-export.md](../recipes/set-up-ci-export.md).

`--headless` is not optional on CI runners: the official CLI docs state it "is required on platforms that do not have GPU access (such as continuous integration)" (`docs.godotengine.org/en/stable/tutorials/editor/command_line_tutorial.html`).

## Headless Export Pipeline

1. **Import first.** `.godot/` is never committed ([../foundations/project-setup.md](../foundations/project-setup.md), [gitignore template](../templates/gitignore.txt)), so a fresh runner has no import cache. Run `godot --headless --import` before exporting — the flag "starts the editor, waits for any resources to be imported, and then quits" (`docs.godotengine.org/en/stable/tutorials/editor/command_line_tutorial.html`).
2. **Export by preset name.** `godot --headless --export-release "Preset Name" build/<artifact>`; use `--export-debug` for QA builds and `--export-pack` to produce only a PCK/ZIP. The preset name "must match the name of an export preset defined in the project's `export_presets.cfg` file" (`docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`).
3. **Fail loudly.** After export, assert the output file exists and is non-trivially sized before uploading it. Do not trust job status alone: a misconfigured preset or missing template can leave errors in the log without a usable artifact.
4. **Install engine and templates reproducibly.** Do not download editors ad hoc in shell steps. Established building blocks: `github.com/chickensoft-games/setup-godot` (installs editor and matching export templates, including .NET builds) and the `godot-ci` Docker image (`github.com/marketplace/actions/godot-ci`). Whichever you choose, the version comes from the single pin above.

## Export Presets And Templates

- **Commit `export_presets.cfg`.** The official export docs say it "contains the vast majority of the export configuration and can be safely committed to version control" (`docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`). CI export is impossible without it.
- **Never commit credentials.** Sensitive export options (keystores, passwords, encryption keys) live in `.godot/export_credentials.cfg`, which stays out of the repo because `.godot/` is gitignored. Inject signing material through CI secrets at job time; a secret that reaches the repo or the log is an incident.
- **Templates must match the pinned engine version.** Export templates are versioned artifacts installed per release (Editor > Manage Export Templates, or by the setup action in CI); a template/editor mismatch is the first thing to check when an export fails only in CI.
- **Choose the resource export mode deliberately.** Presets export all resources, selected scenes/resources, an exclude list, or **dedicated server** mode, which strips visual assets for headless server builds. Dot-prefixed files and folders are never exported (`docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`).
- **PCK by default, ZIP when mods matter.** PCK packing is fast and opaque; ZIP is compressed and OS-readable, which suits moddable projects.

## CI Workflow Stages

Order stages cheapest-first so failures are fast: format and lint (seconds, no engine), then headless tests, then import + export as a per-platform matrix. The committed [CI workflow template](../templates/github-workflows-ci.yml) encodes this order; change the template and this doc together, not the workflow file in one repo by hand.

- Lint tooling is pinned (`pip3 install "gdtoolkit==4.*"`, matching the engine major) per [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md).
- Test jobs run the framework's CLI runner under `--headless` on every push per [../quality/testing.md](../quality/testing.md); tests are not an export-day activity.
- Export jobs run on every PR, not only on tags. An export that only happens at release time is an export that breaks at release time.

## Versioning And Release Artifacts

- Tag releases `v<MAJOR>.<MINOR>.<PATCH>`; pushing the tag triggers the release export jobs for every shipping preset.
- Name artifacts `<game>-<version>-<platform>.<ext>` and publish checksums alongside them, so an operator or storefront upload can be traced back to an exact commit.
- Ship only `--export-release` artifacts. `--export-debug` builds are for internal QA and crash triage, never for players.
- Treat engine upgrades as dependency updates with their own PR through the full pipeline. Godot's release policy states minor versions may include "minor compatibility breakage in very specific areas" and that stable branches are supported "at least until the next stable branch is released and has received its first patch update" (`docs.godotengine.org/en/stable/about/release_policy.html`) — so schedule upgrades; do not let the pin rot until support lapses.

## Per-Platform Concerns

- **Web**: "projects written in C# cannot be exported to the web platform" (`docs.godotengine.org/en/stable/tutorials/scripting/c_sharp/index.html`). If web is a target, that constraint is decided at language-selection time — see [../foundations/gdscript-vs-csharp.md](../foundations/gdscript-vs-csharp.md).
- **Mobile**: Android and iOS need per-platform binaries and different texture compression (ETC1/ETC2 vs desktop S3TC), so mobile presets are separate matrix entries with their own import output, plus platform signing material injected as CI secrets.
- **Text rendering**: ICU data is required for emoji and CJK/Thai/Khmer/Lao/Burmese text (`docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`); if the game localizes into those scripts ([../systems/localization.md](../systems/localization.md)), verify the exported build renders them, not just the editor.
- **Dedicated servers**: use the dedicated-server resource export mode for headless multiplayer builds ([../systems/multiplayer.md](../systems/multiplayer.md)) instead of shipping the full client asset set to a server host.

## Common Mistakes And Forbidden Patterns

- exporting on a fresh runner without a prior `--headless --import` run, then debugging "missing resource" errors that are really a cold import cache
- committing `.godot/export_credentials.cfg`, a keystore, or any signing secret — or echoing one into a CI log
- editor, export-template, and local-tool versions drifting apart instead of deriving from one pinned version
- producing release artifacts by hand-clicking Export in the editor, so the shipped build drifts from what CI proves
- shipping `--export-debug` artifacts to players
- treating a green export job as proof without asserting the artifact exists and has plausible size
- export jobs that run only on tags, so export breakage is discovered on release day
- a web export preset in a C# project
- renaming a preset in the editor without updating the workflow, so CI exports a stale or missing preset name

## Verification And Proof

- CI is green on the release commit, and every shipping preset's export job uploaded an artifact
- each exported artifact launches on its target platform and shows the expected version string
- `git check-ignore .godot/export_credentials.cfg` succeeds, and no keystore or credential file is tracked
- preset names referenced in the workflow all appear in the committed `export_presets.cfg`
- the engine version pin, installed export templates, and local editor versions agree
- release tags are `v`-prefixed and each published artifact name embeds the same version
