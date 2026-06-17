# Maintainer Reference

Purpose: hold slower-path architecture, package-map, lifecycle, and rationale guidance that is useful but not worth loading for every task.
Audience: maintainers and agents working in Go repositories that use this handbook.
Read [AGENTS.md](AGENTS.md) first. Use this file when you need the fuller background behind the fast-path rules.

## Architecture Snapshot

This handbook assumes a single Go module with a small public surface and a larger private implementation under `internal/`. The dominant shape is:

```text
repo/
  go.mod
  go.sum
  cmd/
    app/
      main.go
  internal/
    api/
      http/
      grpc/
    core/
    db/
    config/
    telemetry/
    runtime/
    testutil/
  api/
```

This follows the official module-layout guidance from `go.dev/doc/modules/layout`. Community project-layout repos may offer ideas, but they are not the primary authority and they should not override the simpler `cmd/` plus `internal/` baseline unless a real codebase need appears. A complete, compiling instance of this architecture lives at [reference/exampleservice/](reference/exampleservice/) (`make verify`-green); read it alongside this map to see the boundaries embodied in real code.

Three compiling reference modules under [reference/](reference/) embody this architecture for the main service shapes — [exampleservice](reference/exampleservice/) (HTTP+Postgres), [examplegrpc](reference/examplegrpc/) (gRPC), and [exampleworker](reference/exampleworker/) (event-driven worker) — and are all `make verify`-green.

## Two-Speed Documentation Model

- Fast path: [AGENTS.md](AGENTS.md) for invariants, task loop, and baseline proof.
- Routing path: [maintainer-map.md](maintainer-map.md) for change type to file set mapping.
- Slow path: this file for architecture, package map, test taxonomy, lifecycle, and rationale.

Use the fast path for most tasks. Use this file when a change crosses layers, introduces new runtime behavior, or challenges an existing default.

## Package Map

| Package Area | Owns | Must Not Own |
|---|---|---|
| `cmd/<app>` | startup, flags, config wiring, dependency injection, signal handling, process exit | business rules, raw SQL, request validation details |
| `internal/core` | domain behavior, orchestration, interfaces consumed from the outside | HTTP types, gRPC transport details, SQL text |
| `internal/api/http` | handlers, encoding/decoding, HTTP status mapping, middleware composition | business rules, schema migrations |
| `internal/api/grpc` | proto-to-core mapping, status codes, interceptors | SQL, domain state mutation outside core |
| `internal/db` | SQL, repositories, transaction helpers, migrations, database-specific mapping | HTTP or gRPC concerns |
| `internal/httputil` | JSON helpers, error-response helpers, request-size and timeout helpers shared across transport adapters | domain/business rules, handler-specific logic, a dumping ground for unrelated helpers |
| `internal/buildinfo` | version, commit, and build-time metadata stamped via `-ldflags` and surfaced in logs and `version` output | runtime behavior, config loading, anything beyond build provenance |
| `internal/config` | env and flag loading, defaults, validation, startup errors | lazy runtime lookups spread through the codebase |
| `internal/telemetry` | logger setup, metrics registration, tracing helpers, health/readiness primitives | domain decisions about what a request means |
| `internal/runtime` | application assembly helpers that keep `main` thin | transport or persistence logic |
| `internal/testutil` | reusable test harnesses, fixtures, helpers that improve tests without hiding behavior | production behavior |

Repos with significant async work often add an `internal/messaging` or ownership-specific package for producers, consumers, relay jobs, and settlement logic. The same boundary rules apply: business behavior stays in core packages, while broker and delivery mechanics stay in adapters.

## Lifecycle Model

For services and workers, the normal process lifecycle is:

1. Parse flags and load env-driven config.
2. Validate config and fail fast before opening listeners or background loops.
3. Construct logger, metrics registry, tracer provider, DB pools, and external clients.
4. Wire core services and adapters.
5. Start servers or workers under a root context created with `signal.NotifyContext`.
6. On cancellation, stop accepting new work, drain in-flight work within a bounded timeout, and close shared resources.

If a repository shape does not fit this lifecycle, it should document the exception explicitly.

## Test Taxonomy

