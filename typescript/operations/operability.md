# Operability

SLO, alert, dashboard, runbook, capacity, and ownership rules for production TypeScript systems.

## Default Approach

Every deployed process has named owners, user-centered service objectives, actionable alerts, and exercised runbooks.

### Ownership And Service Catalog

Record service purpose, repository, deployable, owners, on-call rotation, dependencies, data classification, dashboards, alerts, runbooks, SLOs, and escalation paths. Ownership changes update access and handoff evidence before responsibility moves.

Expose build version and safe runtime identity to operators. The running artifact must be traceable to source, lockfile, image digest, and release record.

### Service-Level Objectives

Define indicators from user-visible outcomes: successful eligible requests, end-to-end latency, durable job completion, freshness, or correctness. State population, exclusions, measurement point, window, and target.

Use an error budget to govern reliability tradeoffs. Do not define 100% objectives without a hard external mandate and a costed design. Dependency objectives inform diagnosis but do not replace the service's user-facing SLO.

### Alerts

Page on urgent user-impacting symptoms with actionable response. Use multi-window burn-rate or equivalent sustained-impact logic where the platform supports it. Ticket on slow trends; dashboard informational signals.

Every alert names severity, owner, SLO or risk, dashboard, runbook, and safe first action. Avoid pages on raw CPU, one exception, pod restart, or downstream health without demonstrated user impact.

### Dashboards

Start with traffic, errors, latency, saturation, and current release markers. Add dependency and business panels that answer known diagnostic questions. Use bounded dimensions and consistent units.

Show readiness, queue/backlog age, pool wait, event-loop delay, retry/shed rate, and deploy/config changes where material. A dashboard is not a substitute for an alert or runbook.

### Runbooks

Use [../templates/runbook.md](../templates/runbook.md). Include symptom, impact, prerequisites, safe diagnosis, mitigations, rollback, verification, escalation, data sensitivity, and follow-up.

Commands are copy-pasteable, bounded, read-only by default, and do not print secrets. Destructive, privacy-sensitive, or high-blast-radius steps require explicit warnings and approval paths.

### Capacity And Failure Modes

Document expected throughput, tail latency, queue/pool limits, memory bounds, dependency quotas, scaling signal, and saturation behavior. Review capacity after material traffic, payload, dependency, or runtime changes.

Maintain known failure modes for dependency outage, credential expiry, certificate expiry, cache loss, broker backlog, database saturation, telemetry failure, bad config, and deploy rollback. Exercise the high-risk paths.

### Operational Changes

New config, ports, migrations, contracts, limits, SLOs, alerts, or manual actions are operator-visible and require changelog/runbook updates. Rollout defines observation window and abort criteria.

## Common Mistakes And Forbidden Patterns

- No named owner, stale contacts, or a runbook reachable only during healthy authentication.
- Infrastructure alerts presented as user-impact alerts.
- High-cardinality dashboards that cannot be queried reliably.
- Runbooks containing unbounded queries, secret-printing commands, or destructive first steps.
- SLOs without population/exclusions or targets chosen without product impact.
- Capacity assumptions without load evidence.
- Operator-visible changes absent from changelog and rollout notes.

## Verification And Proof

- Service catalog entry links current owners, SLOs, dashboards, alerts, and runbooks.
- Synthetic or controlled failure proves each page routes to an actionable runbook.
- Dashboard and SLO queries use bounded dimensions and match documented populations.
- A game day exercises dependency failure, bad rollout, backlog, or restore at the appropriate cadence.
- Capacity test demonstrates configured queue, pool, memory, and latency behavior.
- On-call can identify version, rollback, and validate recovery without repository authors.

Related: [observability.md](observability.md), [resilience.md](resilience.md), and [deployment.md](deployment.md).
