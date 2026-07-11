# exampleservice

A small, complete HTTP + PostgreSQL service that is the **keystone reference**
for the C# engineering handbook. It manages an "orders" resource end to end and
exists to prove the handbook's guidance and committed templates work on real
code. The language baseline is **.NET 10 / C# 14** (`global.json` pins the SDK).

> **Note:** the root files (`global.json`, `Directory.Build.props`,
> `Directory.Packages.props`, `nuget.config`, `.editorconfig`, `verify.ps1`,
> `Makefile`, `Dockerfile`, `.dockerignore`, `docker-compose.yml`,
> `.gitattributes`, `.gitignore`) are copies of [templates/](../../templates/).
> Building this module surfaced a handful of real template gaps (a
> `GenerateDocumentationFile` requirement for IDE0005, `.editorconfig`
> overrides for CA1032/CA1034/CA1515 that the handbook's own patterns require,
> and a `Set-StrictMode` array pitfall in `verify.ps1`); the fixes live in this
> module's copies, commented inline, and are flagged for upstreaming.

## What it is

- ASP.NET Core **Minimal APIs** with `TypedResults`, one endpoint group
  (`/orders`), route-group endpoint filters, and a thin `Program.cs`
  composition root — the only file where `Orders.Api` touches
  `Orders.Infrastructure`.
- Three projects with compiler-enforced boundaries:
  `Orders.Core` (domain: aggregate, typed IDs, ports, business rules — no
  ASP.NET, no EF Core), `Orders.Infrastructure` (EF Core + Npgsql, migrations,
  repositories, idempotency runner), `Orders.Api` (host, endpoints, wire DTOs,
  middleware, telemetry).
- **EF Core migrations** applied by an explicit step
  (`dotnet Orders.Api.dll --migrate`), never on normal startup; the compose
  file runs the same image as a one-shot migrate job before the service starts.
- **JWT bearer authentication against the issuer's JWKS** (OIDC discovery,
  pinned issuer/audience, algorithm allowlist), deny-by-default authorization
  (fallback policy + `orders.reader` / `orders.writer` role policies), and
  tenant scoping from the token's claims — a cross-tenant read is a 404, never
  another tenant's data. `Auth:Enabled=false` (the committed local default)
  swaps in a synthetic local-dev principal so the service boots offline.
- **Idempotency-Key on POST /orders** with single-transaction durability: the
  claim is a unique-constraint insert, and the domain write plus the captured
  response commit in ONE database transaction, so a replay can only exist if
  the write committed. Byte-identical replay, 409 for in-flight duplicates,
  422 for key reuse with a different body, TTL-bounded records.
- **RFC 9457 ProblemDetails** as the only wire error shape: `requestId` on
  every problem body, stable `type` URIs per domain error, field-level `errors`
  from the built-in minimal-API validation (`AddValidation`).
- A dedicated **audit logger** on its own `Orders.Audit` category recording
  authentication failures, authorization denials, and successful mutations —
  who/what/when/result/requestId, never tokens or payloads.
- **OpenTelemetry** traces, metrics, and logs via one `AddServiceTelemetry()`
  extension, OTLP export; `/livez` + `/readyz` (readiness checks the database);
  request-ID middleware outermost; per-client rate limiting answering
  429 + `Retry-After`; request timeouts and a body-size cap from config.
- Keyset (cursor) pagination over the stable `(CreatedAt, Id)` sort key using
  PostgreSQL row-value comparison; opaque URL-safe cursors; page size clamped
  server-side.
- Optimistic concurrency on PostgreSQL's `xmin` system column surfaced as a
  `version` the client echoes back; a stale version is a 409, never a silent
  lost update. Unique violations (SQLSTATE 23505) are matched on constraint
  name and translated to typed domain exceptions.

## Run it

