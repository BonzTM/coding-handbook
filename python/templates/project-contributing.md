<!-- Copy to CONTRIBUTING.md and replace every <placeholder>. -->

# Contributing To <project-name>

This repo follows the [Python Project Handbook](<handbook-url>). This file is the local entry point; the handbook is the full contract.

## Setup

```bash
git clone <repo-url>
cd <repo-directory>
uv sync --frozen
make verify
```

`make verify` runs lock-check, frozen sync, format-check, Ruff lint, Import Linter, strict mypy, pytest, and pip-audit. `make test-integration` separately requires <Docker/external prerequisites>.

## Branches, Commits, And PRs

- Branch from protected `main`; keep the branch short-lived and single-purpose.
- Use a Conventional Commits PR title because squash merge makes it the mainline commit.
- Put rationale and compatibility/operational impact in the PR body; link an ADR for hard-to-reverse decisions.
- Merge only after green CI and <required approvals/owners>.

## Definition Of Done

- Behavior is implemented at the right boundary and proved by focused tests; real external semantics have integration proof.
- Contracts, settings, migrations, operator docs, and `CHANGELOG.md` stay synchronized.
- `make verify` passes from the committed lock and compatibility floor proof remains green.
- New dependencies have rationale and a reviewed `uv.lock` diff.

## Where To File Issues

- Bugs/features: <issue-tracker-url>
- Security: report privately through <security-policy-path>; never open a public issue.
- Architecture: create an ADR using <adr-process-path> before implementation.

## Maintainers

Ownership: <codeowners-path>. Questions: <contact-channel>.
