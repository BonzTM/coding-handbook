# Recipe: Add External Client

Use this when a backend calls an external HTTP service.

## Files To Touch

- `src/lib/http/<name>-client.ts`
- the consumer-owned port under `src/core/`
- config schema and `.env.example`
- local-server client tests
- telemetry, runbook, and dependency contract fixtures

## Steps

1. Define the port from the consuming use case, not from the vendor SDK.
2. Parse and allowlist the configured origin; reject credentials and unsafe schemes.
3. Resolve owned relative paths against that origin; never accept an arbitrary request URL.
4. Compose caller cancellation with `AbortSignal.timeout` and preserve abort classification.
5. Bound redirects, response bytes, content type, retries, attempts, and concurrency.
6. Validate every success and problem body from `unknown` with Zod.
7. Retry only classified transient failures when the operation is safe to repeat.
8. Emit one boundary log and bounded route/dependency/result telemetry.
9. Inject the client through composition and close any owned resources on shutdown.

```bash
npm test -- --runInBand src/lib/http/<name>-client.test.ts
npm run typecheck
npm run verify
```

## Invariants To Preserve

- Core imports no fetch, URL-policy, vendor, or transport types.
- Caller cancellation and the remaining deadline flow to every attempt.
- Redirects and DNS/address policy cannot escape the approved destination set.
- Credentials, full URLs, raw bodies, and identifiers stay out of metrics and logs.
- Unknown or malformed responses never enter domain logic.
- Retry amplification is bounded at one owned layer.

## Proof

- A local HTTP server proves request, auth, timeout, abort, redirect, and size policy.
- Tests cover malformed JSON, wrong content type, non-success status, and retry exhaustion.
- SSRF tests reject loopback/private/link-local or unapproved destinations as policy requires.
- Captured telemetry has bounded attributes and safe values.
- `npm run verify` is green.
