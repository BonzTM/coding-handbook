# Security Review Checklist

Review checklist for any security-sensitive change: a new boundary, auth or crypto code, secret handling, an outbound client, or a dependency or build change. Distills [../operations/security.md](../operations/security.md) into actionable gates.

## Input And Boundaries

- [ ] Every new or changed boundary (HTTP, gRPC, CLI, file, external callback) validates shape, size, and authorization before any core logic runs.
- [ ] Authorization is checked at the boundary, not buried in shared helpers where it can be silently skipped.
- [ ] Request and payload size limits are bounded; unbounded reads, decoders, or allocations from untrusted input are rejected.
- [ ] File and path operations constrain an allowed root and reject traversal; `filepath.Clean` alone is not treated as sufficient.
- [ ] Outbound clients constrain allowed destinations and set explicit timeouts to limit SSRF-style abuse (see [../recipes/add-external-client.md](../recipes/add-external-client.md) and [../operations/resilience.md](../operations/resilience.md)).

## Secrets

- [ ] No secret value appears in source, tests, examples, logs, panic output, debug endpoints, build args, or image layers.
- [ ] Required secret material is read in the `internal/config` load path and fails fast at startup if missing or empty, naming the secret without logging its value.
- [ ] A rotation path exists: the process picks up new material via signal re-read or rolling restart without a code redeploy, and the mechanism is documented in the runbook (per [../operations/security.md](../operations/security.md) ### Secrets).
- [ ] No real `.env` is committed; only `.env.example` with placeholder values.

## Crypto And Auth

- [ ] Crypto uses standard-library or well-known primitives; no homegrown crypto, token signing, or session code.
- [ ] Authorization logic stays at the boundary and is not mixed into unrelated helper packages.
- [ ] Comparison of secrets, tokens, or MACs uses constant-time comparison where timing matters.
- [ ] Any `unsafe` usage has a documented, measured justification and review.

## Supply Chain

- [ ] `go tool govulncheck ./...` is clean, or each finding has a documented, justified exception.
- [ ] Every newly introduced non-trivial dependency was reviewed for maintenance health and security posture (see [../recipes/bump-dependency.md](../recipes/bump-dependency.md)).
- [ ] Tool-only dependencies are tracked via `go.mod` `tool` directives (`go get -tool`, run with `go tool`); no `tools.go` blank-import pattern.
- [ ] No `replace` directive is committed pointing at a local or fork path.
- [ ] A PRIVATE vulnerability report path exists (GitHub security advisory or `security@<org>` alias) — externally-facing services and published libraries ship a `SECURITY.md` from the [security-policy template](../templates/security-policy.md); internal-only services name an owner and private triage path in the runbook (per [../operations/security.md](../operations/security.md) ### Vulnerability Disclosure). No report path is itself a finding.

## Build Hardening

- [ ] Release builds use `-trimpath`; no local filesystem paths are embedded in the artifact.
- [ ] Binary is built pure-Go (`CGO_ENABLED=0`); any cgo dependency is ADR-justified.
- [ ] No secret material or embedded paths are present in the final artifact or image layers.

## Verification

```bash
go tool govulncheck ./...
make verify
```

- Targeted negative tests pass for the affected boundary: rejected authorization, rejected malformed or oversized input, rejected path traversal, and rejected disallowed outbound destination.
- Image and log scan shows no secret values in any layer, build arg, or emitted log line.
- Startup fails fast and names the missing secret when required material is absent or empty.
