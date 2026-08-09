# Git Workflow

Branch, commit, pull-request, and history rules for reviewable TypeScript changes.

## Default Approach

Ship one logical, green, reviewable change at a time through a protected main branch.

### Branches

Branch from current main and keep branches short-lived. Use descriptive names such as `feat/idempotent-widget-create` or the repository's established issue convention.

Rebase or merge main according to project policy before final verification. Never rewrite shared history without coordination. Release and hotfix branches exist only when the release model requires them.

### Commits

Use Conventional Commits subjects: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `build:`, `ci:`, `chore:`, or `revert:`. Keep the subject imperative and specific.

Each commit should build and explain one coherent step when practical. Do not mix generated artifacts, dependency changes, formatting churn, and unrelated refactors into feature work. Never add a co-author trailer unless explicitly instructed for that commit.

Breaking changes use `!` or a `BREAKING CHANGE:` footer and include migration guidance. Reference issues and ADRs in the body where they explain intent.

### Pull Requests

The PR description states the problem, contract change, approach, proof, security/privacy impact, operational impact, rollout, and rollback. Mark irrelevant concerns explicitly for consequential changes.

Keep the diff small enough to reason about. Split independent behavior, migrations, framework adoption, and cleanup when separation reduces risk without leaving main broken.

Update tests, schemas, migrations, config examples, changelog, docs, and generated artifacts in the same PR as the behavior they describe.

### Review And Merge

Require green `npm run verify`, required integration jobs, and approval from the owning reviewers. Resolve comments with code or evidence; do not dismiss correctness concerns as style.

Use the repository's chosen merge strategy consistently. The resulting main-branch history must preserve the Conventional Commit information required for release automation.

### Dependency And Generated Diffs

Dependency PRs review `package.json`, the full lockfile diff, scripts, licenses, maintenance, advisories, and runtime compatibility. Generated changes identify their source command and are reproducible in CI.

### Emergency Changes

Hotfixes keep the same review and verification bar except where an incident commander explicitly records a temporary exception. Follow immediately with the omitted proof, root-cause work, and rollback or cleanup.

## Common Mistakes And Forbidden Patterns

- Direct pushes to protected main or merging a red gate.
- Huge commits labeled “cleanup” that mix behavior and formatting.
- Force-pushing a shared branch without coordination.
- Dependency-bot auto-merge without lockfile and advisory review.
- Generated output edited by hand or without its source change.
- Operator-visible changes without changelog, rollout, or rollback notes.
- Co-author trailers added automatically by tools or agents.

## Verification And Proof

- PR title and commits follow the repository's Conventional Commit policy.
- `npm run verify` passes on the final commit from a lockfile-honest install.
- Required integration and security checks are green.
- The diff contains synchronized tests, docs, contracts, config, migrations, and changelog.
- An accepted ADR precedes any changed invariant or hard-to-reverse architecture choice.
- Release notes identify breaking and operator-visible changes with rollback guidance.

Related: [style-and-review.md](style-and-review.md), [../operations/ci-and-release.md](../operations/ci-and-release.md), and [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md).
