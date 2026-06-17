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

### Outbound HTTP Transport

For an HTTP upstream, never use `http.DefaultClient`, `http.Get`, or any package-level helper: they share one process-global client with no timeout and a transport tuned for nobody. Construct exactly one `*http.Client` per upstream, store it on the client struct, and reuse it for every request — `http.Client` and `http.Transport` are safe for concurrent use and pool connections, so a per-request client defeats keep-alives and leaks file descriptors.

Set both layers of deadline. The client gets an explicit `Timeout` as a backstop covering the whole request including body read; every call also carries the caller's `context` deadline so cancellation propagates and the timeout is per-operation, not just per-client. Tune a non-default `http.Transport` so a slow upstream cannot stall a connection forever:

```go
func newUpstreamClient(base string) *Client {
	t := &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,              // default is 2 — too low for a busy upstream
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second, // cap time-to-first-byte
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &Client{
		base: base,
		http: &http.Client{Transport: t, Timeout: 30 * time.Second},
	}
}
```

Build requests with `http.NewRequestWithContext(ctx, ...)`, never `http.NewRequest`. On every response, `defer resp.Body.Close()` and drain the body to EOF (`io.Copy(io.Discard, resp.Body)`) before closing so the connection returns to the pool instead of being dropped — the `bodyclose` linter (per [../quality/linting.md](../quality/linting.md)) enforces the close. Layer retries, backoff, and circuit breakers on top of this client per [../operations/resilience.md](../operations/resilience.md#composition-order); the transport handles connection hygiene, not policy.

## Invariants To Preserve

- caller context and deadlines flow into every outbound request
- secrets stay out of logs
- SSRF or arbitrary-destination risk is constrained by config and validation
- retries are bounded and idempotent-safe
- no `http.DefaultClient` / `http.Get`; one reused `*http.Client` per upstream with an explicit `Timeout` and a tuned non-default `Transport`
- every response body is drained and closed exactly once

## Proof

- client tests with `httptest.Server` or protocol-specific test server
- timeout and cancellation tests, including an `httptest.Server` that stalls (`time.Sleep` past the deadline) to prove the client `Timeout` and context deadline fire instead of hanging
- a test or review check asserting the client carries its own `*http.Client` and does not fall back to `http.DefaultClient`
- `bodyclose` clean under `make lint` — proof every response body is closed
- negative tests for auth failures and malformed upstream responses
- config validation proving required client settings fail fast at startup

If the external system interaction is asynchronous publish or subscribe rather than request/response, use [add-event-publisher.md](add-event-publisher.md) or [add-event-consumer.md](add-event-consumer.md) instead.
