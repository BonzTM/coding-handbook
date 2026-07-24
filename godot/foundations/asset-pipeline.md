# Asset Pipeline

How imported assets are governed: import settings treated as reviewed code, texture compression chosen by render target, the WAV/Ogg audio split, glTF plus inherited scenes for 3D, project-wide import defaults set on day one, and a cold import cache rebuilt in CI — never committed.

## Default Approach

Every asset dropped into the project gets a `.import` sidecar file next to it; the editor compiles the source into the gitignored `.godot/imported/` cache (`docs.godotengine.org/en/stable/tutorials/assets_pipeline/import_process.html`). The source asset and its `.import` file are committed and reviewed together; the compiled cache never is (see [project-setup.md](project-setup.md) and the [gitignore template](../templates/gitignore.txt)). An import-parameter change is a behavior change to the shipped asset — review it like code, because on export it is the asset.

### Import Settings Ownership

- **`.import` files are committed, always.** The official import-process doc is explicit: commit them, "as these files contain important metadata" — while `.godot/` stays ignored (`docs.godotengine.org/en/stable/tutorials/assets_pipeline/import_process.html`).
- **The asset's feature owner owns its import settings.** Assets live beside the scenes that use them per [project-setup.md](project-setup.md); whoever changes an import parameter commits the modified `.import` in the same change, or teammates and CI import the asset differently than the author saw it.
- **Change settings through the Import dock, then Reimport** — selecting another file first discards the pending changes (`docs.godotengine.org/en/stable/tutorials/assets_pipeline/import_process.html`). Batch changes by multi-selecting files; the checkboxes apply only the checked parameters across the selection.
- Directories that must never generate `.import` metadata (tooling, source art workfiles) get a `.gdignore`, per [project-setup.md](project-setup.md).

### Texture Compression By Target

Pick the compress mode by where the texture renders, not by habit (`docs.godotengine.org/en/stable/tutorials/assets_pipeline/importing_images.html`):

| Mode | Use for | Cost |
|---|---|---|
| Lossless | 2D default; mandatory for pixel art | High VRAM at large sizes (a 4096×4096 texture: ~85 MiB vs ~21 MiB VRAM-compressed) |
| Lossy | Large 2D assets where disk size matters | Compression artifacts; VRAM as high as Lossless |
| VRAM Compressed | 3D default — the docs call it "a must-have for 3D games with high-resolution textures" | "Should be avoided for 2D as it exhibits noticeable artifacts"; even in 3D, pixel-art textures disable it |
| VRAM Uncompressed | Raw floating-point formats only | Maximum memory |
| Basis Universal | Smaller files at VRAM-Compressed memory cost | Lower quality, slower import, no float formats |

Mipmaps: on for 3D (prevents distance graininess), off in 2D unless visibly beneficial (same source). Mobile targets need different compressed formats than desktop (ETC1/ETC2 vs S3TC), which is why mobile export presets are separate CI matrix entries with their own import output — owned by [../operations/ci-and-release.md](../operations/ci-and-release.md).

### Audio Import

The official recommendation is the rule: "Consider using WAV for short and repetitive sound effects, and Ogg Vorbis for music, speech, and long sound effects. MP3 is useful for mobile and web projects where CPU resources are limited" (`docs.godotengine.org/en/stable/tutorials/assets_pipeline/importing_audio_samples.html`).

- **WAV for SFX**: cheap to decode — "hundreds of simultaneous voices in this format are fine" — at the price of disk size; IMA ADPCM or Quite OK Audio compression recovers most of that.
- **Ogg Vorbis for music and long clips**: smallest files, highest decode cost; one or two streaming music voices, not dozens of gunshots.
- **Loop points**: WAV supports forward, ping-pong, and backward looping with explicit loop points; Ogg Vorbis and MP3 only loop forward from the start. A seamlessly looping ambience with a lead-in must be WAV or restructured.
- Bus routing, mixing, and player lifetime are owned by [../systems/audio.md](../systems/audio.md); this doc owns only the format decision at import.

### Mesh And Scene Import

