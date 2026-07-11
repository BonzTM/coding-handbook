# examplegrpc

A compiling gRPC reference service for the C# engineering handbook. It is the
gRPC analog of [exampleservice](../exampleservice/) (the HTTP + PostgreSQL
keystone) and mirrors its structure and conventions: a thin `Program.cs`
composition root, an `Orders.Grpc.Core` domain with a consumer-defined store
seam, config-gated auth and TLS, an injectable `TimeProvider`, and the single
`pwsh ./verify.ps1` safety gate. The language baseline is **.NET 10 / C# 14**
(`global.json` pins the SDK). Its Go twin is
`golang/reference/examplegrpc` (widgets there, orders here - same shape).

Governing docs: [grpc-services.md](../../services/grpc-services.md),
[add-grpc-method.md](../../recipes/add-grpc-method.md).

> **Note:** the root files (`global.json`, `Directory.Build.props`,
> `Directory.Packages.props`, `nuget.config`, `.editorconfig`, `verify.ps1`,
> `Makefile`, `Dockerfile`, `.dockerignore`, `.gitattributes`, `.gitignore`)
> are copies of [templates/](../../templates/), taken from the keystone
> module's fixed copies (its README lists the template gaps they close).
> `Directory.Packages.props` is trimmed to this module's packages plus the
> gRPC family (`Grpc.AspNetCore`, `Grpc.AspNetCore.HealthChecks`,
> `Grpc.AspNetCore.Server.Reflection`, `Grpc.StatusProto`, `Grpc.Net.Client`,
> `Grpc.HealthCheck`, `Grpc.Reflection` - all 2.80.0, verified on nuget.org
> 2026-07-11), flagged for upstreaming into the template.

## Layout

```text
api/orders/v1/orders.proto     versioned wire contract (orders.v1); the ONLY
                               committed contract artifact - Grpc.Tools
                               generates stubs into obj/ at build time
src/Orders.Grpc.Core/          domain: Order aggregate (collects ALL field
                               violations), typed OrderId/TenantId, keyset
                               OrderCursor, IOrderStore port + the in-memory
                               store, OrderService (role checks live here),
                               CallerPrincipal identity
src/Orders.Grpc.Api/           transport host
  Grpc/OrdersGrpcService.cs    thin adapter: decode -> one Core call -> encode
  Grpc/GrpcErrorMapping.cs     domain exception -> StatusCode.* + google.rpc
                               BadRequest details (mapped once, at the boundary)
  Grpc/RpcDeadlineGuard.cs     server-side ceiling when the client sends no
                               deadline
  Grpc/Interceptors/           the pinned chain: logging -> auth -> exception
                               mapping
  Auth/                        IAuthenticator seam: static-token (Auth:Enabled
                               =true) or synthetic local-dev principal
  Tls/                         config-gated TLS/mTLS: options trio validation,
                               PEM loading, CA-bundle client-cert validator
  Telemetry/                   AddServiceTelemetry() (OTel over OTLP), metrics
  Program.cs                   composition root: options + ValidateOnStart,
                               two Kestrel listeners, health, reflection,
                               bounded shutdown
tests/Orders.Grpc.UnitTests/   domain tests, mapping/deadline-guard/TLS unit
                               tests, and in-proc transport tests (GrpcChannel
                               over WebApplicationFactory's TestServer)
```

## Run

```bash
export DOTNET_ROOT="$HOME/.dotnet"; export PATH="$HOME/.dotnet:$PATH"
dotnet run --project src/Orders.Grpc.Api     # in-memory store, local-dev auth
```

The gRPC server listens HTTP/2-only on `:5001`; the probes listener serves
`/livez` and `/readyz` over HTTP/1.1 on `:8080` (h2c and HTTP/1.1 cannot share
a plaintext listener - the same two-port shape as the Go reference's metrics
sidecar). Server reflection and the standard gRPC health service are
registered, so `grpcurl` works out of the box:

