# Security

Practical hardening guidance for .NET repos that need to reduce attack surface without becoming security theater.

## Default Approach

- Validate input at the boundary: HTTP, gRPC, CLI, files, and external callbacks.
- Authenticate with JWT bearer tokens validated against the issuer's JWKS; authorize with deny-by-default policies.
- Keep secrets out of source, logs, and exception output.
- Prefer platform primitives (ASP.NET Core auth, Data Protection, `System.Security.Cryptography`) over bespoke wrappers.
- Run NuGet supply-chain tooling as part of normal development and release flow.
- Browser-facing security headers (CSP, HSTS, frame options) and antiforgery live in [../services/web-apps.md](../services/web-apps.md); this doc governs everything behind them.

### Boundary Rules

- Endpoints validate shape, size, and authorization before calling core logic. Request DTOs carry DataAnnotations enforced by the built-in minimal-API validation (see [../services/http-services.md](../services/http-services.md)); Kestrel body-size limits stay on.
- File and path operations normalize and constrain allowed roots; never trust a user-supplied path:

```csharp
public static string ResolveUnder(string root, string userPath)
{
    string fullRoot = Path.GetFullPath(root);
    string full = Path.GetFullPath(Path.Combine(fullRoot, userPath));
    return full.StartsWith(fullRoot + Path.DirectorySeparatorChar, StringComparison.Ordinal)
        ? full
        : throw new UnauthorizedAccessException("path escapes allowed root");
}
```

  Ordinal comparison is correct for the Linux containers we ship; see [../foundations/cross-platform.md](../foundations/cross-platform.md) for Windows-dev caveats.
- Outbound clients constrain destinations and timeouts to reduce SSRF-style abuse: typed clients get a fixed `BaseAddress` from validated config (see [../recipes/add-external-client.md](../recipes/add-external-client.md)); never build an outbound URL from user input. Where user-supplied URLs are the feature, allow-list scheme and host and reject private/link-local address ranges after resolution.

### Authentication

JWT bearer is the service default. The handler fetches signing keys from the issuer's OIDC discovery document (JWKS) and refreshes them automatically — never embed or hand-roll key material.

```csharp
builder.Services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
    .AddJwtBearer(o =>
    {
        o.Authority = authOptions.Authority;   // OIDC discovery + JWKS live here
        o.TokenValidationParameters.ValidAudience = authOptions.Audience;
        o.MapInboundClaims = false;            // keep raw JWT claim names
    });
```

- Always set and validate the audience; a token minted for another API must not work here.
- `RequireHttpsMetadata` stays at its secure default outside local development.
- Service-to-service callers use short-lived tokens (client credentials), never shared static API keys.

### Authorization

Deny by default: a fallback policy requires an authenticated user on every endpoint, and anonymous access is an explicit, reviewable opt-in.

```csharp
builder.Services.AddAuthorizationBuilder()
    .SetFallbackPolicy(new AuthorizationPolicyBuilder().RequireAuthenticatedUser().Build())
    .AddPolicy("orders:write", p => p.RequireRole("orders-admin"));
```

- Name policies after capabilities (`orders:write`), attach them to route groups with `RequireAuthorization`, and keep role/claim mapping in one place.
- Ownership and tenancy checks are resource-based: an `AuthorizationHandler<TRequirement, TResource>` invoked via `IAuthorizationService.AuthorizeAsync(user, order, requirement)` after the resource loads — never a string comparison scattered through endpoint bodies.
- Authorization decisions live in handlers and policies, not in unrelated helper classes; a denial returns 403 through the standard `ProblemDetails` envelope and emits an audit event (below).

### Secrets

Define both PROVENANCE (where the material comes from) and ROTATION (how new material reaches a running process) for every secret. "Keep secrets out of source and logs" is necessary but not sufficient.

