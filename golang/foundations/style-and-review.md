# Style and Review

Idioms and review heuristics that keep Go code obvious, boring, and resilient.

## Default Approach

Use `gofmt` and standard Go idioms first. Prefer clarity over cleverness.

### Formatting And Comments

- `gofmt -s` is mandatory.
- `goimports` is a good editor integration, but the repo contract is still `gofmt`-clean code.
- Every exported type, function, method, and package should have a comment that explains the contract, not the obvious implementation.

### Naming

- Initialisms stay uppercase: `ID`, `URL`, `HTTP`, `API` — `userID` not `userId`, `ServeHTTP` not `ServeHttp`.
- Getters carry no `Get` prefix: `Name()` not `GetName()`. Setters, when a type needs them, are `SetName(...)`.
- Receiver names are one or two letters abbreviating the type (`s` for `*Server`, `wc` for `*WidgetCache`), never `self` or `this`, and the same name on every method of the type.
- Variable name length scales with scope: `i` and `ok` are right in a three-line loop; anything package-level or long-lived gets a descriptive name.
- Exported names are read with their package qualifier — avoid stutter (`orders.Service`, not `orders.OrderService`); the package-naming side lives in [package-design.md](./package-design.md) ### Naming Rules.
- The lint gate enforces part of this mechanically — revive's `var-naming` (initialisms) and `receiver-naming` (receiver consistency) rules, routed via [../quality/linting.md](../quality/linting.md). The rest is review.

### Documentation Standards

Docs are part of the contract. Treat them with the same rigor as the code they describe.

- Every package has a package comment that states its purpose and boundaries. Put it on a dedicated `doc.go` when the prose runs more than a few lines or the package has no single obvious "main" file; otherwise a package comment on one source file is fine. The comment names what the package is for and what it deliberately does not do.
- Every exported identifier is documented, and the comment begins with the identifier's name (`// Store persists ...`, `// ErrNotFound is returned when ...`). This is the godoc convention, not a style preference: tooling and readers rely on the leading name to render and grep docs.
- A comment states the contract — preconditions, error and nil behavior, concurrency safety, ownership of returned values — not a restatement of the signature. If the comment only repeats the name, delete it or replace it with the contract.
- Non-trivial exported APIs ship runnable `Example_` functions that double as living documentation: they appear inline in godoc and fail the build if the API drifts. Write them for anything a caller would otherwise have to guess at. The test mechanics (naming, `// Output:` blocks, when to add them) live in [../quality/testing.md](../quality/testing.md) — do not duplicate them here.
- Deprecate in-band with the godoc-recognized form `// Deprecated: <reason>; use <replacement>.` as the first line of a paragraph in the symbol's comment, so editors and `staticcheck` surface it at call sites. The full deprecation and removal sequence lives in [contracts-and-compatibility.md](./contracts-and-compatibility.md).
- Doc-as-code: architecture notes, design rationale, and diagrams live in the repo (`docs/` or ADRs under [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md)), versioned and reviewed in the same PR as the change they describe. A wiki or slide deck that drifts from `main` is worse than no doc.

### API Shape

- Accept interfaces when the caller benefits; return concrete types when you own the implementation.
- Prefer zero-value-friendly types when practical.
- Use constructors when invariants must hold or dependencies must be injected.
- Keep boolean parameters rare; if a call site is hard to read, the API is probably wrong.

### Struct Embedding

- Embedding to satisfy or forward an interface is fine: embed the interface (or a base implementation) and override the methods you care about.
- Never embed a type in an exported API struct when it leaks methods you do not mean to promise — the embedded type's method set becomes your public API, and removing it later is a breaking change.
- Prefer a named field unless you deliberately want method promotion; a field keeps the contract explicit at every call site.
- Never embed to "save typing" in domain types. Embedding is a statement about the API, not a shortcut.

### Values, Pointers, And Receivers

- Use pointer receivers when the method mutates state or the type contains mutexes, slices, maps, or large fields.
- Use value receivers for small immutable value types.
- Be consistent within a type; mixed receiver styles are a smell unless the distinction is deliberate.

### Copying And Immutability

Copy semantics are part of the contract, not an afterthought. Decide deliberately what a value means when it is passed by copy.

- Large structs copy on every value pass and assignment. Pass big aggregates by pointer; keep small value types (timestamps, IDs, money — see [data-modeling.md](./data-modeling.md)) as values so they stay comparable and allocation-free.
- A pointer means one of two things — be intentional about which. Pointer-as-optional (`*int` for "maybe absent") and pointer-as-shared-mutable ("we all see each other's writes") are different contracts; do not let one field carry both. For optionality prefer a clear sentinel, a separate `ok bool`, or a dedicated optional type over an overloaded pointer.
- Never copy a struct that contains a `sync.Mutex`, `sync.WaitGroup`, `sync.Once`, or `atomic.*` value: the copy and the original share nothing, and the lock no longer protects what you think. Hold such types behind a pointer, or embed them and only ever pass `*T`. Do not return a locked struct by value or range over a slice of them by value.
- Prefer constructing immutable value objects: validate once in a constructor, expose fields read-only (unexported + accessor, or document "do not mutate"), and hand out copies rather than shared mutable references. See [data-modeling.md](./data-modeling.md) for the value-object and copy-ownership rules.
- No mutable package-level globals for state. Package-level `var` is for genuinely constant, immutable data (lookup tables, sentinel errors). Wire everything else through constructors and pass it explicitly; `init()` is not for configuration, registration, or dependency wiring (see the global-state and `init()` rules below and in [package-design.md](./package-design.md)).

### Everyday Idioms

- Keep the happy path on the left margin with early returns.
- Error strings are lower-case and do not end with punctuation.
- Use `time.Time` and `time.Duration`, not raw integers, for temporal values.
- Prefer explicit field names in large literals and public-facing config structs.

## Common Mistakes And Forbidden Patterns

- Naked returns in non-trivial functions.
- Ignored errors without a comment explaining why the ignore is safe.
- Global mutable state for config, pools, or loggers.
- Package-global mutable state of any kind, including `init()` that wires, configures, or registers.
- Copying a value that contains a lock (`sync.Mutex`, `sync.WaitGroup`, `atomic.*`) — pass it by pointer instead.
- Speculative generic APIs: type parameters added before a second concrete type or to wrap a single type.
- Giant option structs or constructors that are really several concepts glued together.
- Comments that restate the function name but not its contract.
- A package with no package comment, or a `doc.go` that exists but says nothing about purpose or boundaries.
- Exported identifiers whose comment does not start with the identifier name, breaking the godoc convention.
- Design docs, architecture diagrams, or rationale kept outside the repo (wiki, slides, chat) where they silently drift from `main`.
- Marking something obsolete in prose or a changelog without the in-band `// Deprecated:` form, so call sites get no warning.

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
go doc ./...
```

`go vet ./...` runs the `copylocks` analyzer, which flags any value copy of a type containing a `sync.Mutex`, `sync.WaitGroup`, `sync.Once`, or `atomic.*` — including silent copies via value receivers, value passing, value returns, and ranging. Treat a `copylocks` finding as a hard failure, not a lint suggestion; these are routed through `make verify`.

For exported APIs, also read the generated docs or comments as if you were a first-time caller. If the contract is unclear in prose, it will be unclear in code review too. Read `go doc ./...` intentionally: every package should have a purpose line, and every exported symbol a contract that starts with its name. Confirm the package's `Example_` tests pass (they run under `go test`, which is inside `make verify`) — a failing example means the documented usage no longer compiles or no longer produces the documented output.