- **glTF 2.0 (`.glb`) is the interchange default** — the format the official docs recommend, with the most complete import pipeline (`docs.godotengine.org/en/stable/tutorials/assets_pipeline/importing_3d_scenes/available_formats.html`). FBX imports via ufbx; OBJ and Collada are supported but limited (no skeletons/PBR for OBJ). Committing `.blend` files directly means "*all* team members" — and every CI runner — need Blender installed (same source); that is a team-tooling decision recorded in an ADR ([../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)), not a default.
- **Never edit an imported scene in place — edits are lost on reimport.** Customize through a **New Inherited Scene**, where "nodes from the base scene can't be removed, but additional nodes can be added anywhere", so the artist can re-export the source and the game-side additions survive (`docs.godotengine.org/en/stable/tutorials/assets_pipeline/importing_3d_scenes/import_configuration.html`).
- Per-mesh and per-material overrides go through the Advanced Import Settings dialog; extract materials and textures to external files so each gets its own `.import` settings and reviewable diff (same source).
- Attach gameplay scripts, collision, and signals to the inherited scene, keeping the imported artifact a pure art asset the DCC tool can regenerate.

### Import Presets And Project-Wide Defaults

Per-type baselines are set once, before assets accumulate: the **Preset...** menu in the Import dock saves and applies named option sets, and the **Import Defaults** tab in Project Settings changes the project-wide baseline per resource type (`docs.godotengine.org/en/stable/tutorials/assets_pipeline/import_process.html`). A pixel-art project sets its texture defaults (Lossless, filtering off) on day one in [../checklists/new-project.md](../checklists/new-project.md) territory; retrofitting a default later means a mass-reimport diff across every `.import` file. Importer defaults are stored in `project.godot`, so they are reviewed like any settings change per [../templates/project-settings-conventions.md](../templates/project-settings-conventions.md). Land a default change plus its mechanical `.import` fallout as one dedicated commit, separate from feature work.

### Reimport In CI

Because `.godot/` is never committed, a fresh clone or CI runner has a cold import cache; the import stage (`godot --headless --import`) rebuilds it before tests and exports run. The stage ordering, the exact command, and the "missing resource means cold cache" failure mode are owned by [../operations/ci-and-release.md](../operations/ci-and-release.md) — do not restate the pipeline here. The local corollary: a clean clone must import headless with no errors, or the repo is broken for everyone who is not the author's machine.

## Common Mistakes And Forbidden Patterns

- `.import` files missing from a commit that adds or retunes an asset — the author's import settings silently diverge from everyone else's.
- VRAM Compressed on 2D or pixel-art textures (visible artifacts), or Lossless on high-resolution 3D textures (VRAM blowout) — inverting the table above.
- Music or long ambience shipped as WAV (disk bloat for nothing), or dozens of simultaneous SFX as Ogg Vorbis (decode cost the profiler will find later).
- Editing an imported 3D scene directly instead of through an inherited scene, then losing the work on the next artist export.
- Committing `.blend` sources as the shipped format without the ADR making Blender a team-and-CI-wide dependency.
- Retrofitting import defaults months in and burying the mass `.import` diff inside a feature PR.
- Hand-editing a `.import` file without reimporting, leaving the sidecar claiming settings the cached artifact does not have.
- Adding binary assets before Git LFS covers their types — owned by [project-setup.md](project-setup.md), broken most often by asset-pipeline work.

## Verification And Proof

```bash
git check-ignore .godot                                  # cache ignored
git ls-files --others --exclude-standard '*.import'      # expect empty: no untracked sidecars
git status --porcelain '*.import'                        # expect empty after a reimport: no drift
godot --headless --import                                # clean clone imports without errors
```

The pipeline is healthy when a fresh clone imports headless with zero errors, every imported asset has a committed `.import` sidecar and no sidecar is dirty after reimport, texture and audio choices match the tables above (spot-check `compress/mode` in `.import` diffs at review), and every customized 3D asset is an inherited scene rather than an edited import.

## Where To Go Next

- [../operations/ci-and-release.md](../operations/ci-and-release.md) — the headless import and export stages this doc's CI behavior lives in.
- [project-setup.md](project-setup.md) — VCS hygiene, Git LFS, `.gdignore`, and where assets live.
- [../systems/audio.md](../systems/audio.md) — what happens to a sound after import: buses, players, mixing.
- [../operations/performance-and-profiling.md](../operations/performance-and-profiling.md) — evidence before changing compression or pooling decode-heavy audio.
- [../templates/project-settings-conventions.md](../templates/project-settings-conventions.md) — the committed home of importer defaults.
