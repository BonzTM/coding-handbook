# Style And Review

TypeScript readability, documentation, and review rules for maintainable changes.

## Default Approach

Prefer obvious code, focused functions, explicit failure, and repository-owned automation over personal style.

### Formatting And Naming

Prettier owns formatting; reviewers do not negotiate whitespace. ESLint owns correctness and selected consistency rules.

Use `camelCase` for values and functions, `PascalCase` for types and React components, and `UPPER_SNAKE_CASE` only for true constants. Name booleans as predicates and functions as actions. Include units and domain meaning in names.

File names use the repository's established kebab-case convention. Avoid `I` prefixes for interfaces, Hungarian notation, and abbreviations not defined in the glossary.

### Function Design

Give each function one primary responsibility, target no more than 30 logical lines, and treat 60 lines as a review gate. Use guard clauses and extracted decisions to keep nesting shallow.

Separate orchestration from parsing, authorization, calculation, mapping, and persistence. Pass dependencies explicitly. Return observable values or typed failures rather than mutating hidden global state.

Loops and external work are bounded. Prefer straightforward control flow over recursion, clever chaining, or metaprogramming. A smaller line count does not justify combining unrelated responsibilities.

### Comments And JSDoc

Comments explain why, invariants, operational consequences, and non-obvious constraints. Delete comments that merely restate syntax.

Document exported library APIs with JSDoc covering behavior, parameters whose meaning is not obvious, failures, cancellation, side effects, and compatibility. Examples must compile or run in verification when practical.

TODOs name an owner or issue and a removal condition. Do not use comments to excuse unsafe assertions or missing error handling.

### Review Order

Review contract and architecture first, then correctness, security/privacy, failure behavior, bounded work, observability, tests, and maintainability. Review generated and lockfile diffs as artifacts, not noise.

Ask for evidence at the changed boundary. A green broad suite does not replace a missing negative-path test. Confirm configuration, migrations, contracts, docs, and operational notes move with behavior.

### Suppressions

Suppressions are line-scoped, name the exact rule, and explain the proven invariant. File-wide disables, `@ts-ignore`, unowned TODO suppressions, and warning-only CI are forbidden. Prefer `@ts-expect-error` in a negative type test because it fails when no error exists.

## Common Mistakes And Forbidden Patterns

- Formatting debates that contradict Prettier.
- Large handlers, components, or scripts mixing orchestration and decisions.
- Clever conditional types or chains whose diagnostics obscure behavior.
- Comments that narrate code, stale TODOs, or undocumented public failure behavior.
- Drive-by refactors hiding the requested behavior change.
- Review based only on happy-path tests or aggregate coverage.
- Broad lint/type suppressions and `any` introduced for convenience.

## Verification And Proof

- Prettier check, ESLint, and typecheck pass with zero warnings or diagnostics.
- Every touched function remains focused; any function over the target has a cohesion justification.
- Public API documentation matches exports and tested behavior.
- Review identifies negative paths, cancellation, bounds, and sensitive-data handling where relevant.
- Lockfile, generated artifact, schema, migration, and changelog diffs are reviewed.
- Suppressions are minimal, local, justified, and searchable.

Related: [module-design.md](module-design.md), [../quality/linting.md](../quality/linting.md), and [git-workflow.md](git-workflow.md).
