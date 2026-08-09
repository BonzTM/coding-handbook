# Checklist: Incident Response

Use during detection, stabilization, recovery, and follow-up. Preserve evidence and protect sensitive data.

## Declare And Stabilize

- [ ] Incident lead, severity, start time, affected user promise, and communication channel are recorded.
- [ ] Current release/config/migration identity and last known good state are captured.
- [ ] User impact, scope, data/security implications, and SLO burn are bounded with evidence.
- [ ] First mitigation is reversible and follows the runbook; destructive actions require explicit approval.
- [ ] Change freeze and communication cadence are set when severity requires them.

## Diagnose And Mitigate

- [ ] Timeline separates observed facts, hypotheses, decisions, actions, and results.
- [ ] Logs/traces/metrics are queried with bounded dimensions and no secret exposure.
- [ ] Dependency, database, queue, event-loop, pool, config, credential, and rollout hypotheses are tested deliberately.
- [ ] Rollback compatibility covers schema, messages, caches, config, and irreversible external effects.
- [ ] Retries, replay, backfill, or traffic shift is bounded, rate-limited, idempotent, and observed.
- [ ] Potential security/privacy incident follows the private escalation and evidence-retention path.

## Recover And Verify

- [ ] User-facing SLI, data consistency, readiness, backlog, and dependency health show recovery.
- [ ] Recovery holds through the documented observation window.
- [ ] Temporary access, flags, scaling, routing, and debug changes have owners and expiry.
- [ ] Stakeholders receive impact, mitigation, current state, and next update without speculation.

## Follow-Up

- [ ] End time, duration, impact, detection gap, contributing conditions, and recovery evidence are recorded.
- [ ] Blameless review produces owned, prioritized, dated corrective actions.
- [ ] Tests, alerts, runbook, capacity, rollback, and handbook/ADR surfaces are updated as needed.
- [ ] Incident is closed only after temporary mitigations are removed or tracked.
