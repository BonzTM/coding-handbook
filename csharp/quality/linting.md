# Linting

The single source of truth for static analysis policy: what runs, why, and how it is pinned and invoked.

## Default Approach

The built-in .NET analyzers are the mandated linter. The Roslyn compiler ships correctness, security, performance, and code-style analyzers with the SDK; this handbook turns all of them on, makes every finding a build error, and drives every severity and style decision from one file. There is no separate lint tool to install, version, or drift — **the build is the lint**.

Three MSBuild properties, set once in [../templates/Directory.Build.props](../templates/Directory.Build.props) so every project in the solution inherits them, define the policy:

```xml
<PropertyGroup>
  <AnalysisLevel>latest-all</AnalysisLevel>
  <EnforceCodeStyleInBuild>true</EnforceCodeStyleInBuild>
  <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
</PropertyGroup>
```

- `AnalysisLevel=latest-all` — run the newest analyzer set the SDK ships, with **all** rule categories enabled, including the many CA rules that are off by default. This is the "enable the high-signal linters" decision made once, at the strictest baseline.
- `EnforceCodeStyleInBuild=true` — the IDE-style rules (`IDExxxx`: naming, `var` policy, file-scoped namespaces, unused usings) fail the command-line build, not just squiggle in an editor. Style stops being an IDE opinion and becomes a contract.
- `TreatWarningsAsErrors=true` — every compiler and analyzer warning is a build failure, in every configuration, from day one. A codebase that tolerates warnings trains people to ignore them.

Severities and style choices live in **`.editorconfig` and nowhere else**. The canonical copy is [../templates/.editorconfig](../templates/.editorconfig); copy it to the repo root. Never scatter severity policy across `<NoWarn>` properties, ruleset files, or per-project props — one file, reviewable in one diff, is the whole policy.

### Versioning And Pinning

The analyzers ship inside the SDK, so the pin is [../templates/global.json](../templates/global.json): every developer and CI leg resolves the same SDK feature band (`rollForward: latestFeature`), hence the same analyzer set. There is no globally installed lint binary to drift between machines.

Consequence: bumping `global.json` can introduce new rules — that is `latest-all` working as intended. Fix the new findings (or suppress with justification, below) in the same PR as the SDK bump; the [dependency-upgrade checklist](../checklists/dependency-upgrade.md) carries this step.

### Running It

The gate runs through the one canonical entrypoint, which both humans and CI invoke (see [../templates/verify.ps1](../templates/verify.ps1)):

```powershell
pwsh ./verify.ps1        # restore (locked), format-check, build (warnings-as-errors), test, audit
dotnet format            # apply formatting and style fixes in place
dotnet format --verify-no-changes   # the read-only format gate, as verify.ps1 runs it
dotnet build -c Release  # the raw analyzer gate: analyzers run as part of compilation
```

Formatting and analysis are deliberately separate stages: `dotnet format --verify-no-changes` fails fast on whitespace and style drift without writing anything, so a build failure always means a real defect, not indentation. `dotnet format` is SDK-bundled, so `global.json` pins it too — no formatter drift. The `.editorconfig` sets `end_of_line = lf` for source, which together with the committed [.gitattributes](../templates/.gitattributes) keeps the format gate green on all three CI OSes (see [../foundations/cross-platform.md](../foundations/cross-platform.md)).

### What The Enabled Set Buys

`latest-all` is a curated set — Microsoft maintains it, and each category maps to a handbook contract:

- **Correctness and reliability** (`CA1806`, `CA2007`-family async rules, `CA2016`, `CA2200`, …) — unused results, unforwarded `CancellationToken`s, rethrows that destroy stack traces. These enforce [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md) and [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md) mechanically.
- **Resource and dispose tracking** (`CA1001`, `CA2000`, `CA2213`) — undisposed fields, locals, and members: the static half of the leak-detection story in [testing.md](testing.md).
- **Security** (`CA2100`, `CA3xxx`, `CA5xxx`) — SQL built by concatenation, weak crypto, insecure deserialization; the code-level companion to [../operations/security.md](../operations/security.md).
- **Globalization** (`CA1304`, `CA1305`, `CA1310`) — culture-implicit parse/format/compare, the exact bug class [../foundations/cross-platform.md](../foundations/cross-platform.md) bans at trust boundaries.
- **Performance** (`CA18xx`) — including `CA1848`, which steers logging to the source-generated `[LoggerMessage]` pattern [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md) mandates.
- **Style** (`IDExxxx`, enforced in build) — naming conventions, file-scoped namespaces, unused code removal, per [../foundations/style-and-review.md](../foundations/style-and-review.md). The `.editorconfig` is where each style rule's severity is chosen.
- **Nullability** (compiler `CS86xx`/`CS87xx` warnings) — with `<Nullable>enable</Nullable>` and warnings-as-errors, a nullability warning is a build failure. Never "fix" one with `!` unless the justification would survive review as a comment.

