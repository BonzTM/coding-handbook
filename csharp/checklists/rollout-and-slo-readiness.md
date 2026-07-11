# Rollout And SLO Readiness Checklist

Gate for shipping a release safely: [release.md](release.md) produces a deployable artifact; this checklist gates putting it in front of traffic without breaching the service's SLOs. Walk it per rollout, not per build.

## Pre-Rollout

- [ ] [release.md](release.md) checklist is green for this exact artifact; the `InformationalVersion` stamped in the assembly matches the image label and the canonical `v1.2.3` tag.
- [ ] Schema changes are expand/contract and deploy-safe: the new code runs against the old schema and the old code runs against the new schema, so a roll-forward and a roll-back are both safe — see [../recipes/add-migration.md](../recipes/add-migration.md).
- [ ] The migration step runs explicitly (the `--migrate` flag or init job) before instances of the new version take traffic; the service never auto-migrates on normal startup — see [../services/database.md](../services/database.md).
- [ ] Contracts (HTTP, gRPC, events) are backward-compatible, or the break is documented with a migration window and consumers are notified.
- [ ] New behavior ships behind a feature flag that defaults **off**; the rollout enables it separately from deploying the code, so code and behavior roll back independently.
- [ ] A rollback plan is written and concrete: the prior known-good version/tag, the exact command, and what (if anything) must be reverted in data or flags.
- [ ] The artifact's readiness and graceful-shutdown behavior were smoke-tested (`/readyz` gates traffic, `SIGTERM` drains in-flight work within `HostOptions.ShutdownTimeout`) — see [../operations/deployment.md](../operations/deployment.md).

## SLO & Observability Readiness

- [ ] Any new user-visible behavior has an SLI defined as a ratio over valid events and an SLO stated as `target over window` — see [../operations/operability.md](../operations/operability.md).
- [ ] Dashboards exist for the SLIs that move during this rollout, are versioned in the repo, and are open before traffic shifts.
- [ ] Symptom-based, burn-rate alerts cover the new behavior (fast burn pages, slow burn tickets); no new page fires on a cause or a single spike.
- [ ] The error budget for the affected SLO has room to absorb this rollout's expected risk; if it is exhausted, the SLO owner has explicitly approved spending against a deficit.
- [ ] Metric labels added for this change are low-cardinality (no per-user, per-tenant, or per-request-id labels on SLI series).

## Rollout

- [ ] Rollout is progressive: canary / single-instance-first or a percentage ramp, never all instances at once.
- [ ] Explicit ABORT criteria are written **before** starting and tied to SLO burn (e.g. "abort if availability burn rate exceeds X over Y minutes on the canary"), not to a vibe check.
- [ ] Readiness gates traffic at every step: a new instance receives traffic only after `/readyz` passes, and a failing instance is pulled from rotation — see [../operations/deployment.md](../operations/deployment.md).
- [ ] The shutdown-grace ordering holds: `terminationGracePeriodSeconds` > `HostOptions.ShutdownTimeout` > the longest request or message-handling time. The platform's grace exceeds the app's drain window, and the drain window exceeds the longest in-flight work, so draining instances finish and exit instead of being `SIGKILL`ed mid-request — see [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md) and [../operations/deployment.md](../operations/deployment.md). Never invert this ordering.
- [ ] During the canary stage, old and new versions run concurrently and both are healthy — the proof that the expand/contract migration and contract compatibility actually hold in production.

## Post-Rollout

- [ ] SLOs and the dashboards from above are watched through a defined bake period before declaring the rollout done; do not start the bake clock until traffic is fully shifted.
- [ ] Logs and metrics are confirmed healthy at full traffic: no new error class, no latency regression, no error-log spike, no unexpected restart loop.
- [ ] The rollout is explicitly closed (flag fully enabled, canary promoted, change logged) **or** rolled back via the written plan; it is never left half-shipped.
- [ ] If rolled back, the cause is captured for the postmortem and the SLO owner is told what budget was spent.

## Verification

- [ ] Canary / first stage is healthy on the SLIs before any further traffic is shifted — confirmed on the dashboard, not assumed.
- [ ] ABORT criteria are written and the rollback was rehearsed (dry-run the prior-version deploy and any flag/data revert), so the path is proven before it is needed.
- [ ] SLOs are steady (no abnormal burn) across the full bake period at 100% traffic before the rollout is declared complete.

```powershell
# The artifact gate must already be green before this checklist begins.
pwsh ./verify.ps1

# Confirm the running canary reports the expected version
# (stamped InformationalVersion / image label matches the v1.2.3 tag).
kubectl get pods -l app=<app> -o jsonpath='{.items[*].spec.containers[*].image}'
```
