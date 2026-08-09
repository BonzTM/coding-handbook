# HTTP Services

FastAPI service defaults for predictable adapters, explicit lifecycle ownership, and framework-free core behavior.

## Default Approach

Build one FastAPI application through `create_app()`. Keep HTTP code under `src/<app>/api/http/`; routers translate the wire contract, call one core use case, and translate the result.

### Service Layout

```text
src/<app>/
  main.py                 # composition and create_app()
  core/                   # domain values, use cases, Protocol ports
  api/http/
    routers/              # endpoint adapters
    dto.py                # request/response Pydantic models
    errors.py             # exception-to-problem mapping
    dependencies.py       # Depends providers for core ports
    middleware.py         # request ID and HTTP cross-cutting behavior
```

`main.py` wires concrete adapters. A router never imports a SQLAlchemy session, migration, broker client, or domain-specific HTTPX implementation.

### App Factory And Lifespan

`create_app(settings: Settings | None = None) -> FastAPI` returns an unstarted app with routers and exception handlers registered. It performs no import-time I/O. Tests call the factory and override dependencies.

One `@asynccontextmanager` lifespan acquires app-scoped resources before yielding and closes them in reverse dependency order afterward: database engine/session factory, one outbound `httpx.AsyncClient`, worker supervisors, and telemetry exporters. FastAPI documents lifespan as the recommended startup/shutdown mechanism for shared resources in its [lifespan guide](https://fastapi.tiangolo.com/advanced/events/). Do not mix lifespan with legacy startup/shutdown event handlers.

### Router Contract

A router does five things:

1. receive validated transport values and authenticated request context
2. map DTOs and primitive route values into domain values
3. resolve one core-owned port or use case through `Depends`
4. invoke it under the request deadline
5. map the domain result or exception into the published response

Use `Annotated[Port, Depends(provider)]` aliases for repeated wiring. Providers obtain lifespan-owned objects from typed application state and return core-owned `Protocol` surfaces. FastAPI's [dependency guide](https://fastapi.tiangolo.com/tutorial/dependencies/) defines `Depends`; it is adapter wiring, not a container imported by core.

### DTO And Validation Boundary

Request and response DTOs are separate Pydantic v2 models. Inbound command models set `ConfigDict(extra="forbid")`, declare bounds and strict coercion policy, and map once into plain domain values. Response models select and alias fields explicitly. Never return a domain dataclass or SQLAlchemy model and rely on framework serialization.

Override FastAPI's request-validation handler so Pydantic implementation details do not become the public contract. Emit stable field paths and codes through the problem response defined below. See [serialization](../foundations/serialization.md).

### Problem Responses

Register exception handlers once on the app. Known domain exceptions map to stable problem `type`, `title`, `status`, and safe extensions; unknown exceptions become an opaque 500 containing only the request identifier. Use `application/problem+json` and keep the HTTP status equal to the body `status`, as required by [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html).

The `type` URI names a durable problem class, never a Python exception. Validation extensions expose no Pydantic URLs, raw input, secrets, SQL, dependency names, or tracebacks. Log an unexpected failure once in the outer exception boundary.

### Request IDs And Middleware

The outer request-ID middleware validates a caller-provided identifier against a small length/character policy or replaces it with a generated opaque value. Store it in request state/context, return it in the response header, include it in logs/traces, and place it in opaque 5xx problems. Never trust an unbounded header or use request IDs as metric labels.

Wire middleware in a documented stable order: request ID, exception/recovery boundary, trace context, security headers/CORS, authentication and limits, access logging/metrics, router. FastAPI notes that the last added middleware is outermost in its [middleware guide](https://fastapi.tiangolo.com/tutorial/middleware/); test the actual order rather than relying on visual registration order.

### Timeouts And Limits

Every request receives an overall monotonic deadline. Scope core work with `asyncio.timeout()` and pass the remaining budget to database and external-client adapters; retries share it. Map deadline exhaustion to the published timeout problem and continue propagating cancellation during disconnect or shutdown.

Configure Uvicorn concurrency, backlog, keep-alive, maximum-request recycling when justified, and graceful-shutdown timeout from bounded settings. Its [settings reference](https://uvicorn.dev/settings/) defines `--limit-concurrency` and `--timeout-graceful-shutdown`. The ingress/proxy owns header/body/connection limits it can enforce; the app additionally bounds decoded request bodies, uploads, collection sizes, and expensive fan-out. Streaming endpoints require an explicit timeout and disconnect design.

### Security Headers And CORS

Set transport-appropriate headers once. JSON APIs return `X-Content-Type-Options: nosniff`, a restrictive `Referrer-Policy`, and cache controls appropriate to sensitivity. Browser-rendered responses add CSP and framing policy under [web apps](web-apps.md). TLS/HSTS is owned by the actual HTTPS edge; do not claim HSTS on an internal plaintext listener without platform proof.

Service-to-service APIs have no CORS middleware. When browser clients are part of the contract, configure `CORSMiddleware` with an explicit origin, method, and header allowlist. Credentialed CORS never uses wildcard origins; OWASP recommends specific origins in its [HTTP headers guidance](https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html#access-control-allow-origin).

### Idempotent Writes

Require an idempotency key for externally retried writes that must produce one effect. Validate and scope the key by tenant, authenticated actor, route/operation, and normalized request fingerprint. Atomically reserve it with the domain write, then store the final status, safe headers, and response body for replay.

The same key plus same fingerprint returns the stored result. The same key plus a different fingerprint returns a stable conflict problem. Concurrent duplicates collapse behind a unique constraint/transaction; they do not both execute. Define expiry, in-progress recovery, retryable failure treatment, and storage limits before shipping. Never use an in-process dictionary for cross-replica correctness.

### Health And Metrics Endpoints

- `/livez` is cheap, local, and answers whether the process/event loop can continue. It does not query dependencies.
- `/readyz` is bounded and reports whether the instance can accept new traffic. It becomes unready during startup, drain, or loss of a required dependency.
- `/metrics` exposes Prometheus text through `prometheus-client`, separate from health and protected according to platform exposure policy.

Probe responses contain no secrets, DSNs, dependency topology, or raw exceptions. Cache expensive readiness checks briefly and bound their fan-out. A dependency that is not required for serving traffic does not fail readiness automatically.

### Graceful Shutdown

On SIGTERM, stop readiness and new work first. Uvicorn drains accepted requests within its configured grace; the lifespan owner then cancels/awaits workers, closes clients and the database engine, and flushes telemetry last. The application shutdown budget stays below the platform termination grace. A smoke test sends SIGTERM during a slow request and proves bounded exit.

## Common Mistakes And Forbidden Patterns

- Process-global app whose import loads settings, opens resources, or configures logging.
- Router querying SQLAlchemy or calling a domain-specific HTTPX client directly.
- FastAPI, Pydantic DTOs, or request objects crossing into core.
- Raw Pydantic validation errors, bare error JSON, or exception text in a 5xx body.
- Caller request IDs accepted without bounds, or IDs used as metric labels.
- Per-request engine/client creation, implicit timeout defaults, or a fresh timeout for every retry.
- CORS enabled by reflex, wildcard credentialed origins, or HSTS asserted at the wrong TLS boundary.
- Idempotency based on key alone without tenant/operation/fingerprint scope or durable atomic storage.
- Liveness calling dependencies, readiness remaining true during drain, or probes exposing internals.
- Unbounded lists, bodies, uploads, fan-out, request drain, or background tasks.

## Verification And Proof

```bash
uv run pytest tests/api/http
uv run pytest -k "problem or request_id or idempot"
make verify
```

Prove exact success/problem wire shapes, unknown-field rejection, dependency overrides, middleware order, request-ID replacement, deadline expiry, cancellation, body/concurrency limits, CORS allow/deny behavior when enabled, duplicate-write collapse, probe semantics, lifespan resource order, and SIGTERM drain. Smoke-test `/livez`, `/readyz`, and `/metrics` on the built artifact.

Related: [serialization](../foundations/serialization.md), [errors and logging](../foundations/errors-and-logging.md), [resilience](../operations/resilience.md), and [add HTTP endpoint](../recipes/add-http-endpoint.md).