### The GenerateDocumentationFile Trade-Off

`IDE0005` (remove unnecessary usings) only fires during the build when the project generates an XML documentation file, so the committed [Directory.Build.props](../templates/Directory.Build.props) sets `GenerateDocumentationFile=true` repo-wide — without it, the compiler emits the `EnableGenerateDocumentationFile` diagnostic and warnings-as-errors fails every build. The side effect is `CS1591` (missing XML doc on public member), which is a policy decision, not noise: services set `CS1591` to `none` in `.editorconfig` (a service documents its public seams, not every DTO property), libraries flip it to `warning` so an undocumented public member fails the build. The doc-comment contract itself lives in [../foundations/style-and-review.md](../foundations/style-and-review.md).

### Considered And Excluded

Deliberate exclusions, so nobody re-litigates them one PR at a time. Extra analyzer packages are ADR-level adoptions routed via [../decisions/framework-selection.md](../decisions/framework-selection.md):

- **StyleCop.Analyzers** — excluded. Its useful rules (naming, ordering, documentation) now overlap the built-in IDE analyzers plus `dotnet format`; adopting it reintroduces double-reporting, conflicting fixers, and a second style authority next to `.editorconfig`.
- **Roslynator** — via ADR. A large, decent ruleset, but mostly refactoring suggestions; adopt only if a repo demonstrates recurring defect classes the built-ins miss.
- **SonarAnalyzer.CSharp** — via ADR, and normally only when the org already runs SonarQube and wants parity between local build and server findings; otherwise its overlap with the CA rules is noise.
- Per-repo additions (e.g. a banned-API analyzer for a specific internal rule) follow the same ADR path with the same one-line rationale recorded here.

### Suppression Policy

When a finding is genuinely wrong (not merely inconvenient), suppress it at the narrowest possible scope, with the rule ID and a reason:

