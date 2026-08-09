# Recipe: Add Config Key

Use this when a process gains a runtime setting, feature flag, limit, endpoint, or secret.

## Files To Touch

- `src/<app>/config.py`
- `.env.example`, README config table, deployment manifests
- `tests/test_config.py`

## Steps

1. Add a typed pydantic-settings field with range/shape constraints; use `SecretStr` for secrets and no secret default.
2. Keep one environment naming convention and nested delimiter. Do not add an `os.getenv` read elsewhere.
3. Validate the complete settings graph before opening listeners or resources; malformed values fail instead of falling back.
4. Pass the validated value explicitly to its owner.
5. Add the key, safe example, required/default status, and secret classification to every operator surface.

## Invariants To Preserve

- Settings are constructed once in composition; no lazy reads or global settings singleton.
- Required and malformed values fail startup with the key name, never a secret value.
- `.env.example`, README, deployment configuration, and settings model stay synchronized.
- A feature flag has an owner, safe default, rollout purpose, and removal date.

## Proof

```bash
uv run pytest tests/test_config.py
env -u <REQUIRED_KEY> uv run <app>
make verify
```

Tests cover valid loading, omission, malformed input, precedence, nested mapping, and secret redaction. Governing doc: [configuration](../foundations/configuration.md).
