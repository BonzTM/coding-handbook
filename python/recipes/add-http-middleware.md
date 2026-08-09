# Recipe: Add HTTP Middleware

Use this when HTTP-wide envelope behavior cannot live in one route.

## Files To Touch

- `src/<app>/api/http/middleware.py` or the dependency module
- `src/<app>/main.py` for ordering and scope
- middleware/dependency tests and config when tunable

## Steps

1. Use an ASGI middleware for request/response envelope work that must wrap many routes: request IDs, recovery, access telemetry, headers, or body limits.
2. Use FastAPI `Depends` for route-aware auth, tenant resolution, and injected use-case ports. Dependencies may reject before core; they do not become a general middleware framework.
3. Inject state through constructors, `app.state`, or typed dependency providers owned by lifespan; never module globals.
4. Document order: recovery/trace → request identity → auth/rate limit → access telemetry → route. Keep `/livez`, `/readyz`, and `/metrics` outside heavy/auth middleware as policy requires.
5. Short-circuit rejects with the standard error shape and do not call the downstream app.

## Invariants To Preserve

- No business decisions or direct database/client calls in middleware.
- Request-scoped state stays request-scoped; app resources are lifespan-owned.
- Response body wrapping remains streaming-safe and does not buffer unbounded content.
- Labels remain finite; credentials, payloads, and identities never enter logs or metrics.

## Proof

```bash
uv run pytest tests/api/http -k 'middleware or dependency'
uv run pytest tests/api/http -k 'reject or probe'
make verify
```

Prove wrapper order, one short-circuit path, exception shielding, and probe exclusion. Governing doc: [HTTP services](../services/http-services.md).
