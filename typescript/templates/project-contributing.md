# Contributing

## Before You Start

- Read `AGENTS.md` and route the change.
- Link the issue, acceptance criteria, and ADR when required.
- Keep one logical change per pull request.

## Local Proof

```bash
npm ci
npm run verify
```

Run the narrowest relevant test first. Run `npm run test:integration` when PostgreSQL or another real dependency changes.

## Pull Requests

- Use a Conventional Commits subject.
- Explain user/operator impact, failure behavior, compatibility, rollout, and rollback.
- Include tests, config, schemas, migrations, changelog, and runbook changes with behavior.
- Review manifest, lockfile, generated artifacts, and packed/image contents.
- Do not merge with warnings, unexplained suppressions, or unresolved required checks.

## Security

Do not open a public issue for an unpatched vulnerability. Follow `SECURITY.md`. Never commit or paste secrets, production data, or sensitive logs.
