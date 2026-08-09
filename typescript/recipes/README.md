# Recipes

Step-by-step contracts for common TypeScript changes. Every recipe uses **Files To Touch / Steps / Invariants To Preserve / Proof**. Start with [../AGENTS.md](../AGENTS.md) for routing.

## Backend And Data

| Recipe | Purpose | Governing docs |
|---|---|---|
| [add-http-endpoint.md](add-http-endpoint.md) | Add one Fastify route and core use case. | [HTTP services](../services/http-services.md) |
| [add-http-middleware.md](add-http-middleware.md) | Add a scoped Fastify hook or plugin. | [HTTP services](../services/http-services.md) |
| [add-database-feature.md](add-database-feature.md) | Add PostgreSQL queries and mapping. | [Database](../services/database.md) |
| [add-migration.md](add-migration.md) | Add a deploy-safe migration. | [Database](../services/database.md) |
| [add-external-client.md](add-external-client.md) | Add a bounded outbound HTTP client. | [Resilience](../operations/resilience.md) |
| [add-idempotent-write.md](add-idempotent-write.md) | Make a write safe to retry. | [HTTP services](../services/http-services.md) |

## Frontend

| Recipe | Purpose | Governing docs |
|---|---|---|
| [add-react-component.md](add-react-component.md) | Add an accessible component. | [React applications](../services/react-applications.md) |
| [add-react-hook.md](add-react-hook.md) | Add owned stateful behavior. | [React applications](../services/react-applications.md) |
| [add-frontend-route.md](add-frontend-route.md) | Add a lazy route boundary. | [React applications](../services/react-applications.md) |
| [add-form.md](add-form.md) | Add a validated accessible form. | [React applications](../services/react-applications.md) |

## Runtime And Integration

| Recipe | Purpose | Governing docs |
|---|---|---|
| [add-config-key.md](add-config-key.md) | Add typed startup configuration. | [Configuration](../foundations/configuration.md) |
| [add-background-worker.md](add-background-worker.md) | Add bounded process-owned work. | [Async and cancellation](../foundations/async-and-cancellation.md) |
| [add-scheduled-job.md](add-scheduled-job.md) | Add deterministic scheduled work. | [Time](../foundations/time.md) |
| [add-event-publisher.md](add-event-publisher.md) | Add an outbox-backed producer. | [Eventing](../services/eventing-and-messaging.md) |
| [add-event-consumer.md](add-event-consumer.md) | Add an idempotent consumer. | [Eventing](../services/eventing-and-messaging.md) |
| [add-metric.md](add-metric.md) | Add a bounded-cardinality metric. | [Observability](../operations/observability.md) |
| [add-cli-command.md](add-cli-command.md) | Add a CLI command or flag. | [Configuration](../foundations/configuration.md) |

## Contracts And Lifecycle

| Recipe | Purpose | Governing docs |
|---|---|---|
| [bump-dependency.md](bump-dependency.md) | Upgrade one dependency deliberately. | [Framework selection](../decisions/framework-selection.md) |
| [deprecate-and-remove-contract.md](deprecate-and-remove-contract.md) | Retire a public contract safely. | [Contracts](../foundations/contracts-and-compatibility.md) |
| [release-library-version.md](release-library-version.md) | Publish a packed library release. | [CI and release](../operations/ci-and-release.md) |

Use [../checklists/README.md](../checklists/README.md) for lifecycle gates.
