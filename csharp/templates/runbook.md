<!--
Destination: docs/runbook.md

Operational runbook for <SERVICE_NAME>. Contract:
- Every service ships this file and keeps it current; a stale runbook is a release defect.
- A new on-call engineer must be able to resolve any alert below using ONLY this file,
  without paging the owner. If they cannot, the runbook has a gap — fix it.
- Update this file in the SAME change that alters deploy, rollback, alerts, SLOs, or behavior.
- See csharp/operations/operability.md for the SLO/alerting contract this template implements.
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
- **Entrypoints:** `src/<APP>.Api` — <WHAT IT LISTENS ON / CONSUMES>

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

### Runtime Deep-Dive Tools

When dashboards do not explain the symptom, go to the process (see
csharp/operations/observability.md):

- **Live counters:** `dotnet-counters monitor --process-id <PID>` — CPU, GC pause/heap, ThreadPool queue length, Kestrel connections/requests. First stop for latency and memory alerts.
- **Dump capture:** `dotnet-dump collect -p <PID>` then `dotnet-dump analyze <dump>` (`clrstack`, `dumpheap -stat`) — hangs, leaks, deadlocks.
- **In-cluster:** the chiseled runtime image has no shell; use `kubectl debug <pod> --image=mcr.microsoft.com/dotnet/sdk:10.0 --target=<container>` to attach an ephemeral container that carries the tools, or run `dotnet monitor` as a sidecar if the platform standardizes on it.

### Alert: <ALERT NAME — e.g. Availability fast burn>

- **Fires when:** <BURN-RATE CONDITION AGAINST WHICH SLO>.
- **User impact:** <WHAT USERS SEE>.
- **Diagnose:**
  1. Check <DASHBOARD/QUERY> for error-class breakdown (ProblemDetails `status`/type distribution).
  2. Check recent deploys: <HOW TO SEE LAST DEPLOY/RELEASE TAG>.
  3. Check dependency health: <DOWNSTREAM CHECKS> — see [Dependencies And Failure Modes](#dependencies-and-failure-modes).
- **Remediate:**
  - If caused by the latest release: [Rollback](#rollback).
  - If a downstream dependency is down: <FALLBACK / DEGRADE STEP>.
  - If load-driven: [Scale](#scale).
- **Escalate if:** <CONDITION UNRESOLVED AFTER X>.

### Alert: <ALERT NAME — e.g. Latency slow burn (ticket)>

- **Fires when:** <CONDITION>.
- **Diagnose:** <STEPS — start with `dotnet-counters` (ThreadPool starvation, GC pauses) before guessing>.
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
- Pre-deploy gate: `pwsh ./verify.ps1` must pass (the release workflow re-runs it on the tag). Post-deploy: <SMOKE CHECK / READINESS CONFIRMATION>.
- Migrations run as the explicit pre-rollout Job (`--migrate`) — confirm it completed before the Deployment rolls.

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

- Safe bounds: <MIN..MAX>. Connection-pool limits: <SEE csharp/services/database.md GUIDANCE; MAX POOL SIZE = ?>.

### Drain

```bash
<DRAIN / GRACEFUL-SHUTDOWN STEP>
```

- The service drains in-flight work on SIGTERM (`IHostApplicationLifetime`; Kestrel stops accepting, in-flight requests finish). Drain timeout: `HostOptions.ShutdownTimeout` = <SECONDS>, which must stay under the platform's `terminationGracePeriodSeconds`.
- For workers: <STOP CONSUMING (`stoppingToken`), FINISH IN-FLIGHT, ACK/NACK BEHAVIOR>.

## Dependencies And Failure Modes

| Dependency | Purpose | If it fails | Mitigation |
|---|---|---|---|
| <DATABASE> | <PRIMARY STORE> | <SYMPTOM> | <RETRY/TIMEOUT/READ-REPLICA — see resilience> |
| <DOWNSTREAM API> | <WHAT IT PROVIDES> | <SYMPTOM> | <CIRCUIT BREAKER / FALLBACK / DEGRADE> |
| <QUEUE/BROKER> | <ASYNC WORK> | <SYMPTOM> | <DLQ / REPLAY> |

- Readiness (`/readyz`) reflects critical-dependency health and goes red when <X> is unavailable.

## Configuration And Secrets

- **Config source:** `appsettings.json` (committed, no secrets) + environment-variable overrides; options validated fail-fast at startup (`ValidateOnStart`). Required keys: <LIST OR LINK the README config table>.
- **Where config lives:** <DEPLOY CONFIG LOCATION>.
- **Where secrets live:** <SECRET STORE + PATHS + SCOPE>. Access granted by <WHO>. Local dev uses `dotnet user-secrets` only.
- **Rotation:** <PROCEDURE PER SECRET, WHO OWNS IT>. See csharp/operations/security.md.

## Recovery Procedures

- **Service won't start:** <DIAGNOSE: options validation failure (the startup exception names the bad key), missing secret, port conflict; FIX>.
- **Data corruption / bad write:** <RESTORE-FROM-BACKUP STEPS, BACKUP LOCATION, RPO>.
- **Stuck/poison work item:** <IDENTIFY, REMOVE/REPLAY, SKIP>.
- **Full outage:** <COLD-START / DISASTER-RECOVERY STEPS, DEPENDENCIES TO BRING UP FIRST>.
- **Backups:** location <...>, retention <...>, last verified restore <DATE>.

## Related

- Operational contract this runbook implements: csharp/operations/operability.md
- Signals behind the SLIs: csharp/operations/observability.md
- Failure handling: csharp/operations/resilience.md
- Deploy/rollback packaging: csharp/operations/deployment.md
- Taking over operations: csharp/onboarding-and-handoff.md
