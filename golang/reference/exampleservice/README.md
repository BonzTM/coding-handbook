# exampleservice

A small, complete HTTP service that is the **keystone reference** for the Go
engineering handbook. It manages a "widgets" feature end to end and exists to
prove the handbook's guidance and committed templates work on real code.

It is its own Go module (`github.com/example/exampleservice`) so editor tooling
resolves the internal imports cleanly when you open it in isolation. The
language baseline is **Go 1.24+**.

> **Note:** the module's `go.mod` `go` directive is `go 1.26.0`, not `1.24`.
> Pinning `golangci-lint` and `sqlc` as `go tool` dependencies pulls in a tool
> graph that requires Go 1.26 (the sqlc tool graph sets the floor), which raises
> the directive (`go mod tidy` re-asserts it). This is a real, instructive
> consequence of treating linters and generators as versioned tool dependencies;
> the 1.24+ baseline still describes the service code itself.

## What it is

- `net/http` + Go 1.22 `ServeMux` HTTP service, no web framework.
- An in-memory store by default; a runnable `database/sql` Postgres store that
  links the **pure-Go** pgx driver (blank import in `main`), so setting `DB_DSN`
  opens the pool, runs the embedded goose migrations (when
  `DB_MIGRATE_ON_STARTUP=true`), and serves from Postgres — all while the binary
  stays `CGO_ENABLED=0` static.
- A dedicated **audit logger** on its own stream (separate from the app log) that
  records security-relevant actions — authentication failure, authorization
  denial, and successful data-mutating writes — with who/what/when/where and no
  secrets or PII, per [operations/security.md](../../operations/security.md).
- Structured logging with `log/slog`, a readiness flag, real Prometheus metrics
  (`GET /metrics`), and config-gated OpenTelemetry tracing with W3C context
  propagation.
- Typed, validated, fail-fast configuration from flags and the environment.
- A thin `main` that wires lifecycle and performs ordered, bounded shutdown.

## Run it

```bash
make run            # go run ./cmd/exampleservice (in-memory store)
# or directly:
go run ./cmd/exampleservice

# Run against Postgres (pure-Go pgx driver; self-migrate on startup):
DB_DSN='postgres://user:pass@localhost:5432/exampleservice?sslmode=disable' \
  DB_MIGRATE_ON_STARTUP=true go run ./cmd/exampleservice

# One-shot migration mode: apply the embedded goose migrations against DB_DSN,
# then exit 0 (non-zero with a clear error on failure; DB_DSN is required).
# This is the production path for schema changes — a deployment runs the SAME
# image as a migration Job (args: ["-migrate"]) ahead of the rollout, so the
# startup self-migrate above stays a dev/CI convenience:
DB_DSN='postgres://user:pass@localhost:5432/exampleservice?sslmode=disable' \
  go run ./cmd/exampleservice -migrate
```

Then:

```bash
curl -s localhost:8080/livez
curl -s localhost:8080/readyz
curl -s localhost:8080/metrics            # Prometheus exposition
curl -s -XPOST localhost:8080/widgets -d '{"id":"w1","name":"Widget One"}'
curl -s localhost:8080/widgets/w1

# Keyset (cursor) pagination. The list envelope is {"items":[...],"next_cursor":"..."}.
# page_size is clamped server-side (default 20, max 100); next_cursor is "" on the
# last page. Pass it back verbatim to fetch the next page:
curl -s 'localhost:8080/widgets?page_size=2'
curl -s 'localhost:8080/widgets?page_size=2&cursor=<next_cursor-from-previous-response>'
```

By default the service runs in **local/dev mode** (`AUTH_ENABLED=false`): requests
need no token and act as a synthetic principal (tenant `local-dev`, reader+writer
roles), so the commands above work as-is. With `AUTH_ENABLED=true` every API
request must carry a Bearer JWT and the tenant/roles come from its claims:

```bash
curl -s -H 'Authorization: Bearer <jwt>' localhost:8080/widgets

# Idempotent create: retrying with the same Idempotency-Key replays the first
# response byte-for-byte instead of creating a second widget.
curl -s -XPOST localhost:8080/widgets \
  -H 'Authorization: Bearer <jwt>' \
  -H 'Idempotency-Key: 2f1c... (a client-generated unique key)' \
  -d '{"id":"w1","name":"Widget One"}'
```

Configuration keys are documented in [.env.example](.env.example). Precedence is
flags > environment > defaults.

## Verify