```bash
grpcurl -plaintext localhost:5001 list
grpcurl -plaintext localhost:5001 grpc.health.v1.Health/Check
grpcurl -plaintext -d '{"external_reference":"ord-1001","customer_id":"cust-42","quantity":3}' \
  localhost:5001 orders.v1.OrdersService/CreateOrder
grpcurl -plaintext -d '{"id":"<id-from-create>"}' localhost:5001 orders.v1.OrdersService/GetOrder
grpcurl -plaintext -d '{"page_size":2}' localhost:5001 orders.v1.OrdersService/ListOrders
grpcurl -plaintext -d '{}' localhost:5001 orders.v1.OrdersService/StreamOrders
curl -s localhost:8080/livez
curl -s localhost:8080/readyz
```

Configuration keys are documented in
[appsettings.json](src/Orders.Grpc.Api/appsettings.json) (committed,
secret-free - it doubles as the `.env.example`); environment variables
override it (`Section__Key`).

## Interceptor chain (the order contract)

Registration order = execution order
([grpc-services.md](../../services/grpc-services.md)):

1. **`RequestLoggingInterceptor`** - outermost, so it records the status the
   client actually saw; one line per RPC (method, code, duration, request id -
   a well-formed inbound `x-request-id` is adopted, a malformed one replaced).
2. **`AuthInterceptor`** - rejects before any handler work. Registered
   globally so a future service is authenticated by default (fail closed);
   health and reflection are the two documented exemptions.
3. **`ExceptionMappingInterceptor`** - innermost: nothing unmapped escapes.
   Everything above it observes a proper `RpcException`.

Tracing is not an interceptor - OpenTelemetry's ASP.NET Core instrumentation
already covers gRPC calls.

## Errors

Domain exceptions are mapped to `StatusCode.*` exactly once, at the transport
boundary (`GrpcErrorMapping`). A validation failure maps to `InvalidArgument`
**with** a `google.rpc.BadRequest` detail listing every offending
`{field, description}` pair - Core carries the violations structurally
(`FieldViolation` on `OrderValidationException`); the proto detail is built at
the boundary via `Grpc.StatusProto` (`Google.Rpc.Status` + `ToRpcException()`).
Other exceptions keep their codes (`NotFound`, `AlreadyExists`,
`PermissionDenied`, `InvalidArgument` for a malformed page token); an
unexpected exception is logged once server-side and becomes a generic
`Internal` so internals never leak.

Decode the detail on the client with `exception.GetRpcStatus()` then
`.GetDetail<BadRequest>()` - never by scraping the message string.

## Deadlines and cancellation

`context.CancellationToken` fires on client cancellation *and* client deadline
expiry; every Core and I/O call receives it. When the client sends **no**
deadline, `RpcDeadlineGuard` links the configured `Server:MaxRpcDuration`
ceiling into the token - a missing deadline is not permission to run forever.
A cancellation the client did not cause maps to `DEADLINE_EXCEEDED`; a client
cancellation maps to `CANCELLED`. The guard is deterministic under test
(`FakeTimeProvider` drives the ceiling).

## Transport security (TLS / mTLS)

TLS is **config-gated** (`Tls` section, validated at startup):

- `Tls:CertPath` + `Tls:KeyPath` set ⇒ the gRPC listener serves TLS
  (PEM pair, TLS 1.2 minimum).
- Additionally `Tls:ClientCaPath` ⇒ **mutual TLS**: the server requires and
  verifies the client certificate against that CA bundle and ONLY that bundle
  (`CustomRootTrust` - the machine root store is ignored). mTLS is the default
  posture for internal service-to-service traffic unless a mesh terminates TLS.
- Neither set ⇒ an **insecure** plaintext listener for **local/dev only**; the
  process logs a loud warning at startup. **Production requires TLS.**

A configured-but-unloadable key pair is **fail-fast** at startup (never a
silent downgrade to plaintext). Cert and key must be set together, and a
client CA requires server TLS - both enforced by `TlsOptionsValidator` via
`ValidateOnStart`. The probes listener stays plaintext HTTP/1.1 (the kubelet
probes it in-cluster).

