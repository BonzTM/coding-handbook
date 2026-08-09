<!-- Install at .github/pull_request_template.md. -->

## Summary

- <what changed and why; link the issue>

## Change Type

- [ ] Feature
- [ ] Bug fix
- [ ] Refactor
- [ ] Performance
- [ ] Documentation/tooling
- [ ] Dependency update
- [ ] Breaking public/wire/schema/config/event change

## Gates

- [ ] `make verify` is green locally: lock-check, frozen sync, format-check, lint, imports, types, test, audit.
- [ ] Focused tests prove the behavior and negative/error paths; real boundaries have integration proof.
- [ ] Settings, `.env.example`, README, deployment, migrations, and contracts are synchronized where affected.
- [ ] `CHANGELOG.md` records every operator/caller-visible change.
- [ ] Dependency rationale and `uv.lock` diff are reviewed where affected.
- [ ] ADR is linked for a hard-to-reverse architecture/default change.

## Compatibility / Migration / Rollback

<impact, staged migration, rollback, or None.>

## Security / Operations

<trust-boundary, secret, telemetry, capacity, rollout, and alert impact, or None.>

## ADR

<link or N/A>
