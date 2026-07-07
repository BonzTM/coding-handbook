# Incident Response Checklist

> **Team-process document.** This guides humans operating the service in production. It is not part of the app-generation contract; agents building or changing code do not read it.

What an on-call engineer does, in order, when paged. Stabilize before you understand: the first job is to stop the bleeding with the most reversible action available, not to find root cause. Drive every step from the service runbook ([../templates/runbook.md](../templates/runbook.md)) and the SLO/alerting contract in [../operations/operability.md](../operations/operability.md).

## Acknowledge & Assess

- [ ] Acknowledge the page within the rotation's SLA so it stops escalating and the team knows it is owned.
- [ ] Declare a severity using the project's scale; when in doubt, declare higher — downgrading is cheap, a missed major is not.
- [ ] Open the incident channel and incident doc/timeline now; record every action with a UTC timestamp from this point on.
- [ ] Name yourself incident commander (IC) for now; hand off explicitly if scope outgrows you, and escalate per the runbook's order if you are stuck.
- [ ] Identify user impact in one sentence: who is affected, doing what, how badly.
- [ ] Pull up the SLO dashboard and read the actual burn: which SLI, how fast the error budget is draining, fast-burn vs slow-burn (per [../operations/operability.md](../operations/operability.md)).
- [ ] Confirm the page is a real symptom, not a cause-based or flapping alert; if it is the latter, file a ticket to fix the alert and stand down.

## Stabilize

Mitigate first. Pick the most reversible action that restores users, even before you know why. Consult the service [runbook](../templates/runbook.md) Key Operations and Common Alerts blocks.

- [ ] Check the timeline of recent changes first: last deploy/release tag, migration, feature-flag flip, config change, dependency bump.
- [ ] If symptoms started at the last deploy, roll it back — redeploy the previous good tag per the runbook's [Rollback](../templates/runbook.md) step (packaging/image-identity mechanics in [../operations/deployment.md](../operations/deployment.md)); this is the default first move and the most reversible.
- [ ] If a recent feature flag caused it, disable the flag (faster and more reversible than a redeploy).
- [ ] If load-driven, scale out within the runbook's safe bounds (mind connection-pool limits before adding replicas).
- [ ] If overloaded and scaling is not enough, shed load: enable rate limiting / load shedding so the service degrades instead of collapsing (per [../operations/resilience.md](../operations/resilience.md)).
- [ ] If a single dependency or region is down, fail over or activate the documented fallback / degraded mode.
- [ ] Do not run a forward-only migration as a mitigation; if the trigger was a migration, follow the runbook's contract/expand caveats, not an ad-hoc schema change.
- [ ] State the mitigation and its expected effect in the timeline before applying it, then confirm the burn rate actually responds.

## Diagnose

Only after users are stable, find the cause from signals — not by guessing or by changing production live. Use the views in [../operations/observability.md](../operations/observability.md).

- [ ] Correlate the incident start with the change timeline from Stabilize: deploys, migrations, dependency bumps, infra events.
- [ ] Read metrics: which SLI broke, error-class breakdown, latency tail, saturation of the suspected resource.
- [ ] Read logs filtered to the incident window and the affected component; pivot on `trace_id` rather than guessing.
- [ ] Follow traces across service boundaries to find where latency or errors originate (self vs downstream).
- [ ] Check critical-dependency health (DB, downstream APIs, broker) and `/readyz` to see what the service itself considers unhealthy.
- [ ] Form one falsifiable hypothesis, test it against signals; do not apply speculative fixes to a service that is already stable.

## Communicate

- [ ] Post an initial status update: impact, severity, that it is being worked, and the time of the next update.
- [ ] Update on a fixed cadence (e.g. every 30 min for a major, sooner if it changes) even when there is "nothing new" — silence reads as the incident being abandoned.
- [ ] Notify the stakeholders the severity requires (affected teams, support, status page, leadership for major) per the runbook's escalation/communication plan.
- [ ] Keep customer-facing wording impact-focused and free of internal jargon, blame, or unverified root cause.
- [ ] Announce mitigation, then recovery, then all-clear as distinct messages so readers know which phase you are in.

## Recover & Verify

- [ ] Confirm the SLI has recovered and the error budget has stopped burning on the dashboard, not just that the alert auto-resolved.
- [ ] Confirm `/readyz` is green and the service is taking traffic normally; run the runbook's post-deploy smoke check.
- [ ] Verify no data loss or corruption from the incident or the mitigation: reconcile counts, drain/replay any DLQ, check for half-applied writes.
- [ ] If you mitigated by rolling back or flagging off, confirm a forward fix is tracked so the mitigation does not become permanent by accident.
- [ ] Let the alert clear on its own (do not silence it manually) and resolve the page only after the symptom is genuinely gone.
- [ ] Close the incident, record end time in the timeline, and stand down with a one-line summary of impact and mitigation.

## Postmortem

- [ ] Write a blameless postmortem within the project's SLA (e.g. N business days); focus on systems and contributing factors, not individuals.
- [ ] Include the full timeline (detection -> mitigation -> recovery), measured user impact, and the contributing causes.
- [ ] File action items as tracked issues, each with a single named owner and a due date; "improve monitoring" without an owner is not an action item.
- [ ] Capture detection gaps: was the page timely, was it a symptom alert, did the SLO/burn-rate threshold behave? Update alerts per [../operations/operability.md](../operations/operability.md).
- [ ] Update the [runbook](../templates/runbook.md) in the same effort: a missing/wrong diagnosis or remediation step that slowed you is a runbook defect — fix it now.
- [ ] Share the postmortem with the team and review action items to closure; an action item that never lands means the next incident is identical.

## Verification

The incident is fully handled only when all of the following exist and can be pointed to:

- [ ] A timestamped timeline in the incident doc from acknowledgement to all-clear.
- [ ] A named mitigation with the moment the burn rate responded.
- [ ] A verified recovery: SLI back within SLO, `/readyz` green, and a no-data-loss check recorded.
- [ ] A blameless postmortem filed within the SLA, with action items as tracked issues, each owned and dated.
- [ ] Any runbook or alert changes the incident exposed are merged, not just noted.

## Related

- SLO/alerting contract and severity/escalation expectations: [../operations/operability.md](../operations/operability.md)
- Signals you diagnose from (logs, metrics, traces, health): [../operations/observability.md](../operations/observability.md)
- The per-service runbook this checklist drives: [../templates/runbook.md](../templates/runbook.md)
- Rollback and deploy-safety mechanics: [../operations/deployment.md](../operations/deployment.md)
- Mechanisms that stop a cause becoming an incident: [../operations/resilience.md](../operations/resilience.md)
