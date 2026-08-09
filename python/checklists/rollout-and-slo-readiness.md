# Rollout And SLO Readiness Checklist

Gate for putting a verified artifact in front of traffic without breaching its service objectives.

## Pre-Rollout

- [ ] [Release checklist](release.md) is green for this exact wheel/image version and immutable digest.
- [ ] Alembic changes are expand/contract: old code accepts new schema and new code accepts the rollout schema; migration job and recovery are rehearsed.
- [ ] HTTP/gRPC/event/library contracts remain compatible or have a communicated deprecation/migration window.
- [ ] Risky behavior is independently controllable by a typed, owned, expiring feature flag where product semantics permit it.
- [ ] Rollback names the prior version/digest, exact command, config/flag actions, data constraints, and owner.
- [ ] Exact artifact passed readiness and SIGTERM drain smoke proof under the platform grace.

## SLO & Observability Readiness

- [ ] Changed user journeys have good/valid ratio SLIs and `target over rolling window` SLOs with owner and error-budget action.
- [ ] Versioned dashboards show SLI burn, traffic, latency, error classes, loop lag, task/queue/pool saturation, dependencies, memory, and deploy markers as relevant.
- [ ] Multi-window symptom burn alerts cover new behavior; fast burn pages and slow burn tickets.
- [ ] Error budget has room, or its owner explicitly approves the risk and recovery action.
- [ ] New metric labels have a finite inventory and contain no request/user/tenant/message/raw-path/PII values.
- [ ] Runbook covers deploy, rollback, scale, drain, migrations, dependency failure, and the new alerts.

## Rollout

- [ ] Shift progressively—canary or single replica, then bounded percentage/instance stages; never all replicas at once.
- [ ] Abort criteria are written before traffic moves and use SLO burn, error/latency, saturation, or correctness thresholds with exact windows.
- [ ] `/readyz` gates each stage and turns false before termination; `/livez` remains local.
- [ ] App/Uvicorn graceful timeout plus propagation/preStop/headroom remains below platform termination grace.
- [ ] Old and new versions run concurrently long enough to prove contract/schema compatibility.
- [ ] Migration, HTTPX/SQLAlchemy pool, semaphore/queue, and replica scaling remain within downstream capacity.

## Post-Rollout

- [ ] Observe the full-traffic artifact for a defined bake period; SLO burn, error classes, loop lag, pool wait, queues, memory, and restarts remain normal.
- [ ] Validate one critical path and required background/event behavior against production-safe evidence.
- [ ] Close explicitly: promote canary, settle/remove rollout flags according to plan, record version, and update change record.
- [ ] On rollback, verify user recovery and data consistency, record budget spent, and file the forward fix/postmortem trigger.

## Verification

```bash
make verify
uv run pytest -m integration
# Run the project's immutable-image inspect, canary smoke, and rollback dry-run commands.
```

- [ ] First stage is healthy on the named SLI/dashboard before further traffic shifts.
- [ ] Abort thresholds and prior-version rollback were rehearsed before rollout.
- [ ] SLO burn remains within policy for the complete bake period at full traffic.
