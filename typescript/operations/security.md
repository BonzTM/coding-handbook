# Security

OWASP-anchored security rules for TypeScript services and React applications.

## Default Approach

Use deny-by-default authorization, parsed trust boundaries, least privilege, secure deployment defaults, and threat-driven negative tests.

The baseline follows the [OWASP Application Security Verification Standard](https://owasp.org/www-project-application-security-verification-standard/) and relevant [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/). Project threat models and platform policy may strengthen it.

### Trust Boundaries

Parse HTTP, environment, database rows, messages, files, browser storage, and third-party responses with Zod before business use. Enforce byte, field, depth, count, pagination, and time bounds before expensive work.

Use parameterized SQL and context-appropriate output encoding. Constrain file paths to owned roots, validate uploads by content and policy, and never build shell commands from input.

### Authentication And Sessions

Use the organization's proven identity provider and protocol library. Validate issuer, audience, signature, expiry, not-before, algorithm, and key rotation according to the protocol. Do not write custom token or cryptographic schemes.

Cookie sessions use `Secure`, `HttpOnly`, appropriate `SameSite`, narrow path/domain, rotation, and CSRF protection for state changes. Regenerate or revoke sessions after authentication and privilege changes.

Rate-limit authentication paths without creating an account-enumeration oracle. Error messages and timing do not reveal whether a principal exists beyond the accepted product contract.

### Authorization

Authorize every action and resource server-side after authentication. Default deny and derive tenant/resource scope from trusted identity plus loaded resource, not from client-supplied ownership fields.

Central policy may support decisions, but enforcement remains at each boundary. Frontend route guards and hidden controls are UX only, never authorization.

Test horizontal and vertical privilege escalation, cross-tenant identifiers, stale privileges, and bulk endpoints. Cache authorization only with explicit identity, resource, policy-version, expiry, and invalidation semantics.

### Secrets And Cryptography

Secrets are runtime-only, least-privilege, rotatable, and excluded from source, images, build arguments, frontend bundles, logs, traces, snapshots, and error responses. Validate presence at startup without printing values.

Use platform TLS and established cryptographic libraries with approved algorithms and key management. Encryption requires a threat, key owner, rotation, recovery, and integrity design; encoding or hashing is not encryption.

### Outbound Requests And SSRF

Centralize outbound URL resolution. Prefer configured origins and allowlisted destinations. Validate scheme, hostname, port, redirect destinations, resolved address classes, credentials, timeout, response size, and content type.

Do not accept arbitrary URLs for server-side fetch. Protect against DNS and redirect changes; network egress controls provide defense in depth.

### Browser Security

Use semantic encoding and avoid unsafe HTML. If rich HTML is required, sanitize through one reviewed adapter and test known XSS shapes. Configure CSP without routine `unsafe-inline`/`unsafe-eval`, plus frame, MIME-sniffing, referrer, and transport controls appropriate to deployment.

CORS is a narrow origin allowlist and never authentication. Vite variables and source-delivered JavaScript are public. Do not store bearer tokens or sensitive records in localStorage by default.

### Dependency And Supply Chain

Commit the lockfile, use `npm ci`, review install scripts and transitive changes, and run the repository `npm audit` policy. Pin CI actions and container bases according to platform policy; produce provenance/SBOM where release requirements demand it.

An audit advisory is triaged for reachability, severity, exploitability, and fix. A temporary exception has an owner, rationale, compensating control, expiry, and tracking issue. Never silence the audit globally.

### Audit Logging

Audit records answer who did what, to which resource, when, from where, and with what result. They are separate from operational/access logs.

- Emit events for authentication success/failure, authorization denial, privileged or sensitive mutations, permission/config/secret changes, and regulated reads when policy requires them.
- Use a stable schema with principal and tenant, action, target type/ID, UTC time, result, request/trace reference, and source context.
- Record denial and failure as well as success.
- Send audit events to a separate access-controlled sink with defined retention and, where required, append-only or tamper-evident guarantees.
- Record facts and identifiers, not secret values or sensitive payloads.
- Define delivery failure behavior; a critical audit event must not disappear silently.

Audit volume, retention, access, export, and deletion follow legal/privacy requirements. Operational logs cannot substitute because their sampling, access, and lifecycle differ.

### Vulnerability Disclosure

Externally facing services and published libraries provide `SECURITY.md` with supported versions, a private reporting path, acknowledgement/triage expectations, and coordinated disclosure policy. Do not require reporters to open a public issue for an unpatched flaw.

## Common Mistakes And Forbidden Patterns

- Compile-time types, ORM models, or client validation treated as boundary security.
- Authentication without resource-level authorization.
- CORS, hidden UI, or opaque IDs presented as access control.
- Secrets in logs, images, build args, `.env`, frontend config, or snapshots.
- User-controlled outbound URLs, SQL interpolation, shell execution, or unsafe HTML.
- Homegrown crypto, tokens, password storage, or sanitizer.
- Audit records mixed into short-retention operational logs.
- Audit payloads containing secrets or raw PII.
- `npm audit` disabled or exceptions without owners and expiry.

## Verification And Proof

- Threat review identifies assets, trust boundaries, abuse cases, and mitigations for sensitive changes.
- Negative tests cover malformed/oversized input, authn, authz, cross-tenant access, injection, XSS, CSRF, SSRF, and replay as applicable.
- Secret scans cover source, history policy, artifacts, images, source maps, logs, and fixtures.
- Dependency/lockfile review and `npm audit` policy pass or record a time-bounded exception.
- Cookie, CSP, CORS, TLS/proxy, and security-header behavior is tested in the deployed topology.
- Every relevant security action emits a separate audit record with who/what/when/where/result and no sensitive payload.
- Rotation and private vulnerability-reporting paths are exercised.

Related: [data-handling.md](data-handling.md), [../foundations/configuration.md](../foundations/configuration.md), and [../services/http-services.md](../services/http-services.md).
