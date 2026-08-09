# Security

Practical hardening rules for Python boundaries, dependencies, secrets, and security evidence.

## Default Approach

Validate and normalize untrusted data once at each boundary, authorize before effects, use safe library APIs, and fail closed. Pydantic v2 owns HTTP/config/message parsing; core receives plain typed values. Apply the [OWASP Input Validation Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Input_Validation_Cheat_Sheet.html) allowlist posture and keep limits explicit.

### Boundary Validation And Authorization

Bound body, collection, string, upload, decompression, pagination, and fan-out sizes before allocation or work. Validate syntax and semantics, normalize once, then map to domain types. A schema-valid value is not automatically authorized.

Authentication produces a typed principal at the adapter. Authorization is deny-by-default and checked for the operation, tenant, and resource before core effects. Core may enforce domain ownership rules through explicit principal/scope values; it never reads FastAPI request state. Authentication and authorization providers remain replaceable seams selected through [framework selection](../decisions/framework-selection.md).

### Injection And Process Execution

SQL uses SQLAlchemy expressions or `text()` with bound parameters. Dynamic identifiers map through an allowlist; values never enter SQL through f-strings, `%`, `.format()`, or concatenation.

Subprocesses receive an argv list, bounded input/output, timeout, checked return code, and a minimal environment. Async paths use `asyncio.create_subprocess_exec`; synchronous paths use `subprocess.run(..., check=True, timeout=...)`. `shell=True` with external data is forbidden. `shlex.quote()` formats a shell token; it is never the primary defense when an argument-list API exists. The OWASP [OS Command Injection Defense Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/OS_Command_Injection_Defense_Cheat_Sheet.html) prefers avoiding shell command construction.

### Deserialization

Treat serialized input as hostile. JSON is decoded within byte/depth/collection limits and parsed into an explicit boundary model. `pickle` and equivalent object reconstruction formats never cross a trust boundary. YAML uses `yaml.safe_load` only when YAML is required; unsafe `yaml.load` is forbidden. Untrusted XML uses `defusedxml`, with DTDs/external entities disabled according to OWASP [XXE prevention guidance](https://cheatsheetseries.owasp.org/cheatsheets/XML_External_Entity_Prevention_Cheat_Sheet.html).

Never choose a decoder from an untrusted type name, import path, or class discriminator. Versioned payloads map through an allowlisted contract registry.

### Files And Paths

Resolve the configured allowed root and candidate path with `Path.resolve()`, then require the candidate to be inside the root with `Path.is_relative_to()` before opening it. Reject absolute user paths, NULs, unexpected encodings, and symlink escapes. Recheck policy at the final operation; validation before an attacker-controlled filesystem change is not sufficient for a high-risk write.

Generate server-owned filenames, allowlist extensions/content types where relevant, bound size, and open with the narrowest permissions. Archive extraction validates every member path and total expanded size before writing.

### SSRF And Outbound Destinations

Outbound fetches use a destination allowlist derived from configuration, not an arbitrary caller URL. Parse with a URL library; require approved schemes, hosts, ports, and path policy. Resolve DNS and reject loopback, link-local, private, multicast, unspecified, and platform metadata destinations unless explicitly required. Revalidate every redirect and pin the client to a bounded redirect count; do not forward caller credentials to a new origin.

Network egress policy is the second boundary. Application checks do not replace firewall/service-mesh controls, and DNS rebinding requires connection-time enforcement by the platform or a reviewed resolver/transport design. OWASP's [SSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html) is the governing threat model.

### Secrets

Production secrets arrive at runtime as environment variables or mounted files from the platform. They are never committed, baked into images/build args, embedded in `uv.lock`, or fetched into an unmanaged process-global cache. `pydantic.SecretStr`/`SecretBytes` prevent casual representation, but code must call their reveal methods deliberately and must still avoid logs, errors, telemetry, debug endpoints, and reprs.

Validate required secret presence at startup while naming only the missing setting. Prefer short-lived credentials. Rotation defaults to a rolling restart; file re-read or a manager client requires an explicit lifecycle, bounded refresh failure behavior, and runbook. Generate tokens, nonces, and reset links with the stdlib `secrets` module, never `random`.

### Audit Logging

Audit records answer who did what, to which resource, when, from where, and with what result. Emit them at the action boundary for authentication success/failure, authorization denial, privileged/data-mutating operations, and permission/configuration/secret changes. Audit regulated reads only when policy requires them; indiscriminate read auditing buries the signal.

Use a dedicated structured logger and sink, separate from operational/access logs. Every record carries an aware UTC timestamp, principal and tenant, stable action, resource type/ID, allowed/denied and success/failure result, request/correlation ID, and safe source identity. Store facts and identifiers, never payloads, credentials, or PII. Retention, access control, integrity/tamper evidence, export, and failure behavior are explicit. If audit durability is required, a failed audit write must fail the protected action or enter a documented durable fallback; silently dropping it is forbidden.

### Supply Chain And Disclosure

Applications commit and review `uv.lock`; `make verify` runs `uv lock --check` and `uv run --with pip-audit pip-audit`. Every new package receives maintenance, license, typing, vulnerability, and transitive-impact review. Vulnerability exceptions name an owner, expiry, exposure analysis, and compensating control.

Externally facing services and published libraries ship `SECURITY.md` from [the template](../templates/security-policy.md), naming a private reporting path, supported versions, acknowledgement/triage expectations, and coordinated-disclosure posture. Never direct an unpatched report to a public issue.

## Common Mistakes And Forbidden Patterns

- Raw dictionaries, coerced strings, or unbounded collections passed beyond a trust boundary.
- Authorization that defaults allow, runs after effects, or trusts a caller-supplied tenant/resource owner.
- SQL interpolation, `shell=True`, command strings assembled from external data, or `shlex.quote()` treated as sufficient isolation.
- `pickle`, unsafe YAML, stdlib XML parsers on hostile data, or decoder/class selection from input.
- `Path.resolve()` without containment proof, archive extraction without member checks, or filename extension treated as content validation.
- Caller-controlled outbound URLs, redirect destinations, DNS results, or credential forwarding without SSRF controls.
- Secrets in source, `.env`, images, build args, logs, exceptions, reprs, traces, or audit payloads.
- Security tokens generated with `random`, homegrown cryptography, or comparison code where standard primitives exist.
- Audit events mixed with short-retention application logs, recording only successes, or silently dropped.
- An ignored `pip-audit` finding or dependency bot PR merged without lockfile review.

## Verification And Proof

```bash
uv lock --check
uv run --with pip-audit pip-audit
uv run pytest -k "auth or security or traversal or ssrf or audit"
make verify
```

Negative tests reject malformed/oversized input, unauthorized tenant/resource access, SQL/shell metacharacters, unsafe serialized input, traversal/symlink/archive escape, and disallowed/redirected/rebound outbound destinations. Scan artifacts, image history, logs, traces, and audit records for seeded secret/PII values. Rotate a credential end to end. Prove each protected action emits the complete audit schema to the dedicated sink and exercises the documented sink-failure policy.

Related: [configuration](../foundations/configuration.md), [database](../services/database.md), [data handling](data-handling.md), and [security review](../checklists/security-review.md).