```bash
# Full local stack: Postgres -> one-shot migrate -> service on :8080
docker compose up --build

# Or on the host against the compose Postgres (start it first):
docker compose up -d postgres
export DOTNET_ROOT="$HOME/.dotnet"; export PATH="$HOME/.dotnet:$PATH"
dotnet run --project src/Orders.Api -- --migrate   # apply schema, exit
dotnet run --project src/Orders.Api                # serve
```

Then (local-dev mode — `Auth:Enabled=false` — needs no token):

```bash
curl -s localhost:8080/livez
curl -s localhost:8080/readyz

# Idempotent create: retrying with the same Idempotency-Key replays the first
# response byte-for-byte (look for the Idempotency-Replayed: true header).
curl -s -XPOST localhost:8080/orders \
  -H 'Content-Type: application/json' \
  -H "Idempotency-Key: $(uuidgen)" \
  -d '{"externalReference":"ord-1001","customerId":"cust-42","quantity":3}'

curl -s localhost:8080/orders/<orderId-from-create>

# Keyset pagination: {"items":[...],"nextCursor":"..."} - nextCursor is null on
# the last page; pass it back verbatim:
curl -s 'localhost:8080/orders?pageSize=2'
curl -s 'localhost:8080/orders?pageSize=2&cursor=<nextCursor>'

# Optimistic concurrency: echo the version you read; a stale one is a 409.
curl -s -XPUT localhost:8080/orders/<orderId> \
  -H 'Content-Type: application/json' \
  -d '{"quantity":5,"status":"Confirmed","version":<version-from-read>}'
```

With `Auth__Enabled=true` (plus `Auth__Authority` and `Auth__Audience`) every
request must carry `Authorization: Bearer <jwt>`; roles and the tenant come
from the token. Configuration keys are documented in
[appsettings.json](src/Orders.Api/appsettings.json) (committed, secret-free —
it doubles as the `.env.example`); environment variables override it
(`Section__Key`).

## Verify

`pwsh ./verify.ps1` is the single ordered gate — restore (locked),
format-check, build (warnings-as-errors), test, audit. Humans, the Makefile
shim, and CI run the same script.

```bash
pwsh ./verify.ps1                # offline gate (unit tests only)
pwsh ./verify.ps1 -Integration   # + Testcontainers PostgreSQL suite (needs Docker)
```

The integration suite starts a disposable PostgreSQL, applies the committed
migrations, and proves the things the in-memory fakes cannot: SQLSTATE 23505 →
typed conflict, real `xmin` bumps, row-value keyset SQL translation,
transactional idempotency replay/TTL takeover, and the end-to-end HTTP path
including `/readyz`.

Migrations are scaffolded with the pinned local tool
(`dotnet tool restore`, then
`dotnet ef migrations add <Name> --project src/Orders.Infrastructure --output-dir Data/Migrations`);
see [../../recipes/add-migration.md](../../recipes/add-migration.md).

## Project map

Each piece embodies a specific handbook doc:

