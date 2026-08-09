# Recipe: Add CLI Command

Use this when adding a command, subcommand, option, or flag.

## Files To Touch

- `src/cli/<name>.ts` adapter and parser
- focused core use case under `src/core/`
- `src/index.ts` or CLI composition root
- config schema, help text, README, and shell examples
- command tests and process smoke tests

## Steps

1. Define command syntax, defaults, precedence, output contract, and exit codes.
2. Parse argv as untrusted input and reject unknown options and partial numeric parses.
3. Keep parsing and terminal formatting in the CLI adapter; call a focused core use case.
4. Pass `AbortSignal` through I/O and handle `SIGINT` with bounded cleanup.
5. Write ordinary results to stdout and diagnostics to stderr.
6. Avoid shell evaluation; subprocesses use fixed executables and argument arrays.
7. Bound input files, output, pagination, concurrency, retries, and execution time.
8. Keep secret values out of argv where process listings expose them; prefer runtime injection.
9. Update `--help`, README examples, completion surfaces, and compatibility notes together.

```bash
npm test -- --runInBand src/cli/<name>.test.ts
node dist/index.js <name> --help
npm run verify
```

## Invariants To Preserve

- CLI adapters contain no business or persistence decisions.
- Invalid input fails before external work begins.
- Exit codes are stable, documented, and tested.
- Cancellation releases files, processes, pools, and telemetry.
- Machine-readable output stays stable and contains no incidental logs.
- Help examples are copy-pasteable and use safe placeholders.

## Proof

- Parser tests cover help, unknown flag, missing value, invalid value, and precedence.
- Core tests prove behavior without process globals.
- Process smoke tests assert stdout, stderr, exit code, signal, and built artifact behavior.
- Shell-injection-shaped input remains an argument, not executable syntax.
- `npm run verify` is green.
