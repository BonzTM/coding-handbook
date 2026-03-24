# Recipe: Add External Client

Use this when the repo needs to call another HTTP or gRPC service.

## Files To Touch

- `internal/client/<name>/...`
- the consuming package in `internal/core` or another adapter
- config docs and loader if new endpoints, credentials, or timeouts are introduced
- client tests using a test server

## Steps

1. Define the dependency seam from the consumer's perspective.
2. Implement the client with explicit base URL, auth, timeout, and retry policy.
3. Add request and response mapping in the client package only.
4. Instrument requests with logs, metrics, and tracing at the client boundary.
5. Decide what is retryable and what is terminal before adding automatic retries.

## Invariants To Preserve

- caller context and deadlines flow into every outbound request
- secrets stay out of logs
- SSRF or arbitrary-destination risk is constrained by config and validation
- retries are bounded and idempotent-safe

## Proof

- client tests with `httptest.Server` or protocol-specific test server
- timeout and cancellation tests
- negative tests for auth failures and malformed upstream responses
- config validation proving required client settings fail fast at startup

If the external system interaction is asynchronous publish or subscribe rather than request/response, use [add-event-publisher.md](add-event-publisher.md) or [add-event-consumer.md](add-event-consumer.md) instead.
