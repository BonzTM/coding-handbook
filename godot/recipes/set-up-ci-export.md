# Recipe: Set Up CI Export

Use this when you stand up the CI pipeline for a Godot repo: install the pinned engine and export templates on the runner, run format, lint, and tests headless, and produce export artifacts on every PR. Pipeline policy is owned by [../operations/ci-and-release.md](../operations/ci-and-release.md); this recipe is the setup path from the committed [CI workflow template](../templates/github-workflows-ci.yml).

## Files To Touch

- `.github/workflows/ci.yml` — copied from [../templates/github-workflows-ci.yml](../templates/github-workflows-ci.yml)
- `export_presets.cfg` — created in the editor Export dialog, committed
- `.gdlintrc` — from the [gdlintrc template](../templates/gdlintrc.txt), if not already present
- `.gitignore` — must already ignore `.godot/`, per the [gitignore template](../templates/gitignore.txt)
- repository CI secrets (keystores, signing passwords) — configured in the forge UI, never as files

## Steps

1. Copy the workflow template to `.github/workflows/ci.yml` and set the engine version variable to the pinned current 4.x stable release (releases at `godotengine.org/download/archive/`). That one variable is the only place the version appears — editor install, export templates, and tooling all derive from it per [../operations/ci-and-release.md](../operations/ci-and-release.md).
2. Install the engine and matching export templates with an established building block — `github.com/chickensoft-games/setup-godot` (handles editor plus export templates, including .NET builds) or the `godot-ci` Docker image (`github.com/marketplace/actions/godot-ci`) — not ad hoc download steps in shell.
3. Wire the cheap stages first: `pip3 install "gdtoolkit==4.*"` (major pinned to the engine major), then `gdformat --check <first-party dirs>` and `gdlint <first-party dirs>` against the committed `.gdlintrc` — scoped to first-party directories, never vendored `addons/`, per [../quality/linting-and-formatting.md](../quality/linting-and-formatting.md).
4. Wire the test stage: the project's framework CLI runner (GUT or gdUnit4, chosen in [../quality/testing.md](../quality/testing.md)) under `--headless`. The flag is not optional — the official CLI docs state it "is required on platforms that do not have GPU access (such as continuous integration)" (`docs.godotengine.org/en/stable/tutorials/editor/command_line_tutorial.html`).
5. Create one export preset per shipping platform in the editor Export dialog and commit `export_presets.cfg`; the docs confirm it "can be safely committed to version control" (`docs.godotengine.org/en/stable/tutorials/export/exporting_projects.html`). Sensitive options land in `.godot/export_credentials.cfg`, which the gitignore already excludes — inject those values from CI secrets at job time.
6. In each export job, run `godot --headless --import` before exporting: `.godot/` is gitignored, so a fresh runner has no import cache and the export would otherwise fail on missing imported resources.
7. Export by exact preset name: `godot --headless --export-release "Preset Name" build/<artifact>`. The name must match `export_presets.cfg` character for character; add one matrix entry per platform preset.
8. After each export, assert the output file exists and is non-trivially sized (`test -s`) before uploading it as a CI artifact. A missing template or misconfigured preset can log errors without failing the process.
9. Trigger the full pipeline — including export jobs — on every PR, not only on tags, and push a PR to watch it go green on a fresh runner.

## Invariants To Preserve

- exactly one engine version pin; editor, export templates, and gdtoolkit major all derive from it
- `.godot/` (and with it `export_credentials.cfg`) stays gitignored; no keystore, password, or key reaches the repo or a CI log
- every engine invocation in CI carries `--headless`, and every export job imports before it exports
- every preset name referenced in the workflow exists in the committed `export_presets.cfg`
- export jobs run on every PR; players only ever receive `--export-release` artifacts
- the workflow file stays in sync with the [template](../templates/github-workflows-ci.yml) — divergence goes through the template and [../operations/ci-and-release.md](../operations/ci-and-release.md) together

## Proof

- the pipeline is green on a fresh runner: format, lint, tests, import, and every per-platform export job
- each export job uploaded an artifact, and the size assertion would fail if it had not been produced
- `git check-ignore .godot/export_credentials.cfg` succeeds, and no credential file is tracked
- preset names in `.github/workflows/ci.yml` all grep-match entries in `export_presets.cfg`
- a deliberately broken lint or test on a scratch branch fails the corresponding stage, proving the gates bite
- one exported artifact launches on its target platform

The ongoing gate order, release tagging, and per-platform concerns are owned by [../operations/ci-and-release.md](../operations/ci-and-release.md) — this recipe only stands the pipeline up.
