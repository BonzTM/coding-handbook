# Recipes

Step-by-step implementation guides for common Python changes. Each recipe is a fixed-shape contract — **Files To Touch / Steps / Invariants To Preserve / Proof**. Route changes through [AGENTS.md](../AGENTS.md).

## HTTP And gRPC Transport

- [add-http-endpoint.md](add-http-endpoint.md) — FastAPI route, DTO mapping, core call, and tests.
- [add-http-middleware.md](add-http-middleware.md) — ASGI middleware or dependency at the correct boundary.
- [add-grpc-method.md](add-grpc-method.md) — versioned proto, generated stubs, servicer, and proof.
- [add-idempotent-write.md](add-idempotent-write.md) — tenant-scoped write dedupe and byte-identical replay.

## Data And Migrations

- [add-database-feature.md](add-database-feature.md) — SQLAlchemy repository behavior with real PostgreSQL proof.
- [add-migration.md](add-migration.md) — reviewed Alembic revision and explicit apply step.

## Eventing

- [add-event-publisher.md](add-event-publisher.md) — stable event contract and transactional outbox.
- [add-event-consumer.md](add-event-consumer.md) — validation, inbox dedupe, settlement, and DLQ behavior.

## CLI

- [add-cli-command.md](add-cli-command.md) — argparse command and PEP 621 entry point.

## Workers And Scheduled Jobs

- [add-background-worker.md](add-background-worker.md) — TaskGroup-owned worker with bounded shutdown.
- [add-scheduled-job.md](add-scheduled-job.md) — interval loop with jitter, non-overlap, and deterministic tests.

## External Clients

- [add-external-client.md](add-external-client.md) — HTTPX adapter with deadlines, retries, and a core-owned port.

## Config And Metrics

- [add-config-key.md](add-config-key.md) — pydantic-settings field, example, deployment wiring, and tests.
- [add-metric.md](add-metric.md) — Prometheus instrument with reviewed cardinality.

## Contracts, Dependencies, And Releases

- [deprecate-and-remove-contract.md](deprecate-and-remove-contract.md) — observed multi-release retirement.
- [bump-dependency.md](bump-dependency.md) — uv lock upgrade, audit, and diff review.
- [release-library-version.md](release-library-version.md) — version, build, artifact inspection, tag, and publish.

## Where To Go Next

- Handbook overview: [../README.md](../README.md)
- Lifecycle gates: [../checklists/README.md](../checklists/README.md)
- Binding stack: [../decisions/framework-selection.md](../decisions/framework-selection.md)
