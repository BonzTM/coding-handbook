# Recipe: Add CLI Command

Use this when an installed command gains a subcommand or user-visible option.

## Files To Touch

- `src/<app>/cli.py` or a focused command module
- `[project.scripts]` and thin `src/<app>/__main__.py` when entrypoints change
- core use case, README/help text, and command tests

## Steps

1. Use stdlib `argparse`; a larger CLI framework requires the escalation in framework selection.
2. Build a parser in a focused function and dispatch to a thin synchronous entry callable.
3. Validate/normalize arguments, load settings once, and pass typed values to core.
4. Map expected failures to stable stderr text and non-zero exit codes; never log-only a command failure.
5. Keep `python -m <app>` and the installed `[project.scripts]` command on the same path.

## Invariants To Preserve

- Argument parsing and process exit stay outside business logic.
- Secrets do not appear in argv examples, help, logs, or errors.
- Subprocesses use argv lists, checked status, timeout, and bounded output; no `shell=True`.
- Help/version output works from the built wheel.

## Proof

```bash
uv run pytest tests -k cli
uv run <command> --help
uv build --no-sources
uv run --with ./dist/<wheel> --no-project -- <command> --help
make verify
```

Exercise one success and one failure exit. Governing docs: [project setup](../foundations/project-setup.md) and [framework selection](../decisions/framework-selection.md).