`make verify` is the single ordered safety gate (tidy, fmt-check, lint, vet,
test, race, vuln, build); humans and CI run the same target.

```bash
make verify
```

The `lint`, `fmt`, and `vuln` targets use `go tool` directives
(`go tool golangci-lint`, `go tool govulncheck`); add the tools with
`go get -tool ...` per
[golang/foundations/project-setup.md](../../foundations/project-setup.md).
Pinning the tools this way is what raises the `go.mod` directive to 1.26 (see
the note above). The core build/test loop needs no extra tools:

```bash
go build ./...
go test ./...
go test -race ./...
```

## Package map

Each package embodies a specific handbook doc:

| Package / file | Responsibility | Governing handbook doc |
|---|---|---|
| `cmd/exampleservice/main.go` | thin main: signal context, config load, slog, wiring, errgroup, ordered bounded shutdown; one-shot `-migrate` mode for migration Jobs | [foundations/context-and-concurrency.md](../../foundations/context-and-concurrency.md), [templates/cmd-app-main.go.txt](../../templates/cmd-app-main.go.txt) |
| `internal/config` | env+flags load, fail-fast `Validate`, no globals/init | [foundations/configuration.md](../../foundations/configuration.md) |
| `internal/core` | widgets domain service; defines the `Store` interface it consumes (interface-at-consumer); injected `Clock` | [foundations/package-design.md](../../foundations/package-design.md), [foundations/data-modeling.md](../../foundations/data-modeling.md), [foundations/time.md](../../foundations/time.md) |
| `internal/db/memory.go` | in-memory `Store` test/dev double (default + tests); same keyset-pagination contract as Postgres | [services/database.md](../../services/database.md) |
| `internal/db/postgres.go` | runnable `database/sql` repository (pure-Go pgx driver); explicit pool sizing; delegates SQL to the sqlc-generated layer; maps the typed `*pgconn.PgError` SQLSTATE `23505` to `core.ErrAlreadyExists` | [services/database.md](../../services/database.md) |
| `internal/db/migrations/*.sql` | goose-tagged (`-- +goose Up`/`Down`) schema migrations creating `widgets` (composite `(tenant_id, id)` key + `(tenant_id, created_at, id)` keyset index) and `idempotency_keys`, embedded via `//go:embed` | [services/database.md](../../services/database.md), [recipes/add-idempotent-write.md](../../recipes/add-idempotent-write.md) |
| `internal/db/migrate.go` | `embed.FS` + goose runner (`goose.SetBaseFS`/`UpContext`); config-gated by `DB_MIGRATE_ON_STARTUP` at startup and the engine behind the one-shot `-migrate` mode | [services/database.md](../../services/database.md) |
| `internal/db/queries.sql`, `sqlc.yaml`, `internal/db/sqlcgen` | sqlc source queries, config (v2, `database/sql`, postgresql), and committed generated code; regenerate with `go tool sqlc generate` | [services/database.md](../../services/database.md) |
| `internal/auth` | bearer-token (JWT) verification against a JWKS key source; pins iss/aud, allowlists RS*/ES* algorithms (rejects `alg=none`); `Verifier` seam with a JWKS impl + an injectable static impl for offline tests | [services/http-services.md](../../services/http-services.md), [operations/security.md](../../operations/security.md) |
| `internal/api/http` | transport adapter: server hardening, middleware order (recovery → request-id/trace → authn → logging/metrics → authz → idempotency → handler), decode→validate→core→map→encode, DTOs, error mapping; Idempotency-Key middleware; emits audit events on authn failure, authz denial, and create | [services/http-services.md](../../services/http-services.md), [foundations/serialization.md](../../foundations/serialization.md), [foundations/errors-and-logging.md](../../foundations/errors-and-logging.md), [operations/security.md](../../operations/security.md), [recipes/add-idempotent-write.md](../../recipes/add-idempotent-write.md) |
| `internal/telemetry` | `slog` logger construction, readiness flag, metrics seam (no-op + expvar + Prometheus), config-gated OTel tracer provider; dedicated `AuditLogger` (separate slog handler/sink) for who/what/when/result audit events | [operations/observability.md](../../operations/observability.md), [operations/security.md](../../operations/security.md), [foundations/errors-and-logging.md](../../foundations/errors-and-logging.md) |
| `internal/buildinfo` | `Name`/`Version`/`Commit` stamped via `-ldflags` | [foundations/project-setup.md](../../foundations/project-setup.md) |
| `Makefile`, `.golangci.yml`, `Dockerfile`, `.dockerignore`, `.env.example` | adapted copies of [golang/templates/](../../templates/) with the real module path | [quality/linting.md](../../quality/linting.md), [operations/deployment.md](../../operations/deployment.md) |

