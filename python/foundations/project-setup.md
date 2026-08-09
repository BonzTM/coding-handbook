# Project Setup

Default repository shape, packaging policy, and bootstrap expectations for new Python projects.

## Default Approach

Start with one PEP 621 project, one import package under `src/`, and tests outside the package. The committed [project template](../templates/pyproject.toml) is authoritative for exact dependencies and tool settings.

### Bootstrap Commands

```bash
uv init --build-backend uv --python 3.11
mkdir -p src/<app> tests
uv lock
uv sync
```

Replace placeholders before the first lock. Applications commit `uv.lock`; published libraries may omit it only when their documented consumer workflow tests dependency ranges separately.

### Preferred Tree

```text
repo/
  .python-version
  pyproject.toml
  uv.lock
  Makefile
  src/
    <app>/
      __init__.py
      __main__.py
      main.py
  tests/
```

The `src/` layout prevents the repository root from making an uninstalled package accidentally importable; tests therefore exercise the installed artifact shape. This is the isolation benefit documented by the [Python Packaging User Guide](https://packaging.python.org/en/latest/discussions/src-layout-vs-flat-layout/).

### pyproject.toml Is The Project Contract

`pyproject.toml` contains three kinds of configuration:

- `[build-system]` selects `uv_build` and its build requirement.
- `[project]` holds PEP 621 name, description, `requires-python = ">=3.11"`, dependencies, optional dependency groups, and `[project.scripts]`.
- `[tool.*]` holds repo policy: `[tool.ruff]`, `[tool.mypy]`, `[tool.pytest.ini_options]`, and `[tool.coverage.*]`. Import Linter configuration stays in its supported project configuration surface.

The standardized anatomy follows the [pyproject.toml specification](https://packaging.python.org/en/latest/specifications/pyproject-toml/). Do not split tool policy across ad hoc dotfiles when the tool supports `pyproject.toml`.

### Runtime Floor And Development Pin

`requires-python = ">=3.11"` is the compatibility promise. `.python-version` is the current stable interpreter used by developers and CI. They serve different purposes: the pin may advance without changing the floor, but code must remain valid on 3.11 until the floor moves through an ADR and compatibility proof.

### uv Owns The Workflow

- `uv lock` resolves and updates the lock deliberately.
- `uv lock --check` proves metadata and lock agree without rewriting either.
- `uv sync` creates or updates the project environment; CI and verification use `uv sync --frozen`.
- `uv run <command>` executes inside that environment.
- No ad hoc `pip install`, Poetry state, or hand-maintained requirements export participates in the gate.

Use `uv_build` for the default pure-Python, single-package distribution. Hatchling requires an ADR showing a build hook or layout that `uv_build` cannot express; see [framework selection](../decisions/framework-selection.md).

### Entrypoints Stay Thin

Declare installed commands under `[project.scripts]`, whose values point to synchronous callables as specified by the [packaging guide](https://packaging.python.org/en/latest/guides/writing-pyproject-toml/#creating-executable-scripts). `src/<app>/__main__.py` delegates to the same callable so `python -m <app>` and the installed command share one path.

For an async service, the synchronous entry callable starts uvicorn or calls one `asyncio.run(...)`; it owns no business rules. `create_app()` builds the FastAPI application, registers routers, and defines the lifespan owner. Tests call the factory instead of importing a process-global app with startup side effects.

### Tests Are Outside src

`tests/` mirrors behavior and boundaries without becoming part of the wheel. Test through installed package imports. Shared fakes and fixtures belong under `tests/` unless production packages genuinely consume them; see [testing](../quality/testing.md).

## Common Mistakes And Forbidden Patterns

- A flat layout that lets tests pass against the checkout but fail after installation.
- A local interpreter using post-3.11 syntax or stdlib APIs while metadata still promises 3.11.
- Uncommitted or stale `uv.lock` in an application; CI that resolves instead of using `--frozen`.
- `setup.py` as primary metadata, or duplicated metadata in multiple files.
- Business logic, resource creation, or argument parsing spread through `__main__.py`.
- A module-level FastAPI app whose import opens clients, reads config, or configures logging.
- Tool settings scattered across files without a tool limitation or documented reason.

## Verification And Proof

```bash
uv lock --check
uv sync --frozen
uv build
make verify
```

Proof is complete when the wheel contains only the intended package, an installation smoke test can run the declared script and `python -m <app>`, tests import through `src/`, and both the 3.11 floor and current development pin pass the supported verification matrix.

Related: [package design](package-design.md), [CI and release](../operations/ci-and-release.md), and [templates](../templates/README.md).
