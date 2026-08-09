# Checklist: Security Review

Apply the OWASP-aligned controls in [security](../operations/security.md) to the changed threat surface.

## Threats And Boundaries

- [ ] Assets, actors, trust boundaries, abuse cases, and accepted residual risk are documented.
- [ ] HTTP, env, database, message, file, browser storage, and third-party input is parsed with Zod.
- [ ] Byte, field, nesting, count, pagination, concurrency, time, and response bounds are tested.
- [ ] SQL uses parameters; paths, subprocess arguments, redirects, and outbound destinations are constrained.
- [ ] Unsafe HTML is absent or passes one reviewed sanitizer with XSS tests.

## Identity And Browser

- [ ] Authentication validates protocol-required issuer, audience, signature, time, algorithm, and key rotation.
- [ ] Every resource action is authorized server-side with deny-by-default and tenant scope.
- [ ] Horizontal, vertical, stale-privilege, bulk, and cross-tenant negative tests exist.
- [ ] Cookie, CSRF, CORS, CSP, framing, MIME, referrer, and transport policy fit the deployed topology.
- [ ] Frontend route guards and hidden controls are documented as UX, not authorization.

## Secrets, Data, And Supply Chain

- [ ] Secrets are absent from source, history policy, images, build args, `VITE_*`, logs, traces, errors, and fixtures.
- [ ] Data classification, minimization, retention, deletion, access, and telemetry treatment are proven.
- [ ] Required audit events reach a separate controlled sink with safe fields and failure policy.
- [ ] Lockfile, transitive graph, install scripts, licenses, maintenance, and advisories are reviewed.
- [ ] `npm audit` passes or each exception has owner, reachability rationale, control, expiry, and issue.

## Operations And Proof

- [ ] Least-privilege runtime, non-root image, read-only filesystem, and secret injection are verified.
- [ ] Rate limit, timeout, cancellation, overload, retry, and idempotency cannot amplify abuse.
- [ ] Private vulnerability reporting and supported versions are current.
- [ ] Negative security tests and secret/artifact scans are linked.
- [ ] Reviewer records approval, conditions, and unresolved owned risks.
