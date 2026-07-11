# Cross-Platform

Rules that make one codebase behave identically on Windows, Linux, and macOS dev machines and in Linux containers — enforced by running the same gate on all three.

## Default Approach

Development happens on any of the three OSes; production is Linux containers. Every rule below exists because code that only ever ran on the author's OS has a latent bug on the other two. The gate is not "works on my machine" — it is `pwsh ./verify.ps1` green on the ubuntu/windows/macos CI matrix.

### Paths

- Build filesystem paths with `Path.Combine` or `Path.Join`; never concatenate with a hardcoded `\` or `/`.
- `Path.DirectorySeparatorChar` when a separator must be named; `Path.GetFullPath` to normalize before comparing.
- URLs, MSBuild item paths, and container paths always use `/` — the rule is about *filesystem* paths on the host.

```csharp
var exportPath = Path.Combine(exportRoot, "orders", $"{orderId}.json");
```

### Case Sensitivity

- Linux (and CI on ubuntu) is case-sensitive; Windows and default macOS are not. Every file reference — project references, `Content` items, config file names, embedded resource names, paths in code — must match the on-disk casing exactly.
- Naming discipline prevents the class of bug: no two files or directories whose names differ only by case, ever.
- The ubuntu leg of the CI matrix is the enforcement: a casing mismatch that Windows tolerates fails the Linux build. Do not "fix" it by renaming on Windows alone — `git mv` in two steps so the case change actually commits.

### Line Endings

- [templates/.gitattributes](../templates/.gitattributes) is the source of truth: `* text=auto` with `eol=lf` for source files; CRLF only for `.cmd`/`.bat`.
- [templates/.editorconfig](../templates/.editorconfig) matches it with `end_of_line = lf`.
- Developers rely on `.gitattributes` (or set `core.autocrlf=false`); no per-machine line-ending opinions reach the repo.
- `dotnet format --verify-no-changes` in the gate fails on drift, on every OS. Setup details: [project-setup.md](project-setup.md).

### Scripts: pwsh Only

- PowerShell 7 (`pwsh`) is the only blessed script runtime — one script, three OSes. `verify.ps1`, tooling scripts, and CI steps are all pwsh.
- Bash-only scripts are forbidden in the verify path; `.cmd`/`.bat` wrappers are forbidden everywhere except as a documented last resort for Windows-only tooling.
- The `Makefile` is a one-line shim delegating to `pwsh ./verify.ps1` for muscle memory, not a second implementation — copy both from [templates/verify.ps1](../templates/verify.ps1) and [templates/Makefile](../templates/Makefile).

```powershell
# Works unmodified on Windows, Linux, macOS:
$repoRoot = Split-Path -Parent $PSCommandPath
dotnet build (Join-Path $repoRoot 'Orders.slnx') -c Release
```

### Signals And Process Lifecycle

- `IHostApplicationLifetime` is the shutdown abstraction: the host translates SIGTERM/SIGINT on Unix and console close events on Windows into the same `ApplicationStopping` sequence — see [cancellation-and-async.md](cancellation-and-async.md).
- Never P/Invoke signal handling. If a component genuinely needs a raw signal (e.g. SIGHUP-triggered reload), use `PosixSignalRegistration` — in `Orders.Infrastructure` only, with a documented Windows fallback path, because the signal does not exist there.
- Exit codes are the portable status channel; do not encode status in signals or OS-specific mechanisms.

### File Locking And Handle Disposal

- Windows locks open files: deleting, renaming, or re-opening a file that still has an undisposed handle throws. Code that passes on Linux and fails only on the Windows CI leg almost always has a leaked handle.
- Dispose deterministically on every path: `await using` for streams and writers, `using` for everything else. Never rely on finalizers to release a file.
- Tests that create temp files must dispose before deleting, and clean up in `finally`/`Dispose`, not after an assertion that may throw.

```csharp
await using (var stream = File.Create(exportPath))
{
    await JsonSerializer.SerializeAsync(stream, report, ReportContext.Default.Report, cancellationToken);
}
// stream is closed here — safe to move/delete on Windows too
File.Move(exportPath, finalPath);
```

### Globalization And String Comparison

- Machine-facing parse and format is always `CultureInfo.InvariantCulture`. Culture-implicit `ToString()` and `Parse` on numbers and dates at trust boundaries are banned — `double.ToString()` under a comma-decimal culture writes `"1,5"` into your JSON.
- Every string comparison that matters states its rules: `StringComparison.Ordinal` for identifiers, keys, protocol tokens, and paths; `OrdinalIgnoreCase` when case-insensitivity is part of the contract. Culture-sensitive comparison is only for human-facing sort/display, chosen explicitly.
- The analyzers enforce this: CA1305/CA1310/CA1307-family rules require an explicit `IFormatProvider`/`StringComparison`, and run as errors in the gate ([../quality/linting.md](../quality/linting.md)).

```csharp
var amount = decimal.Parse(raw, CultureInfo.InvariantCulture);
if (string.Equals(header, "idempotency-key", StringComparison.OrdinalIgnoreCase)) { /* ... */ }
```

- The runtime uses ICU on all three OSes; keep invariant-globalization mode OFF for services that handle user data, and note that chiseled container images need the `-extra` variant to ship ICU and tzdata — the image decision lives in [../operations/deployment.md](../operations/deployment.md).

### Time Zones

- `TimeZoneInfo.FindSystemTimeZoneById` with IANA IDs (`"Europe/Berlin"`) works on every OS — since .NET 6, ICU maps IANA IDs on Windows. Store IANA IDs, never Windows zone names.
- Full time rules, including the conversion helpers for legacy Windows IDs: [time.md](time.md).

### Filesystem Locations

- Temp files: `Path.GetTempPath()` (+ `Path.GetRandomFileName()` for collision-free names). Never hardcode `/tmp` or `C:\Temp`.
- User-scoped data and config: `Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData)` / `LocalApplicationData` — the framework maps them per OS.
- Long paths: Windows tooling can still choke past 260 characters; keep repo-relative paths short (shallow directories, no generated deep nesting) rather than relying on machine-level long-path opt-ins.
- Reserved Windows device names (`aux`, `con`, `nul`, `prn`, `com1`–`com9`, `lpt1`–`lpt9`) are never used as file or directory names in the repo — a Linux-created `nul.cs` makes the repo un-checkout-able on Windows.

### Process Invocation

- Launch processes with `ProcessStartInfo`, `UseShellExecute = false`, and `ArgumentList` — one argument per entry, no shell string concatenation, no quoting bugs, no shell injection, no dependence on `cmd.exe` vs `/bin/sh` semantics.

```csharp
var psi = new ProcessStartInfo("dotnet") { UseShellExecute = false, RedirectStandardOutput = true };
psi.ArgumentList.Add("ef");
psi.ArgumentList.Add("database");
psi.ArgumentList.Add("update");
using var process = Process.Start(psi) ?? throw new InvalidOperationException("dotnet failed to start");
```

- Resolve tool names without extensions (`"dotnet"`, not `"dotnet.exe"`); the OS resolver appends the right one.
- Every spawned process gets a timeout and cancellation (`WaitForExitAsync(cancellationToken)`) per [cancellation-and-async.md](cancellation-and-async.md).

## Common Mistakes And Forbidden Patterns

- Hardcoded `\` or `/` in filesystem paths, or hardcoded roots like `/tmp` and `C:\`.
- File references whose casing does not match the on-disk name (builds on Windows, fails on ubuntu CI).
- Two files differing only by name casing.
- Bash-only scripts in the verify path; committing CRLF source files or bypassing `.gitattributes`.
- P/Invoked signal handlers; `PosixSignalRegistration` outside Infrastructure or without a Windows fallback.
- Deleting or moving a file while a handle is still open; relying on finalizers instead of `await using`.
- Culture-implicit `ToString()`/`Parse` on numbers or dates at trust boundaries; string comparisons without an explicit `StringComparison`.
- Windows time zone names in stored data or config.
- `aux`, `con`, `nul`, or other reserved device names as file/directory names.
- Shell-string process launching (`cmd /c ...`, `/bin/sh -c ...` with concatenated input) instead of `ArgumentList`.
- Treating a green build on one OS as proof — only the three-OS matrix is proof.

## Verification And Proof

```powershell
pwsh ./verify.ps1
```

The same script — restore (locked), format-check, build (warnings-as-errors), test, audit — runs locally on whatever OS you develop on and in CI on the ubuntu/windows/macos matrix ([templates/github-workflows-ci.yml](../templates/github-workflows-ci.yml)). Cross-platform correctness is proven, not assumed:

- The ubuntu leg catches casing and line-ending drift; the windows leg catches leaked file handles and path assumptions; macos catches the rest of the POSIX-but-not-Linux surface.
- Analyzer rules for globalization (CA1305/CA1307/CA1310-family) fail the build on culture-implicit calls.
- `dotnet format --verify-no-changes` fails on line-ending or style drift on every leg.
- Tests involving temp files, process launching, or time zones must pass on all three legs before merge — an OS-specific skip requires a comment naming the OS limitation and a tracking issue.
