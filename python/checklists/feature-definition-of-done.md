# Feature Definition of Done Checklist

Author-facing gate for declaring a Python feature complete. Every checked box points to reviewable evidence.

## Behavior & Tests

- [ ] Code is at the correct boundary: adapter parses/maps, core decides, repository/client performs I/O, composition wires.
- [ ] Pydantic validates each trust boundary and authorization runs before effects per [security](../operations/security.md).
- [ ] Unit tests prove core decisions; real integration tests prove each changed database, HTTP, broker, filesystem, or process boundary.
- [ ] Negative paths cover invalid input, denial, empty/boundary values, partial failure, timeout, cancellation, and not-found behavior.
- [ ] Async tests use events/queues/fake clock, own every task, and contain no sleep-based synchronization.
- [ ] Each behavior change has a test that fails without it and passes with it.

## Contracts & Data

- [ ] Pydantic DTOs/message models, protobuf, CLI, or public Python types expose only intentional fields and map once into domain values.
- [ ] Change is additive/backward compatible or follows a documented deprecation/removal plan.
- [ ] Alembic changes are reviewed expand/contract revisions; old/new code compatibility and migration recovery are proven on real PostgreSQL.
- [ ] Event changes preserve versioning, idempotency, ordering, replay, settlement, and DLQ behavior.
- [ ] New data fields have classification, retention, telemetry, deletion, and export decisions.

## Config & Docs

- [ ] New settings are typed/validated in pydantic-settings, added to `.env.example` and README, and fail fast when required values are absent.
- [ ] No secret appears in source, lockfile, examples, logs, errors, reprs, telemetry, build args, or image layers.
- [ ] Public APIs and `Protocol` contracts have concise docstrings; `py.typed` remains present for libraries.
- [ ] Operator-visible changes have a Keep a Changelog entry and current runbook/procedure.

## Observability & Resilience

- [ ] Logs, low-cardinality metrics, and traces answer whether the feature works and why it failed without leaking data.
- [ ] New required dependencies affect readiness; liveness remains local.
- [ ] Every external call has total/attempt timeouts, bounded idempotent-only retries, resource closure, and deterministic failure proof.
- [ ] Concurrency, queue, pool, body, fan-out, and drain limits are explicit; overload sheds predictably.
- [ ] SLI/dashboard/alert/runbook are updated when the user-visible SLO surface changes.

## Proof & Release

- [ ] `make verify` passed: lock-check, frozen sync, format-check, lint, imports, types, test, audit.
- [ ] `uv run pytest -m integration` exercised every changed real boundary with no unexpected skip.
- [ ] Branch coverage did not regress on core, parsing, auth, mapping, retry, and cancellation decisions.
- [ ] Package/container artifact installs or starts and proves the changed path, probes, and shutdown.
- [ ] Rollout/rollback, compatibility, migration, config, and contract notes are actionable.

## Verification

```bash
make verify
uv run pytest -m integration
uv run pytest --cov=src/<app> --cov-branch --cov-report=term-missing
```

- [ ] Every box above links to a test name, diff, command output, schema, metric, changelog line, or runbook section—not assurance.
- [ ] The proof was run from a clean checkout/artifact with the committed `uv.lock`.
