# exampleworker

A complete broker-neutral event worker that proves the TypeScript handbook's messaging, cancellation, resilience, health, and telemetry rules compile, test, audit, and build together.

## What It Is

- A Zod 4-parsed event envelope with stable UUID event ID, event type, schema version, RFC 3339 occurrence time, producer, optional trace context, and type-specific payload.
- An in-memory broker seam modeling at-least-once delivery and bounded topic capacity, plus a consumer with configurable concurrency and explicit intake and handler lifetimes.
- Consumer-scoped inbox keys that collapse concurrent and replayed duplicates before projection effects, with in-memory atomic implementations standing in for a database transaction.
- Explicit validation, permanent, transient, cancellation, and retry-exhaustion behavior; only transient failures receive bounded exponential backoff with full jitter.
- A DLQ store retaining safe replay metadata, attempt count, classification, reason, original body, and the injected-clock parking time.
- A transactional-outbox relay that publishes before marking sent. A crash or broker failure leaves the record pending; a crash after publish can duplicate the event, which the inbox tolerates by stable event ID.
- Ordered SIGINT/SIGTERM shutdown: readiness false, stop intake and relay polling, drain in-flight handlers within the configured deadline, final outbox flush, close probes, close broker, then flush logs.
- Pino JSON logs with construction-time redaction and minimal low-cardinality consume/publish counters exposed with `/livez`, `/readyz`, and `/metrics`.

The worker uses Fastify only for the small probe/metrics sidecar. This keeps the handbook's standard HTTP framework and lets probe behavior use `app.inject()` tests; no application route or HTTP business logic is present.

## Requirements

- Node.js 24.18.0 (pinned in `.nvmrc` and the container)
- npm and the committed `package-lock.json`
- No broker, database, collector, or network service for the default gate

## Setup And Run

```bash
npm ci
cp .env.example .env
npm run build
node --env-file=.env dist/main.js
```

Probe the running worker:

```bash
curl -s localhost:3001/livez
curl -s localhost:3001/readyz
curl -s localhost:3001/metrics
```

The runtime composition deliberately wires in-memory broker, inbox, DLQ, outbox, and projection adapters. A production deployment replaces those adapters with an approved broker and durable stores while preserving the consumer-owned ports and delivery contract.

## Configuration

| Key                        | Required | Default            | Meaning                                               |
| -------------------------- | -------: | ------------------ | ----------------------------------------------------- |
| `NODE_ENV`                 |       no | `development`      | `development`, `test`, or `production`                |
| `HOST`                     |       no | `127.0.0.1`        | probe listener address                                |
| `HEALTH_PORT`              |       no | `3001`             | probe listener port                                   |
| `LOG_LEVEL`                |       no | `info`             | Pino level                                            |
| `TOPIC`                    |       no | `widget.events`    | logical subscription and publish destination          |
| `CONSUMER_NAME`            |       no | `widget-projector` | identity used with event ID for inbox dedupe          |
| `CONSUMER_CONCURRENCY`     |       no | `4`                | concurrent handlers, 1 through 32                     |
| `CONSUMER_MAX_ATTEMPTS`    |       no | `5`                | total transient attempts, 1 through 20                |
| `CONSUMER_BASE_BACKOFF_MS` |       no | `100`              | first full-jitter backoff ceiling                     |
| `CONSUMER_MAX_BACKOFF_MS`  |       no | `30000`            | maximum full-jitter backoff ceiling                   |
| `OUTBOX_POLL_INTERVAL_MS`  |       no | `1000`             | relay scan interval                                   |
| `OUTBOX_BATCH_SIZE`        |       no | `100`              | maximum records published per scan                    |
| `SHUTDOWN_TIMEOUT_MS`      |       no | `15000`            | in-flight drain, final flush, and cleanup time budget |

Only `src/config/index.ts` reads environment state. All numeric work and memory bounds are parsed before the worker listens or consumes.

## Package Map

| Path             | Responsibility                                                            |
| ---------------- | ------------------------------------------------------------------------- |
| `src/main.ts`    | fail-fast startup, process signals, and terminal failure ownership        |
| `src/index.ts`   | composition, task supervision, readiness, and ordered bounded shutdown    |
| `src/config/`    | strict environment selection, defaults, parsing, and cross-field bounds   |
| `src/core/`      | widget event types, processor port, clock, failures, and projection logic |
| `src/messaging/` | envelope, broker, consumer, backoff, inbox, DLQ, outbox store, and relay  |
| `src/health/`    | Fastify-only probe and metrics sidecar                                    |
| `src/telemetry/` | redacted Pino logger, readiness state, and low-cardinality counters       |
| `src/testutil/`  | deterministic fake clock                                                  |

## Verification

The canonical offline gate runs formatting, ESLint with zero warnings, strict type checking, Jest with at most two workers, the high-severity npm audit policy, and the production TypeScript build:

```bash
npm run verify
# equivalent shim
make verify
```

Tests use an injected fake clock, recording waiters, controlled promises, and in-process Fastify injection. They use no real sleeps or external services.

## Release And Recovery

Build the multi-stage image from the committed lockfile and promote one digest. The image runs emitted ESM as numeric user `10001`; rollout configuration supplies a termination grace period longer than `SHUTDOWN_TIMEOUT_MS`. During recovery, replay preserves the original event ID and passes through current envelope validation and inbox idempotency. DLQ inspection and replay remain operator-controlled actions, not an automatic loop.

## Related Exemplars And Handbook Docs

- [exampleservice](../exampleservice/) implements the synchronous widget HTTP and PostgreSQL boundary.
- [examplefrontend](../examplefrontend/) provides the React administration surface for the same widget domain.
- [Eventing and messaging](../../services/eventing-and-messaging.md), [resilience](../../operations/resilience.md), and [testing](../../quality/testing.md) govern this package's boundaries and proof.
