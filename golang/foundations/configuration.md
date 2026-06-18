# Configuration

Configuration loading and validation rules for repos that should fail early instead of drifting into bad runtime state.

## Default Approach

- Load config in one place, usually `internal/config`.
- Prefer env vars for services and shared infrastructure settings.
- Use flags for local overrides and CLI-specific inputs.
- Keep precedence explicit. A good default is `flags > environment > hard-coded defaults`.
- Validate the full config before opening listeners, starting workers, or creating external clients.

### Config Shape

- group related settings into nested structs when it improves readability
- keep field names aligned with the concepts they control, not the source format alone
- use `time.Duration` for timeouts and intervals
- separate secret values from optional tuning knobs in both code and docs

### Feature Flags

The default is **static, validated config — a typed field**, not a flag system. A boolean or enum loaded at startup, validated with the rest of the config, and read through the typed config struct covers the overwhelming majority of "toggle" needs without new machinery. A flag is debt; do not introduce a flag system to avoid a restart that you could simply perform.

When a flag genuinely needs to change without a redeploy (a gradual rollout, a kill switch for a risky path), keep it in `internal/config` behind a **typed accessor backed by an atomic snapshot** updated on reload:

```go
// internal/config
type Flags struct {
	v atomic.Pointer[flagState] // whole snapshot swapped atomically on reload
}

func (f *Flags) NewCheckoutFlow() bool { return f.v.Load().newCheckoutFlow }
```

- Swap the **whole snapshot** atomically (`atomic.Pointer`), never mutate fields in place — readers must always see a consistent set, and there are no torn reads or locks on the hot path.
- Validate the new snapshot before swapping; a malformed reload keeps the last good value and is logged once (see [errors-and-logging.md](errors-and-logging.md)).
- Callers read through the typed accessor (`cfg.Flags.NewCheckoutFlow()`), never a raw lookup. The flag's name, type, and default live in exactly one place.
- **Flags are short-lived.** A flag exists to de-risk a rollout, then dies. The work is not done when the rollout completes — it is done when the flag *and the dead branch it guarded* are deleted. Track each flag's removal like any other follow-up; a flag that outlives its rollout is permanent branching debt and a source of dead, untested code paths.
- A **dynamic flag service** (LaunchDarkly, Unleash, a config server) is an external dependency with its own SLO and failure modes — adopt one only when you need targeting, percentage rollouts, or runtime changes a reload cannot give you, and route the pick through [../decisions/framework-selection.md](../decisions/framework-selection.md). Even then, the typed accessor in `internal/config` is the seam; the service is its backing source, with a static fallback when it is unreachable.

### Documentation Expectations

- Commit `.env.example`, never real `.env` secrets.
- Document every supported config key in the repo README or operator docs.
- Keep example values safe and realistic enough to exercise the startup path.

## Common Mistakes And Forbidden Patterns

- Lazy config lookups spread through handlers, repositories, or workers.
- Silent fallback when a required value is malformed.
- Hidden defaults that differ between local dev, CI, and production.
- Config loading in `init()`.
- Committed secrets, local token files, or environment-specific overrides in source control.
- Raw env or flag lookups scattered through handlers and workers to implement a toggle, instead of one typed field in `internal/config`.
- A flag system (or external flag service) introduced to dodge a restart that a static config value plus a deploy would handle.
- Feature flags that outlive their rollout: the flag and its dead branch left in place after the decision is permanent, accreting untested code paths.
- Mutating a live flag field in place under concurrent reads instead of swapping a validated snapshot atomically.

## Verification And Proof

- Unit tests for successful loading, missing required values, malformed values, and precedence rules.
- Startup smoke test with a missing required variable and a clearly actionable failure.
- Review `.env.example` and deployment docs whenever new keys are added.
- For dynamic flags: a test that a malformed reload is rejected and the last good snapshot is retained, and a `-race` test that reads through the accessor while a reload swaps the snapshot.
- Each flag has a tracked removal item; a periodic review confirms no flag has outlived its rollout.
