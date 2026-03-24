# Framework Selection

Rules for deciding when a dependency earns its complexity cost.

## Default Approach

Start with the standard library and add third-party packages only when they clearly improve correctness, interoperability, or operator experience.

### Approval Questions

Before adding a dependency, answer all of these:

1. What concrete problem does the stdlib or current stack fail to solve well enough?
2. What maintenance, upgrade, and security cost does this add?
3. Does the package introduce hidden magic, global state, or framework lock-in?
4. Is it widely used, actively maintained, and easy to replace later if needed?

## Default Choices By Concern

| Concern | Default | Acceptable escalation | Avoid by default |
|---|---|---|---|
| HTTP routing | `net/http` with `ServeMux` | `chi` for more complex routing/middleware shape | framework-first stacks that hide stdlib handlers |
| CLI | stdlib `flag` | `cobra` for real subcommand trees and shell completion | `viper`-driven global config magic |
| config loading | explicit env plus flags in `internal/config` | a small parsing helper if it stays explicit | global config frameworks with implicit precedence |
| logging | `log/slog` | thin adapters only when the sink requires them | bespoke logging frameworks |
| metrics | Prometheus client | org-mandated backend SDKs | high-level wrappers that hide metric names and labels |
| tracing | OpenTelemetry | none if the repo is local-only and simple | ad hoc trace systems |
| persistence | `database/sql`, then `sqlc` | small query builders when they stay transparent | ORMs as the day-one default |
| messaging | broker-specific client only after contract, ordering, and retry needs are clear | thin clients or libraries that do not hide delivery semantics | frameworks that obscure ack, retry, DLQ, or partition behavior |
| testing helpers | stdlib `testing` | `go-cmp`, `testify/require`, `goleak` where they clearly improve signal | assertion DSLs that obscure behavior |
| release automation | simple scripts or CI | GoReleaser when matrix packaging becomes real work | heavyweight tooling nobody on the team understands |

## Hard Warnings

- No committed `replace` directives for production builds.
- No dependency added only because it is familiar from another language ecosystem.
- No ORM, DI container, or web framework just to avoid writing explicit Go code.
- No tool dependency in runtime code when it belongs in `tools.go`.
- No messaging library adopted before the repo documents idempotency, ordering, retry, and DLQ expectations.

## Decision Record

When a repo chooses an exception, write down:

- the package name and why the default was insufficient
- which repo area is allowed to depend on it
- the operational risk or lock-in tradeoff accepted
- what would trigger re-evaluation or removal later
