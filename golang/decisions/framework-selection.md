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
| gRPC / RPC framework | `google.golang.org/grpc` (grpc-go) with `buf` for codegen | `connectrpc.com/connect` (connect-go) for HTTP/1.1 + gRPC + gRPC-Web browser-friendly endpoints | hand-rolled RPC, or gateway/proxy sprawl before it is needed |
| request validation | explicit validation in the handler/core after decode (see [../foundations/serialization.md](../foundations/serialization.md), [../foundations/data-modeling.md](../foundations/data-modeling.md)) — no library | `github.com/go-playground/validator/v10` for large struct-tag-driven validation | reflection-heavy validation frameworks as the day-one default |
| CORS | none — service-to-service APIs need no CORS layer (see [../services/http-services.md](../services/http-services.md) ### CORS) | `github.com/rs/cors` (pure Go), or a small hand-rolled handler for a trivial static allowlist, when spec intake identifies browser clients | wildcard origin combined with credentials; middleware that reflects arbitrary `Origin` values |
| HTML templating | `html/template` parsed once from `embed.FS` (see [../services/web-apps.md](../services/web-apps.md)) | a compile-time-checked template generator (e.g. `a-h/templ`, ADR-level) when template volume and type-safety pressure outgrow stdlib | `text/template` for browser output; string-built HTML |
| sessions | server-side sessions via `github.com/alexedwards/scs/v2` with a store the repo already runs (Postgres) | stateless signed cookies only for tiny, non-revocable UI state (ADR-level) | hand-rolled session/cookie crypto; JWTs in cookies as a session substitute |
| CSRF | stdlib `net/http.CrossOriginProtection` (Go 1.25+) wrapped around the mux (see [../services/web-apps.md](../services/web-apps.md)) | token-based CSRF middleware (ADR-level) only when clients send neither `Sec-Fetch-Site` nor `Origin` | disabling protection app-wide to accommodate one endpoint; CSRF-exempting state-changing GETs instead of removing them |
| CLI | stdlib `flag` | `cobra` for real subcommand trees and shell completion | `viper`-driven global config magic |
| config loading | explicit env plus flags in `internal/config` | a small parsing helper if it stays explicit | global config frameworks with implicit precedence |
| logging | `log/slog` | thin adapters only when the sink requires them | bespoke logging frameworks |
| metrics | Prometheus client | org-mandated backend SDKs | high-level wrappers that hide metric names and labels |
| tracing | OpenTelemetry | none if the repo is local-only and simple | ad hoc trace systems |
| persistence | `database/sql`, then `sqlc` | small query builders when they stay transparent | ORMs as the day-one default |
| schema migrations | `goose` (SQL-first, embeddable via `embed.FS`, runs from code or CLI; pairs with `database/sql` + `sqlc`) | `golang-migrate` for many drivers, `atlas` for declarative/diff-based schemas (ADR-level) | hand-applied SQL with no migration tool or version table |
| money / decimal | integer minor units (`int64` cents plus an explicit currency code) in domain types (see [../foundations/serialization.md](../foundations/serialization.md) ### Numeric Precision) — no library | `github.com/shopspring/decimal` (ADR-level) only for genuine arbitrary-precision arithmetic such as rates, allocations, or compounding | `float64` for money anywhere; a decimal library imported ad hoc without an ADR |
| retries / backoff | hand-rolled bounded exponential backoff with full jitter behind an injectable clock/sleep seam, per [../operations/resilience.md](../operations/resilience.md) — the exemplar is [../reference/exampleworker/](../reference/exampleworker/) `internal/messaging/backoff.go`; no library | a retry library (ADR-level) only when policy complexity genuinely outgrows the hand-rolled loop | unbounded or zero-jitter retry loops; resilience frameworks that hide attempt state |
| circuit breaker | none — timeouts plus bounded retries first, per [../operations/resilience.md](../operations/resilience.md) | `github.com/sony/gobreaker` (ADR-level) when a dependency's failure mode demands one; it must expose state transitions and trip metrics | a breaker on every client by reflex; libraries that hide breaker state |
| job scheduling | stdlib `time.Ticker` for fixed intervals | a cron-expression library (e.g. `robfig/cron`) only for calendar schedules; advisory-lock or leader election for multi-replica | unmanaged goroutines with `time.Sleep` loops and no overlap guard |
| API deprecation signaling | `Sunset` header (RFC 8594) plus a documented `Deprecation` header form, recorded in an ADR | an org-standard deprecation registry or policy | removing a contract with no deprecation signal or window |
| in-process caching | a bounded LRU/TTL cache (e.g. `hashicorp/golang-lru/v2`) plus `golang.org/x/sync/singleflight` to collapse duplicate loads | an external cache (Redis/memcached) only when the working set or cross-instance sharing demands it | unbounded in-memory maps used as caches |
| feature flags | static typed config in `internal/config` | a typed accessor over an atomic snapshot for runtime toggles | a managed flag/experimentation service before targeting genuinely needs it; scattered raw env lookups; long-lived flags left as debt |
| messaging | broker-specific client only after contract, ordering, and retry needs are clear | thin clients or libraries that do not hide delivery semantics | frameworks that obscure ack, retry, DLQ, or partition behavior |
| testing helpers | stdlib `testing` | `go-cmp`, `testify/require`, `goleak` where they clearly improve signal | assertion DSLs that obscure behavior |
| test doubles | hand-rolled fakes at consumer-defined seams (see [../quality/testing.md](../quality/testing.md) ### Test Doubles) | a mock framework (`mockgen`/`gomock`, `moq`) via ADR when interface churn makes fakes unmaintainable | mock frameworks as the default; over-specified call expectations |
| benchmark comparison | `golang.org/x/perf/cmd/benchstat` over repeated `-bench` runs | — | single-run before/after deltas presented as proof |
| release automation | simple scripts or CI | GoReleaser when matrix packaging becomes real work | heavyweight tooling nobody on the team understands |
| binary linkage | pure-Go / `CGO_ENABLED=0` static | cgo ONLY with an ADR, after ruling out a pure-Go alternative | cgo pulled in transitively unnoticed |
| secrets manager | injected env vars or mounted files from an external manager (the app reads injected material at startup) | Vault, a cloud KMS / Secrets Manager, or sealed-secrets when the platform provides one | embedding plaintext in source/image/build args, or the app fetching and caching long-lived plaintext itself |
| audit / log sink | structured `log/slog` to a dedicated audit stream the platform collects | a SIEM, managed audit service, or append-only store when compliance requires tamper-evidence | mixing audit events into the shared application log; no retention or access control on the sink |

## Mandated Frameworks

Sometimes the spec or the requester mandates a framework this table would not choose — Gin, Echo, Fiber, an org-standard stack. A mandate is honored, not fought, and not silently absorbed:

- Record an ADR stating the framework was **mandated by the requester**, which repo area may depend on it, and what the handbook default would have been. The Approval Questions still get written answers; "mandated" answers question 1.
- The framework-independent invariants survive unchanged: `cmd/`+`internal/` layout, thin `main`, `ctx` as first parameter, `%w` error wrapping, `log/slog` (through an adapter if the framework insists on its own logger), the structured error envelope, the testing bar, and the full `make verify` gate.
- Confine framework types to the transport package (`internal/api/http`). Handlers translate the framework's context into plain arguments for `internal/core`; core never imports the framework. This keeps the mandate reversible and the domain testable without it.
- Flag `net/http`-incompatible frameworks explicitly in the ADR. Fiber is `fasthttp`-based: stdlib middleware, `http.Handler` composition, and `httptest` do not apply, and the handbook's HTTP guidance holds only in spirit. That cost belongs in writing before code starts.
- A mandate covers the named framework only — it is not a license to relax the rest of this table.

## Common Mistakes And Forbidden Patterns

- No committed `replace` directives for production builds.
- No dependency added only because it is familiar from another language ecosystem.
- No ORM, DI container, or web framework just to avoid writing explicit Go code.
- No tool dependency in runtime code when it belongs in a `go.mod` `tool` directive (managed with `go get -tool` and run with `go tool`).
- No messaging library adopted before the repo documents idempotency, ordering, retry, and DLQ expectations.
- No dependency added without the approval questions answered in writing and the `go.mod`/`go.sum` diff understood line by line.
- No adopting a cgo-only library when a pure-Go one exists (e.g. `modernc.org/sqlite` over `mattn/go-sqlite3`); cgo forfeits `CGO_ENABLED=0` static builds and needs an ADR.
- No returning a bare `{"error":"..."}` string for validation failures; emit a structured field-error envelope so clients can map errors to fields (see [../foundations/serialization.md](../foundations/serialization.md) ### Error Responses).
- No exception to a default in this doc without an ADR recorded.

## Verification And Proof

A dependency choice is proven, not asserted. Before a dependency lands, demonstrate all of:

- The Approval Questions above are answered in writing, in the PR description or the ADR — not left implicit.
- The `go.mod` and `go.sum` diff is reviewed and understood: every added direct and indirect module is accounted for, and the size of the transitive blast radius is acceptable.
- `go tool govulncheck ./...` is clean against the new dependency set (this is part of `make verify`).
- An ADR is recorded for any choice that departs from the Default Choices By Concern table, cross-linking [architecture-decision-records.md](architecture-decision-records.md).

### Decision Record

When a repo chooses an exception, the ADR (see [architecture-decision-records.md](architecture-decision-records.md)) must write down:

- the package name and why the default was insufficient
- which repo area is allowed to depend on it
- the operational risk or lock-in tradeoff accepted
- what would trigger re-evaluation or removal later
