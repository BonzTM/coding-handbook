<!-- Copy to docs/runbook.md. Replace every <placeholder>; every production alert needs a procedure. -->

# Runbook: <service-name>

<one-line purpose and consumers.>

## Overview And Owners

- **Shape:** <service | worker | CLI | library>
- **Repo:** <repo-url>
- **Owner/on-call:** <team, rotation, links>
- **Escalation:** <ordered escalation path>
- **Environments:** <names, regions, URLs>
- **Entrypoint:** `src/<app>/__main__.py` — <surface>

## SLOs And Dashboards

| SLI | SLO / window | Error budget | Owner |
|---|---|---|---|
| availability | <target/window> | <budget> | <owner> |
| latency | <target/window> | <budget> | <owner> |
| <freshness/lag> | <target/window> | <budget> | <owner> |

- **Budget exhaustion policy:** <action>.
- **Dashboards/logs/traces:** <links and repository definitions>.

## On-Call And Escalation

1. First responder: <rotation/contact>.
2. Escalate after <condition> to <owner>.
3. Declare an incident at <impact/threshold>; process: <link>.

## Common Alerts: Diagnosis And Remediation

### Alert: <symptom and burn rate>

- **User impact:** <impact>.
- **Diagnose:** check <dashboard>, recent releases, saturation, and dependency health.
- **Remediate:** <rollback, degrade, scale, or dependency procedure>.
- **Escalate if:** <condition/time>.

### Alert: <worker lag or readiness failure>

- **Diagnose:** <queue/DB/upstream/log/trace steps>.
- **Remediate:** <bounded replay, poison handling, scale, or rollback>.

## Key Operations

### Deploy

```bash
<deploy command or pipeline trigger>
```

Confirm `make verify`, the explicit Alembic migration job, rollout readiness, and smoke checks.

### Rollback

```bash
<redeploy previous digest or disable feature command>
```

Migration caveat: <forward-only/expand-contract recovery>.

### Scale And Drain

```bash
<scale command>
<drain or rollout command>
```

Safe replica/pool/concurrency bounds: <numbers and evidence>. Shutdown grace: <seconds>, below platform termination grace.

## Dependencies And Failure Modes

| Dependency | Purpose | Failure symptom | Mitigation |
|---|---|---|---|
| <PostgreSQL> | <state> | <symptom> | <timeout/degrade/restore> |
| <upstream> | <capability> | <symptom> | <bounded retry/fallback> |
| <broker> | <work> | <symptom> | <DLQ/replay> |

## Configuration And Secrets

- Settings source/table: <README/.env.example link>; pydantic-settings validates before startup.
- Runtime config/secret store and access: <locations and owners>.
- Rotation: <per-secret procedure without exposing values>.

## Recovery Procedures

- **Startup failure:** <settings, secret, bind, migration diagnosis>.
- **Bad data/write:** <backup restore, RPO/RTO, validation>.
- **Poison/stuck work:** <identify, quarantine, replay>.
- **Full outage:** <cold-start order and dependency prerequisites>.
- **Backups:** <location, retention, last restore proof>.
