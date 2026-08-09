# Dependency Upgrade Checklist

Recurring maintenance gate for Python, direct/transitive packages, tooling, and automation. For one bump, follow [the dependency recipe](../recipes/bump-dependency.md); automation posture lives in [CI and release](../operations/ci-and-release.md).

## Cadence & Scope

- [ ] Review Dependabot/Renovate on the owned cadence; the committed [dependabot template](../templates/dependabot.yml) does not accumulate an ignored queue.
- [ ] Group compatible patch/minor updates only when their lock diff and proof remain reviewable.
- [ ] Split each major, Python floor/current-pin change, build backend change, or lint/type policy change into its own PR.
- [ ] Security updates follow exposure/urgency, not the ordinary batch schedule.

## Per Update

- [ ] Read official changelog/release/security notes; record breaking behavior, deprecations, Python floor, typing, migration, and operational changes.
- [ ] Update the direct requirement with `uv add <package>@<constraint>` or `uv add --dev <package>@<constraint>`; do not hand-edit the resolved graph.
- [ ] Run `uv lock`, then inspect `git diff -- pyproject.toml uv.lock`; explain every direct change and material transitive/build change.
- [ ] Reject unrelated package, source/index, Python requirement, or resolution changes and investigate them separately.
- [ ] Update callers, config, adapters, docs, and tests; real DB/HTTP/broker boundaries prove changed semantics.

## Runtime And Tools

- [ ] `.python-version` remains a current stable pin and code still runs at `requires-python = ">=3.11"`; a floor change has ADR and compatibility notes.
- [ ] Ruff, mypy, pytest/pytest-asyncio, Import Linter, pip-audit invocation, uv/build backend, and CI actions/images are reviewed on cadence.
- [ ] New Ruff/mypy diagnostics are fixed or narrowly justified; rules/gates are not disabled to land an upgrade.
- [ ] Native wheels/build requirements are checked for all supported OS/architectures and do not silently add compilers/system libraries to runtime.

## Safety

- [ ] `make verify` passes from a frozen sync; bot PRs have no fast lane.
- [ ] `uv run pytest -m integration` and artifact/container smoke tests cover the dependency's affected boundary.
- [ ] `pip-audit` is clean; security exceptions are explicit, owned, expiring, and reviewed against reachable code.
- [ ] No unreviewed index/source override, editable/path dependency, VCS branch, or credential entered `pyproject.toml`/`uv.lock`.
- [ ] Major/runtime changes include rollout, rollback, compatibility, changelog, and observability evidence.

## Verification

```bash
uv lock --check
uv sync --frozen
uv run --with pip-audit pip-audit
uv run pytest -m integration
make verify
git diff -- pyproject.toml uv.lock
```

- [ ] The dependency queue has named owners and no stale untriaged security update.
- [ ] The lock diff, upstream notes, verification output, and rollback constraint are attached to the PR.
