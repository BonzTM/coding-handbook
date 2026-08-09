# Security Review Checklist

Gate for a new trust boundary, authentication/authorization, secret, parser, file/process operation, outbound client, dependency, or build change.

## Input And Boundaries

- [ ] Every HTTP/gRPC/message/CLI/file/callback boundary bounds size and parses into an explicit Pydantic/protobuf/typed contract before core work.
- [ ] Authentication yields a typed principal; authorization is deny-by-default and checks operation, tenant, and resource before effects.
- [ ] SQL values are parameters and dynamic identifiers are allowlisted; no f-string/format/concatenated SQL.
- [ ] Subprocesses use argv APIs, timeouts, checked exits, bounded output, and never `shell=True` with external data.
- [ ] `pickle`, unsafe YAML, and unsafe XML are absent from trust boundaries; XML uses `defusedxml` and YAML uses `safe_load` only when required.
- [ ] Paths resolve beneath an allowed root and tests cover absolute, `..`, symlink, archive, encoding, and race-relevant cases.
- [ ] Outbound clients constrain scheme/host/port/path, DNS/IP ranges, redirects, credential forwarding, egress, and explicit timeouts against SSRF.

## Secrets And Tokens

- [ ] Secret provenance is runtime environment or mounted file; no value appears in source, `.env`, lockfile, test, log, error, repr, telemetry, build arg, or image.
- [ ] Pydantic settings fails startup for missing/invalid secret while revealing only the key name; `SecretStr` is not treated as complete redaction.
- [ ] Rotation through rolling restart or bounded re-read is documented and proven.
- [ ] Tokens/nonces use `secrets`; cryptographic and comparison operations use maintained standard primitives, not homegrown code.

## Audit And Data

- [ ] Required authn/authz/privileged/data events emit who/action/resource/time/source/result to a dedicated access-controlled audit sink.
- [ ] Audit events include denials/failures and contain identifiers only—no credentials, payloads, or PII.
- [ ] Data classification, minimization, retention, deletion/export, encryption hops, and synthetic-test-data rules cover every new field/store.
- [ ] Audit sink failure follows the documented fail-closed or durable-fallback policy; evidence is not silently dropped.

## Supply Chain And Disclosure

- [ ] `pyproject.toml` and `uv.lock` diffs are understood; new dependencies were reviewed for maintenance, license, typing, vulnerabilities, and transitive/build impact.
- [ ] `pip-audit` is clean or each exception has owner, expiry, exposure analysis, and compensating control.
- [ ] Build/release workflows pin trusted actions/images, isolate untrusted PRs, use least permissions, and keep publish identity unavailable to PR code.
- [ ] External services/published libraries have a private `SECURITY.md` reporting path and coordinated-disclosure expectations.

## Verification

```bash
uv lock --check
uv run --with pip-audit pip-audit
uv run pytest -k "auth or security or traversal or ssrf or audit"
make verify
```

- [ ] Negative tests reject malformed/oversized input, authorization bypass, injection, unsafe decode, traversal, and disallowed outbound destinations.
- [ ] Seeded canary secrets/PII are absent from logs, traces, metrics, audit events, failure output, wheel/sdist, image history, and layers.
- [ ] Credential rotation and audit sink failure were exercised end to end.
