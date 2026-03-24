# Style and Review

Idioms and review heuristics that keep Go code obvious, boring, and resilient.

## Default Approach

Use `gofmt` and standard Go idioms first. Prefer clarity over cleverness.

### Formatting And Comments

- `gofmt -s` is mandatory.
- `goimports` is a good editor integration, but the repo contract is still `gofmt`-clean code.
- Every exported type, function, method, and package should have a comment that explains the contract, not the obvious implementation.

### API Shape

- Accept interfaces when the caller benefits; return concrete types when you own the implementation.
- Prefer zero-value-friendly types when practical.
- Use constructors when invariants must hold or dependencies must be injected.
- Keep boolean parameters rare; if a call site is hard to read, the API is probably wrong.

### Values, Pointers, And Receivers

- Use pointer receivers when the method mutates state or the type contains mutexes, slices, maps, or large fields.
- Use value receivers for small immutable value types.
- Be consistent within a type; mixed receiver styles are a smell unless the distinction is deliberate.

### Everyday Idioms

- Keep the happy path on the left margin with early returns.
- Error strings are lower-case and do not end with punctuation.
- Use `time.Time` and `time.Duration`, not raw integers, for temporal values.
- Prefer explicit field names in large literals and public-facing config structs.

## Common Mistakes And Forbidden Patterns

- Naked returns in non-trivial functions.
- Ignored errors without a comment explaining why the ignore is safe.
- Global mutable state for config, pools, or loggers.
- Giant option structs or constructors that are really several concepts glued together.
- Comments that restate the function name but not its contract.

## Review Questions

- Is the package boundary still clear after this change?
- Does the API expose only what callers need?
- Are error paths, logging, and telemetry consistent with the rest of the repo?
- Would a new contributor understand where this behavior belongs by reading the call site?
- Did the change add complexity that a smaller refactor or simpler type could avoid?

## Verification And Proof

```bash
gofmt -s -l .
go vet ./...
```

For exported APIs, also read the generated docs or comments as if you were a first-time caller. If the contract is unclear in prose, it will be unclear in code review too.
