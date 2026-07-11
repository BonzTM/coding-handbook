# Recipe: Add Config Key

Use this when a feature needs a new runtime setting (a tuning knob, a feature flag, a timeout, or a secret) that the process reads at startup.

Governing doc: [`csharp/foundations/configuration.md`](../foundations/configuration.md).

## Files To Touch

- the owning options class — `src/Orders.Api/Options/<Area>Options.cs` for host-level settings, or the Infrastructure options class for client/database settings
- `src/Orders.Api/Program.cs` — only if this is a brand-new options class that needs registration
- `appsettings.json` — the key with a safe, committed default (never a secret)
- options tests (happy path + a missing/malformed failure case)
- the project README config-key table (same key, same description)

## Steps

1. Add the typed property to the right options class. Use `TimeSpan` for timeouts/intervals, `int`/`long` for counts/sizes, `Uri` for addresses, an enum for closed sets — never a bare `string` for something with a fixed range. Group it into the options class that owns the area (`FulfillmentOptions`, `DatabaseOptions`); create a new class only for a genuinely new area.
2. Put the default in exactly one place: the property initializer in the options class (`public int WorkerConcurrency { get; set; } = 4;`), so defaults stay a single reviewable source of truth, not scattered literals. Omit the initializer only for genuinely required values (including all secrets) that must have no fallback.
3. Bind it in ONE place. A new options class is registered once in the composition root:

   ```csharp
   builder.Services.AddOptions<FulfillmentOptions>()
       .BindConfiguration(FulfillmentOptions.SectionName)
       .ValidateDataAnnotations()
       .ValidateOnStart();
   ```

   Do not add `builder.Configuration["..."]` or `Environment.GetEnvironmentVariable` lookups anywhere else. Environment variables override `appsettings.json` by the standard provider order using the `Section__Key` naming (`Fulfillment__DispatchTimeout=00:00:10`).
4. VALIDATE it fail-fast. Annotate with DataAnnotations (`[Required]`, `[Range(1, 64)]`, `[Url]`); `ValidateOnStart` aborts the host before it listens, with an `OptionsValidationException` naming the member. Cross-property invariants (`MaxIdle <= MaxOpen`) go in an `IValidateOptions<T>` implementation whose failure message names both keys. Never silently correct a bad value.
5. Separate secrets from tuning knobs. A secret has no default, appears in `appsettings.json` **only as documentation of the key path with an empty value — or not at all**, is `[Required]`, and its VALUE is never logged or echoed — only its key name on failure. Locally it comes from `dotnet user-secrets`; in deployment, from env vars or mounted files. Define provenance and rotation per [../operations/security.md](../operations/security.md).
6. Document it in `appsettings.json` AND the README config-key table in the same change, with a safe, realistic example value. The two must describe the same key identically.

## Invariants To Preserve

- no lazy or scattered `IConfiguration`/`Environment.GetEnvironmentVariable` reads in endpoints, repositories, or workers — consumers take `IOptions<T>` (or `IOptionsMonitor<T>` where live reload is deliberate) via constructor injection
- required values (and all secrets) fail fast at startup, before the host listens, with a message naming the key; no silent fallback on missing or malformed input
- `appsettings.json` and the README config-key table stay in sync — same keys, same descriptions
- secrets carry no default, never live in `appsettings.json`, and their values never reach a log line, error message, or exception dump
- no static config classes and no config reads in static initializers

## Proof

- an options test for the happy path (defaults applied, env-shaped override wins) using `ConfigurationBuilder` + `AddInMemoryCollection` and the same `Bind`/validate chain as production
- a failure test asserting a malformed or missing-required value throws `OptionsValidationException` with the member-named message
- run `pwsh ./verify.ps1`
- startup smoke: launch the service with the required key unset and confirm it aborts before opening listeners with an actionable, key-named message — e.g. `Fulfillment__WorkerConcurrency=0 dotnet run --project src/Orders.Api` exits non-zero and the error names `WorkerConcurrency` and its valid range (substitute your new key)
