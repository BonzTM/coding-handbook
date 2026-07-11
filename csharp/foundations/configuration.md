# Configuration

Configuration loading and validation rules for repos that should fail early instead of drifting into bad runtime state.

## Default Approach

- Compose configuration in one place: `Program.cs` via `WebApplication.CreateBuilder(args)`. Nothing outside the composition root touches `IConfiguration`; everything else consumes typed options.
- Use the default provider stack and its precedence deliberately: `appsettings.json` → `appsettings.{Environment}.json` → user secrets (Development only) → environment variables → command-line args, last one wins. Env vars overriding committed JSON is the deployment contract.
- `appsettings.json` is committed and secret-free. It documents every supported key with a safe default; secrets arrive only through env vars or mounted files at runtime (see [../operations/security.md](../operations/security.md)).
- Bind every section to a sealed options class with `ValidateDataAnnotations()` plus `ValidateOnStart()`, so a missing or malformed value kills the process at startup — before listeners open — with an error naming the section and field.
- Env-var mapping uses double underscores: `Payments__BaseAddress` overrides `Payments:BaseAddress`. This is the `.env`-file counterpart: local, CI, and production all override through the same mechanism, so there is no separate loader and no environment-specific code path.
- Adding a key follows [../recipes/add-config-key.md](../recipes/add-config-key.md).

### Config Shape

- one options class per section, named `<Section>Options`, with a `SectionName` constant
- group related settings into one section rather than scattering flat keys
- use `TimeSpan` for timeouts and intervals — the binder parses `"00:00:30"`; never raw integer seconds
- keep property names aligned with the concepts they control, not the source format alone
- separate secret values from tuning knobs in both code and docs

```csharp
namespace Orders.Api;

public sealed class PaymentsOptions
{
    public const string SectionName = "Payments";

    [Required, Url]
    public string BaseAddress { get; init; } = string.Empty;

    [Range(1, 10)]
    public int MaxRetries { get; init; } = 3;

    public TimeSpan Timeout { get; init; } = TimeSpan.FromSeconds(30);
}
```

```csharp
builder.Services
    .AddOptions<PaymentsOptions>()
    .BindConfiguration(PaymentsOptions.SectionName)
    .ValidateDataAnnotations()
    .ValidateOnStart();
```

### Options Lifetimes

- `IOptions<T>` is the default: a singleton snapshot fixed at startup. Config that changes requires a redeploy, which is the correct default posture.
- `IOptionsMonitor<T>` is for the rare value that genuinely must change at runtime (see Feature Flags below). It is a singleton and safe to inject anywhere, but every consumer inherits reload semantics — do not spread it by reflex.
- `IOptionsSnapshot<T>` (scoped, recomputed per request) is almost never justified and crashes as a captive dependency inside singletons. Reach for it only when a per-request recompute is a stated requirement.

### Feature Flags

The default is **static, validated config — a typed property**, not a flag system. A boolean or enum bound at startup, validated with the rest of the config, and read through the typed options class covers the overwhelming majority of "toggle" needs without new machinery. A flag is debt; do not introduce a flag system to avoid a restart that you could simply perform.

When a flag genuinely needs to change without a redeploy (a gradual rollout, a kill switch for a risky path), keep it behind a **typed accessor backed by a snapshot** swapped on reload:

```csharp
public sealed class FeatureFlags
{
    private readonly ILogger<FeatureFlags> _logger;
    private FlagsOptions _current;   // whole snapshot swapped atomically on reload

    public FeatureFlags(IOptionsMonitor<FlagsOptions> monitor, ILogger<FeatureFlags> logger)
    {
        _logger = logger;
        _current = monitor.CurrentValue;
        monitor.OnChange(next =>
        {
            if (!TryValidate(next, out var error))
            {
                _logger.LogError("Rejected flags reload: {Error}", error);
                return; // keep the last good snapshot
            }
            Volatile.Write(ref _current, next);
        });
    }

    public bool NewCheckoutFlow => Volatile.Read(ref _current).NewCheckoutFlow;
}
```

- Swap the **whole snapshot** (`Volatile.Write` of the new instance), never mutate properties of the live instance — readers must always see a consistent set, with no torn reads and no locks on the hot path.
- The accessor owns reload validation, because an `IOptionsMonitor` change callback must never throw: a malformed reload keeps the last good snapshot and is logged once (see [errors-and-logging.md](errors-and-logging.md)).
- Only file providers hot-reload; env-var overrides do not. A reloadable flag therefore lives in the JSON config the deployment can rewrite — if your platform can only change env vars, the flag is static and the answer is a restart.
- Callers read through the typed accessor (`flags.NewCheckoutFlow`), never `IConfiguration["Features:NewCheckoutFlow"]`. The flag's name, type, and default live in exactly one place.
- **Flags are short-lived.** A flag exists to de-risk a rollout, then dies. The work is not done when the rollout completes — it is done when the flag *and the dead branch it guarded* are deleted. Track each flag's removal like any other follow-up; a flag that outlives its rollout is permanent branching debt and a source of dead, untested code paths.
- A **dynamic flag service** (LaunchDarkly, Azure App Configuration, a config server) is an external dependency with its own SLO and failure modes — adopt one only when you need targeting, percentage rollouts, or runtime changes a config reload cannot give you, and route the pick through [../decisions/framework-selection.md](../decisions/framework-selection.md). Even then, the typed accessor is the seam; the service is its backing source, with a static fallback when it is unreachable.

### Documentation Expectations

- The committed `appsettings.json` plays the `.env.example` role: every supported key present, with values safe and realistic enough to exercise the startup path.
- `dotnet user-secrets` is for local development only. The `UserSecretsId` in the `.csproj` is safe to commit; the secrets themselves never are, and user secrets are never a production mechanism.
- Document every supported key — and its env-var override form — in the repo README or operator docs.
- Never commit real secrets, `appsettings.Production.json` with credentials, or environment-specific override files.

## Common Mistakes And Forbidden Patterns

- `IConfiguration` injected into endpoints, repositories, or workers for lazy lookups.
- Silent fallback when a required value is missing or malformed — options bound without `ValidateOnStart()`.
- Committed secrets: connection strings or API keys in any `appsettings*.json`.
- Hidden defaults that differ between local dev, CI, and production, usually via a sprawling `appsettings.Development.json`.
- Raw `Environment.GetEnvironmentVariable` calls scattered through handlers and workers instead of one bound options class.
- `IOptionsSnapshot<T>` injected into a singleton (captive-dependency crash), or `IOptionsMonitor<T>` used everywhere "just in case".
- A flag system (or external flag service) introduced to dodge a restart that a static config value plus a deploy would handle.
- Feature flags that outlive their rollout: the flag and its dead branch left in place after the decision is permanent, accreting untested code paths.
- Mutating a live options instance in place under concurrent reads instead of swapping a validated snapshot.

## Verification And Proof

- Unit tests for successful binding, missing required values, malformed values, and precedence rules — build an in-memory `IConfiguration` (`AddInMemoryCollection` layered under env-var-shaped overrides) and assert on the bound options.
- Startup smoke test with a required variable removed: the process must die at startup with an `OptionsValidationException` naming the section and field, not at first use.
- Review `appsettings.json` and deployment docs whenever new keys are added.
- For dynamic flags: a test that a malformed reload is rejected and the last good snapshot is retained, and a stress test that reads through the accessor while reloads swap the snapshot.
- Each flag has a tracked removal item; a periodic review confirms no flag has outlived its rollout.
