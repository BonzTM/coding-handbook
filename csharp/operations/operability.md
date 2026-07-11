# Operability

What "good" looks like operationally: turning the signals a service already emits into targets, alerts, and a runbook a stranger can act from at 3 a.m.

[observability.md](observability.md) makes a service emit logs, metrics, traces, and health endpoints. It never says what those signals should be *worth*. This doc closes that gap: it defines the Service Level Indicators (SLIs) you derive from those signals, the Service Level Objectives (SLOs) you hold them to, the only conditions worth waking a human for, and the runbook that makes a page actionable.

## Default Approach

### SLIs Derived From The Signals You Already Emit

An SLI is a ratio of good events to valid events, computed entirely from telemetry [observability.md](observability.md) already mandates. Do not invent new instrumentation for SLIs; if a signal is missing, that is an observability gap, not an SLO problem.

| Service shape | SLI | Built from |
|---|---|---|
| Request/response (HTTP, gRPC) | Availability: non-5xx (or non-`Internal`/`Unavailable`) responses ÷ valid requests | request counter labeled by status class |
| Request/response | Latency: requests served under a threshold ÷ valid requests | latency histogram |
| Request/response | Error rate: the inverse of availability, expressed against the same denominator | request counter |
| Worker / consumer | Freshness: work completed within its deadline ÷ work due | processing-latency histogram, backlog-age gauge |
| Worker / consumer | Lag: backlog age (queue depth in time, not count) below a threshold | backlog-age gauge from [observability.md](observability.md) |

Rules:

- Define SLIs as a **ratio against valid events**, never as a raw count or an average. Averages hide the tail; the tail is what users feel.
- Latency SLIs are threshold-based ("fraction served under 300 ms"), driven by histogram buckets, so pick buckets that straddle your target before you set the SLO.
- Exclude requests the service is right to reject (client `4xx`, malformed input) from the *good*-events numerator but keep them honest in the denominator only where they reflect real user experience; document the exclusion in the runbook.
- One SLI measures one user-visible promise. If you cannot state the promise in a sentence, you do not yet have an SLI.

### Setting SLOs With An Error Budget

An SLO is a target for an SLI over a rolling window. The gap between the SLO and 100% is the **error budget** — the amount of failure you have explicitly agreed to tolerate.

- Express every SLO as `target over window`, e.g. "99.9% of requests succeed over 30 rolling days" or "99% of jobs complete within their deadline over 7 rolling days". Window and target are both load-bearing; never state one without the other.
- The error budget is `(1 - target) × valid events`. At 99.9% over 30 days that is roughly 43 minutes of full outage, or its equivalent spread as a low error rate. Spend it deliberately on releases and risk; do not aim for 100% — 100% is the wrong target because it forbids change.
- **Every SLO has a named owner** (a person or a rotation, not "the team"). The owner decides what happens when the budget is exhausted: typically freeze risky change and spend the next cycle on reliability until the budget recovers.
- Set SLOs from observed baseline plus user need, not aspiration. Start slightly looser than current performance, tighten as you earn it. An SLO no one can meet is noise; an SLO the service never threatens is theater.
- SLOs and their error-budget policy live in the project runbook ([templates/runbook.md](../templates/runbook.md)) and are surfaced at handoff ([../checklists/handoff.md](../checklists/handoff.md)).

### Alerting: Page On Symptoms, Not Causes

Alert on the user-visible symptom measured against the SLO, never on the cause and never on every spike.

- **Symptom, not cause.** Alert "error rate is burning the budget", not "CPU is high" or "pod restarted" or "queue has 5000 items". Causes are diagnostics for the runbook; symptoms are what users feel. A cause-based alert fires when nothing is wrong and stays silent when something new breaks.
- **Burn rate, not single spikes.** Drive the page off the rate at which the rolling window is consuming the error budget. A fast burn (budget gone in hours) pages now; a slow burn (budget will run out this window) opens a ticket. Multi-window, multi-burn-rate conditions catch both real outages and slow degradations without firing on momentary blips. This is the alerting expression of the error budget defined above.
- **Page only when both true: user-impacting AND actionable.** If users are not affected, it is not a page. If the responder can do nothing about it right now, it is not a page. Everything that fails one of those tests is a **ticket**, not a page.
- **Tiers.** Page = burning the budget fast, on-call must act. Ticket = slow burn, capacity trend, expiring cert, flaky dependency — triaged in business hours. Log/dashboard only = everything else.
- Resilience mechanisms (retries, timeouts, circuit breakers, fallbacks) are what *prevent* a cause from becoming a symptom; see [resilience.md](resilience.md). Operability assumes they exist and alerts on the symptom that escapes them.

### Every Service Ships A Runbook, Kept Current

A page is useless without a runbook. Every service ships one from [templates/runbook.md](../templates/runbook.md) at `docs/runbook.md`, and it is a release artifact, not a nice-to-have.

