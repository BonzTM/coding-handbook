# Shared Constructs

Recommended reusable building blocks for mature Go repos. These exist to remove repeated wiring without creating a junk drawer.

## Default Approach

Prefer a small number of explicit internal packages that solve common cross-cutting needs well.

| Package | Owns | Do not turn it into |
|---|---|---|
| `internal/runtime` | assembly helpers, startup utilities, process lifecycle glue | a second `main` full of business logic |
| `internal/buildinfo` | version, commit, build-time metadata shown in logs or `version` output | a catch-all constants package |
| `internal/telemetry` | logger creation, metrics registries, tracing setup, health helpers | a wrapper around every stdlib call |
| `internal/testutil` | builders, fixtures, fake clocks, shared test harnesses | hidden production logic or giant assertion DSLs |
| `internal/httputil` | JSON helpers, error response helpers, request-size or timeout helpers | a dumping ground for unrelated helpers |

### Constructor Pattern

- Constructors should make dependencies explicit.
- Long constructor signatures are acceptable when they reveal real dependencies; a magical container is not inherently cleaner.
- Group dependencies in a small struct only when they naturally belong together.

### Shutdown Pattern

- Create the root context in `main` with `signal.NotifyContext`.
- Pass that context to servers, workers, and background loops.
- Close shared resources in a bounded, ordered shutdown path.

## Common Mistakes And Forbidden Patterns

- A `util` or `common` package that gradually becomes the real architecture.
- Reflection-heavy DI added before manual wiring is actually painful.
- Hidden dependencies passed through `context.Value`.
- Test helpers that assert too much and make failures harder to read.

## Verification And Proof

- A new contributor should be able to read `cmd/<app>/main.go` and understand the dependency graph.
- Shared helpers should reduce duplication without making call sites mysterious.
- If a shared package cannot state exactly what it owns, it is probably too broad.
