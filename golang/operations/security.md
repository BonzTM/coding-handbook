# Security

Practical hardening guidance for Go repos that need to reduce attack surface without becoming security theater.

## Default Approach

- Validate input at the boundary: HTTP, gRPC, CLI, files, and external callbacks.
- Keep secrets out of source, logs, and panic output.
- Prefer well-known standard-library crypto and transport primitives over bespoke wrappers.
- Run official Go supply-chain tooling as part of normal development and release flow.

### Boundary Rules

- HTTP and gRPC handlers validate shape, size, and authorization before calling core logic.
- File and path operations must normalize and constrain allowed roots; never trust user-supplied paths.
- External clients should constrain destinations and timeouts to reduce SSRF-style abuse.

### Secrets

Define both PROVENANCE (where the material comes from) and ROTATION (how new material reaches a running process) for every secret. "Keep secrets out of source and logs" is necessary but not sufficient.

- Provenance: in production, secrets are injected at runtime as environment variables or mounted files from an external secrets manager / KMS / vault. The app reads the injected material; it does not embed plaintext in source, image, or build args, and does not fetch-and-cache long-lived plaintext itself. Route the specific manager (Vault, cloud KMS/Secrets Manager, sealed config) to [decisions/framework-selection.md](../decisions/framework-selection.md).
- Validate required secret material at startup and fail fast if it is missing or empty, in the same `internal/config` load path as the rest of configuration (see [foundations/configuration.md](../foundations/configuration.md)). Report which secret is missing by name; never log its value.
- Prefer short-lived, rotatable credentials over static long-lived ones. A credential that can be rotated cheaply is one you can revoke after a leak.
- Rotation: if rotation matters for a secret, the process must not pin it for its whole lifetime. Make new material reachable either by re-reading the mounted file / env on a signal (e.g. `SIGHUP`) or by a rolling restart that re-runs the startup load. Decide the mechanism per secret and document it in the [runbook](../templates/runbook.md); rolling restart is the default unless re-read on signal is required. See [operations/deployment.md](deployment.md) for how injection and rolling restarts are wired.
- Never log, echo into exec output, or include secret values in panic dumps, debug endpoints, or error messages. Wrap or redact before anything reaches a sink.

### Supply-Chain Rules

- run `govulncheck ./...`
- keep Go and direct dependencies on supported, patched versions
- track tool-only dependencies with `go.mod` `tool` directives (add via `go get -tool`, run via `go tool`); the legacy `tools.go` / `//go:build tools` blank-import pattern is obsolete as of Go 1.24
- review new dependencies for maintenance health and security posture before adoption

### Vulnerability Disclosure

Decide HOW a vulnerability reaches you before one does, so a reporter never has to choose between a public issue and silence.

- Every externally-facing service and every published library ships a `SECURITY.md` (root or `.github/`) from the [security-policy template](../templates/security-policy.md). Internal-only services may ship a trimmed version (or skip the file) but must still name an owner and a private triage path (in the [runbook](../templates/runbook.md) or service README).
- Give a PRIVATE report path — a GitHub private security advisory or a `security@<org>` alias — and say plainly: no public issue, PR, or discussion for an unpatched vulnerability. A missing report path is itself a risk; reporters fall back to public issues.
- Commit to an acknowledgement and triage SLA (e.g. ack within 2 business days, triage within 5) and keep the supported-versions table honest about what you actually patch.
- Hold a coordinated-disclosure posture: fix privately, agree an embargo/disclosure date with the reporter, then publish an advisory (with a CVE where applicable) and credit the reporter. Pair the fix with the dependency response in [recipes/bump-dependency.md](../recipes/bump-dependency.md) and the [dependency-upgrade checklist](../checklists/dependency-upgrade.md) when the vuln is in a dependency.

### Build Hardening

- use `-trimpath`
- keep `-buildvcs` enabled for traceability unless reproducibility requirements force an explicit alternative
- prefer static binaries when cgo is unnecessary

## Common Mistakes And Forbidden Patterns

- secrets committed anywhere in the repo, including examples or tests
- a real secret baked into image layers or passed as a Docker `ARG`/build arg (it persists in layer history even if a later layer removes it)
- a secret value logged in startup, debug, or error output, or dumped on panic
- no rotation path, so a leaked credential is valid forever and cannot be cheaply revoked
- `.env` files used outside local development, or a real `.env` committed (commit `.env.example` only)
- auth logic mixed into unrelated helper packages
- path handling that assumes `filepath.Clean` alone makes untrusted input safe
- homegrown crypto or token code when the stdlib already solves it
- `unsafe` usage without a documented, measured justification and review
- publicly disclosing an unpatched vulnerability (issue, PR, commit message, or release note) before a fix ships
- no private report path, so a reporter's only option is a public issue that exposes the flaw

## Verification And Proof

- `govulncheck ./...`
- targeted negative tests for auth, validation, path traversal, or outbound-client restrictions
- release build audit showing no local paths or secret material embedded in the artifact
- image and log scan confirming no secret values appear in any layer, build arg, or emitted log line
- rotation exercised end to end: rotate the credential and confirm the running process picks up the new value (via signal re-read or rolling restart) without a redeploy of code
- startup fails fast and names the missing secret when required material is absent or empty
- dependency review for every newly introduced non-trivial package
- `SECURITY.md` present for externally-facing services and published libraries, naming a private report path (no public issue) and an acknowledgement/triage SLA
- disclosure path known and tested: a reporter (or internal tester) can reach the private channel and gets the documented acknowledgement
