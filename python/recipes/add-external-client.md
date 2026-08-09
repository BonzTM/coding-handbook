# Recipe: Add External Client

Use this when the application calls an outbound HTTP service.

## Files To Touch

- core-owned consumer `Protocol`
- `src/<app>/clients/<upstream>.py`
- settings/lifespan composition and fake/stub-server tests

## Steps

1. Define the port from the consumer's domain perspective; return domain values, not HTTPX responses.
2. Create one lifespan-owned `httpx.AsyncClient` with validated base URL, bounded limits, and explicit `httpx.Timeout` values.
3. Map request/response schemas inside the adapter and close/consume responses deterministically.
4. Add a small bounded exponential-backoff loop with full jitter, explicit retryable outcomes, and injected sleep/random seams.
5. Constrain destinations against the SSRF policy; instrument attempt, latency, outcome, and trace propagation without secrets.

## Invariants To Preserve

- No per-call client, implicit timeout, unbounded body, or retry of unsafe effects without idempotency.
- Caller cancellation and deadline remain authoritative.
- Redirects and operator-supplied destinations cannot escape the approved host/scheme policy.
- Core imports neither HTTPX nor upstream wire DTOs.

## Proof

```bash
uv run pytest tests/clients -k '<upstream>'
uv run pytest tests/clients -k 'timeout or cancel or retry or malformed'
make verify
```

Use an in-process transport or stub server and prove timeout, cancellation, retry exhaustion, auth failure, malformed/oversized response, and SSRF rejection. Governing doc: [resilience](../operations/resilience.md).
