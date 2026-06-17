# Recipes

Step-by-step implementation guides for the common Go changes this handbook governs. Each recipe is a fixed-shape contract — **Files To Touch / Steps / Invariants / Proof** — and links the topical doc that owns the rules it applies. Use a recipe when you know what kind of change you are making and want the exact file set and proof steps without rediscovering them.

For routing a change to its recipe and related obligations, start at [../maintainer-map.md](../maintainer-map.md). For the handbook overview, see [../README.md](../README.md).

## HTTP And gRPC Transport

- [add-http-endpoint.md](add-http-endpoint.md) - add or change one HTTP route, with DTOs, error mapping, and route wiring. Governed by [../services/http-services.md](../services/http-services.md).
- [add-http-middleware.md](add-http-middleware.md) - add a cross-cutting HTTP middleware without leaking transport concerns into core. Governed by [../services/http-services.md](../services/http-services.md).
- [add-grpc-method.md](add-grpc-method.md) - add a gRPC method or proto change, with generation, interceptors, and error mapping. Governed by [../services/grpc-services.md](../services/grpc-services.md).
- [add-idempotent-write.md](add-idempotent-write.md) - make a non-GET endpoint safe to retry via an `Idempotency-Key` header, with a key store, atomic commit, and byte-identical replay. Governed by [../services/http-services.md](../services/http-services.md) and [../services/database.md](../services/database.md).

## Data And Migrations

- [add-database-feature.md](add-database-feature.md) - add or change queries, schema, or transaction behavior at the storage boundary. Governed by [../services/database.md](../services/database.md).
- [add-migration.md](add-migration.md) - author and apply a schema migration with `goose`, kept backward-safe. Governed by [../services/database.md](../services/database.md).

## Eventing

- [add-event-publisher.md](add-event-publisher.md) - publish a new event or message with a stable payload contract and outbox semantics. Governed by [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).
- [add-event-consumer.md](add-event-consumer.md) - consume a message with idempotency, retry, and DLQ handling. Governed by [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).

## CLI

- [add-cli-command.md](add-cli-command.md) - add a CLI command or flag with help output and config wiring. Governed by [../foundations/configuration.md](../foundations/configuration.md) and [../decisions/framework-selection.md](../decisions/framework-selection.md).

## Workers And Scheduled Jobs

- [add-background-worker.md](add-background-worker.md) - add a long-running background worker with context-driven shutdown and leak-free goroutines. Governed by [../foundations/context-and-concurrency.md](../foundations/context-and-concurrency.md).
- [add-scheduled-job.md](add-scheduled-job.md) - add a periodic job using an injectable clock and deterministic tests. Governed by [../foundations/time.md](../foundations/time.md).

## External Clients

- [add-external-client.md](add-external-client.md) - add an outbound HTTP or gRPC client with timeouts, retries, and a testable seam. Governed by [../operations/resilience.md](../operations/resilience.md).

## Config And Metrics

- [add-config-key.md](add-config-key.md) - add a configuration key with fail-fast validation and synced documentation. Governed by [../foundations/configuration.md](../foundations/configuration.md).
- [add-metric.md](add-metric.md) - add a metric with low-cardinality labels and a clear name. Governed by [../operations/observability.md](../operations/observability.md).

## Contracts And Deprecation

- [deprecate-and-remove-contract.md](deprecate-and-remove-contract.md) - deprecate then remove a wire contract, endpoint, or public surface without breaking clients. Governed by [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md).

## Dependencies

- [bump-dependency.md](bump-dependency.md) - upgrade a dependency with an understood diff and proof it still earns its cost. Governed by [../decisions/framework-selection.md](../decisions/framework-selection.md).

## Library Release

- [release-library-version.md](release-library-version.md) - cut a tagged library release with a canonical `v1.2.3` tag and changelog. Governed by [../operations/ci-and-release.md](../operations/ci-and-release.md).

## Where To Go Next

- Routing a change to the right files: [../maintainer-map.md](../maintainer-map.md)
- Handbook overview: [../README.md](../README.md)
- Checklists for lifecycle gates: [../checklists/README.md](../checklists/README.md)
