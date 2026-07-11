# Handoff Checklist

Ownership-transfer checklist for a .NET repo built from this handbook. Walk it with both the outgoing and incoming owner present. See [../onboarding-and-handoff.md](../onboarding-and-handoff.md) for the day-one reading path and the questions the new owner must be able to answer.

## Ownership And Access

- [ ] `CODEOWNERS` updated so reviews and notifications route to the new owner; old owner removed where no longer warranted.
- [ ] Deploy and production access granted to the new owner.
- [ ] Deploy and production access revoked from the outgoing owner where it should not outlive their ownership.
- [ ] Any cloud, registry, or infra console roles transferred or reassigned.
- [ ] New owner can authenticate to every system needed to build, deploy, and operate the repo, verified live.

## On-Call And Escalation

- [ ] New owner added to the on-call rotation; outgoing owner removed.
- [ ] Escalation path documented (who to page, in what order) and confirmed reachable by the new owner.
- [ ] Alert routes and dashboard ownership transferred so pages and metrics reach the new owner.

## Secrets And Configuration

- [ ] Location of every secret documented (store, path, scope).
- [ ] Rotation procedure documented for each secret, with who is now responsible.
- [ ] The README config table and `appsettings.json` confirmed in sync with the bound options classes — every required key, env-var override, and default is accurate today.
- [ ] New owner has access to read and rotate secrets, verified live.

## Decisions And Knowledge

- [ ] All open and proposed ADRs in `decisions/` surfaced to the new owner with current status.
- [ ] Undocumented decisions and tribal knowledge converted into an ADR or runbook entry before transfer.
- [ ] Rationale behind the load-bearing architectural choices walked through and understood.

## Operations And Maintenance

- [ ] Runbook confirmed current: deploy steps, rollback, SLOs, alerts, and dashboards match reality today.
- [ ] SLOs and their dashboards identified and accessible to the new owner.
- [ ] Known failure modes and first fixes reviewed.
- [ ] Dependency-update ownership transferred, including who drains the Dependabot queue and acts on `dotnet list package --vulnerable --include-transitive` findings per [dependency-upgrade.md](dependency-upgrade.md).
- [ ] Release and tagging procedure (`v1.2.3`, who triggers, how to roll back) walked through.

## Verification

The handoff is complete only when the new owner can do the following unaided, from a clean checkout, without the outgoing owner present:

```powershell
# 1. Clean checkout passes the canonical safety gate locally.
pwsh ./verify.ps1

# 2. Vulnerability and dependency posture is current.
dotnet list package --vulnerable --include-transitive

# 3. New owner performs a deploy dry-run independently.
#    (Use the project's documented dry-run / staging deploy command.)
```

- [ ] New owner ran `pwsh ./verify.ps1` from a clean clone and it passed on their machine.
- [ ] New owner performed a deploy dry-run independently.
- [ ] New owner can answer every day-one question in [../onboarding-and-handoff.md](../onboarding-and-handoff.md) without help.