## Observability

The telemetry the handbook describes is real and demonstrated, per
[operations/observability.md](../../operations/observability.md).

- **Metrics (Prometheus)**: `telemetry.NewPromMetrics` owns a private registry
  (not the global default) and implements the same `telemetry.Metrics` seam the
  rest of the service consumes, so `NopMetrics`/`ExpvarMetrics` remain drop-in
  alternatives — swap the adapter in `main`, never the call sites. It records a
  request counter and a latency histogram with **low-cardinality** labels only
  (matched route pattern + status class), never request/user/tenant IDs or raw
  paths. `GET /metrics` is served by `promhttp` and mounted on the probe-side
  mux, ahead of the heavy middleware, so scrapes are neither access-logged nor
  counted as requests — the same probe split that protects `/livez` and
  `/readyz`. The endpoint also exposes the standard Go runtime and process
  collectors.
- **Tracing (OpenTelemetry)**: `telemetry.NewTracerProvider` builds an OTel
  `TracerProvider` with a batch span processor and an OTLP/HTTP exporter, gated
  on config: with no `OTLP_ENDPOINT` it installs a never-sampling provider so
  the service runs and tests pass offline; with one it batches and ships spans
  under a parent-based head sampler at `TRACE_SAMPLE_RATIO`. It sets the global
  W3C TraceContext + Baggage propagator so trace context crosses service
  boundaries. The HTTP server is instrumented with `otelhttp`, so every API
  request is a span (named `METHOD /route-pattern`, low cardinality); `/livez`,
  `/readyz`, and `/metrics` are deliberately not traced. The logging middleware
  pulls the active span context off the request and attaches `trace_id` and
  `span_id` to every access-log line so logs and traces join. Spans are flushed
  in the ordered shutdown as the final telemetry step.

Configure both via the `OTLP_*` / `TRACE_SAMPLE_RATIO` keys in
[.env.example](.env.example).

## Security & identity

The handbook defers the identity *scheme* to an ADR but ships ONE copyable
default here: a Bearer **JWT** validated against a **JWKS** key source, with
per-tenant scoping, route-level RBAC, per-resource ownership, and idempotent
unsafe writes. All libraries are pure-Go (`github.com/golang-jwt/jwt/v5`,
`github.com/MicahParks/keyfunc/v3`) and build with `CGO_ENABLED=0`.

- **Authentication (JWT/JWKS)**: the `authMiddleware` validates the
  `Authorization: Bearer <jwt>` token via `internal/auth`. The verifier pins the
  issuer and audience from config, requires an expiry, and **allowlists** the
  signature algorithms (`RS*`/`ES*`) so `alg=none` and HS/RS confusion attacks
  are rejected. On success it builds a typed `core.Principal`
  (subject, tenant, roles) and puts it on the request context via a typed key;
  any failure is a uniform **401** whose body never leaks which check failed
  (the detail stays in the boundary log) and emits an **audit event** (see
  Audit logging below). Auth is **config-gated**
  (`AUTH_ENABLED`): with it off the service boots offline in local/dev mode with
  a synthetic principal. The `Verifier` is a seam — production wires the
  JWKS-backed impl; tests sign with a local key and wire a static verifier, so
  unit tests never touch a real JWKS URL.
- **Authorization (RBAC + ownership)**: authorization is enforced **at the
  boundary**, not in helpers. A route-scoped `requireRole` check rejects a
  caller lacking the route's role with **403** (`POST /widgets` needs
  `widgets.writer`; the reads need `widgets.reader`), and the per-resource
  ownership check is the tenant-scoped store: a cross-tenant read is a **404**,
  never another tenant's data. The core service re-checks role and tenant as
  defense-in-depth. A role denial emits an **audit event** (see below).
- **Multi-tenancy**: the `tenant_id` is resolved from the principal (never the
  request body) and threaded in context. **Every** store query is scoped by
  `tenant_id` — the in-memory store keys on `(tenant_id, id)` and the SQL store
  filters on the `tenant_id` column with a composite primary key — so one tenant
  cannot observe another's rows. Cross-tenant isolation is proven by unit tests
  (and a live integration test).
