# Release Checklist

Release gate for traceable, reproducible Python distributions and service images.

## Source And CI

- [ ] Release commit is protected and `make verify` passed with committed `pyproject.toml` and `uv.lock`.
- [ ] Docker integration job ran `uv run pytest -m integration` against real required dependencies with no unexpected skip.
- [ ] `uv run --with pip-audit pip-audit` is clean or approved time-bounded exceptions are linked.
- [ ] Migrations, compatibility, config, ports, timeouts, contracts, replay/DLQ, and operator steps are documented.
- [ ] `CHANGELOG.md` moved `Unreleased` content into `[X.Y.Z] - YYYY-MM-DD` and opened a fresh `Unreleased`.

## Artifact Quality

- [ ] Static `[project].version`, `v<version>` tag, changelog, wheel/sdist metadata, image labels, and startup version agree.
- [ ] Library build uses `uv build --no-sources`; wheel is built through the sdist and both contents were reviewed.
- [ ] Wheel installs in an isolated environment; import, `[project.scripts]`, `py.typed`, licenses, and package data are correct.
- [ ] Service image is multi-stage, digest-pinned, non-root, contains only runtime needs, and carries immutable source/revision/version labels.
- [ ] Artifact/image scans contain no source secrets, `.env`, caches, tests, local paths, credentials, or unexpected native/runtime packages.
- [ ] Registry publication uses protected least-privilege identity; PyPI libraries use Trusted Publishing by default.

## Deploy Safety

- [ ] `/livez`, `/readyz`, `/metrics`, startup, and SIGTERM drain were smoke-tested on the exact artifact.
- [ ] Alembic migration runs as one explicit pre-deploy step and has expand/contract plus rollback/forward-recovery evidence.
- [ ] API/event/library changes have compatibility, deprecation, replay, and consumer notification plans.
- [ ] Prior known-good package/image remains addressable by immutable version/digest and rollback commands are current.
- [ ] Release notes identify every operator-visible change and link the rollout/runbook steps.

## Verification

```bash
make verify
uv run pytest -m integration
uv build --no-sources
uv run --with ./dist/<wheel> --no-project -- python -c "import <package>"
```

- [ ] Published registry artifact was fetched by version/digest and independently installed or started.
- [ ] Recorded hashes/digests, provenance, release notes, and tag point to the same reviewed commit.
