# Configuration

Configuration loading and validation rules for repos that should fail early instead of drifting into bad runtime state.

## Default Approach

Construct one `pydantic_settings.BaseSettings` graph in the composition root. Environment variables are the service contract; validated settings are passed explicitly. Pydantic Settings defines the supported source model and precedence in its [official documentation](https://docs.pydantic.dev/latest/concepts/pydantic_settings/).

### Config Shape

- Group related settings in nested models or focused settings classes.
- Use typed URLs, durations, enums, and bounded numeric fields rather than raw strings.
- Use `SecretStr` for secrets so normal display and representation do not reveal the value; unwrap only at the client-construction boundary.
- Give environment keys one documented naming convention and explicit nested delimiter when nested models are used.
- Load once before listeners, workers, engines, or clients start. Catch the single Pydantic validation failure only to render an actionable startup error; preserve all field errors rather than reporting the first.

`.env` loading is local-development convenience only. Production receives environment variables or mounted secret files from the platform. Real `.env` files never enter source control.

### Source And Precedence Policy

Constructor overrides are for tests, environment variables are the deployment source, `.env` is local-only, and code defaults are safe non-secret defaults. Do not add another source without an ADR defining precedence and failure behavior. Never read `os.environ` outside settings construction.

### Feature Flags

The default flag is a static typed boolean or enum on settings, validated at startup and changed by redeploy. Every flag has an owner, safe default, rollout purpose, and tracked removal date; delete the flag and dead branch after rollout.

A dynamic flag service is an external dependency. Adopt one only for targeting, percentage rollout, or a kill switch that cannot tolerate restart, through [framework selection](../decisions/framework-selection.md). Callers still use a typed accessor with a last-known-safe fallback; vendor values do not flow through core as untyped lookups.

### Documentation Expectations

Every supported key, purpose, required/default status, secret classification, and example value appears in `.env.example` and operator documentation. Adding a key follows [the config recipe](../recipes/add-config-key.md); documentation and deployment manifests change in the same PR.

## Common Mistakes And Forbidden Patterns

- Scattered `os.getenv` or `os.environ` reads in handlers, repositories, clients, or workers.
- Lazy validation at first use, partial settings construction, or fallback after malformed required input.
- Secrets in source, `.env.example`, defaults, exception text, logs, or model dumps.
- Treating `.env` as a production secret store.
- Raw string feature-flag lookups or flags with no removal owner.
- Multiple configuration sources with undocumented precedence.

## Verification And Proof

```bash
uv run pytest tests/test_config.py
make verify
```

Tests cover valid loading, every required-key omission, malformed values, nested environment mapping, precedence, secret redaction, and aggregated startup diagnostics. Review `.env.example`, deployment configuration, and operator docs with every settings change.