- **Idempotency-Key** ([recipes/add-idempotent-write.md](../../recipes/add-idempotent-write.md)): `POST /widgets`
  **requires** an `Idempotency-Key` header scoped to **(tenant, route, key)**; a
  missing key is rejected with **400** (the recipe's recommended stance for
  resource creation). The first use processes the write and persists the response
  (status + body); a duplicate completed key **replays** that response
  byte-identically without re-running the side effect; an in-flight duplicate is
  **409**; the same key with a different request body is **422**; records are
  **TTL-bounded** (`IDEMPOTENCY_TTL`). All four rejection bodies use the
  structured error envelope with a machine-readable `code`.

  **Durability model (honest):** the shipped reference does
  **capture-and-replay**, not single-transaction atomicity. The middleware lets
  the handler commit the domain write, then calls the store to persist the
  captured response **after** the write has committed — both the in-memory store
  and the SQL store's `Complete` are a separate write, not part of the write's
  transaction. So a crash between the committed write and the persisted record
  re-executes on retry. The **production-grade** pattern that makes the response
  and the write commit atomically is a single `*sql.Tx` that claims the key,
  performs the write, and completes the record before `COMMIT` (the SQL path
  *should* adopt this; `sqlcgen.Queries.WithTx` exists for it). The in-memory
  store has no transaction and is documented as not single-transaction. See
  [recipes/add-idempotent-write.md](../../recipes/add-idempotent-write.md) (Steps 4 and the Atomic-commit invariant) and
  `internal/db/idempotency_postgres.go`.
- **Audit logging** ([operations/security.md](../../operations/security.md) ### Audit Logging): a **dedicated**
  `telemetry.AuditLogger` — a SEPARATE `log/slog` handler and sink, not the
  application logger — records the security-relevant actions: an authentication
  failure (`authMiddleware`), an authorization denial (`requireRole`), and a
  successful data-mutating write (create widget). Each record carries the fixed
  who/what/when/where schema — `actor` (`sub`) + `tenant`, `action`, `resource`,
  `result`, UTC `time`, and `request_id` — so a denial or failure is as traceable
  as a success. Audit records **never** contain secrets or PII payloads: the bare
  token is never logged, and the widget name (application payload) is excluded —
  the resource is referenced by id only. The reference routes the audit stream to
  **stderr** (the app log goes to stdout) so the two streams have independent
  retention and access controls; a deployment points stderr at the org's audit
  sink. Negative and positive tests in `internal/api/http/audit_test.go` assert
  the emitted fields for an authn failure, an authz denial, and a create, and that
  no token/payload leaks into the stream.

## Notes on deliberate stdlib-only choices

- **Database**: `internal/db/postgres.go` talks to Postgres through
  `database/sql`. `main` blank-imports the **pure-Go** pgx stdlib driver
  (`github.com/jackc/pgx/v5/stdlib`), so setting `DB_DSN` opens the pool via
  `db.OpenDB` and serves from Postgres while the binary stays `CGO_ENABLED=0`
  static; an empty `DB_DSN` keeps the offline in-memory store. The store maps a
  unique-violation to `core.ErrAlreadyExists` by matching the driver's typed
  `*pgconn.PgError` SQLSTATE `23505`, not an error-string. SQL is
  generated by sqlc (`go tool sqlc generate`, pinned as a module tool) from
  `internal/db/queries.sql` against the goose migration schema; the generated
  package `internal/db/sqlcgen` is committed. Migrations live in
  `internal/db/migrations`, are embedded via `//go:embed`, and apply on startup
  when `DB_MIGRATE_ON_STARTUP=true` (DB-backed builds only). The list endpoint
  uses keyset/cursor pagination over the stable `(created_at, id)` key with an
  opaque base64 `next_cursor`; the in-memory store honors the identical contract
  so offline tests cover it. The Postgres store and migrations are exercised by
  `//go:build integration` tests (`go test -tags=integration ./internal/db/...`
  with `TEST_DATABASE_DSN` set) that open the real store, migrate, round-trip a
  widget + the keyset list, and assert the unique-violation → `ErrAlreadyExists`
  mapping; they are kept out of the default offline `make verify`.
- **errgroup**: lifecycle orchestration uses `golang.org/x/sync/errgroup`,
  matching the canonical
  [templates/cmd-app-main.go.txt](../../templates/cmd-app-main.go.txt).
- **Runtime dependencies**: beyond `errgroup`, the service links the
  OpenTelemetry SDK + OTLP/HTTP trace exporter + `otelhttp`, the Prometheus
  client (`client_golang`), and the pgx stdlib driver (`jackc/pgx/v5`) for the
  Postgres store. All are pure-Go and build with `CGO_ENABLED=0`.
