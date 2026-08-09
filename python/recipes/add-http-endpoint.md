# Recipe: Add HTTP Endpoint

Use this when a feature adds or changes one FastAPI route.

## Files To Touch

- `src/<app>/api/http/<resource>.py` for the router and Pydantic DTOs
- `src/<app>/core/` for the use case and domain types
- composition/router registration and transport/core tests

## Steps

1. Define strict request and response DTOs in `api/http`; map them once to plain core values.
2. Add or update one core use case behind a core-owned `Protocol` when an adapter is required.
3. Keep the route thin: validate, authorize, call core, map domain failures, return the declared response.
4. Register the router with a stable prefix, operation ID, status code, and response model.
5. Confirm existing middleware/dependencies cover request identity, auth, logging, tracing, and metrics.

## Invariants To Preserve

- Routers never query SQLAlchemy or call HTTPX directly; core never imports FastAPI or Pydantic DTOs.
- Request bodies and collection sizes are bounded; external input is parsed before core runs.
- Cancellation and deadlines flow to I/O; status/error mappings remain contract-compatible.
- Route metric labels use route templates and status classes, never raw paths or identities.

## Proof

```bash
uv run pytest tests/api/http -k '<route_or_use_case>'
uv run pytest tests/api/http -k 'validation or authorization'
make verify
```

Smoke the route against a local server and inspect its OpenAPI operation, structured error, log, metric, and trace. Governing doc: [HTTP services](../services/http-services.md).
