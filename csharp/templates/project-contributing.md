<!--
Template: downstream CONTRIBUTING.md for a repo built on the C# handbook.
Copy to the repo root as CONTRIBUTING.md and replace every <PLACEHOLDER>.
Keep it short — it routes contributors to the handbook; it does not restate it.
-->

# Contributing To <PROJECT_NAME>

Thanks for contributing. This repo follows the [C# Project Handbook](<HANDBOOK_URL>); this file is the local entry point and points at the handbook for the detail. When in doubt, the handbook is the contract.

## Setup

```bash
git clone <REPO_URL>
cd <REPO_DIR>
pwsh ./verify.ps1
```

`pwsh ./verify.ps1` is the single gate: it runs restore (locked), format-check, build (warnings-as-errors), test, and the vulnerability audit. It must be green before you open a PR. The integration suite is a separate `pwsh ./verify.ps1 -Integration` (requires Docker for Testcontainers). Requires the .NET SDK pinned in `global.json` and PowerShell 7 (`pwsh`) — the only blessed script runtime, one script on all three OSes.

A complete worked example of the layout and proof style lives in <REFERENCE_SERVICE_PATH_OR_LINK>.

## Branches, Commits, And PRs

Follow [foundations/git-workflow.md](<HANDBOOK_URL>/foundations/git-workflow.md):

- Trunk-based: branch off `main`, keep the branch short-lived, delete it after merge. `main` is protected.
- Keep PRs small and single-purpose (guideline: under ~400 lines of human-authored diff). Split refactors from behavior changes.
- Write the PR title as a [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) subject — `type(scope): description`, imperative mood — because we **squash-merge** and that title becomes the commit on `main` and feeds the CHANGELOG and SemVer bump.
- Put the WHY in the commit/PR body. Link the issue or ADR for architectural changes.
- A PR merges only after CI `pwsh ./verify.ps1` is green on the full OS matrix and at least one non-author approval. Fill in the PR template at <PR_TEMPLATE_PATH>.

## Definition Of Done

A change is not done until it clears the project's done bar — see [checklists/feature-definition-of-done.md](<HANDBOOK_URL>/checklists/feature-definition-of-done.md). In short:

- Behavior is implemented at the right boundary and covered by tests that fail without the change; DB and external boundaries have real integration tests.
- Contracts, config keys, and public API surface are documented; backward-incompatible changes carry a deprecation plan.
- `pwsh ./verify.ps1` is green from a clean tree; coverage did not regress on mandatory paths.
- Operator-visible changes (new config keys, migrations, ports, contract changes) have a `CHANGELOG.md` entry — see <CHANGELOG_PATH>.

## Where To File Issues

- Bugs and feature requests: <ISSUE_TRACKER_URL>
- Security reports: do **not** open a public issue — follow <SECURITY_POLICY_PATH> (`SECURITY.md`).
- Proposing a change to an architectural decision: open an ADR per <ADR_PROCESS_PATH> before the implementing PR.

## Maintainers

Code ownership and review routing: <CODEOWNERS_PATH>. Questions: <CONTACT_CHANNEL>.
