# Style and Review

Idioms and review heuristics that keep Python code explicit, typed, and resilient.

## Default Approach

PEP 8 is the baseline and Ruff is the formatter/linter that enforces the mechanical policy. Do not reproduce formatter output or the rule catalog in prose; [linting](../quality/linting.md) owns selection and suppression.

### Formatting And Imports

Run `ruff format` and `ruff check`; never hand-format around the tool. Imports are absolute across package boundaries, ordered by Ruff, and dependency-direction compliant. Relative imports may be used narrowly within a cohesive subpackage only where the configured `TID` policy permits them. Wildcard imports are forbidden.

### Naming

- Modules, functions, local variables, and parameters use `snake_case`.
- Classes, exceptions, Protocols, and type aliases use `CapWords`; exception names end in `Error`.
- Constants use `UPPER_SNAKE_CASE` only for genuinely immutable module-level policy/data.
- Boolean names read as predicates (`is_ready`, `has_access`, `should_retry`).
- Avoid mechanics-only names: `utils`, `helpers`, `common`, `manager`, `processor`, and `data` require a more precise owner.
- Leading underscore marks package-private support policy; double-leading underscores are not a general privacy mechanism.

### Docstrings Are Public Contracts

Public modules, classes, functions, methods, Protocol members, and exported constants have docstrings when callers need behavior beyond the signature. Use one repository format consistently; the default is concise Google-style sections for parameters, returns, raises, and examples because the shape stays readable in source. Do not mix Google, NumPy, and Sphinx field styles in one repo.

A docstring states preconditions, invariants, side effects, error/cancellation behavior, ownership/lifecycle, concurrency safety, and compatibility commitments. It does not paraphrase the function name or repeat obvious annotations. Private code receives a docstring only when the contract is non-obvious.

### Comments Explain Why

Comments record constraints, rejected simpler choices, security assumptions, algorithm invariants, or suppression rationale. Delete stale narration. TODOs name an owner/issue and removal condition. Workarounds link the upstream defect or ADR.

### Public Surface And __all__

Keep `__init__.py` empty unless it defines an intentional facade. A facade uses explicit imports and `__all__`; every name is typed, documented, and treated as a compatibility commitment. Do not use `__all__` to disguise a large module or compensate for wildcard imports.

### Function And Module Shape

Each function has one primary responsibility, guard clauses keep nesting shallow, and orchestration is separated from parsing/validation/decisions. Target at most 30 logical lines; 60 is the hard review gate absent a documented cohesive exception. Split modules by responsibility before they become navigation problems; around 300–500 non-generated lines triggers review, not an automatic rewrite.

Prefer keyword-only parameters for several same-typed or optional arguments. Avoid boolean mode switches; use named functions or an enum/config value. Keep decorators transparent and typed. No mutable module globals or import-time I/O.

### Review Questions

- Does the dependency still point inward and is the public surface intentional?
- Is untrusted input parsed once before domain logic?
- Are tasks, clients, sessions, files, and generators owned and closed on every path?
- Are failure, cancellation, logging, and telemetry observable without leaking sensitive data?
- Does the test prove behavior and negative paths rather than implementation call order?
- Did the change add a dependency, suppression, compatibility obligation, or operational failure mode that needs explicit review?

## Common Mistakes And Forbidden Patterns

- Style arguments already decided by Ruff, or formatter directives used as local preference.
- Wildcard imports, implicit re-exports, or a broad `__all__` facade.
- Comments/docstrings that restate code while omitting errors, side effects, and ownership.
- Generic names that conceal ownership; single-letter names outside tiny local scopes.
- Functions over the 60-line gate, deep nesting, or one function mixing I/O, validation, and decisions.
- Boolean parameters whose call sites do not explain the behavior.
- Mutable module globals, import-time registration/configuration, or hidden dependencies.
- Unbounded TODOs, blanket suppressions, or workarounds without an upstream reference.

## Verification And Proof

```bash
uv run ruff format --check .
uv run ruff check .
uv run mypy .
make verify
```

Read changed public APIs as a first-time caller. Every exported contract explains what annotations cannot; every suppression states why; touched functions pass the size/responsibility gate; and imports, ownership, error paths, and tests remain obvious at the call site.

Related: [package design](package-design.md), [typing discipline](typing-discipline.md), and [PR review checklist](../checklists/pr-review.md).
