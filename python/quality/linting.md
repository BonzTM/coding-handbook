# Linting

The single source of truth for formatting, static linting, typing, and import-boundary policy.

## Default Approach

Ruff owns formatting and linting. Mypy owns static typing. Import Linter owns architectural dependency rules. The committed [pyproject template](../templates/pyproject.toml) is the executable policy; prose explains why the gate exists and does not replace configuration. Keep Import Linter contracts under `[tool.importlinter]`, one of its documented [configuration surfaces](https://import-linter.readthedocs.io/en/stable/get_started/configure/).

### Running The Gate

Humans and CI invoke the same Make targets:

```bash
make format        # uv run ruff format .
make format-check  # uv run ruff format --check .
make lint          # uv run ruff check .
make imports       # uv run lint-imports
make types         # uv run mypy .
make verify        # ordered repository gate
```

`make verify` runs lock-check, frozen sync, format-check, lint, imports, types, test, and audit. Individual targets stay runnable for fast feedback. CI calls the Make target; it does not reconstruct the commands in workflow YAML.

### Ruff Rule-Family Policy

The [Ruff rule catalog](https://docs.astral.sh/ruff/rules/) is authoritative for every selector. The template pins the tool and exact selection. Enable stable rules deliberately; preview, deprecated, and removed rules require separate review.

| Selector | Source family | Why it is enabled |
|---|---|---|
| `E`, `W` | pycodestyle errors and warnings | Catch syntax-adjacent mistakes and mechanical style defects not owned by the formatter. |
| `F` | Pyflakes | Catch undefined names, unused imports, and invalid bindings. |
| `I` | isort | Keep imports deterministic and reviewable. |
| `UP` | pyupgrade | Keep code on idioms supported by the declared Python floor. |
| `B` | flake8-bugbear | Catch mutable defaults, closure traps, assertion misuse, and likely defects. |
| `SIM` | flake8-simplify | Remove control flow that obscures a simpler equivalent. |
| `C4` | flake8-comprehensions | Reject wasteful or misleading collection construction. |
| `DTZ` | flake8-datetimez | Enforce aware datetime construction and use. |
| `T20` | flake8-print | Keep `print()` out of services and libraries. |
| `PT` | flake8-pytest-style | Enforce consistent pytest fixtures, markers, and assertions. |
| `RET` | flake8-return | Catch inconsistent or unnecessary return paths. |
| `RUF` | Ruff-specific | Catch Python defects and suppression/configuration mistakes not covered elsewhere. |
| `ASYNC` | flake8-async | Catch blocking calls and unsafe timeout/task patterns in async code. |
| `S` | flake8-bandit | Surface common injection, subprocess, crypto, deserialization, and secret-handling risks. |
| `PL` | Pylint | Retain high-signal correctness and refactoring checks; the template explicitly curates noisy complexity/style checks. |
| `TRY` | tryceratops | Retain exception-construction and raise-path checks that support the error contract; the template curates opinionated stylistic findings. |
| `N` | pep8-naming | Keep public naming consistent with the Python contract. |
| `A` | flake8-builtins | Prevent shadowing builtins where it makes code ambiguous or unsafe. |
| `ISC` | flake8-implicit-str-concat | Catch accidental string concatenation and formatter-sensitive literals. |
| `PIE` | flake8-pie | Catch unnecessary or error-prone Python constructs. |
| `PERF` | Perflint | Catch clear allocation and iteration inefficiencies without substituting for measurement. |
| `FURB` | refurb | Prefer clearer modern-library constructs after stable-rule review. |
| `TID` | flake8-tidy-imports | Enforce banned and relative-import policy at the statement level. |

Do not select every Ruff rule. `PL`, `TRY`, and `S` especially need explicit template curation because some findings conflict with the function-size contract, test assertions, exception policy, or framework-required signatures. A selector change updates the template, this rationale, and reference exemplar together.

### Per-File Ignores

Per-file ignores encode a real context, never convenience. Tests may relax rules whose production threat model does not apply, including `S101` for pytest assertions and narrowly justified security findings for fixed test credentials or subprocess fixtures. Generated protobuf modules and migration revisions may receive only the exclusions their generator or migration contract requires.

Do not ignore an entire family for `tests/**`, migrations, or generated code when a narrower code suffices. Production packages receive no broad security, async, typing, or import-boundary exemption.

### Suppression Discipline

A Ruff suppression names the exact rule and states why the finding is false or why the safer form cannot be used:

```python
value = trusted_fixture[name]  # noqa: S105 -- fixed non-secret test token
```

Bare `# noqa`, family-wide suppressions, file-level suppression for handwritten code, and reasons such as “false positive” are forbidden. Put the directive on the smallest statement. A repeated suppression is evidence that configuration or design needs review.

Every `# type: ignore[code]` follows [typing discipline](../foundations/typing-discipline.md): narrow error code, adjacent justification, and removal when the upstream defect is fixed. `cast()` is not a suppression substitute.

### Formatter Policy

Ruff format is the only formatter. Do not run Black beside it, add Black-compatible directives as local preference, or let an editor commit a second formatting contract. `ruff format --check .` is read-only in CI; `ruff format .` is the local write command. Ruff documents its formatter as a direct replacement workflow in the [formatter guide](https://docs.astral.sh/ruff/formatter/).

Import sorting is a lint fix, not a second formatter. Apply safe mechanical changes deliberately with `uv run ruff check --fix .`, review the diff, then run the normal gate. CI never applies fixes.

### Mypy Strict Policy

Mypy is the repository type gate; Pyright may run in editors but does not replace it. Configure `strict = true`, `warn_unreachable = true`, and `disallow_any_unimported = true`.

The current [mypy strict reference](https://mypy.readthedocs.io/en/stable/command_line.html#cmdoption-mypy-strict) says strict enables `disallow_any_generics`, `disallow_subclassing_any`, `disallow_untyped_calls`, `disallow_untyped_defs`, `disallow_incomplete_defs`, `check_untyped_defs`, `disallow_untyped_decorators`, `warn_redundant_casts`, `warn_unused_ignores`, `warn_return_any`, `no_implicit_reexport`, `strict_equality`, and `extra_checks`. The set can change, so the lockfile pin and upgrade review matter. `warn_unreachable` and `disallow_any_unimported` are explicit additions.

Do not enable global `ignore_missing_imports`, exclude handwritten packages, or accept an untyped third-party boundary silently. Add a narrow adapter, reviewed stub, or dependency replacement.

### Import Boundaries

Import Linter is the architecture gate. Its forbidden/layer contracts enforce that `core` imports no `api`, `db`, `clients`, `config`, `telemetry`, or `workers`, and that adapters do not shortcut through one another. Ruff `TID` handles banned or relative import statements; it cannot prove the whole package graph.

An import failure is a design failure. Move a consumer-owned `Protocol` inward, split mixed responsibility, or repair composition. Do not hide the cycle with local imports, `TYPE_CHECKING`, or an ignore.

## Common Mistakes And Forbidden Patterns

- Black and Ruff format both configured, or CI rewriting files.
- Selecting all Ruff rules and suppressing the resulting noise instead of curating policy.
- Bare `# noqa`, `# type: ignore`, family-wide suppressions, or directives without a concrete reason.
- Disabling security or correctness families across all tests instead of narrow per-file codes.
- Global `ignore_missing_imports`, broad mypy exclusions, or Pyright replacing the mypy gate.
- Treating `TID` as a replacement for Import Linter contracts.
- Local imports or `TYPE_CHECKING` branches used to conceal a circular dependency.
- CI invoking tool commands that diverge from `make verify`.

## Verification And Proof

```bash
uv run ruff format --check .
uv run ruff check .
uv run lint-imports
uv run mypy .
make verify
```

Linting is done when all commands exit zero; the formatter produces no diff; every selected family has a stated purpose; every suppression is narrow, coded, justified, and reviewed; strict mypy has no broad escape hatch; and the import graph matches [package design](../foundations/package-design.md).

Related: [style and review](../foundations/style-and-review.md), [typing discipline](../foundations/typing-discipline.md), and [CI and release](../operations/ci-and-release.md).