- The runbook maps each alert to a diagnosis path and remediation steps, and documents the key operations: deploy, rollback, scale, drain.
- It is kept current as part of the change that alters behavior. A deploy that changes a remediation step and not the runbook is incomplete. Handoff explicitly verifies the runbook matches reality today ([../checklists/handoff.md](../checklists/handoff.md)).
- The bar: a new on-call engineer can resolve a fired alert from the runbook alone, without paging the owner. If they cannot, the runbook has a defect.

### Live Diagnosis: dotnet-counters, dotnet-trace, dotnet-dump

Runbook diagnosis paths name tools, and for a .NET service the toolbox is the first-party diagnostics CLI suite. These are for *diagnosis after a symptom alert fires* — they are not monitoring, and no alert is ever built on them.

- **`dotnet-counters`** — first look. Live view of runtime and Kestrel metrics (GC heap size and pause time, thread-pool queue length and starvation, exception rate, request rate) attached to the running process. If the symptom is latency, thread-pool starvation shows here before anything else does.
- **`dotnet-trace`** — CPU and allocation sampling without a debugger or restart, the profiling counterpart of Go's pprof. Capture a trace during the incident, analyze offline in speedscope or PerfView.
- **`dotnet-dump`** (and `dotnet-gcdump` for heap-only) — full process capture for leaks, deadlocks, and stuck async chains; `dotnet-dump analyze` walks heaps and stacks with SOS commands after the fact.
- These tools attach through the runtime's diagnostic IPC socket in the target's `/tmp` and need process visibility. The chiseled runtime image ([deployment.md](deployment.md)) has no shell and no SDK, so **plan the attach path before the incident**: a Kubernetes ephemeral debug container sharing the pod's process namespace and diagnostic socket, or a `dotnet monitor` sidecar exposing diagnostics over an authenticated endpoint. Which one is a per-repo choice; write it into the runbook, because 3 a.m. is not when you design it.

### On-Call, Escalation, And Dashboards Are Repo-Specific

The handbook fixes the contract; the project fills in the specifics.

- Every service names an on-call rotation and an escalation path (who to page when the first responder is stuck, in what order). These live in the runbook and are transferred at handoff ([../onboarding-and-handoff.md](../onboarding-and-handoff.md), [../checklists/handoff.md](../checklists/handoff.md)).
- Dashboards are **code in the repo**, versioned and reviewed like any other artifact, so they survive handoff and cannot silently drift from the SLOs they display.
- The specific monitoring provider, alerting backend, paging tool, and dashboard format are vendor choices. Route them to [../decisions/framework-selection.md](../decisions/framework-selection.md) and record the pick in an ADR; do not hardcode a vendor into the handbook.

## Common Mistakes And Forbidden Patterns

- Alerting on causes or on every spike (high CPU, single pod restart, one slow request) — produces pager fatigue, trains responders to ignore pages, and goes silent on the failure mode you did not predict.
- An SLO with no error budget or no named owner — there is nothing to spend, no one to decide on exhaustion, and the SLO becomes decoration.
- Aiming for 100% — forbids change, guarantees the budget is always "violated", and makes every release a fight.
- No runbook, or a stale one — every incident is improvised from scratch and the new on-call must page the previous owner, which is the exact failure [../onboarding-and-handoff.md](../onboarding-and-handoff.md) exists to prevent.
- Per-entity SLO labels (per-user, per-tenant, per-request-id) on SLI metrics — a cardinality blowup that breaks the metrics backend; SLIs are aggregate ratios, not per-entity counters. Same rule as [observability.md](observability.md).
- Paging on conditions a human cannot act on right now — move them to tickets; a page must be both user-impacting and actionable.
- SLIs built from averages instead of ratios over valid events — averages hide the tail users actually experience.
- Treating the diagnostics CLI suite as monitoring — building alerts on `dotnet-counters` output, or discovering during the incident that nothing can attach to the chiseled container.

## Verification And Proof

- Each service has documented SLOs (`target over window`) with a named owner and an error-budget policy, recorded in `docs/runbook.md`.
- Every page corresponds to a symptom alert tied to an SLO burn rate; no page fires on a cause or a single spike. Inspect the alert definitions and confirm each maps to an SLI, not a resource metric.
- Each SLI metric is an aggregate ratio with low-cardinality labels only; no per-entity labels appear on SLO series.
- `docs/runbook.md` exists, lists every alert with diagnosis and remediation, and matches current behavior.
- The runbook's diagnosis paths name a working attach path for `dotnet-counters`/`dotnet-trace`/`dotnet-dump` against the production container — exercise it once outside an incident.
- A new on-call engineer can resolve a representative fired alert using the runbook alone, without escalating to the owner — run this as a drill before sign-off.
- On-call rotation and escalation path are documented and confirmed reachable ([../checklists/handoff.md](../checklists/handoff.md)).

## Related

- Signals these targets are built from: [observability.md](observability.md)
- What stops a cause becoming a symptom: [resilience.md](resilience.md)
- The container the tools must attach to: [deployment.md](deployment.md)
- The runbook template: [../templates/runbook.md](../templates/runbook.md)
- Taking over a service's operations: [../onboarding-and-handoff.md](../onboarding-and-handoff.md)
- Confirming operability at transfer: [../checklists/handoff.md](../checklists/handoff.md)
