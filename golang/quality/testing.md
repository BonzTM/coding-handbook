# Testing

Testing strategy for Go repos that need trustworthy behavior, not just green checkmarks.

## Default Approach

Use the stdlib `testing` package first. Keep tests close to the code they prove.

### Test Taxonomy

| Test Type | Use for | Default tools |
|---|---|---|
| unit | pure business logic, edge cases, small adapters | `testing`, `t.Run` |
| transport | handler decoding, status mapping, middleware | `httptest`, fake services |
| integration | database, migrations, external clients, multi-package behavior | real DB or protocol test server |
| contract | payload compatibility, schema validation, event examples | schema fixtures, golden examples, generated validators when used |
| fuzz | parsers, decoders, protocol edges, untrusted input | `testing.F` |
| benchmark | hot paths, serialization, allocation-sensitive code | `testing.B` |

### Table-Driven Tests

Use table-driven tests when the behavior is truly the same shape repeated across inputs. Do not force everything into a table if branching setup becomes harder to read than separate tests.

### Integration Defaults

- Database code should get real database tests.
- External HTTP clients should usually use `httptest.Server` before they use mocks.
- Contract-heavy systems may need a smaller unit suite plus one or two high-value end-to-end tests.
- Event-driven systems should add duplicate-delivery, replay, retry-exhaustion, and ordering tests where those semantics matter.

### TDD Guidance

When practical, start behavior changes with the proving test first. The goal is not ritual; it is making the contract explicit before the implementation hardens.

### Eventing-Specific Proof

- Contract tests prove payload shape, metadata, and compatibility expectations.
- Duplicate-delivery tests prove idempotency.
- Replay and out-of-order tests prove the real ordering contract instead of an assumed one.
- DLQ or parked-message tests prove terminal failures stop retrying and preserve operator context.

## Common Mistakes And Forbidden Patterns

- Mocking repositories, HTTP clients, and workers so aggressively that no boundary behavior is exercised.
- Mocking brokers so heavily that replay, settlement, duplicate delivery, or exhaustion behavior is never exercised.
- Treating coverage percentage as the goal instead of edge-case confidence.
- Huge shared fixtures that make failures opaque.
- Benchmarks without stable inputs or with hidden setup inside the hot loop.
- Fuzz targets that mutate shared state or depend on external side effects.

## Verification And Proof

```bash
go test ./...
go test -race ./...
```

Add the following when relevant:

- fuzz targets for parsing or untrusted input paths
- event handlers and decoders that accept untrusted payloads are good fuzz targets when the decode path is non-trivial
- integration suites against a real database or test server
- `GOCOVERDIR` plus `go tool covdata` when integration coverage matters
- benchmarks compared with a stable baseline when performance is part of the change
- contract validation plus replay, duplicate-delivery, and DLQ tests when the repo publishes or consumes messages

Testing is done when the chosen proof matches the risk of the change, not when a single unit suite turns green.
