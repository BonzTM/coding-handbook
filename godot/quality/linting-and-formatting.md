# Linting And Formatting

The single source of truth for GDScript lint and format policy: what runs, how it is pinned, and where it is enforced.

## Default Approach

**gdtoolkit** (`github.com/Scony/godot-gdscript-toolkit`) is the mandated toolchain for GDScript: an independent parser plus a linter (`gdlint`), a formatter (`gdformat`), and cyclomatic-complexity metrics. `gdformat` keeps whitespace and layout mechanical; `gdlint` catches naming, structure, and design violations statically before review.

This doc owns the tooling policy only. The style rules themselves — tabs, snake_case naming, the prescribed script member order, static typing — are owned by [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md); gdlint is how the mechanically checkable subset of them is enforced. Engine-side enforcement of typing (the `UNTYPED_DECLARATION` warning) is also owned there — gdlint does not type-check, so the two gates complement rather than replace each other. C# scripts in a Godot .NET project follow [../../csharp/quality/linting.md](../../csharp/quality/linting.md), not this doc; the language split itself is governed by [../foundations/gdscript-vs-csharp.md](../foundations/gdscript-vs-csharp.md).

### Toolchain

- Install the Godot 4 line: `pip3 install "gdtoolkit==4.*"`. gdtoolkit's major version tracks the Godot major version — a 3.x gdtoolkit cannot parse Godot 4 GDScript. Pin one exact version (in `requirements.txt` or the pre-commit `rev`) and use that same version locally and in CI; an unpinned install is how local-green/CI-red churn starts.
- `gdformat` is deliberately opinionated: per its docs, "The only configurable thing is max line length allowed (`--line-length`). The rest will be taken care of by `gdformat` in a one, consistent way." Do not fight it; there is no style debate to have below the line-length knob.
- GDQuest's GDScript-formatter (`github.com/GDQuest/GDScript-formatter`) is a faster alternative formatter for Godot 4. Swapping to it is a repo-wide decision recorded via [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md), never a per-developer choice — two formatters in one repo means permanent diff noise.

### Configuration

- Generate the default lint config with `gdlint -d`, rename the output to `.gdlintrc`, and commit it at the repo root. The handbook's starting point is [../templates/gdlintrc.txt](../templates/gdlintrc.txt); copy it rather than regenerating from scratch so repos share one baseline.
- gdlint's rule set mechanizes the official GDScript style guide (`docs.godotengine.org/en/stable/tutorials/scripting/gdscript/gdscript_styleguide.html`): naming (`function-name`, `class-name`, `constant-name`, `enum-name`, `signal-name`), formatting (`max-line-length`, `trailing-whitespace`, `mixed-tabs-and-spaces`), structure (`class-definitions-order`, `no-else-return`), and design limits (`function-arguments-number: 10`, `max-public-methods`). Keep `class-definitions-order` enabled — it enforces the script member order the style doc prescribes.
- Set `max-line-length` in `.gdlintrc` and `gdformat --line-length` to the same value — default to 100, the official style guide's ceiling. If they diverge, freshly formatted code fails lint.
- Scope the tools to first-party code only. `addons/` is third-party vendored code ([../foundations/project-setup.md](../foundations/project-setup.md) owns the layout); pass explicit first-party directories to `gdlint`/`gdformat` and exclude `addons/` (gdformat supports excluded directories via a `gdformatrc` dumped with `--dump-default-config`). Linting code you do not own produces findings you can only suppress.
- Suppress false positives at the narrowest scope: same-line `# gdlint:ignore = rule-name`, or a `# gdlint: disable=rule-name` … `# gdlint: enable=rule-name` pair for a block — always closed with the matching `enable`. If the same rule needs suppressing across files, the rule is misconfigured for this codebase: change `.gdlintrc` (and the template, in the same PR) instead of scattering ignores.

### Editor And Pre-Commit Integration

Format on save, lint on commit, gate in CI — each layer catches what the previous one missed.

- Editor: run `gdformat` on save via your editor's format-on-save hook for `.gd` files. The gdLinter editor addon (`godotengine.org/asset-library/asset/2520`) surfaces `gdlint` findings on save inside the Godot editor.
- Pre-commit: gdtoolkit ships hooks for the `pre-commit` framework. Commit this in `.pre-commit-config.yaml`, formatter before linter so mechanical fixes never surface as lint findings, with `rev` pinned to the same version CI installs:

```yaml
repos:
  - repo: https://github.com/Scony/godot-gdscript-toolkit
    rev: 4.3.2  # keep identical to the CI-installed gdtoolkit version
    hooks:
      - id: gdformat
        require_serial: true
      - id: gdlint
        require_serial: true
```

- gdformat rewrites files in place and its docs warn that "formatting may lead to data loss, so it's highly recommended to use it along with Version Control System (VCS)". Run it on committed or staged work, never as the only copy of an uncommitted change.

### CI Enforcement

CI runs both tools read-only on every push, pinned to the committed version:

```bash
pip3 install "gdtoolkit==<pinned-version>"
gdformat --check <first-party dirs>   # fails if formatting would produce a diff; writes nothing
gdlint <first-party dirs>             # must exit 0
```

`gdformat --check` is the read-only counterpart of the in-place formatter — it is gdtoolkit's own documented CI invocation. Keep the format check and the lint run as separate CI steps so a lint failure always means a real finding, not whitespace. Where these steps live in the pipeline, and the workflow that runs them, are owned by [../operations/ci-and-release.md](../operations/ci-and-release.md); the committed workflow is [../templates/github-workflows-ci.yml](../templates/github-workflows-ci.yml). Lint and format gates run before tests ([testing.md](testing.md)) — they are seconds-cheap and fail fast.

## Common Mistakes And Forbidden Patterns

- Installing gdtoolkit unpinned (`pip install gdtoolkit`) or at a different version than the pre-commit `rev`, so local and CI results diverge.
- Using a 3.x gdtoolkit against a Godot 4 project — the parser does not match the language version.
- Hand-formatting against `gdformat` output, or carrying two formatters in one repo without an ADR.
- Setting `.gdlintrc` `max-line-length` and `gdformat --line-length` to different values, so the formatter emits code the linter rejects.
- Running the tools over `addons/` and then mass-suppressing the third-party findings.
- Open-ended `# gdlint: disable=` blocks with no matching `enable`, or repo-wide rule disables to dodge one finding instead of a scoped `# gdlint:ignore`.
- Treating a green `gdlint` as proof of typed GDScript — typing enforcement is the `UNTYPED_DECLARATION` warning owned by [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md), not a gdlint rule.
- Putting the format check and the lint run in one CI step, so a red gate cannot tell a defect from whitespace.

## Verification And Proof

```bash
gdformat --check <first-party dirs>   # exit 0, no would-be diff
gdlint <first-party dirs>             # exit 0
pre-commit run --all-files            # proves the committed hook config actually runs
```

The policy holds when all three exit 0 at the pinned version, no suppression was added to dodge a real finding, and the CI workflow runs the identical commands.

## Where To Go Next

- Style and typing rules the linter mechanizes: [../foundations/gdscript-style-and-typing.md](../foundations/gdscript-style-and-typing.md)
- The committed lint config baseline: [../templates/gdlintrc.txt](../templates/gdlintrc.txt)
- Where the gates run in the pipeline: [../operations/ci-and-release.md](../operations/ci-and-release.md)
- Dynamic proof beyond static checks: [testing.md](testing.md)
