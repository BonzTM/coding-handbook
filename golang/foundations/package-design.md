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

### Export Policy

- Export the smallest API surface that serves real callers.
- Start unexported. Make things public only when another package truly needs them.
- Treat exported symbols as support commitments. The cost is future compatibility, not just one capital letter.

### Naming Rules

- Package names should be short, lower-case, and descriptive: `auth`, `config`, `orders`, `telemetry`.
- Avoid stutter: prefer `orders.Service`, not `orders.OrderService`.
- Avoid packages named after mechanics rather than purpose: `util`, `helpers`, `common`, `base`, and `misc` are red flags.

## Common Mistakes And Forbidden Patterns

- A helper package that becomes the real architecture.
- Interfaces created for every struct before a second implementation exists.
- Transport-layer DTOs leaking into domain packages.
- Circular imports solved by moving code into a junk drawer package.
- `init()` side effects for registration, configuration, or dependency wiring.

## Verification And Proof

- `go test ./...` should compile cleanly without import cycles.
- `go doc ./...` should show a public surface that looks intentional rather than accidental.
- Review a proposed package by asking: what owns this behavior, who imports it, and what contract does that create?
