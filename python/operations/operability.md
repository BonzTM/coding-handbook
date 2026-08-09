# Operability

Operational targets, alerts, dashboards, and runbooks that make emitted telemetry useful under pressure.

## Default Approach

[Observability](observability.md) provides signals. Operability turns them into user-visible objectives, symptom alerts, and a current [runbook](../templates/runbook.md) that an on-call engineer can execute without tribal knowledge.

### Service Level Indicators

| Service shape | SLI | Evidence |
|---|---|---|
| HTTP/gRPC | availability: valid requests completed successfully / valid requests | request counter by route/status class |
| HTTP/gRPC | latency: valid requests below threshold / valid requests | request-duration histogram |
| worker | timeliness: work completed inside deadline / work due | handler-duration and deadline counters |
| worker | freshness: oldest ready work below threshold / observation windows | backlog-age gauge/histogram |
| CLI/library | successful owned operation / attempted operation, when operated as a product | exit/result telemetry or release evidence |

One SLI expresses one user-visible promise as good events divided by valid events. Do not use averages, pod health, or raw counts as availability. Document denominator exclusions such as malformed requests and planned maintenance; exclusions never hide server failure. Histogram buckets straddle the stated latency thresholds before an SLO depends on them.

### SLOs And Error Budgets

State every SLO as `target over rolling window`, with a named owner and data source. The error budget is the allowed bad-event fraction over that window. The owner defines what happens at warning and exhaustion—usually slowing risky change and prioritizing reliability—not a ceremonial target with no consequence.

Set starting targets from user need and measured baseline, then review them on a scheduled cadence. Never promise 100%. Separate critical journeys when their promises differ; do not create per-user or per-tenant metric series to do so.

### Symptom-Based Alerting

Page on fast error-budget burn for user-visible, actionable symptoms. Open a ticket for slow burn, capacity trends, certificate expiry, dependency deprecation, or other business-hours work. CPU, restart count, GC activity, and a single failed probe are diagnostic causes, not pages by themselves.

Use multi-window burn-rate alerts so sustained fast and slow failures surface without paging on spikes. Every page names the SLO, observed impact, dashboard, runbook section, owner, and escalation path. If the responder cannot act now, downgrade the notification channel.

### Python Runtime Signals

Measure event-loop lag for async services, owned task count, in-flight work, semaphore utilization, queue depth/oldest age, HTTP/DB pool checkout wait, process RSS, file descriptors, and restart/OOM state. Persistent loop lag indicates blocking or saturation; a growing task/queue count indicates lost ownership or backpressure failure.

GC pause metrics are not a baseline requirement. Python GC counts alone rarely explain user symptoms; add pause/allocation profiling only after latency or CPU evidence points there. Memory growth and retained objects matter: graph RSS/heap trends during load and soak tests and investigate monotonic growth instead of masking it with restarts.

### Dashboards

Version dashboard and alert definitions with the repo. The primary dashboard shows SLI/error-budget burn, traffic/work rate, latency distribution, error classes, saturation/backpressure, dependency behavior, deploy/version markers, and relevant resource trends. Labels remain bounded; dashboards link request/trace IDs into logs or traces instead of metric dimensions.

Every rollout opens the affected dashboard before traffic shifts. A panel without an operator decision or diagnostic use is removed.

### Runbooks And On-Call Handoff

Every operated service copies [the runbook template](../templates/runbook.md) to `docs/runbook.md`. It records ownership, SLOs/error-budget policy, alerts, dashboards, dependencies, deploy/rollback, scaling bounds, drain, migrations, replay/DLQ, secret rotation, common failure modes, communication, and escalation.

Update the runbook in the same change that alters an operational procedure. Walk [the handoff checklist](../checklists/handoff.md) with incoming and outgoing owners. The acceptance test is a new on-call engineer diagnosing and mitigating a representative alert from the runbook alone.

## Common Mistakes And Forbidden Patterns

- SLI expressed as an average, raw count, pod state, or infrastructure utilization instead of a valid-event ratio.
- SLO without target, window, owner, data source, or error-budget action.
- Per-user/tenant/request SLO labels that explode cardinality or expose PII.
- Paging on CPU, GC, restart, queue count, or single spikes without proven user impact.
- Event-loop lag, task/queue growth, pool wait, or RSS absent from an async service capacity view.
- GC dashboards added by reflex while memory growth and blocking remain unmeasured.
- Dashboard or alert edited only in a provider UI and lost at handoff.
- Missing/stale runbook, unreachable escalation path, or procedure known only to the previous owner.

## Verification And Proof

```bash
make verify
# Run the project's alert-rule and dashboard validation commands.
# Run one game-day alert and execute docs/runbook.md unaided.
```

Each service has a ratio SLI, `target over window` SLO, owner, budget policy, versioned dashboard, and burn-rate alert. Inspect alert definitions: every page is user-impacting and actionable. Load/soak evidence shows loop lag, task/queue/pool saturation, and memory growth remain within documented bounds. An incoming on-call owner can receive, diagnose, mitigate, roll back, and escalate a representative alert using only the current runbook.

## Related

- [observability](observability.md)
- [resilience](resilience.md)
- [rollout and SLO readiness](../checklists/rollout-and-slo-readiness.md)
- [incident response](../checklists/incident-response.md)