- Provenance: in production, secrets are injected at runtime as environment variables or mounted files from an external secrets manager / KMS / vault (env-var configuration overrides and `AddKeyPerFile` are the two supported shapes). The app reads the injected material; it never embeds plaintext in source, image, or build args, and never appears in `appsettings.json`. `dotnet user-secrets` is for local development only. Route the specific manager (Vault, cloud KMS/Secrets Manager, sealed config) to [../decisions/framework-selection.md](../decisions/framework-selection.md).
- Validate required secret material at startup and fail fast if it is missing or empty, through the same options-pattern load path as the rest of configuration (`ValidateDataAnnotations` + `ValidateOnStart`; see [../foundations/configuration.md](../foundations/configuration.md)). Report which secret is missing by name; never log its value.
- Prefer short-lived, rotatable credentials over static long-lived ones. A credential that can be rotated cheaply is one you can revoke after a leak.
- Rotation: if rotation matters for a secret, the process must not pin it for its whole lifetime. Make new material reachable either by configuration reload (`reloadOnChange` on mounted files + `IOptionsMonitor<T>`) or by a rolling restart that re-runs the startup load. Decide the mechanism per secret and document it in the [runbook](../templates/runbook.md); rolling restart is the default unless live re-read is required. See [deployment.md](deployment.md) for how injection and rolling restarts are wired.
- Never log, serialize, or include secret values in exception messages, `ProblemDetails` responses, debug endpoints, or crash dumps. Redact before anything reaches a sink (see [data-handling.md](data-handling.md)).

### Data Protection

Anything the service must hold encrypted at rest under its own key — auth cookies, antiforgery tokens, password-reset and email-confirmation tokens, stored refresh tokens — goes through the ASP.NET Core Data Protection API, not hand-rolled AES.

- Set `SetApplicationName` explicitly and persist the key ring to shared storage when the service runs more than one replica; the in-container filesystem default silently breaks cookie and token round-trips across instances and restarts.
- Protect the persisted key ring itself (KMS/certificate protection per platform); route the concrete store to [../decisions/framework-selection.md](../decisions/framework-selection.md).
- Use purpose strings (`CreateProtector("Orders.RefreshTokens")`) so a payload protected for one purpose cannot be unprotected by another.
- Data Protection is for the service's own at-rest artifacts. Database-column encryption for user data is governed by [data-handling.md](data-handling.md).

### Audit Logging

Audit logs answer "who did what, to what, when, and with what result" for security- and compliance-relevant actions. They are distinct from operational/access logs (see [observability.md](observability.md)): access logs serve debugging and traffic analysis and are sampled, rotated, and discarded freely; audit logs are evidence and are governed accordingly.

- Emit an audit event for security-relevant actions: authentication success and failure, authorization denials, privileged or data-mutating actions, and changes to configuration, secrets, or permissions. Emit it at the action boundary — the endpoint or command handler that performs the action — not deep in shared helpers. Do not audit ordinary reads at the same volume as writes; audit reads only where compliance demands it (e.g. access to a regulated record).
- Use a structured schema with the full WHO / WHAT / WHEN / WHERE on every record: WHO is the principal (actor) plus tenant/org; WHAT is the action, the target resource, and the result (allowed/denied, success/failure); WHEN is a UTC timestamp (see [../foundations/time.md](../foundations/time.md)); WHERE is the request id and source (caller IP / service identity). A denial or failure is as important to record as a success.
- Keep audit logs on their OWN stream and sink, separate from operational logs, so they carry their own retention, access controls, and integrity guarantees. Where compliance requires it, the sink is append-only and tamper-evident; define a retention period aligned to the governing regime rather than the default operational log retention.
- Never put secrets or PII payloads in audit records. Audit the fact and identity of the action (resource id, principal, result), not the sensitive contents — redact or reference by id. See [data-handling.md](data-handling.md) for what counts as sensitive and how to redact it.
- Use a DEDICATED `ILogger` category (e.g. `Orders.Audit`, via source-generated `[LoggerMessage]` methods with `actor`, `action`, `resource`, `result` properties) routed by the logging pipeline to the audit sink — not the shared application logger — or the audit backend the org mandates. Route the specific backend (SIEM, managed audit service, append-only store) to [../decisions/framework-selection.md](../decisions/framework-selection.md).

### Supply-Chain Rules

- NuGetAudit runs on every restore (`NuGetAuditMode=all` from the `Directory.Build.props` template) and fails the build on high/critical advisories.
- `packages.lock.json` is committed and CI restores `--locked-mode`, so the dependency graph cannot drift silently; versions are centralized in `Directory.Packages.props`.
- Dependabot (from the [dependabot template](../templates/dependabot.yml)) keeps patches flowing; upgrades follow [../recipes/bump-dependency.md](../recipes/bump-dependency.md) and the [dependency-upgrade checklist](../checklists/dependency-upgrade.md).
- Review new dependencies for maintenance health, license, and security posture before adoption.

### Vulnerability Disclosure

Decide HOW a vulnerability reaches you before one does, so a reporter never has to choose between a public issue and silence.

