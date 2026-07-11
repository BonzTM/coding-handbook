# gRPC Services

gRPC defaults for repos that need strong contracts, typed APIs, and predictable transport boundaries.

## Default Approach

- Use gRPC when the repo benefits from strong schemas, streaming, or polyglot clients. The stack is `Grpc.AspNetCore` on Kestrel — same host, same DI, same telemetry wiring as HTTP endpoints ([http-services.md](http-services.md)).
- Put wire contracts under `api/<service>/v1/`. The repo owns its protos; they are source, reviewed like source.
- Keep the service implementation thin: decode, call `Orders.Core`, map errors. Domain logic never lives in generated types or service methods.

### Protocol Layout

```text
api/orders/v1/
  orders.proto
src/Orders.Api/Grpc/
  OrdersGrpcService.cs
  Interceptors/
    RequestLoggingInterceptor.cs
    AuthInterceptor.cs
    ExceptionMappingInterceptor.cs
```

- Always version the proto package and directory, even at `v1`:

```protobuf
syntax = "proto3";

package orders.v1;

option csharp_namespace = "Orders.Api.Grpc.V1";
```

- Keep messages and services small and explicit. Proto files are transport contracts, not database tables.
- Proto package versioning is separate from release versioning. `orders.v1` states wire compatibility; the NuGet/container release version states what shipped. Breaking a wire contract means a new proto package (`orders.v2`), not a new release number. See [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md).

### Generated Code Policy

Generation happens at build, never by hand, never committed.

- Reference `Grpc.AspNetCore` (it carries `Grpc.Tools` and `Google.Protobuf`) and declare protos as `Protobuf` items in the csproj:

```xml
<ItemGroup>
  <Protobuf Include="../../api/orders/v1/*.proto"
            ProtoRoot="../../api"
            GrpcServices="Server" />
</ItemGroup>
```

- `Grpc.Tools` ships a pinned `protoc`, so generation is deterministic and `dotnet build` is the only tool contributors need. Generated code lands in `obj/` — never commit it, never edit it.
- One `ProtoRoot` per repo (`api/`). Imports resolve relative to it; no ad-hoc include paths.
- Buf-style hygiene still applies even without buf: one package per directory, versioned directories, lint-clean field naming (`snake_case` fields, `PascalCase` messages), and a backward-compatibility review before any change to a shipped proto. If the org already runs `buf`, add `buf lint` and `buf breaking` to CI — but generation stays with `Grpc.Tools`.

### Server Rules

- Service methods decode transport data, call Core with the call's `CancellationToken`, and map errors to `StatusCode.*`. Nothing else.

```csharp
public sealed class OrdersGrpcService(IOrderService orders) : OrdersService.OrdersServiceBase
{
    public override async Task<GetOrderResponse> GetOrder(
        GetOrderRequest request, ServerCallContext context)
    {
        var order = await orders.GetAsync(
            OrderId.Parse(request.OrderId), context.CancellationToken);
        return order.ToResponse();
    }
}
```

- Interceptors own cross-cutting behavior. Registration order is execution order — the first registered interceptor is the outermost. Pin the order and keep it stable:

```csharp
builder.Services.AddGrpc(options =>
{
    // Registration order = execution order. Outermost first.
    options.Interceptors.Add<RequestLoggingInterceptor>();   // observes the final mapped status
    options.Interceptors.Add<AuthInterceptor>();             // rejects before any handler work
    options.Interceptors.Add<ExceptionMappingInterceptor>(); // innermost: nothing unmapped escapes
});
```

  Logging sits outermost so it records the status the client actually saw; exception mapping sits innermost so every interceptor above it observes a proper `RpcException`, not a raw domain exception. Tracing is not an interceptor — OpenTelemetry's ASP.NET Core instrumentation already covers gRPC calls ([../operations/observability.md](../operations/observability.md)).
- Deadlines: clients set them, servers respect them. `context.CancellationToken` fires on client cancellation *and* deadline expiry — pass it through every Core and I/O call ([../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md)). For service-to-service fan-out, register clients through the gRPC client factory with `EnableCallContextPropagation()` so the inbound deadline and cancellation flow to outbound calls automatically. Servers still protect themselves from unbounded work; a missing client deadline is not permission to run forever.
- Expose the gRPC health service for deployability, and reflection so `grpcurl` works without local protos:

```csharp
builder.Services.AddGrpcHealthChecks();
builder.Services.AddGrpcReflection();

app.MapGrpcService<OrdersGrpcService>();
app.MapGrpcHealthChecksService();
app.MapGrpcReflectionService(); // internal services; config-gate off at public edges
```

  `AddGrpcHealthChecks` reuses the same `Microsoft.Extensions.Diagnostics.HealthChecks` registrations as `/livez` and `/readyz` — one set of checks, two transports.

### Error Details

A bare `new RpcException(new Status(StatusCode.InvalidArgument, msg))` tells a client *that* the request failed but not *what* to fix. For validation failures, attach a machine-readable detail so clients map errors back to fields instead of scraping a message string.

- Use the richer error model from `Grpc.StatusProto`: build a `Google.Rpc.Status` carrying a `BadRequest` detail with one `FieldViolation` per failed field, then `ToRpcException()`:

