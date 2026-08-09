# gRPC Services

gRPC defaults for strong versioned contracts, async lifecycle ownership, and framework-free domain behavior.

## Default Approach

Use `grpcio`'s `grpc.aio` API when schemas, streaming, or polyglot clients justify gRPC. Put authoritative `.proto` files under `api/<service>/v1/`; keep generated modules separate from handwritten adapters under `src/<app>/api/grpc/`.

### Protocol Layout

```text
api/orders/v1/orders.proto
src/<app>/api/grpc/
  server.py
  service.py
  errors.py
  interceptors.py
src/<app>/generated/orders/v1/
  orders_pb2.py
  orders_pb2.pyi
  orders_pb2_grpc.py
```

Version the proto package and directory from `v1`. Messages model transport contracts, not tables or SQLAlchemy entities. Core owns use cases and plain values; the service maps protobuf requests/results at the edge.

### Generated Stubs

The default generator is pinned `grpcio-tools`. Invoke `python -m grpc_tools.protoc` with deterministic include/output paths and generate message Python, `.pyi`, and gRPC service modules using `--python_out`, `--pyi_out`, and `--grpc_python_out`; the official [Python basics guide](https://grpc.io/docs/languages/python/basics/#generating-client-and-server-code) documents those outputs.

Commit generated stubs so consumers can build without a local compiler. CI regenerates into a clean location and fails on diff. Never edit generated code, suppress it as handwritten code, or treat it as the authoritative schema. Buf is an acceptable ADR-backed generation wrapper only when the organization already owns pinned Buf plugins and compatibility policy, matching [framework selection](../decisions/framework-selection.md).

### Async Server Shape

Composition creates one `grpc.aio.server(...)`, registers interceptors, generated servicers, health, and optional reflection, then binds the configured secure listener. A handwritten servicer method validates transport semantics, maps to domain values, calls one core use case, maps the result/error, and returns a generated response.

Set `maximum_concurrent_rpcs` from a bounded capacity decision; the [AsyncIO API](https://grpc.github.io/grpc/python/grpc_asyncio.html#grpc.aio.server) maps excess work to `RESOURCE_EXHAUSTED`. No blocking function runs on the event loop. Start, stop with a bounded grace, and await termination under the composition-owned root task.

### Interceptors

Server interceptors own authentication/context extraction, authorization preconditions common to all methods, request/correlation context, access logging, metrics, tracing, and unexpected-exception shielding. Register a documented order and test it. Method-specific authorization remains explicit near the adapter/use case.

Interceptors do not consume streaming iterators, log payloads, convert cancellation into `INTERNAL`, or duplicate logs emitted by the acting boundary. Core imports no `grpc` types.

### Deadlines And Cancellation

Clients always set a timeout. Servers read `ServicerContext.time_remaining()`, reject impossible work early, and propagate the shorter remaining budget to database/outbound calls. gRPC warns that no client deadline can mean waiting effectively forever in its [deadline guide](https://grpc.io/docs/guides/deadlines/).

Python propagation is explicit: pass the remaining timeout on each downstream RPC rather than claiming automatic propagation. Stop owned work when the context is cancelled; a server is responsible for stopping activity spawned for an expired RPC. Never swallow `CancelledError` or `grpc.aio.AbortError` after cleanup.

### Error Mapping

Map domain outcomes once to stable gRPC statuses:

| Domain condition | Status |
|---|---|
| invalid request/value | `INVALID_ARGUMENT` |
| missing resource | `NOT_FOUND` |
| unauthenticated caller | `UNAUTHENTICATED` |
| authenticated but forbidden | `PERMISSION_DENIED` |
| version/state conflict | `ABORTED` or `FAILED_PRECONDITION`, chosen per contract |
| capacity exhausted | `RESOURCE_EXHAUSTED` |
| caller deadline exhausted | `DEADLINE_EXCEEDED` |
| unavailable dependency | `UNAVAILABLE` only when retry is safe |
| unexpected failure | opaque `INTERNAL` |

Use safe machine-readable standard error details when clients need field violations or retry metadata; build protobuf detail types in the adapter, never core. Do not expose Python exception messages, tracebacks, dependency names, or SQL. The gRPC [status-code guide](https://grpc.io/docs/guides/status-codes/) is authoritative for semantics.

### Health And Reflection

Register the standard `grpc.health.v1.Health` service and update overall/service status: `NOT_SERVING` before startup and during drain, `SERVING` only when required dependencies can accept work. The standard supports unary `Check` and streaming `Watch`, as described in the [health guide](https://grpc.io/docs/guides/health-checking/).

Enable `grpc_reflection` only when operator tooling requires it and exposure policy permits schema discovery. Register the exact public service names plus reflection/health as intended; disable or restrict it at untrusted production edges. Reflection is an operability surface, not the contract source.

### Transport Security

The platform edge/mesh owns TLS/mTLS when present; the app listens plaintext only inside that proven boundary. A directly exposed server uses reviewed server credentials and fails startup when configured credentials cannot load. Never silently downgrade. Authentication/authorization still applies when a mesh authenticates transport identity.

### Testing

Start a real `grpc.aio.Server` on loopback port `0`, register the real servicer/interceptors, connect with an `grpc.aio` channel, and close both through async context/fixture teardown. This in-process network harness proves generated serialization, status mapping, metadata, interceptor order, deadlines, health, reflection, and streaming cancellation without a deployed environment.

Use hand-rolled core fakes behind Protocols. Add contract generation/compatibility proof and one `grpcurl` smoke test against the assembled local service.

## Common Mistakes And Forbidden Patterns

- Domain rules or persistence types in generated messages/servicer methods.
- Unversioned proto package, hand-edited stubs, or unreviewed generation diff.
- Blocking calls, unlimited concurrent RPCs, missing client deadlines, or downstream calls given a fresh budget.
- Raw exceptions/status text exposed; `UNKNOWN`/`INTERNAL` used for known domain outcomes.
- Authentication, logging, or metrics duplicated across every method instead of a tested interceptor.
- Health left `SERVING` during drain, reflection exposed by reflex, or plaintext outside a proven TLS edge.
- Tests calling servicer methods directly and never proving channel serialization/interceptors.

## Verification And Proof

```bash
uv run python -m grpc_tools.protoc <project-generation-arguments>
uv run pytest tests/api/grpc
make verify
```

Prove clean regeneration, compatibility review, every domain-status mapping, safe details, interceptor order, auth deny paths, deadline propagation, cancellation, concurrency exhaustion, streaming backpressure/drain, health transitions, reflection exposure policy, TLS failure behavior where app-owned, and loopback-channel integration. Run a bounded `grpcurl` smoke test against the built service.

Related: [contracts and compatibility](../foundations/contracts-and-compatibility.md), [concurrency and asyncio](../foundations/concurrency-and-asyncio.md), and [add gRPC method](../recipes/add-grpc-method.md).
