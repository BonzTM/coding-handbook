# Recipe: Bump A Dependency

Use this to update a direct dependency or review an automated uv update.

## Files To Touch

- `pyproject.toml` and `uv.lock`
- consuming adapter and tests when API/behavior changes
- ADR/changelog for major, floor, operational, or security-visible changes

## Steps

1. Classify routine versus security-driven and patch/minor versus major. Re-answer framework approval questions for a major or new transitive capability.
2. Read official release notes, Python-floor changes, advisories, typing changes, and license/maintenance status.
3. Update deliberately: `uv lock --upgrade-package <distribution>` or edit the direct constraint then run the same command.
4. Inspect `git diff -- pyproject.toml uv.lock`; account for every direct/transitive movement and source/build change.
5. Update callers/tests, run the targeted surface, then the complete gate and audit.

## Invariants To Preserve

- No unrelated bulk upgrade, surprise Python-floor move, unreviewed source index, or weakened type/security policy.
- Applications remain fully locked; bot PRs clear the same gate as human PRs.
- Major behavior/contract changes carry rationale, migration, and rollback.
- Removed dependencies leave no imports, config, extras, or lock residue.

## Proof

```bash
uv lock --check
uv sync --frozen
uv run --with pip-audit pip-audit
uv tree
git diff -- pyproject.toml uv.lock
make verify
```

Run targeted tests for the consuming boundary. Governing docs: [framework selection](../decisions/framework-selection.md) and [CI and release](../operations/ci-and-release.md).