- `#pragma warning disable <ID> // <reason>` immediately before the offending line(s), with `#pragma warning restore <ID>` immediately after. The reason comment is mandatory; a bare disable fails review.
- `[SuppressMessage(...)]` with a non-empty `Justification` is acceptable on a member when a pragma pair is awkward (e.g. the finding spans a whole method). Never apply it at assembly scope.
- **`GlobalSuppressions.cs` is discouraged**: it moves the suppression away from the code it excuses, so the context rots and nobody dares delete entries. Suppressions live at the finding.
- **Unused suppressions are dead code.** `IDE0079` (remove unnecessary suppression) is enabled in the `.editorconfig`, so a suppression that no longer fires fails the build and gets deleted.
- `<NoWarn>` in a project file to dodge an analyzer finding is forbidden — it is an invisible, file-wide, unjustified suppression. Severity decisions belong in `.editorconfig`.
- If the same false positive recurs across files, the rule is misconfigured for this codebase: change its severity in the template `.editorconfig` (see [Changing The Rule Set](#changing-the-rule-set)) instead of scattering pragmas.

### Suppressing A False Positive

`CA2000` flags a disposable that leaves scope undisposed — correct almost everywhere, but a false positive when ownership transfers to the returned object:

```csharp
public static OrdersExport OpenExport(string path)
{
    ArgumentException.ThrowIfNullOrEmpty(path);

#pragma warning disable CA2000 // ownership transfers to OrdersExport, which disposes the stream
    var stream = File.OpenRead(path);
#pragma warning restore CA2000
    return new OrdersExport(stream);
}
```

One rule ID, one line covered, restore immediately after, and the reason states the ownership fact a reviewer can verify. Not acceptable: disabling `CA2000` file-wide, adding it to `<NoWarn>`, or parking an entry in `GlobalSuppressions.cs`.

### Changing The Rule Set

The template is the source of truth and copies must not drift: any severity change — tightening or loosening — updates [../templates/.editorconfig](../templates/.editorconfig) and the rationale in this document in the same PR, and `pwsh ./verify.ps1` must pass with the change applied. Tightening (raising a rule to error) is the default direction and needs only the finding class it catches; loosening (downgrading or disabling a rule) needs the false-positive evidence written into the PR. Per-directory `.editorconfig` overrides are acceptable only for the pragmatic cases the template already models (e.g. relaxed documentation rules under `tests/`), never as a back door around a repo-wide decision.

### NuGetAudit: The Supply-Chain Gate

NuGetAudit is the dependency-vulnerability gate — the counterpart of a vulnerability scanner in any other stack, and it runs on every restore rather than as a separate tool. Configured in [../templates/Directory.Build.props](../templates/Directory.Build.props):

```xml
<PropertyGroup>
  <NuGetAudit>true</NuGetAudit>
  <NuGetAuditMode>all</NuGetAuditMode>   <!-- direct AND transitive packages -->
  <NuGetAuditLevel>high</NuGetAuditLevel> <!-- report high and critical -->
</PropertyGroup>
```

- Restore checks every resolved package against the vulnerability database of the package source (nuget.org provides one; the committed [nuget.config](../templates/nuget.config) points at it). Advisories surface as `NU1901`–`NU1904` warnings by severity; with `NuGetAuditLevel=high` and warnings-as-errors, a high or critical advisory **fails the restore** — the audit stage of `pwsh ./verify.ps1`.
- Pair it with the lockfile: `RestorePackagesWithLockFile=true` plus CI restoring `--locked-mode` means the audited graph is exactly the shipped graph (see [../foundations/project-setup.md](../foundations/project-setup.md)).
- When the audit fires, the fix is an upgrade via [../recipes/bump-dependency.md](../recipes/bump-dependency.md) and the [dependency-upgrade checklist](../checklists/dependency-upgrade.md) — not a `NoWarn` on `NU1903`.
- A genuinely unreachable advisory with no patched version may be suppressed with an `<NuGetAuditSuppress>` item naming the advisory URL, a justification comment beside it, and a tracking issue to remove it — the same auditability bar as a pragma. Blanket-suppressing audit warnings is forbidden.
- Broader supply-chain policy (provenance, dependabot cadence) lives in [../operations/security.md](../operations/security.md) and [../operations/ci-and-release.md](../operations/ci-and-release.md).

## Common Mistakes And Forbidden Patterns

- Weakening the trio — dropping `AnalysisLevel` below `latest-all`, turning off `EnforceCodeStyleInBuild`, or scoping `TreatWarningsAsErrors` to Release only so Debug builds rot with warnings.
- Severity policy outside `.editorconfig`: `<NoWarn>` accumulating in csproj files, resurrected `.ruleset` files, or per-developer editor settings standing in for the committed contract.
- Bare `#pragma warning disable` — no rule ID, no reason, no matching restore.
- `GlobalSuppressions.cs` as a dumping ground, or `[SuppressMessage]` with an empty `Justification`.
- Suppressing a nullability warning with `!` where a null check or type change is the real fix.
- Fixing format-gate failures by hand instead of running `dotnet format`, or letting editors with different settings fight the committed `.editorconfig`.
- A globally installed or floating tool doing the formatting/analysis instead of the `global.json`-pinned SDK, so local and CI results diverge.
- Disabling a rule repo-wide to dodge one finding instead of a scoped, justified suppression — or the reverse, pragma-spamming a rule that is genuinely misconfigured instead of fixing the template.
- `NoWarn` on `NU19xx` codes to make the audit shut up, or turning `NuGetAuditMode` back to `direct` to hide transitive advisories.
- Excluding `tests/` from analysis entirely; tests still need correctness, async, and dispose rules — only documentation-style rules are relaxed there.

## Verification And Proof

```powershell
dotnet format --verify-no-changes   # format gate: no diff
dotnet build -c Release             # analyzer gate: zero warnings, therefore zero errors
pwsh ./verify.ps1                   # the full ordered gate: restore (locked), format-check, build (warnings-as-errors), test, audit
```

Linting is clean when the build exits 0 with no suppression added to dodge a real defect, `dotnet format` leaves no diff, and restore reports no high/critical advisory. A green `pwsh ./verify.ps1` proves the whole policy held.

## Where To Go Next

- Style rules the IDE analyzers enforce: [../foundations/style-and-review.md](../foundations/style-and-review.md)
- Error and logging contract the CA rules backstop: [../foundations/errors-and-logging.md](../foundations/errors-and-logging.md)
- Culture and platform rules behind the globalization analyzers: [../foundations/cross-platform.md](../foundations/cross-platform.md)
- How the gate fits the delivery pipeline: [../operations/ci-and-release.md](../operations/ci-and-release.md)
- Dynamic proof beyond static analysis: [testing.md](testing.md)
