# exampleservice

A small, complete HTTP service that is the **keystone reference** for the Go
engineering handbook. It manages a "widgets" feature end to end and exists to
prove the handbook's guidance and committed templates work on real code.

It is its own Go module (`github.com/example/exampleservice`) so editor tooling
resolves the internal imports cleanly when you open it in isolation. The
language baseline is **Go 1.24+**.

> **Note:** the module's `go.mod` `go` directive is `go 1.25.0`, not `1.24`.
> Pinning `golangci-lint` as a `go tool` dependency pulls in a tool graph that
> requires Go 1.25, which raises the directive (`go mod tidy` re-asserts it).
> This is a real, instructive consequence of treating linters as versioned tool
> dependencies; the 1.24+ baseline still describes the service code itself.

## What it is

- `net/http` + Go 1.22 `ServeMux` HTTP service, no web framework.
- An in-memory store by default; a `database/sql` reference store that compiles
  with the standard library only (no driver imported) and documents driver
  wiring per the database doc.
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
`go get -tool ...` per `golang/foundations/project-setup.md`. Pinning
`golangci-lint` this way is what raises the `go.mod` directive to 1.25 (see the
note above). The core build/test loop needs no extra tools:

```bash
go build ./...
go test ./...
go test -race ./...
```

## Package map

Each package embodies a specific handbook doc:

| Package / file | Responsibility | Governing handbook doc |
|---|---|---|
| `cmd/exampleservice/main.go` | thin main: signal context, config load, slog, wiring, errgroup, ordered bounded shutdown | `foundations/context-and-concurrency.md`, `templates/cmd-app-main.go.txt` |
| `internal/config` | env+flags load, fail-fast `Validate`, no globals/init | `foundations/configuration.md` |
| `internal/core` | widgets domain service; defines the `Store` interface it consumes (interface-at-consumer); injected `Clock` | `foundations/package-design.md`, `foundations/data-modeling.md`, `foundations/time.md` |
| `internal/db/memory.go` | in-memory `Store` test/dev double (default + tests); same keyset-pagination contract as Postgres | `services/database.md` |
| `internal/db/postgres.go` | real `database/sql` repository; explicit pool sizing; delegates SQL to the sqlc-generated layer; maps driver errors to the core sentinels | `services/database.md` |
| `internal/db/migrations/*.sql` | goose-tagged (`-- +goose Up`/`Down`) schema migrations creating `widgets` (composite `(tenant_id, id)` key + `(tenant_id, created_at, id)` keyset index) and `idempotency_keys`, embedded via `//go:embed` | `services/database.md`, `recipes/add-idempotent-write.md` |
| `internal/db/migrate.go` | `embed.FS` + goose runner (`goose.SetBaseFS`/`UpContext`); config-gated by `DB_MIGRATE_ON_STARTUP` | `services/database.md` |
| `internal/db/queries.sql`, `sqlc.yaml`, `internal/db/sqlcgen` | sqlc source queries, config (v2, `database/sql`, postgresql), and committed generated code; regenerate with `go tool sqlc generate` | `services/database.md` |
| `internal/auth` | bearer-token (JWT) verification against a JWKS key source; pins iss/aud, allowlists RS*/ES* algorithms (rejects `alg=none`); `Verifier` seam with a JWKS impl + an injectable static impl for offline tests | `services/http-services.md`, `services/security.md` |
| `internal/api/http` | transport adapter: server hardening, middleware order (recovery → request-id/trace → authn → logging/metrics → authz → idempotency → handler), decode→validate→core→map→encode, DTOs, error mapping; Idempotency-Key middleware | `services/http-services.md`, `foundations/serialization.md`, `foundations/errors-and-logging.md`, `recipes/add-idempotent-write.md` |
| `internal/telemetry` | `slog` logger construction, readiness flag, metrics seam (no-op + expvar + Prometheus), config-gated OTel tracer provider | `operations/observability.md`, `foundations/errors-and-logging.md` |
| `internal/buildinfo` | `Name`/`Version`/`Commit` stamped via `-ldflags` | `foundations/project-setup.md` |
| `Makefile`, `.golangci.yml`, `Dockerfile`, `.dockerignore`, `.env.example` | adapted copies of `golang/templates/*` with the real module path | `quality/linting.md`, `operations/deployment.md` |

## Observability

The telemetry the handbook describes is real and demonstrated, per
`operations/observability.md`.

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
  (the detail stays in the boundary log). Auth is **config-gated**
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
  defense-in-depth.
- **Multi-tenancy**: the `tenant_id` is resolved from the principal (never the
  request body) and threaded in context. **Every** store query is scoped by
  `tenant_id` — the in-memory store keys on `(tenant_id, id)` and the SQL store
  filters on the `tenant_id` column with a composite primary key — so one tenant
  cannot observe another's rows. Cross-tenant isolation is proven by unit tests
  (and a live integration test).
- **Idempotency-Key** (`recipes/add-idempotent-write.md`): `POST /widgets`
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
  `recipes/add-idempotent-write.md` (Steps 4 and the Atomic-commit invariant) and
  `internal/db/idempotency_postgres.go`.

## Notes on deliberate stdlib-only choices

- **Database**: `internal/db/postgres.go` compiles against `database/sql`
  with no driver linked in the default build. To run it, blank-import a driver in
  `main` and pass its name to `db.OpenDB`; see `main.go`'s wiring comment. SQL is
  generated by sqlc (`go tool sqlc generate`, pinned as a module tool) from
  `internal/db/queries.sql` against the goose migration schema; the generated
  package `internal/db/sqlcgen` is committed. Migrations live in
  `internal/db/migrations`, are embedded via `//go:embed`, and apply on startup
  when `DB_MIGRATE_ON_STARTUP=true` (DB-backed builds only). The list endpoint
  uses keyset/cursor pagination over the stable `(created_at, id)` key with an
  opaque base64 `next_cursor`; the in-memory store honors the identical contract
  so offline tests cover it. The Postgres store and migrations are exercised by
  `//go:build integration` tests (`go test -tags=integration ./internal/db/...`
  with `TEST_DATABASE_DSN` set), kept out of the default offline `make verify`.
- **errgroup**: lifecycle orchestration uses `golang.org/x/sync/errgroup`,
  matching the canonical `templates/cmd-app-main.go.txt`.
- **Runtime dependencies**: beyond `errgroup`, the service links the
  OpenTelemetry SDK + OTLP/HTTP trace exporter + `otelhttp`, and the Prometheus
  client (`client_golang`). All are pure-Go and build with `CGO_ENABLED=0`.
