# Checklist: Rollout And SLO Readiness

Use this before exposing a new release or service to production traffic.

## Ownership And Objectives

- [ ] Service catalog names purpose, owners, on-call, dependencies, data class, repo, and deployable.
- [ ] Each SLI states user-visible event, population, exclusions, measurement point, and unit.
- [ ] Each SLO states target and rolling window with an error-budget owner.
- [ ] Capacity evidence covers throughput, tail latency, event-loop delay, queue, pool, and memory bounds.
- [ ] Version and artifact digest are visible to operators.

## Detection And Response

- [ ] Dashboards show traffic, errors, latency, saturation, release markers, readiness, and backlog where relevant.
- [ ] Pages correspond to urgent user impact and link an actionable runbook.
- [ ] Alerts name severity, owner, SLO/risk, dashboard, first safe action, and escalation.
- [ ] Runbook diagnosis is bounded, read-only by default, and cannot expose secrets.
- [ ] Dependency outage, credential expiry, bad config, saturation, backlog, telemetry loss, and rollback are covered.

## Rollout

- [ ] Readiness stays false until initialization and becomes false before drain.
- [ ] Migration is an explicit job and old/new versions coexist safely.
- [ ] Canary steps, traffic increments, observation windows, and stop/abort thresholds are written.
- [ ] Orchestrator grace exceeds bounded application drain and telemetry flush.
- [ ] Rollback identifies artifact, config, schema, messages, caches, and irreversible effects.

## Proof

- [ ] Synthetic or controlled failure routes the alert to the runbook and owner.
- [ ] Load/fault evidence demonstrates overload response and recovery.
- [ ] Container smoke proves probes, non-root runtime, `SIGTERM`, drain, and exit budget.
- [ ] On-call can deploy, diagnose, rollback, and validate recovery without repository authors.