## Security & identity

- **Authentication** is a seam (`IAuthenticator`), config-gated exactly like
  the keystone: `Auth:Enabled=true` wires a constant-time static-bearer-token
  authenticator (the dependency-free reference implementation, mirroring the
  Go module); `Auth:Enabled=false` (the committed local default) wires a
  synthetic local-dev principal so the service boots offline. Production
  swaps a JWT/JWKS validator in behind the same interface - the keystone
  module shows that wiring end to end. The UNAUTHENTICATED message never says
  which check failed; the reason is logged server-side.
- **Authorization** lives in the domain: `OrderService` requires
  `orders.reader` / `orders.writer` roles on the `CallerPrincipal`, so the
  rules hold no matter which transport fronts the domain.
- **Multi-tenancy:** the tenant comes from the principal, never a request
  field; the store scopes every operation by tenant; a cross-tenant read is
  `NOT_FOUND`, indistinguishable from a missing row.

## Codegen

Nothing generated is committed (unlike the Go module, which commits
buf-generated stubs - Grpc.Tools makes `dotnet build` the only tool
contributors need). The proto under `api/orders/v1/` is the whole contract
diff; `Grpc.Tools` (carried by `Grpc.AspNetCore`) regenerates deterministically
into `obj/` on every build with `GrpcServices="Both"`, so the transport tests
drive the server through the real generated client.

## Verify

`pwsh ./verify.ps1` is the single ordered gate - restore (locked),
format-check, build (warnings-as-errors), test, audit. Humans, the Makefile
shim, and CI run the same script.

```bash
pwsh ./verify.ps1                # the full offline gate
pwsh ./verify.ps1 -Integration   # documented no-op here: no external
                                 # dependency, so there is no IntegrationTests
                                 # project; the switch adds nothing
```

The unit suite covers the domain (roles, tenant scoping, duplicate references,
keyset pagination incl. the equal-timestamp tie-break, stream cancellation),
the mapping table (every `StatusCode` plus the decoded `BadRequest` detail),
the deadline guard (FakeTimeProvider), TLS gating (options trio, PEM pair
fail-fast, CA-bundle client-cert validation with generated certs), and the
full transport through an in-proc `GrpcChannel`: CRUD round trips, auth
on/off, health, reflection, probes, access-log lines, and INTERNAL shielding.

## Observability

Telemetry is wired once in `AddServiceTelemetry()`
([observability.md](../../operations/observability.md)): traces, metrics, and
logs over OTLP (`OTEL_EXPORTER_OTLP_*` environment variables). RED metrics per
RPC come from the built-in ASP.NET Core instrumentation; domain counters
(`orders.created`, `orders.streamed`) go through `IMeterFactory`. No
Prometheus `/metrics` endpoint by design - OTLP push is the handbook default;
swap in the OTel Prometheus exporter when the org scrapes. `/livez` runs no
checks; `/readyz` runs the `"ready"`-tagged checks (none here - no
dependencies); the standard gRPC health service aggregates the SAME
health-check registrations, one set of checks, two transports.

## Intentionally out of scope

- **No database** - the store is honestly in-memory, mirroring the Go module:
  this reference proves the transport patterns; [exampleservice](../exampleservice/)
  proves EF Core + PostgreSQL (migrations, `--migrate`, Testcontainers) behind
  the same kind of port.
- No outbound gRPC clients (`AddGrpcClient` + `EnableCallContextPropagation()`
  for deadline/cancellation fan-out) - see
  [grpc-services.md](../../services/grpc-services.md).
- No messaging/outbox - that is `exampleworker`'s job.
- No docker-compose - there is nothing to compose; the Dockerfile builds the
  standalone image.
- No Kubernetes manifests - see
  [templates/k8s-deployment.yaml](../../templates/k8s-deployment.yaml).
