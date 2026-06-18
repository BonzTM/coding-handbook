# Recipe: Add CLI Command

Use this when a repo-owned binary gains a new subcommand or user-visible flag set.

## Files To Touch

- `cmd/<app>/...`
- `internal/config` if the command introduces new config or env integration
- the owning core package if the command triggers domain behavior
- CLI tests and README/help text when user-visible behavior changes

## Steps

1. Decide whether stdlib `flag` is still enough; only reach for `cobra` if the command tree truly needs it.
2. Add the command wiring in `cmd/<app>` and keep it thin.
3. Parse flags explicitly and pass validated values into core services.
4. Return errors with clear exit behavior; do not hide failures behind logs only.
5. Update help text and usage examples if the command is user-facing.

## Invariants To Preserve

- business logic stays out of `cmd/`
- flags and env precedence remain documented and predictable
- version and help output still work from the built artifact

## Proof

- command parsing tests
- `go build ./cmd/...`
- local `--help` smoke test
- one success and one failure-path execution against the built binary
