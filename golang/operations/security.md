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

### Supply-Chain Rules

- run `govulncheck ./...`
- keep Go and direct dependencies on supported, patched versions
- isolate tool-only dependencies in `tools.go`
- review new dependencies for maintenance health and security posture before adoption

### Build Hardening

- use `-trimpath`
- keep `-buildvcs` enabled for traceability unless reproducibility requirements force an explicit alternative
- prefer static binaries when cgo is unnecessary

## Common Mistakes And Forbidden Patterns

- secrets committed anywhere in the repo, including examples or tests
- auth logic mixed into unrelated helper packages
- path handling that assumes `filepath.Clean` alone makes untrusted input safe
- homegrown crypto or token code when the stdlib already solves it
- `unsafe` usage without a documented, measured justification and review

## Verification And Proof

- `govulncheck ./...`
- targeted negative tests for auth, validation, path traversal, or outbound-client restrictions
- release build audit showing no local paths or secret material embedded in the artifact
- dependency review for every newly introduced non-trivial package