| Project / path | Responsibility | Governing handbook doc |
|---|---|---|
| `Orders.slnx`, root templates | solution layout, pins, the verify gate | [foundations/project-setup.md](../../foundations/project-setup.md), [quality/linting.md](../../quality/linting.md) |
| `src/Orders.Core/Orders/` | order aggregate (invariants in `Create`/`Amend`), typed `OrderId`/`TenantId`, `IOrderRepository` port, cursor/page/list-query, typed domain exceptions | [foundations/data-modeling.md](../../foundations/data-modeling.md), [foundations/errors-and-logging.md](../../foundations/errors-and-logging.md), [foundations/shared-constructs.md](../../foundations/shared-constructs.md) |
| `src/Orders.Core/Idempotency/` | `IIdempotencyRunner` port + closed `IdempotencyResult` set the transport maps to statuses | [recipes/add-idempotent-write.md](../../recipes/add-idempotent-write.md) |
| `src/Orders.Infrastructure/Data/` | `OrdersDbContext` (typed-ID conversions, no-tracking default), entity configurations (explicit lengths, enum-as-text, `xmin` row version, unique + keyset indexes), committed `Data/Migrations/` | [services/database.md](../../services/database.md), [recipes/add-migration.md](../../recipes/add-migration.md) |
| `src/Orders.Infrastructure/Data/Repositories/` | EF Core repository: tenant-scoped queries, 23505→`DuplicateOrderException` (matched on constraint name), row-value keyset pagination, concurrency-token updates | [services/database.md](../../services/database.md) |
| `src/Orders.Infrastructure/Data/PostgresIdempotencyRunner.cs` | claim → execute → complete in ONE transaction; expired-claim takeover; released claims on error outcomes | [recipes/add-idempotent-write.md](../../recipes/add-idempotent-write.md) |
| `src/Orders.Api/Program.cs` | composition root: options + `ValidateOnStart`, `TimeProvider.System`, middleware order contract, probes, `--migrate` mode, bounded shutdown | [foundations/configuration.md](../../foundations/configuration.md), [templates/program-main.cs.txt](../../templates/program-main.cs.txt) |
| `src/Orders.Api/Endpoints/` + `Contracts/` | endpoint group; wire DTOs with DataAnnotations + source-generated `OrdersJsonContext` (camelCase, enum-as-string) | [services/http-services.md](../../services/http-services.md), [foundations/serialization.md](../../foundations/serialization.md) |
| `src/Orders.Api/Auth/` | JWKS bearer authn, deny-by-default policies, tenant-scope filter, local-dev scheme, audited denials | [operations/security.md](../../operations/security.md) |
| `src/Orders.Api/Idempotency/` | Idempotency-Key endpoint filter: required key, fingerprinting, byte-identical `StoredResponseResult` replay | [recipes/add-idempotent-write.md](../../recipes/add-idempotent-write.md) |
| `src/Orders.Api/ErrorHandling/` + `Middleware/` | the ONE domain-exception→ProblemDetails mapper; stable problem `type` URIs; request-ID middleware | [foundations/errors-and-logging.md](../../foundations/errors-and-logging.md), [services/http-services.md](../../services/http-services.md) |
| `src/Orders.Api/RateLimiting/` | per-identity token bucket, 429 + `Retry-After` | [services/http-services.md](../../services/http-services.md) |
| `src/Orders.Api/Telemetry/` | `AddServiceTelemetry()` (OTel three signals over OTLP), `OrdersMetrics` via `IMeterFactory`, dedicated `AuditLogger` | [operations/observability.md](../../operations/observability.md), [operations/security.md](../../operations/security.md) |
| `tests/Orders.UnitTests/` | domain tests + `FakeTimeProvider`; hand-rolled in-memory fakes behind the Core ports; `WebApplicationFactory` transport tests (CRUD, validation, idempotency statuses, request-id, rate limiting, JWKS auth against a stubbed issuer, audit records); golden-file JSON contract tests (`TestData/`, regenerate with `UPDATE_GOLDEN=1`) | [quality/testing.md](../../quality/testing.md), [foundations/time.md](../../foundations/time.md) |
| `tests/Orders.IntegrationTests/` | Testcontainers PostgreSQL (one container per assembly): migrations apply, repository semantics, transactional idempotency runner, full-stack HTTP round trips | [quality/testing.md](../../quality/testing.md), [services/database.md](../../services/database.md) |

## Observability

Telemetry is wired once in `AddServiceTelemetry()`
([operations/observability.md](../../operations/observability.md)):

- **All three signals over OTLP.** Traces (ASP.NET Core + HttpClient
  instrumentation and the `Orders.Api` `ActivitySource`), metrics (the
  `Orders.Api` meter, the `Npgsql` connection-pool meter, runtime + ASP.NET
  instrumentation), and logs all export via the OTLP exporter, configured by
  the standard `OTEL_EXPORTER_OTLP_*` environment variables.
