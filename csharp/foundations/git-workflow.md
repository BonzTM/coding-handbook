# Git Workflow

The version-control contract: how branches, commits, and merges flow into a clean, derivable history.

## Default Approach

Trunk-based development. `main` is always releasable, always green, and always protected. Everything else is a short-lived branch that exists only long enough to land one reviewed change.

### Branching

- Branch off `main`, name it for the work (`feat/order-export`, `fix/readyz-race`), and delete it after merge.
- Keep branches short-lived: hours to a couple of days, not weeks. A branch that lives a week is a branch that drifts from `main` and accretes merge pain.
- Rebase onto `main` to stay current; do not back-merge `main` into the branch. The branch history is throwaway because the merge squashes it (see below), so optimize it for your own review, not for posterity.
- No long-running feature branches. Hide incomplete work behind a [feature flag](configuration.md) and ship it dark instead of holding a branch open.

### Pull Request Size

Small, reviewable PRs. The target is a diff a reviewer can hold in their head in one sitting — as a guideline, **under ~400 lines of non-generated, non-vendored change**, ideally far less. Generated artifacts — gRPC stubs, EF Core migration scaffolding (`*.Designer.cs`, `*ModelSnapshot.cs`), `PublicAPI.*.txt` updates, golden files, and `packages.lock.json` churn — do not count against the budget but should land in their own commits so the human-authored diff stays legible.

- One logical change per PR. A refactor and the feature that motivated it are two PRs: land the mechanical refactor first, then the behavior change on top.
- If a change is unavoidably large (a stub regeneration, a wide rename, a `dotnet format` sweep), say so in the PR body and make it mechanically obvious so review can be a spot-check rather than a line read.
- Splitting is the author's job, not the reviewer's. A PR too large to review is not done.

### Commit Messages — Conventional Commits

The merged commit subject follows [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) so that CHANGELOG entries (Keep a Changelog format, per [../templates/changelog.md](../templates/changelog.md)) and the SemVer bump can be derived mechanically rather than guessed at release time.

```
<type>[optional scope]: <description>

[optional body explaining WHY]

[optional footer(s)]
```

- **Type** is one of the spec's set: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`. `feat` and `fix` are the only two that drive a release; the rest are housekeeping.
- **Scope** is an optional area in parentheses (`feat(api): ...`, `fix(db): ...`), drawn from the project or boundary touched (`api`, `core`, `infrastructure`, `db`).
- **Description** is a terse, imperative-mood summary: "add order export endpoint", not "added" or "adds". Lower-case, no trailing period. It states the change, not the diff.
- **Body** carries the WHY — the motivation, the rejected alternative, the constraint that forced the shape. The diff already shows the what; the body explains why it was worth doing this way.
- **Breaking changes** are marked with a `!` before the colon (`feat(api)!: drop v1 export route`) and/or a `BREAKING CHANGE:` footer. Per the spec, a breaking change drives a MAJOR bump regardless of type; `feat` drives MINOR, `fix` drives PATCH. This is the mechanical link from commit to SemVer in [contracts-and-compatibility.md](contracts-and-compatibility.md) and [../operations/ci-and-release.md](../operations/ci-and-release.md).
- Link the motivating issue or ADR in the footer (`Refs: #123`, `Refs: ADR-0007`) whenever the change is architectural or non-obvious; see [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md).

### Merging — Squash To Main

Squash-merge every PR into `main`. The result is one commit per PR with a clean Conventional Commits subject, so `main` reads as a linear sequence of complete, reviewed changes.

- The PR title is the future squash subject — write it as a Conventional Commits line and keep it accurate as the PR evolves.
- No merge commits on `main`. Merge commits from back-merging or from a non-squash merge interleave half-finished intermediate states into the permanent history and break the one-commit-per-change derivation that the CHANGELOG depends on.
- Force-push freely on your own branch before review converges; never force-push or rewrite `main`.

### Protected Main

`main` is protected by branch-protection rules, not by convention:

- Direct pushes to `main` are blocked. Every change arrives through a PR.
- A PR cannot merge until **the CI run of `pwsh ./verify.ps1` is green on the full OS matrix (ubuntu/windows/macos)** and it has **at least one approving review** from someone other than the author.
- Stale approvals are dismissed when new commits land, so the approval matches the merged code.

## Common Mistakes And Forbidden Patterns

- Giant PRs that bundle a refactor, a feature, and a rename so review degrades to rubber-stamping. Split them.
- Merge commits on `main` from back-merging `main` into a branch or from a non-squash merge — they muddy history and break per-PR CHANGELOG derivation.
- Commit messages that restate the diff (`update OrderService.cs`, `fix bug`, `changes`) instead of the contract and the WHY.
- Non-imperative or capitalized subjects (`Added export`, `Fixes the thing.`) that break Conventional Commits tooling and the SemVer derivation.
- Long-running feature branches instead of trunk + feature flags; they rot against `main` and produce a merge-day surprise.
- Pushing to `main` directly, or relaxing branch protection "just this once" to land an urgent fix without review or a green gate.
- Letting a breaking change merge with a plain `feat:`/`fix:` subject, so the release tool computes the wrong SemVer bump.
- Committing local-only artifacts — `bin/`, `obj/`, `.vs/`, user-specific `*.user` files — instead of relying on the template `.gitignore`; or hand-editing `packages.lock.json` rather than regenerating it via restore.

## Verification And Proof

- Branch protection on `main` requires a green `pwsh ./verify.ps1` CI run (the ubuntu/windows/macos matrix) and at least one non-author approval before merge — verify the rule is enabled, not just documented.
- `git log --first-parent --oneline main` reads as a linear list of Conventional Commits subjects, one per landed PR, with no merge-commit noise.
- Each architectural PR links an issue or ADR in its body or commit footer.
- Spot-check that subjects parse as `type(scope): description`; a release dry-run (the [../recipes/release-library-version.md](../recipes/release-library-version.md) flow) derives the next version and Keep a Changelog section from them without manual editing.

Related: [../operations/ci-and-release.md](../operations/ci-and-release.md) for the release and changelog pipeline, [../templates/changelog.md](../templates/changelog.md) for the Keep a Changelog format, [../templates/pull_request_template.md](../templates/pull_request_template.md) for the PR checklist, and [../checklists/feature-definition-of-done.md](../checklists/feature-definition-of-done.md) for the done bar a PR must clear.