| Test Type | Default Location | What It Proves |
|---|---|---|
| unit tests | next to the package | package-local business rules and edge cases |
| handler or transport tests | next to `internal/api/http` or `internal/api/grpc` | request decoding, status mapping, middleware behavior |
| repository integration tests | next to `internal/db` or in a storage integration package | real SQL, transactions, and migration compatibility |
| external client tests | client package with `httptest.Server` or protocol-specific test server | request construction, timeout handling, response mapping |
| fuzz tests | same package as parser or input boundary | panic resistance and malformed-input handling |
| benchmarks | same package as hot path | allocation and throughput characteristics |

The important principle is not "more tests". It is "the right tests at the right boundary". A mocked repository test does not replace a real migration or transaction test.

## Runtime Contracts Worth Remembering

- Every goroutine must have an owner, a stop condition, and a proof story.
- Every external call must take a context and a timeout budget.
- Every network-facing component should have a clear readiness story distinct from plain liveness.
- Every non-trivial feature should add telemetry where operators will actually need it.
- Every dependency added today becomes part of tomorrow's debugging and patch surface.

## Contract Surfaces

- HTTP and gRPC boundaries should have an obvious source of truth for payload shapes and error semantics.
- Database schema, migration order, and compatibility expectations are data contracts, not incidental implementation details.
- Event payloads should have explicit envelopes, versioning rules, and idempotency expectations when they cross process boundaries.
- Generated code is never the only contract source; the source schema or protocol definition remains authoritative.

Event delivery rules are operational contracts too: whether delivery is at-least-once, what ordering is guaranteed, when retries stop, and what happens at dead-letter boundaries should be written down before a queue-backed feature is considered done.

## Dependency Rationale

- Stdlib-first keeps onboarding, debugging, and long-term maintenance cheaper.
- `database/sql` plus `sqlc` keeps SQL visible while reducing scan boilerplate.
- `log/slog` is the default because it is standard, structured, and easy to adapt.
- Prometheus and OpenTelemetry are acceptable because they are de facto interoperability standards, not framework lock-in.
- ORMs, reflection-heavy DI, and framework-first HTTP stacks should be treated as exceptions that need evidence, not preferences.

## Common Failure Modes

| Symptom | Likely Cause | First Fix |
|---|---|---|
| handlers know too much about storage | business rules leaked out of `internal/core` | move orchestration into core service methods |
| `main.go` grows with every feature | startup and domain logic are mixed | extract runtime assembly helpers or service constructors |
| tests pass but deploys fail | no real integration coverage for DB or external boundaries | add real boundary tests and startup smoke checks |
| workers leak on shutdown | no owner or no cancellation propagation | root context plus bounded shutdown path |
| metrics backend churns or explodes | high-cardinality labels | collapse labels to stable, finite values |
| config bugs show up in production only | lazy validation or hidden defaults | validate every required field at startup |

## Primary Sources Behind These Defaults

- official module layout: `https://go.dev/doc/modules/layout`
- module semantics and versioning: `https://go.dev/ref/mod`
- toolchain guidance: `https://go.dev/doc/toolchain`
- error wrapping and inspection: `https://go.dev/blog/go1.13-errors`
- context rules: `https://pkg.go.dev/context`
- memory model: `https://go.dev/ref/mem`
- standard testing package: `https://pkg.go.dev/testing`
- fuzzing and security guidance: `https://go.dev/doc/security/fuzz/`, `https://go.dev/doc/security/best-practices`
- structured logging: `https://pkg.go.dev/log/slog`, `https://go.dev/blog/slog`

## Related Docs

- Fast path: [AGENTS.md](AGENTS.md)
- Change routing: [maintainer-map.md](maintainer-map.md)
- Project layout: [foundations/project-setup.md](foundations/project-setup.md)
- Package boundaries: [foundations/package-design.md](foundations/package-design.md)
- Contracts and compatibility: [foundations/contracts-and-compatibility.md](foundations/contracts-and-compatibility.md)
- Event delivery guidance: [services/eventing-and-messaging.md](services/eventing-and-messaging.md)
- Proof and testing: [quality/testing.md](quality/testing.md)
