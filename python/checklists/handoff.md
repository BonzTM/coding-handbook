# Handoff Checklist

Ownership-transfer gate for a Python repo. Walk it with outgoing and incoming owners; use [onboarding and handoff](../onboarding-and-handoff.md) for the day-one reading path.

## Ownership And Access

- [ ] `CODEOWNERS`, repository teams, review routes, and escalation contacts name the incoming owner; obsolete access is removed.
- [ ] Build, CI, registry/PyPI, staging, production, database, broker, telemetry, audit, secrets, and incident access is granted and verified live.
- [ ] Outgoing access is revoked where it should not survive ownership.
- [ ] Incoming owner can authenticate from a clean workstation using documented least-privilege paths.

## On-Call And Escalation

- [ ] Incoming owner is active in the rotation; outgoing owner is removed on the agreed date.
- [ ] Page, ticket, dashboard, SLO, status, and incident-channel ownership routes correctly.
- [ ] Primary and backup escalation contacts were tested and acknowledgement expectations are known.

## Secrets And Configuration

- [ ] Every setting/secret names source, scope, consumer, owner, rotation mechanism, and last/next rotation evidence.
- [ ] `.env.example`, pydantic-settings models, README table, deployment manifest, and runbook agree.
- [ ] Incoming owner can rotate one representative credential and verify rolling restart/re-read without exposure.
- [ ] No outgoing personal token remains in CI, registry, PyPI, cloud, or operational tooling.

## Decisions And Knowledge

- [ ] Accepted/proposed/superseded ADRs and unresolved architecture risks are reviewed.
- [ ] Package map, core/adapter boundaries, FastAPI lifespan, asyncio ownership, SQLAlchemy/Alembic, and verify gate are explained from checked-in docs.
- [ ] Tribal procedures are converted into runbook, recipe, ADR, or tracked issue before transfer.
- [ ] Current dependency exceptions, security findings, deprecations, feature flags, migration phases, and compatibility windows have owners/dates.

## Operations And Maintenance

- [ ] Runbook matches current deploy, migration, rollback, scale, drain, replay/DLQ, secret rotation, alert, and incident behavior.
- [ ] Incoming owner can locate SLO/error-budget policy, dashboards, burn alerts, audit evidence, and data inventory/retention jobs.
- [ ] Known failure modes and first mitigations were exercised in a game day or staging drill.
- [ ] Dependency-bot triage, Python pin/floor review, `uv.lock` ownership, `make audit`, and release cadence are transferred.
- [ ] Package/container release, Trusted Publishing or registry identity, immutable digest, changelog, and rollback procedure are walked through.

## Verification

The handoff is complete only when the incoming owner performs these unaided from a clean checkout:

```bash
uv lock --check
uv sync --frozen
make verify
uv run --with pip-audit pip-audit
uv run pytest -m integration
# Run the project's staging deploy and rollback dry-run commands.
```

- [ ] Clean checkout passed the canonical gate with the committed lockfile.
- [ ] Integration lane and one installed-artifact/container smoke path passed.
- [ ] Incoming owner completed a deploy and rollback dry-run, received a test alert, and used the runbook without outgoing-owner help.
- [ ] Incoming owner can answer every day-one question in [onboarding and handoff](../onboarding-and-handoff.md).
