# Package Design

Package boundaries, export rules, and dependency direction for Go code that stays maintainable under growth.

## Default Approach

Prefer small packages with clear ownership and minimal exported surface.

### Dependency Direction

| Layer | Can depend on | Must not depend on |
|---|---|---|
| `cmd/<app>` | any internal package needed for startup | business logic staying in `main` |
| `internal/api/http`, `internal/api/grpc` | `internal/core`, `internal/config`, `internal/telemetry` | direct SQL, migration logic |
| `internal/core` | stdlib, domain packages, narrowly scoped contracts | transport packages, database-specific implementations |
| `internal/db` | `internal/core` contracts, drivers, query helpers | HTTP or gRPC adapters |
| reusable library surface | only what you are willing to support publicly | private app details hidden under `internal` |

### Interface Placement

- Define interfaces where they are consumed, not where they are implemented.
- Return concrete types by default; accept interfaces when a caller needs substitution.
- Put shared cross-package interfaces in `internal/core` only when they are true domain seams rather than convenience abstractions.
- Keep interfaces small: 1-3 methods. A wide interface is a struct wearing a disguise; either split it or pass the concrete type.
- The interface lives in the consumer's package and names what the consumer needs (`Store`, `Notifier`), not what the implementer is.

### Generics And Type Parameters

Reach for type parameters only for type-safe **containers** and **algorithms** that would otherwise duplicate code per type or fall back to `any` plus runtime assertions. Use ordinary **interfaces** for behavioral polymorphism — when callers vary by what a value *does*, not by what type it *holds*.

- Decision rule: generics when the type is the only thing that changes (a `Set[T]`, a `Map`/`Filter`/`Keys` over a slice, an LRU cache); interfaces when behavior changes.
- Do not add type parameters speculatively, and never to force-fit a single concrete type — that is an indirection tax with no payoff. Write the concrete version; generalize on the second real caller.
- Constrain to the minimum: `cmp.Ordered` for things you compare/sort, `comparable` for map keys, a small named `interface` constraint for everything else. Define a custom constraint when the operation set is domain-specific rather than reusing a vague `any`.
- Overly generic APIs hurt readability more than the duplication they remove. If the signature needs a paragraph to explain, prefer two concrete functions. See [data-modeling.md](./data-modeling.md) for choosing the underlying types these operate on.
- Iterators: default to returning slices. Expose `iter.Seq`/`iter.Seq2` (range-over-func, Go 1.23+) only for genuinely streaming, unbounded, or lazy sequences where materializing a slice is wrong. Do not convert existing slice-returning APIs to iterators.

### Functional Options

When a constructor has optional or extensible configuration, use the `Option` pattern rather than a giant option struct or boolean parameters:

```go
type Option func(*Server)

func WithTimeout(d time.Duration) Option { return func(s *Server) { s.timeout = d } }

func NewServer(addr string, opts ...Option) *Server {
	s := &Server{addr: addr, timeout: 30 * time.Second}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
```

- This keeps required arguments positional, defaults sane, and the option set extensible without breaking callers — and it sidesteps the boolean-parameter and giant-option-struct smells called out in [style-and-review.md](./style-and-review.md).
- Prefer a plain config struct when the configuration is **closed and required together** — a fixed set of mandatory fields validated once is clearer as `Config{...}` than as a pile of `WithX` calls. Options earn their keep only when config is genuinely optional and expected to grow.

### Export Policy

- Export the smallest API surface that serves real callers.
- Start unexported. Make things public only when another package truly needs them.
- Treat exported symbols as support commitments. The cost is future compatibility, not just one capital letter.

### Naming Rules

- Package names should be short, lower-case, and descriptive: `auth`, `config`, `orders`, `telemetry`.
- Avoid stutter: prefer `orders.Service`, not `orders.OrderService`.
- Avoid packages named after mechanics rather than purpose: `util`, `helpers`, `common`, `base`, and `misc` are red flags.

### File Organization

- Split files by responsibility, not by type kind — no `interfaces.go`, `types.go`, or `models.go` junk drawers. Each file holds one coherent concern, and stays under a few hundred lines before it is split.
- Domain types pair `<type>.go` with `<type>_test.go` (`widget.go` / `widget_test.go`).
- The canonical transport-package layout is the reference's: `server.go` (construction and route wiring), `handlers.go`, `middleware.go`, `errors.go` (status mapping), as [../reference/exampleservice/](../reference/exampleservice/) `internal/api/http` does. Cross-cutting concerns that outgrow `middleware.go` get their own file (`auth.go`, `idempotency.go`), same rule.

## Common Mistakes And Forbidden Patterns

- A helper package that becomes the real architecture.
- Interfaces created for every struct before a second implementation exists.
- Transport-layer DTOs leaking into domain packages.
- Circular imports solved by moving code into a junk drawer package.
- `init()` side effects for registration, configuration, or dependency wiring.
- Speculative generics: type parameters added before a second concrete type exists, or used to dress up a single concrete type.
- Functional options where a closed, required config struct would be simpler and clearer.
- Interfaces wider than a handful of methods, or defined next to the implementation instead of the consumer.

## Verification And Proof

- `go test ./...` should compile cleanly without import cycles.
- `go doc ./...` should show a public surface that looks intentional rather than accidental.
- For each type parameter, name the concrete duplication or `any`-and-assert code it removes; if you cannot, drop it.
- For each interface, count the methods (target 1-3) and confirm it sits in the consuming package.
- Review a proposed package by asking: what owns this behavior, who imports it, and what contract does that create?
