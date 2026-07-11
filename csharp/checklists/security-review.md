# Security Review Checklist

Review checklist for any security-sensitive change: a new boundary, auth or crypto code, secret handling, an outbound client, or a dependency or build change. Distills [../operations/security.md](../operations/security.md) into actionable gates.

## Input And Boundaries

- [ ] Every new or changed boundary (HTTP, gRPC, CLI, file, external callback) validates shape, size, and authorization before any core logic runs.
- [ ] Authorization is checked at the boundary — `RequireAuthorization` on the route group or an endpoint filter — not buried in shared helpers where it can be silently skipped.
- [ ] Request and payload size limits are bounded (Kestrel `MaxRequestBodySize`, explicit limits on file and stream reads); unbounded reads, deserialization, or allocations from untrusted input are rejected.
- [ ] File and path operations constrain an allowed root and reject traversal; `Path.GetFullPath` alone is not treated as sufficient — the resolved path is verified to stay under the root with an ordinal comparison.
- [ ] Outbound clients constrain allowed destinations and set explicit timeouts to limit SSRF-style abuse (see [../recipes/add-external-client.md](../recipes/add-external-client.md) and [../operations/resilience.md](../operations/resilience.md)).

## Secrets

- [ ] No secret value appears in source, tests, examples, logs, exception output, debug endpoints, build args, or image layers.
- [ ] Required secret material is bound through the options load path and fails fast at startup (`ValidateOnStart`) if missing or empty, naming the configuration key without logging its value.
- [ ] A rotation path exists: the process picks up new material via configuration reload or rolling restart without a code redeploy, and the mechanism is documented in the runbook (per [../operations/security.md](../operations/security.md) ### Secrets).
- [ ] No secret is committed anywhere: `appsettings.json` stays secret-free, local dev uses `dotnet user-secrets`, runtime uses env vars or mounted files.

## Crypto And Auth

- [ ] Crypto uses `System.Security.Cryptography`, ASP.NET Core Data Protection, or the JwtBearer handler; no homegrown crypto, token signing, or session code.
- [ ] Authorization logic stays at the boundary and is not mixed into unrelated helper classes.
- [ ] Comparison of secrets, tokens, or MACs uses `CryptographicOperations.FixedTimeEquals` where timing matters.
- [ ] Any `unsafe` block (or `AllowUnsafeBlocks` in a project file) has a documented, measured justification and review.

## Supply Chain

- [ ] `dotnet list package --vulnerable --include-transitive` is clean, and NuGetAudit (`NuGetAuditMode=all`) passes on restore, or each finding has a documented, justified exception.
- [ ] Every newly introduced non-trivial dependency was reviewed for maintenance health and security posture (see [../recipes/bump-dependency.md](../recipes/bump-dependency.md)), pinned centrally in `Directory.Packages.props`, and reflected in committed lock files.
- [ ] No floating versions (`*` or open ranges) and no package source added outside the committed `nuget.config`; restore runs `--locked-mode` in CI.
- [ ] A PRIVATE vulnerability report path exists (GitHub security advisory or `security@<org>` alias) — externally-facing services and published libraries ship a `SECURITY.md` from the [security-policy template](../templates/security-policy.md); internal-only services name an owner and private triage path in the runbook (per [../operations/security.md](../operations/security.md) ### Vulnerability Disclosure). No report path is itself a finding.

## Build Hardening

- [ ] Release builds are deterministic (`ContinuousIntegrationBuild=true` in CI); no local filesystem paths are embedded in the artifact.
- [ ] The container runs non-root on the chiseled ASP.NET runtime image; no SDK, shell, or package manager in the final stage.
- [ ] No secret material or embedded paths are present in the final artifact or image layers.

## Verification

```powershell
dotnet list package --vulnerable --include-transitive
pwsh ./verify.ps1
```

- Targeted negative tests pass for the affected boundary: rejected authorization, rejected malformed or oversized input, rejected path traversal, and rejected disallowed outbound destination.
- Image and log scan shows no secret values in any layer, build arg, or emitted log line.
- Startup fails fast and names the missing configuration key when required material is absent or empty.
