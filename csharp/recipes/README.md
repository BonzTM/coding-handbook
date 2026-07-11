# Recipes

Step-by-step implementation guides for the common C#/.NET changes this handbook governs. Each recipe is a fixed-shape contract — **Files To Touch / Steps / Invariants / Proof** — and links the topical doc that owns the rules it applies. Use a recipe when you know what kind of change you are making and want the exact file set and proof steps without rediscovering them.

For routing a change to its recipe and related obligations, start at the Change Routing table in [../AGENTS.md](../AGENTS.md). For the handbook overview, see [../README.md](../README.md).

## HTTP And gRPC Transport

- [add-http-endpoint.md](add-http-endpoint.md) - add or change one minimal-API endpoint, with request/response DTOs, ProblemDetails mapping, and route-group wiring. Governed by [../services/http-services.md](../services/http-services.md).
- [add-http-middleware.md](add-http-middleware.md) - add cross-cutting middleware or an endpoint filter without leaking transport concerns into Core. Governed by [../services/http-services.md](../services/http-services.md).
- [add-grpc-method.md](add-grpc-method.md) - add a gRPC method or proto change, with Grpc.Tools generation, interceptors, and error mapping. Governed by [../services/grpc-services.md](../services/grpc-services.md).
- [add-idempotent-write.md](add-idempotent-write.md) - make a non-GET endpoint safe to retry via an `Idempotency-Key` header, with a key store, atomic commit, and byte-identical replay. Governed by [../services/http-services.md](../services/http-services.md) and [../services/database.md](../services/database.md).

## Data And Migrations

- [add-database-feature.md](add-database-feature.md) - add or change queries, schema, or transaction behavior at the storage boundary. Governed by [../services/database.md](../services/database.md).
- [add-migration.md](add-migration.md) - author and apply an EF Core schema migration, kept backward-safe. Governed by [../services/database.md](../services/database.md).

## Eventing

- [add-event-publisher.md](add-event-publisher.md) - publish a new event or message with a stable payload contract and outbox semantics. Governed by [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).
- [add-event-consumer.md](add-event-consumer.md) - consume a message with inbox idempotency, retry, and DLQ handling. Governed by [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).

## CLI

- [add-cli-command.md](add-cli-command.md) - add a System.CommandLine command or option with help output and config wiring. Governed by [../foundations/configuration.md](../foundations/configuration.md) and [../decisions/framework-selection.md](../decisions/framework-selection.md).

## Workers And Scheduled Jobs

- [add-background-worker.md](add-background-worker.md) - add a long-running `BackgroundService` with token-driven shutdown and no orphaned tasks. Governed by [../foundations/cancellation-and-async.md](../foundations/cancellation-and-async.md).
- [add-scheduled-job.md](add-scheduled-job.md) - add a periodic job using `PeriodicTimer`, an injected `TimeProvider`, and deterministic tests. Governed by [../foundations/time.md](../foundations/time.md).

## External Clients

- [add-external-client.md](add-external-client.md) - add an outbound HTTP or gRPC client with timeouts, resilience handlers, and a testable seam. Governed by [../operations/resilience.md](../operations/resilience.md).

## Config And Metrics

- [add-config-key.md](add-config-key.md) - add a configuration key with a validated options class, fail-fast startup, and synced documentation. Governed by [../foundations/configuration.md](../foundations/configuration.md).
- [add-metric.md](add-metric.md) - add a `Meter` instrument with low-cardinality tags and a clear name. Governed by [../operations/observability.md](../operations/observability.md).

## Contracts And Deprecation

- [deprecate-and-remove-contract.md](deprecate-and-remove-contract.md) - deprecate then remove a wire contract, endpoint, or public API surface without breaking clients. Governed by [../foundations/contracts-and-compatibility.md](../foundations/contracts-and-compatibility.md).

## Dependencies

- [bump-dependency.md](bump-dependency.md) - upgrade a NuGet package with an understood lock-file diff and proof it still earns its cost. Governed by [../decisions/framework-selection.md](../decisions/framework-selection.md).

## Library Release

- [release-library-version.md](release-library-version.md) - cut a tagged NuGet release with a canonical `v1.2.3` tag, a reviewed public-API surface, and a changelog. Governed by [../operations/ci-and-release.md](../operations/ci-and-release.md).

## Where To Go Next

- Routing a change to the right files: [../AGENTS.md](../AGENTS.md) (## Change Routing)
- Handbook overview: [../README.md](../README.md)
- Checklists for lifecycle gates: [../checklists/README.md](../checklists/README.md)