- Every externally-facing service and every published library ships a `SECURITY.md` (root or `.github/`) from the [security-policy template](../templates/security-policy.md). Internal-only services may ship a trimmed version (or skip the file) but must still name an owner and a private triage path (in the [runbook](../templates/runbook.md) or service README).
- Give a PRIVATE report path — a GitHub private security advisory or a `security@<org>` alias — and say plainly: no public issue, PR, or discussion for an unpatched vulnerability. A missing report path is itself a risk; reporters fall back to public issues.
- Commit to an acknowledgement and triage SLA (e.g. ack within 2 business days, triage within 5) and keep the supported-versions table honest about what you actually patch.
- Hold a coordinated-disclosure posture: fix privately, agree an embargo/disclosure date with the reporter, then publish an advisory (with a CVE where applicable) and credit the reporter. Pair the fix with the dependency response in [../recipes/bump-dependency.md](../recipes/bump-dependency.md) when the vuln is in a dependency.

### Build Hardening

- Release builds under warnings-as-errors through the one gate: `pwsh ./verify.ps1`.
- Enable `ContinuousIntegrationBuild` in CI so deterministic builds embed no local paths in PDBs or artifacts.
- Ship the chiseled, non-root runtime image from the [Dockerfile template](../templates/Dockerfile) — no shell, no SDK, no package manager in production layers; see [deployment.md](deployment.md).
- `ASPNETCORE_ENVIRONMENT=Production` keeps the developer exception page, detailed errors, and OpenAPI UI off outside Development — never re-enable them in production config.

## Common Mistakes And Forbidden Patterns

- secrets committed anywhere in the repo, including `appsettings.json`, examples, or tests
- a real secret baked into image layers or passed as a Docker `ARG`/build arg (it persists in layer history even if a later layer removes it)
- a secret value logged in startup, debug, or error output, or surfaced in an exception message or `ProblemDetails` body
- no rotation path, so a leaked credential is valid forever and cannot be cheaply revoked
- `.env` files used outside local development, or a real `.env` committed (commit `.env.example` only); `dotnet user-secrets` treated as a production mechanism
- endpoints that skip the fallback policy silently — `AllowAnonymous` without a review-visible justification
- token validation with audience or issuer checks disabled "to make it work"
- auth logic mixed into unrelated helper classes instead of policies and authorization handlers
- path handling that assumes `Path.Combine` alone makes untrusted input safe
- homegrown crypto or token code when Data Protection or `System.Security.Cryptography` already solves it
- multi-replica services with an unconfigured Data Protection key ring, so sessions and protected tokens break on every deploy
- publicly disclosing an unpatched vulnerability (issue, PR, commit message, or release note) before a fix ships
- no private report path, so a reporter's only option is a public issue that exposes the flaw
- conflating audit logs with access/operational logs — routing them to the same stream, where audit evidence inherits short operational retention and loose access controls
- auditing reads at the same volume as writes, drowning the security signal in routine-access noise
- secret values or PII payloads written into audit records instead of redacted or referenced by resource id
- recording only successful actions, so authorization denials and authn failures leave no audit trail

## Verification And Proof

- run `pwsh ./verify.ps1` — restore (locked), format-check, build (warnings-as-errors), test, audit; the audit stage fails on high/critical advisories
- targeted negative tests for auth (missing/expired/wrong-audience token → 401, missing role → 403), validation, path traversal, and outbound-client restrictions
- release build audit showing no local paths or secret material embedded in the artifact
- image and log scan confirming no secret values appear in any layer, build arg, or emitted log line
- rotation exercised end to end: rotate the credential and confirm the running process picks up the new value (via config reload or rolling restart) without a redeploy of code
- startup fails fast and names the missing secret when required material is absent or empty
- dependency review for every newly introduced non-trivial package
- `SECURITY.md` present for externally-facing services and published libraries, naming a private report path (no public issue) and an acknowledgement/triage SLA
- disclosure path known and tested: a reporter (or internal tester) can reach the private channel and gets the documented acknowledgement
- each security-relevant action (authn success/failure, authz denial, privileged or data-mutating change) emits an audit event carrying the full who / what / when / result
- audit logs land on a separate sink from operational logs, with a defined retention period and access controls aligned to the governing compliance regime
- audit records contain no secret or PII payloads — spot-check a denial and a privileged-write event to confirm sensitive data is redacted or referenced by id
