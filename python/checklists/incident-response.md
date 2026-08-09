# Incident Response Checklist

> **Team-process document.** This guides production responders; it is not an app-generation contract.

Stabilize before root cause. Use the service [runbook](../templates/runbook.md), SLO contract in [operability](../operations/operability.md), and signals in [observability](../operations/observability.md).

## Acknowledge & Assess

- [ ] Acknowledge within the rotation SLA and confirm one owner.
- [ ] Declare severity with the project scale; start higher when impact is uncertain.
- [ ] Open the incident channel/document and record every action with an aware UTC timestamp.
- [ ] Name an incident commander and explicit handoff/escalation path.
- [ ] State user impact in one sentence: who, which journey, scope, and severity.
- [ ] Open the SLO dashboard; identify the failing SLI, burn rate, affected version/region/tenant class, and start time.
- [ ] Confirm the alert is a user symptom; file alert repair and stand down if it is only a flapping cause with no impact.

## Stabilize

- [ ] Review recent immutable image/package release, Alembic migration, config/secret rotation, feature flag, dependency update, and platform change.
- [ ] If correlated with the release, execute the documented prior-digest rollback unless schema/data compatibility forbids it.
- [ ] Disable the causal feature flag when that is faster and safer than redeploying.
- [ ] Scale only inside PostgreSQL/downstream pool and resource bounds; do not amplify dependency saturation.
- [ ] Shed work, pause intake, reduce concurrency, or activate documented degraded mode before queues/memory collapse.
- [ ] Fail over only through a rehearsed path; never invent an unbounded retry or replay during the incident.
- [ ] Do not edit production data or run an ad hoc migration as mitigation; use documented recovery/admin tooling.
- [ ] Record the proposed mitigation, expected signal change, owner, and rollback before applying; confirm burn responds.

## Diagnose

- [ ] Correlate impact with deploy/migration/config/dependency/platform timelines and version markers.
- [ ] Read SLI/error-class/latency, event-loop lag, task/queue/pool saturation, dependency, RSS/OOM, and restart signals.
- [ ] Query structured logs for the incident window and pivot through request/correlation/trace IDs without exposing PII.
- [ ] Follow traces across FastAPI, HTTPX, SQLAlchemy, and worker boundaries; distinguish self from downstream latency/failure.
- [ ] Check `/readyz`, PostgreSQL, brokers, downstreams, certificate/DNS, and platform events with bounded safe diagnostics.
- [ ] Form one falsifiable hypothesis and test it against evidence; avoid speculative production changes after stabilization.

## Communicate

- [ ] Send initial impact/severity/ownership and next-update time.
- [ ] Update on the declared cadence even with no change.
- [ ] Notify support/status/affected teams/leadership according to severity and the runbook.
- [ ] Keep external wording impact-focused, verified, blameless, and free of internal sensitive detail.
- [ ] Announce mitigation, recovery, and all-clear as separate states.

## Recover & Verify

- [ ] Confirm SLI recovery and stopped budget burn on the dashboard, not merely alert closure.
- [ ] Confirm `/readyz`, critical path, workers/queues, and dependency health at normal traffic.
- [ ] Reconcile data and effects: partial writes, idempotency/inbox/outbox, cache/index state, backlog, and DLQ/replay need.
- [ ] Replay or repair only through approved bounded idempotent tooling with dry-run, audit, pause, and abort.
- [ ] Track a forward fix for rollback/disabled/degraded mitigations; temporary state has owner and expiry.
- [ ] Let symptom alerts resolve through real recovery; record end time, impact, mitigation, and remaining risk.

## Postmortem

- [ ] Produce a blameless postmortem within SLA with detection-to-recovery timeline, measured impact, contributing factors, and why controls did/did not contain it.
- [ ] Create single-owner, due-dated issues for code, test, capacity, alert, runbook, data, dependency, and process gaps.
- [ ] Review detection quality: symptom/burn alert timing, missing telemetry, sampling, and unsafe/high-cardinality evidence.
- [ ] Update runbook, dashboards, alerts, resilience limits, rollout proof, and tests in the remediation changes.
- [ ] Share findings with affected owners and review actions to closure.

## Verification

```bash
# After stabilization, run the project's production-safe smoke/reconciliation commands.
make verify
uv run pytest -m integration
```

- [ ] Incident record contains UTC timeline, named IC, user impact, mitigation, and the exact point burn responded.
- [ ] Recovery evidence covers SLI, readiness, critical path, queues/workers, and no-loss/no-duplicate reconciliation.
- [ ] Postmortem and owned/due-dated actions exist; exposed runbook/alert/test defects are merged, not merely noted.

## Related

- [operability](../operations/operability.md)
- [observability](../operations/observability.md)
- [deployment](../operations/deployment.md)
- [resilience](../operations/resilience.md)
- [data handling](../operations/data-handling.md)
