<!--
Destination: docs/runbook.md

Operational runbook for <SERVICE_NAME>. Contract:
- Every service ships this file and keeps it current; a stale runbook is a release defect.
- A new on-call engineer must be able to resolve any alert below using ONLY this file,
  without paging the owner. If they cannot, the runbook has a gap — fix it.
- Update this file in the SAME change that alters deploy, rollback, alerts, SLOs, or behavior.
- See golang/operations/operability.md for the SLO/alerting contract this template implements.
- Replace every <PLACEHOLDER>. Delete sections that genuinely do not apply, and say why.
-->

# Runbook: <SERVICE_NAME>

One-line purpose: <WHAT THIS SERVICE DOES AND FOR WHOM>.

## Overview And Owners

- **Shape:** <service | worker/consumer | CLI | library>
- **Repo:** <REPO_URL>
- **Owner (primary):** <PERSON OR ROTATION>
- **On-call rotation:** <ROTATION NAME + LINK>
- **Escalation:** <WHO TO PAGE WHEN STUCK, IN ORDER> — see [On-Call And Escalation](#on-call-and-escalation)
- **Environments:** <prod / staging / ...> at <URLS OR REGIONS>
- **Entrypoints:** `cmd/<APP>` — <WHAT IT LISTENS ON / CONSUMES>

## SLOs And Dashboards

SLOs follow `target over rolling window`; the gap is the error budget. Owner decides budget-exhaustion policy. See operability for the contract.

| SLI | SLO (target / window) | Error budget | Owner |
|---|---|---|---|
| Availability (non-5xx ÷ valid requests) | <99.9% / 30d> | <~43 min/30d> | <OWNER> |
| Latency (served under <300ms> ÷ valid requests) | <99% / 30d> | <...> | <OWNER> |
| <Worker freshness / lag, if applicable> | <99% within deadline / 7d> | <...> | <OWNER> |

- **Error-budget policy:** when the budget is exhausted, <FREEZE RISKY CHANGE / WHAT THE OWNER DOES>.
- **Dashboards (as code in repo):** <PATH TO DASHBOARD DEFINITIONS> rendered at <DASHBOARD URLS>.
- **Metrics / logs / traces:** <LINKS TO QUERIES OR SAVED VIEWS>.

## On-Call And Escalation

1. **First responder:** <ON-CALL ROTATION / HOW TO REACH>.
2. **Escalate to:** <SECONDARY / DOMAIN EXPERT> after <TIME OR CONDITION>.
3. **Last resort:** <MANAGER / INCIDENT COMMANDER>.
4. **Declare an incident when:** <THRESHOLD, e.g. fast-burn page + customer impact>. Process: <LINK>.

## Common Alerts: Diagnosis And Remediation

Each alert is a SYMPTOM tied to an SLO burn rate, not a cause. Diagnose from signals, then remediate.

### Alert: <ALERT NAME — e.g. Availability fast burn>

- **Fires when:** <BURN-RATE CONDITION AGAINST WHICH SLO>.
- **User impact:** <WHAT USERS SEE>.
- **Diagnose:**
  1. Check <DASHBOARD/QUERY> for error-class breakdown.
  2. Check recent deploys: <HOW TO SEE LAST DEPLOY/RELEASE TAG>.
  3. Check dependency health: <DOWNSTREAM CHECKS> — see [Dependencies And Failure Modes](#dependencies-and-failure-modes).
- **Remediate:**
  - If caused by the latest release: [Rollback](#rollback).
  - If a downstream dependency is down: <FALLBACK / DEGRADE STEP>.
  - If load-driven: [Scale](#scale).
- **Escalate if:** <CONDITION UNRESOLVED AFTER X>.

### Alert: <ALERT NAME — e.g. Latency slow burn (ticket)>

- **Fires when:** <CONDITION>.
- **Diagnose:** <STEPS>.
- **Remediate:** <STEPS>.

### Alert: <WORKER: Backlog/lag>

- **Fires when:** <BACKLOG AGE EXCEEDS THRESHOLD>.
- **Diagnose:** <QUEUE DEPTH, CONSUMER HEALTH, DLQ COUNT>.
- **Remediate:** <SCALE CONSUMERS / DRAIN DLQ / FIX POISON MESSAGE>.

<!-- Add one block per alert. Every page in the alerting backend MUST have a block here. -->

## Key Operations

### Deploy

```bash
<DEPLOY COMMAND OR PIPELINE TRIGGER>
```

- Release tagging: `v<MAJOR>.<MINOR>.<PATCH>` (v prefix required). Triggered by <WHO/WHAT>.
- Pre-deploy gate: `make verify` must pass. Post-deploy: <SMOKE CHECK / READINESS CONFIRMATION>.

### Rollback

```bash
<ROLLBACK COMMAND — redeploy previous tag, revert, or feature-flag off>
```

- Previous good version: <HOW TO IDENTIFY>. Expected recovery time: <RTO>.
- Data/migration caveats: <FORWARD-ONLY MIGRATIONS? HOW TO HANDLE>.

### Scale

```bash
<SCALE COMMAND — replicas / pool size / consumer count>
```

- Safe bounds: <MIN..MAX>. Connection-pool limits: <SEE database.md GUIDANCE; POOL SIZE = ?>.

### Drain

```bash
<DRAIN / GRACEFUL-SHUTDOWN STEP>
```

- The service drains in-flight work on SIGTERM; see graceful-shutdown contract. Drain timeout: <SECONDS>.
- For workers: <STOP CONSUMING, FINISH IN-FLIGHT, ACK/NACK BEHAVIOR>.

## Dependencies And Failure Modes

| Dependency | Purpose | If it fails | Mitigation |
|---|---|---|---|
| <DATABASE> | <PRIMARY STORE> | <SYMPTOM> | <RETRY/TIMEOUT/READ-REPLICA — see resilience> |
| <DOWNSTREAM API> | <WHAT IT PROVIDES> | <SYMPTOM> | <CIRCUIT BREAKER / FALLBACK / DEGRADE> |
| <QUEUE/BROKER> | <ASYNC WORK> | <SYMPTOM> | <DLQ / REPLAY> |

- Readiness (`/readyz`) reflects critical-dependency health and goes red when <X> is unavailable.

## Configuration And Secrets

- **Config source:** `internal/config`, loaded from env + flags, validated fail-fast at startup. Required keys: <LIST OR LINK `.env.example`>.
- **Where config lives:** <DEPLOY CONFIG LOCATION>.
- **Where secrets live:** <SECRET STORE + PATHS + SCOPE>. Access granted by <WHO>.
- **Rotation:** <PROCEDURE PER SECRET, WHO OWNS IT>. See operations/security.md.

## Recovery Procedures

- **Service won't start:** <DIAGNOSE: config validation failure, missing secret, port conflict; FIX>.
- **Data corruption / bad write:** <RESTORE-FROM-BACKUP STEPS, BACKUP LOCATION, RPO>.
- **Stuck/poison work item:** <IDENTIFY, REMOVE/REPLAY, SKIP>.
- **Full outage:** <COLD-START / DISASTER-RECOVERY STEPS, DEPENDENCIES TO BRING UP FIRST>.
- **Backups:** location <...>, retention <...>, last verified restore <DATE>.

## Related

- Operational contract this runbook implements: golang/operations/operability.md
- Signals behind the SLIs: golang/operations/observability.md
- Failure handling: golang/operations/resilience.md
- Deploy/rollback packaging: golang/operations/deployment.md
- Taking over operations: golang/onboarding-and-handoff.md
