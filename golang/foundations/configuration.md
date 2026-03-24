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

## Verification And Proof

- Unit tests for successful loading, missing required values, malformed values, and precedence rules.
- Startup smoke test with a missing required variable and a clearly actionable failure.
- Review `.env.example` and deployment docs whenever new keys are added.