```csharp
private static RpcException ToRpcException(OrderValidationException ex)
{
    var badRequest = new BadRequest();
    foreach (var violation in ex.Violations)
    {
        badRequest.FieldViolations.Add(new BadRequest.Types.FieldViolation
        {
            Field = violation.Field,
            Description = violation.Message,
        });
    }

    return new Google.Rpc.Status
    {
        Code = (int)Google.Rpc.Code.InvalidArgument,
        Message = "order validation failed",
        Details = { Any.Pack(badRequest) },
    }.ToRpcException();
}
```

- Carry the violations *structurally* out of Core (a typed `FieldViolation` list on the validation exception), and build the proto detail once, in the exception-mapping interceptor. Core must not reference `Google.Rpc` — the proto detail type is a transport concern, exactly as `ProblemDetails` is for HTTP ([../foundations/errors-and-logging.md](../foundations/errors-and-logging.md)).
- Use the standard `google.rpc.*` detail types (`BadRequest`, `ErrorInfo`, `RetryInfo`, `QuotaFailure`) so any gRPC client can decode them (`GetRpcStatus()` / `GetDetail<BadRequest>()`), rather than inventing a bespoke detail message.
- Never return raw internal errors. An unmapped exception surfaces as `Unknown` with an opaque message — that is the signal the mapping interceptor has a gap.

### Transport Security

gRPC over HTTP/2 must run on TLS in production; plaintext (h2c) is for local development only.

- **TLS by default in production.** Configure the Kestrel endpoint from config so the same binary serves both modes:

```json
{
  "Kestrel": {
    "Endpoints": {
      "Grpc": {
        "Url": "https://0.0.0.0:5001",
        "Protocols": "Http2",
        "Certificate": { "Path": "/certs/tls.crt", "KeyPath": "/certs/tls.key" }
      }
    }
  }
}
```

  Kestrel refuses to start when a configured certificate fails to load — that fail-fast is the contract. Never catch it and fall back to plaintext. Enforce a TLS minimum via `HttpsConnectionAdapterOptions.SslProtocols` (TLS 1.2 or later).
- **mTLS for internal service-to-service traffic** unless a service mesh terminates TLS for you. When a client-CA bundle is configured, require and verify the client certificate:

```csharp
builder.WebHost.ConfigureKestrel(kestrel =>
    kestrel.ConfigureHttpsDefaults(https =>
    {
        https.ClientCertificateMode = ClientCertificateMode.RequireCertificate;
        https.ClientCertificateValidation =
            (cert, _, _) => clientCa.Validates(cert);
    }));
```

  Validate the trio together at startup: a client-CA path without a server cert/key is a config error, not a warning. If a mesh (Istio/Linkerd) already provides mutual TLS at the sidecar, do not double-terminate — run the listener as plaintext HTTP/2 *inside* the pod (`Protocols: Http2` on an `http://` URL) and let the mesh own the edge.
- Keep it config-gated: no certificate configured selects the insecure local/dev listener so the service boots offline — the same gating pattern as auth and tracing. Options validation with `ValidateOnStart` enforces the invariants ([../foundations/configuration.md](../foundations/configuration.md)).

## Common Mistakes And Forbidden Patterns

- domain logic living in proto-generated types or service methods
- missing versioned proto packages, or protos scattered outside `api/`
- committing or hand-editing code generated into `obj/`, or generating with an ad-hoc local `protoc` instead of the `Protobuf` build items
- returning raw internal errors to clients, or letting unmapped exceptions surface as `Unknown`
- ignoring `context.CancellationToken` — Core calls that outlive the client's deadline
- interceptor order left to whoever registered last, instead of the pinned logging → auth → exception-mapping contract
- proto files that mirror database tables instead of transport contracts
- flattening a field-validation failure into a bare `InvalidArgument` message with no `google.rpc.BadRequest` detail, so clients cannot map errors to fields
- building `Google.Rpc` detail types inside Core instead of at the transport boundary, coupling domain logic to the wire format
- serving plaintext gRPC in production, or silently downgrading to plaintext when a configured key pair fails to load instead of failing fast
- double-terminating TLS when a mesh already provides mTLS, or skipping client-cert verification on internal links a mesh does not cover

## Verification And Proof

- `dotnet build` regenerates stubs from `api/` protos — a stale or hand-edited contract cannot survive the gate; run `pwsh ./verify.ps1` (restore (locked), format-check, build (warnings-as-errors), test, audit)
- service tests for status mapping and interceptor behavior, in-proc via `WebApplicationFactory` with a `GrpcChannel` ([../quality/testing.md](../quality/testing.md))
- a test that asserts a validation failure returns `InvalidArgument` *and* a `google.rpc.BadRequest` detail with the expected `{field, description}` violations — decode the detail with `GetRpcStatus()`/`GetDetail<BadRequest>()`, do not match the message string
- a TLS/mTLS test: the server rejects a missing/invalid client cert when a client CA is configured, and startup fails when a configured key pair cannot load
- `grpcurl` smoke test against a local server:

```bash
grpcurl -plaintext localhost:5001 list
grpcurl -plaintext localhost:5001 grpc.health.v1.Health/Check
grpcurl -plaintext -d '{"order_id":"0197b7e2-..."}' \
  localhost:5001 orders.v1.OrdersService/GetOrder
```

- backward-compatibility review before breaking a public proto contract ([../checklists/pr-review.md](../checklists/pr-review.md))
