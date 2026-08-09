# Checklist: Handoff

Use with [onboarding-and-handoff.md](../onboarding-and-handoff.md).

## Repository And Decisions

- [ ] Incoming owner can explain purpose, shape, entrypoint, boundaries, and critical contracts from project docs.
- [ ] `README.md`, `AGENTS.md`, ADRs, changelog, architecture, and Change Routing are current.
- [ ] Open proposals, exceptions, flags, deprecations, migrations, and known debt have owners and dates.
- [ ] CODEOWNERS, repository permissions, review rules, and dependency automation route correctly.

## Build And Delivery

- [ ] Incoming owner completes clean Node 24 `npm ci` and `npm run verify` unaided.
- [ ] Incoming owner builds and runs the production ESM artifact, frontend bundle, image, or packed library.
- [ ] CI/release permissions, artifact provenance, deploy stages, and rollback are understood.
- [ ] A deploy dry-run and rollback walkthrough succeed without the outgoing owner acting.

## Runtime And Data

- [ ] Config keys, secret locations, access grants, and rotation procedures are documented and exercised.
- [ ] Schema, migration, backfill, restore, retention, deletion, event replay, and cache behavior are understood.
- [ ] SLOs, dashboards, alerts, runbooks, capacity, dependencies, and failure modes are current.
- [ ] Incoming owner can identify running version, diagnose a page, mitigate safely, and verify recovery.

## Access And Responsibility

- [ ] On-call, escalation, service catalog, dashboard, alert, registry, cloud, database, and secret ownership transfer.
- [ ] Incoming access is tested; outgoing excess access is revoked on the agreed date.
- [ ] Private vulnerability, security incident, privacy request, and audit-log escalation paths are known.
- [ ] Dependency patch cadence and audit exception ownership transfer.

## Proof

- [ ] Incoming owner answers every day-one question without tribal knowledge.
- [ ] Both owners sign off on open risks, evidence links, and effective transfer time.
- [ ] Missing artifacts are tracked as handoff defects, not silently waived.