- **No `/metrics` endpoint by design.** OTLP push is the handbook default;
  the OTel Prometheus exporter is added only when the org scrapes. That is
  this module's documented decision — swap `AddOtlpExporter()` for
  `AddPrometheusExporter()` + `MapPrometheusScrapingEndpoint()` if your
  platform pulls.
- **Domain metrics through `IMeterFactory`** (`orders.created`,
  `orders.idempotent_replays`) — never a static `Meter`, so tests get isolated
  instruments. RED metrics per endpoint come from the built-in ASP.NET Core
  instrumentation, not hand-rolled counters.
- **Probes:** `/livez` runs no checks (a dead database must not restart the
  pod); `/readyz` runs the `"ready"`-tagged `AddDbContextCheck` and sheds
  traffic instead. Both are anonymous on purpose — the deny-by-default
  fallback policy would otherwise 401 the kubelet.
- **Request IDs:** the outermost middleware adopts a well-formed inbound
  `X-Request-Id` (bounded charset/length — an untrusted header never lands in
  logs raw) or keeps the generated one, echoes it on every response via
  `OnStarting` (so it survives the exception handler's response reset), and
  the same id lands in every ProblemDetails body as `requestId`.

## Security & identity

One copyable default, per [operations/security.md](../../operations/security.md):

- **Authentication:** `Microsoft.AspNetCore.Authentication.JwtBearer` against
  the issuer's JWKS via OIDC discovery. Issuer and audience are pinned from
  config; `ValidAlgorithms` is an allowlist (`RS*`/`ES*`/`PS*`) so `alg=none`
  and HS/RS-confusion tokens are rejected; `MapInboundClaims` is off so raw
  claim names (`sub`, `tenant`, `roles`) survive. Every authentication failure
  emits an audit event with the reason; the 401 body never says which check
  failed. The unit suite proves the whole path with a locally-signed key and a
  stubbed discovery/JWKS transport — no real identity provider anywhere.
- **Authorization:** deny by default — a fallback policy requires an
  authenticated user on every endpoint; the probes opt out explicitly.
  Capability policies (`orders:read`, `orders:write`) map to roles in one
  place. Denials are audited via a decorated
  `IAuthorizationMiddlewareResultHandler`.
- **Multi-tenancy:** the tenant comes from the principal, never a request
  body. A group-level endpoint filter rejects tokens without a tenant claim
  (403, audited); every repository query filters on `TenantId`; cross-tenant
  reads are 404s. Proven offline by fakes and live by the container suite.
- **Idempotency durability (honest):** unlike the Go reference's
  capture-and-replay, this runner is **single-transaction**: the InFlight
  claim commits first (the unique constraint is the concurrency gate), then
  the domain write and the Completed record commit atomically on the shared
  scoped `DbContext`. A crash between claim and commit leaves an InFlight row
  that answers 409 until the TTL expires, then is taken over — never a
  double-applied write. Error outcomes are not recorded, so a retry after a
  4xx/5xx re-executes.
- **Audit stream:** the `Orders.Audit` logger category is separate from the
  application log; production logging config routes it to its own sink with
  its own retention. Records carry actor, tenant, action, resource id, result,
  and requestId — never tokens, payloads, or PII. Unit tests assert the
  records for authn failure, authz denial, and each successful mutation.

## Intentionally out of scope

- No outbound typed HTTP clients (`IHttpClientFactory` + resilience handler) —
  see [operations/resilience.md](../../operations/resilience.md) and the
  worker/gRPC reference modules.
- No caching layer (`HybridCache`) — [services/caching.md](../../services/caching.md).
- No messaging/outbox — that is `exampleworker`'s job.
- No Kubernetes manifests — the committed
  [templates/k8s-deployment.yaml](../../templates/k8s-deployment.yaml) is the
  production rollout shape; compose here is the local stack only.
- TLS terminates at the platform edge; Kestrel serves plaintext HTTP in the
  container ([operations/deployment.md](../../operations/deployment.md)).
