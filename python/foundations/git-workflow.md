# Git Workflow

The version-control contract for branches, commits, reviews, and a clean derivable history.

## Default Approach

Use trunk-based development. `main` stays releasable, protected, and green; short-lived branches carry one reviewed change.

### Branching

- Branch from `main`, name the work (`feat/order-export`, `fix/readyz-race`), and delete after merge.
- Rebase a private branch onto current `main`; do not back-merge `main` into it.
- Hide incomplete behavior behind a short-lived typed [feature flag](configuration.md#feature-flags), not a long-running branch.
- Never rewrite or force-push `main`.

### Pull Request Size

One logical change per PR. Aim below roughly 400 non-generated, non-vendored changed lines so one reviewer can hold the behavior in mind. Generated output, goldens, and lockfile churn do not excuse an unreadable authored diff; isolate mechanical changes and explain unavoidable breadth.

The `uv.lock` diff is reviewed as supply-chain evidence. The PR names direct dependency changes, explains transitive movement, and does not mix unexplained lock regeneration with behavior work.

### Conventional Commits

The squash subject follows [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/):

```text
<type>[optional scope][!]: <imperative description>

[body explaining why]
[footer, including BREAKING CHANGE when applicable]
```

Use the standard types (`feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`). `feat` implies MINOR, `fix` PATCH, and `!` or `BREAKING CHANGE:` MAJOR for published SemVer surfaces. Keep the subject lower-case, imperative, and without a period. Link the issue or ADR for non-obvious decisions.

### Squash To Main

Squash-merge each PR into one complete Conventional Commit. The PR title is the future subject and stays accurate as scope changes. No merge commits on `main`; permanent history is one reviewed change per commit.

### Changelog

Human-facing releases maintain [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) categories and link entries to behavior/compatibility impact, not commit noise. Release automation may derive a draft from Conventional Commits, but a human reviews wording, migration notes, deprecations, and security effects. Libraries follow [release library version](../recipes/release-library-version.md).

### Protected Main

Branch protection blocks direct pushes and requires `make verify` from CI plus at least one non-author approval. Dismiss stale approvals after new commits. Required checks, ownership rules, and emergency procedure are configured, not merely described.

## Common Mistakes And Forbidden Patterns

- A PR combining refactor, feature, dependency churn, and unrelated cleanup.
- Long-running branches that drift from `main`.
- Merge commits on `main`, direct pushes, or bypassing a failed gate for urgency.
- Subjects such as `updates`, `fix bug`, or `added feature`; capitalized/punctuated non-imperative descriptions.
- Breaking behavior merged without `!`/`BREAKING CHANGE:` and migration notes.
- An unexplained `uv.lock` rewrite or application dependency left unpinned.
- Changelog entries that list internals but omit consumer/operator impact.

## Verification And Proof

```bash
make verify
git diff --check
git log --first-parent --oneline main
```

Branch protection proves the green gate and independent review. The first-parent log is linear Conventional Commits, one per PR. Review the `uv.lock` diff explicitly. A release dry-run derives the intended version and CHANGELOG section without discovering an unmarked breaking change.

Related: [CI and release](../operations/ci-and-release.md), [changelog template](../templates/changelog.md), and [definition of done](../checklists/feature-definition-of-done.md).
