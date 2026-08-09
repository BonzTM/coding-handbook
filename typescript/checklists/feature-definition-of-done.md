# Checklist: Feature Definition Of Done

Use this before marking a feature complete.

## Contract And Behavior

- [ ] Acceptance criteria are proven by observable tests, not implementation assertions.
- [ ] Request, response, event, persistence, UI, and error contracts are synchronized.
- [ ] Compatibility and mixed-version behavior are tested where deploys overlap.
- [ ] Inputs are parsed, normalized, bounded, and mapped into domain values.
- [ ] Expected failures have typed outcomes and unknown failures are safely mapped.

## Reliability And Security

- [ ] Authorization, cross-tenant, injection, SSRF, XSS, replay, and size cases are covered as applicable.
- [ ] I/O has timeout, cancellation, retry limit, concurrency limit, and cleanup.
- [ ] Retried or duplicate durable work is idempotent.
- [ ] Secrets and sensitive data are absent from logs, metrics, traces, errors, snapshots, and artifacts.
- [ ] Database migrations work on empty and prior schema and have rollout/recovery plans.
- [ ] UI keyboard, focus, accessible name, loading, empty, error, and success behavior is proven.

## Operations And Maintenance

- [ ] Logs, metrics, traces, health, dashboards, and alerts answer the operational question.
- [ ] Configuration, `.env.example`, deployment values, changelog, and runbook agree.
- [ ] Capacity, queue, pool, batch, response, and memory bounds are explicit.
- [ ] Rollout stages, observation, abort criteria, artifact identity, and rollback are documented.
- [ ] Owners exist for code, dependency, data, alert, runbook, and deferred cleanup.

## Proof

- [ ] Targeted tests and required real-boundary integration tests are green.
- [ ] Production ESM artifact, frontend bundle, image, or packed library smoke test passes.
- [ ] `npm run verify` is green from a clean Node 24 install.
- [ ] No acceptance criterion depends on undocumented manual knowledge.
