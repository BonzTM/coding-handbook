# PR Review Checklist

Review gate for Python changes affecting behavior, boundaries, contracts, or operations.

## Boundaries And Placement

- [ ] `core` remains framework-free and imports no FastAPI, Pydantic DTO, SQLAlchemy, HTTPX, broker, or telemetry SDK type.
- [ ] Routers/workers/CLI adapters parse and map only; repositories/clients own I/O; composition remains thin.
- [ ] New seams are consumer-owned typed `Protocol`s, not global service locators, ambient context, or `utils.py` dumping grounds.
- [ ] New dependency/framework choices match [framework selection](../decisions/framework-selection.md) or carry an ADR and lock review.

## Correctness

- [ ] External input is bounded, parsed, normalized, authorized, and converted to typed/domain values.
- [ ] Exceptions are caught only where mapped/retried/compensated; cancellation is re-raised and failures are logged once.
- [ ] Every task/resource has an owner, lifetime, timeout, concurrency bound, cleanup, and shutdown path.
- [ ] Blocking filesystem/process/library work stays off async paths; no `time.sleep()` in async code or tests.
- [ ] SQL is parameterized, sessions are per unit of work, and network calls do not occur inside transactions.
- [ ] No secret/PII enters source, logs, errors, reprs, metrics, traces, audit payloads, examples, or artifacts.

## Observability And Operations

- [ ] New runtime behavior has safe structured logs, bounded metric labels, useful trace boundaries, and correct readiness semantics.
- [ ] Timeouts, retries, idempotency, backpressure, and shedding follow [resilience](../operations/resilience.md).
- [ ] Contracts, migrations, config, changelog, runbook, dashboards/alerts, and templates remain synchronized.
- [ ] Deployment preserves non-root runtime, pinned images, one-process default, bounded grace, and explicit migrations.

## Proof

- [ ] Targeted tests prove the behavior and its negative/error/cancellation paths.
- [ ] `make verify` is green with no unjustified `noqa`, `type: ignore`, marker skip, warning, or audit exception.
- [ ] `uv run pytest -m integration` proves changed SQL/external/broker semantics against real boundaries.
- [ ] Coverage did not regress on owned decisions; golden/OpenAPI/protobuf/event diffs are intentional and reviewed.
- [ ] Artifact/install/startup/shutdown proof matches the change blast radius.
- [ ] A feature change clears [feature definition of done](feature-definition-of-done.md).

## Verification

```bash
make verify
uv run pytest -m integration
git diff --check
```
