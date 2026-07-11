# Recipe: Add gRPC Method

Use this when a feature adds or changes one RPC on a gRPC service.

## Files To Touch

- `api/orders/v<N>/orders.proto` for the contract change
- `src/Orders.Api/Orders.Api.csproj` — the `<Protobuf Include="..." GrpcServices="Server" />` item, only when adding a new proto file (Grpc.Tools generates stubs at build time into `obj/`; generated code is never committed — the proto is the contract)
- `src/Orders.Api/Grpc/OrdersGrpcService.cs` for the method implementation
- `src/Orders.Api/Grpc/GrpcErrorMapping.cs` if domain-to-`StatusCode` mapping changes
- `src/Orders.Api/Grpc/Interceptors/` only if cross-cutting behavior changes
- `src/Orders.Core/...` for business logic
- transport and core tests under `tests/Orders.UnitTests`

## Steps

1. Edit the proto under the versioned package: add the RPC and its request/response messages, give every field an explicit number, and `reserve` the numbers and names of any removed fields. Do not renumber existing fields.
2. Build to regenerate: `dotnet build src/Orders.Api` — Grpc.Tools regenerates the base class and messages deterministically from the proto. Nothing generated is committed, so there is no stub-drift check to run; the proto diff is the whole contract diff.
3. Add or update the Core method that owns the behavior. Core takes and returns domain types, not generated proto message types — Core has no Grpc or proto references.
4. Implement the override in `OrdersGrpcService.cs` (derived from the generated `Orders.OrdersBase`): map the request message to domain input, call exactly one Core method passing `context.CancellationToken`, map domain errors to a `RpcException` with the right `StatusCode` via `GrpcErrorMapping`, and build the response message. Attach machine-readable detail (field errors, retry hints) via the `google.rpc.Status` detail model built once in the exception-mapping interceptor — the handbook default per [../services/grpc-services.md](../services/grpc-services.md) (### Error Details), exemplified in [../reference/examplegrpc/](../reference/examplegrpc/); do not invent per-method trailer formats.
5. Confirm the registered interceptors already cover exception shielding, auth, access logging, tracing, and metrics; add a metric hook only if the existing interceptor does not capture the new RPC. Use low-cardinality tags (method name, status code) and never request, user, or tenant IDs.
6. Verify the server honors `context.CancellationToken` (deadline expiry and client disconnect both cancel it) and self-protects against unbounded work even when the client sets no deadline — see [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md).

## Invariants To Preserve

- proto package and directory stay versioned (`v<N>`), even at `v1`
- no domain logic in generated types or in the service override; the gRPC service is a thin transport adapter
- `context.CancellationToken` flows into the Core call; deadline and disconnect cancel in-flight work
- domain errors are raised as typed failures in Core and mapped once to `StatusCode` at the service boundary, not logged twice
- the proto change is backward compatible (additive fields, new RPC); any wire break is an explicit, reviewed decision recorded against the contract
- metric tags stay low cardinality

## Proof

- `dotnet build` succeeds with warnings-as-errors (proves the proto generates cleanly)
- table test for status mapping in `GrpcErrorMapping` covering each mapped `StatusCode`: `dotnet test tests/Orders.UnitTests --filter GrpcErrorMapping`
- interceptor test proving exception shielding, auth rejection, and metric emission on the new method: `dotnet test tests/Orders.UnitTests --filter Interceptor`
- Core unit test for the new behavior plus a cancellation test: `dotnet test tests/Orders.UnitTests --filter <Method>`
- `grpcurl -plaintext -proto api/orders/v1/orders.proto -d '{...}' localhost:<port> orders.v1.Orders/<Method>` smoke call against a locally run host
- run `pwsh ./verify.ps1`

Governing doc: [grpc-services.md](../services/grpc-services.md). The HTTP analog is [add-http-endpoint.md](add-http-endpoint.md). If the method calls another service, see [add-external-client.md](add-external-client.md). To retire an RPC or field, see [deprecate-and-remove-contract.md](deprecate-and-remove-contract.md).
