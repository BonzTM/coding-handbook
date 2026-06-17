# Feature Definition of Done Checklist

The author-facing bar a feature must clear before it is "done" — the scattered "not done until" rules from across this handbook, consolidated into one gate. A reviewer must be able to trace every box to evidence.

## Behavior & Tests

- [ ] Behavior is implemented at the right boundary: transport parses and validates, core decides, storage persists — no business logic leaking into handlers or migrations.
- [ ] Input is validated at every boundary the feature touches, and authorization runs before any core logic per [../operations/security.md](../operations/security.md).
- [ ] Tests exist at the right level per [../quality/testing.md](../quality/testing.md): unit tests for core logic, real integration tests at every DB and external boundary the feature crosses (no mocked database, no mocked queue round-trip).
- [ ] Negative and edge cases are covered: invalid input, authz denial, empty/boundary values, partial failure, cancellation, and the not-found path.
- [ ] Tests are deterministic: time comes from an injected clock and concurrency is awaited explicitly — no `time.Sleep` to "let it settle" per [../foundations/context-and-concurrency.md](../foundations/context-and-concurrency.md).
- [ ] Every behavior change has a test that fails without the change and passes with it.

## Contracts & Data

- [ ] Wire and DTO shapes are explicit: every field has an intentional `json` tag, `omitempty` is deliberate, and no struct is serialized by accident per [../foundations/serialization.md](../foundations/serialization.md).
- [ ] The change is backward compatible, or it carries a deprecation plan per [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md) and follows [../recipes/deprecate-and-remove-contract.md](../recipes/deprecate-and-remove-contract.md).
- [ ] Schema changes are deploy-safe: expand/contract, never a destructive in-place rewrite, and old and new code can run against the same schema during rollout per [../recipes/add-migration.md](../recipes/add-migration.md).
- [ ] Event and message contract changes preserve compatibility, idempotency, and replay/DLQ behavior.

## Config & Docs

- [ ] New config keys are loaded and validated in `internal/config`, fail-fast at startup, and added to `.env.example` and the README config table per [../recipes/add-config-key.md](../recipes/add-config-key.md) and [../foundations/configuration.md](../foundations/configuration.md).
- [ ] No secret is introduced in source, logs, panics, or build args; new secrets are sourced from the external manager and documented per [../operations/security.md](../operations/security.md) `### Secrets`.
- [ ] Exported symbols added or changed carry godoc that states the contract, not the implementation.

## Observability

- [ ] Logs and metrics are added where an operator would need them to answer "is it working" and "why did it fail," using `log/slog` and low-cardinality labels per [../operations/observability.md](../operations/observability.md).
- [ ] Readiness reflects any new dependency the feature requires to serve traffic; liveness stays distinct and does not gate on dependencies.
- [ ] If the feature changes an SLO surface, the relevant alert and runbook are updated per [../operations/operability.md](../operations/operability.md).

## Proof & Release

- [ ] `make verify` is green (tidy, fmt-check, lint, vet, test, race, vuln, build).
- [ ] Coverage did not regress and the feature's mandatory paths (domain core, error/status mapping, decode) are covered (`make cover`) per [../quality/testing.md](../quality/testing.md).
- [ ] Operator-visible changes (new env vars, migrations, ports, dependency or contract changes) have a release-notes or CHANGELOG entry per [release.md](release.md).
- [ ] Resilience behavior (timeout, retry, backoff) for new outbound calls matches the policy in [../operations/resilience.md](../operations/resilience.md).

## Verification

```bash
make verify
```

The feature is done only when:

- [ ] `make verify` passed from a clean tree.
- [ ] Every box above is checked.
- [ ] A reviewer can trace each checked box to concrete evidence — a test name, a diff hunk, a config table row, a metric, a changelog line — not to assurance.
