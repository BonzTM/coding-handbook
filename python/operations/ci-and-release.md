# CI and Release

Delivery rules that keep local, CI, package, image, and release proof on one reproducible path.

## Default Approach

CI runs exactly `make verify`; workflow YAML orchestrates the Make targets and never reimplements them. The ordered gate is lock-check, frozen sync, format-check, lint, imports, types, test, and audit.

| Stage | Command | Purpose |
|---|---|---|
| lock | `uv lock --check` | prove `uv.lock` matches metadata without rewriting |
| environment | `uv sync --frozen` | install exactly the committed graph |
| format | `uv run ruff format --check .` | prove canonical source shape |
| lint/security | `uv run ruff check .` | enforce curated correctness/security rules |
| architecture | `uv run lint-imports` | enforce inward dependency contracts |
| types | `uv run mypy .` | prove strict typed boundaries |
| tests | `uv run pytest` | run the configured ordinary suite |
| vulnerabilities | `uv run --with pip-audit pip-audit` | scan installed/locked dependency posture |

The committed [CI workflow](../templates/github-workflows-ci.yml) invokes `make verify` on every pull request and protected-branch push. A stage remains individually runnable through `make lock-check`, `make sync`, `make format-check`, `make lint`, `make imports`, `make types`, `make test`, and `make audit`.

### CI Jobs And Caching

The baseline job uses the committed `.python-version` current stable pin. Cache uv's download/build cache keyed by OS, architecture, Python pin, `pyproject.toml`, and `uv.lock`; never cache `.venv` across incompatible jobs. A cache miss changes performance only, never resolution. Restore before `uv sync --frozen` and let uv verify/install the environment.

Run a separate Docker-enabled integration job with real PostgreSQL and any required broker. It executes `uv sync --frozen` and `uv run pytest -m integration`; unexpected skips fail the lane. Build/package and container smoke jobs run when those artifacts exist. Secrets are unavailable to untrusted fork code and release permissions remain separate from PR verification.

### Python Version Matrix

Applications develop and run CI on the `.python-version` current stable pin while preserving `requires-python = ">=3.11"`. Compatibility-sensitive applications and all published libraries add a floor job on Python 3.11; libraries test every supported minor or at least floor plus current stable according to their support promise. No code uses a newer feature merely because the primary pin accepts it.

Raise the floor through an ADR, metadata/template/CI changes, changelog, and consumer compatibility proof. Exact interpreter and action pins live in templates.

### Changelog Policy

Maintain `CHANGELOG.md` in [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) shape: newest release first, `Unreleased` at the top, and only `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, and `Security` categories. Start from [the changelog template](../templates/changelog.md).

Operator-visible changes—configuration, secrets, ports, probes, limits/timeouts, migrations, contracts, message replay/DLQ behavior, dependency requirements, and deployment steps—receive an entry in the changing PR. Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) subjects for searchable intent and release automation: `feat`, `fix`, and explicit `BREAKING CHANGE`/`!`. Commit categories do not replace human curation of operator impact.

At release, move `Unreleased` content to `[X.Y.Z] - YYYY-MM-DD`, open a fresh `Unreleased`, and make the version, tag, artifact metadata, and changelog agree.

### Version Source

Published distributions use the static PEP 621 `[project].version` as the one source of package version. Update it with `uv version <version>`, review the `pyproject.toml`/`uv.lock` diff, and create the matching `v<version>` VCS tag. uv documents version updates in its [package publishing guide](https://docs.astral.sh/uv/guides/package/).

This default avoids a runtime/build plugin and allows metadata inspection from a source archive. VCS-derived dynamic versioning (`hatch-vcs` or `uv-dynamic-versioning`) is an ADR-level escalation for a release process that can prove full tag history is always available, dirty/unshallow builds fail safely, and local/wheel/sdist metadata agree. Services may derive display/image identity from the immutable release tag, but a built Python distribution still reports its PEP 621 version.

### Build And Publish

Run `uv build --no-sources` for libraries; uv builds an sdist and then a wheel from it, and recommends `--no-sources` before publishing so local source overrides cannot hide a broken distribution. Inspect contents and metadata, install the wheel into an isolated environment, import the package, run the console script, and run applicable tests against the installed artifact.

Publish libraries with `uv publish` from the protected [release workflow](../templates/github-workflows-release.yml). PyPI Trusted Publishing is the default: its [OIDC flow](https://docs.pypi.org/trusted-publishers/) provides short-lived credentials instead of a stored long-lived API token. Protect the exact release workflow/environment, use least permissions and approvals, and never publish from a PR job. Services publish a digest-addressed container image; they do not publish to PyPI unless they also own a public distribution contract.

### Release Flow

1. Resolve the [release checklist](../checklists/release.md); changelog, version, migrations, compatibility, and rollback are ready.
2. Run `make verify`, the Docker integration lane, `uv build --no-sources` for a library, and artifact/container smoke tests on the release commit.
3. Create the signed/protected `v<version>` tag matching `[project].version` and changelog.
4. The release workflow rebuilds or promotes the reviewed artifact, publishes once, records hashes/digests and provenance, and creates release notes.
5. Verify the registry artifact by digest/version and install or start it before rollout.

## Common Mistakes And Forbidden Patterns

- CI duplicating commands instead of invoking `make verify`, or local and CI stages drifting.
- `uv sync` without frozen/lock proof, unreviewed lockfile rewrite, or cached `.venv` masking resolution.
- Integration markers collected but never run with Docker; unexpected skips treated as green.
- Testing only current stable while claiming a 3.11 library floor.
- Version duplicated across modules/files, or package metadata, tag, changelog, image, and startup output disagreeing.
- Dynamic VCS version plugin added without an ADR and source-archive/shallow-clone proof.
- Operator-visible change absent from `CHANGELOG.md`, or invented changelog categories.
- Wheel built directly from a checkout without proving the sdist; artifact not installed and smoke-tested.
- Long-lived PyPI token stored when Trusted Publishing is available, or a release workflow writable/runnable by untrusted code.
- Service release identified only by mutable tag instead of image digest and commit.

## Verification And Proof

```bash
make verify
uv run pytest -m integration
uv build --no-sources
uv run --with ./dist/<wheel> --no-project -- python -c "import <package>"
```

The protected release commit has green baseline, integration, build, and smoke jobs. `pyproject.toml`, `uv.lock`, changelog heading, `v<version>` tag, wheel/sdist metadata, container labels, and startup version agree. Inspect wheel/sdist contents for required `py.typed`, licenses, and absence of tests/secrets/caches. Verify the published package or image from the registry, record hashes/digest, and prove the rollback artifact remains available.

Related: [project setup](../foundations/project-setup.md), [testing](../quality/testing.md), [deployment](deployment.md), and [dependency upgrade](../checklists/dependency-upgrade.md).
