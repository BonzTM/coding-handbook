# exampleservice

A complete Fastify and PostgreSQL widget service that proves the TypeScript handbook's service rules compile, test, audit, and build together.

## What It Is

- Fastify v5 with `fastify-type-provider-zod` and Zod 4 schemas for headers, params, queries, bodies, responses, environment input, cursors, and PostgreSQL rows.
- Widget create, get, cursor-paginated list, optimistic update, and delete operations behind `api -> core <- db` boundaries.
- A branded `WidgetId`, tenant-scoped authorization, duplicate-name and version conflicts, and RFC 9457 problem details.
- Atomic PostgreSQL create idempotency scoped to tenant, operation, and `Idempotency-Key`; matching retries replay the durable widget and mismatched reuse is rejected.
- Pino JSON logs with construction-time redaction, a separate security-audit sink, bounded request bodies, validated request IDs, `/livez`, and dependency-aware `/readyz`.
- Request and process `AbortSignal` propagation. `pg` 8 does not expose an AbortSignal query API, so the adapter checks cancellation before and after each call and bounds database work with pool, statement, query, lock, and idle-transaction timeouts.
- An injectable telemetry lifecycle seam with an offline no-op default. A deployment can supply an approved OpenTelemetry SDK/exporter adapter without coupling core code or the default verification gate to a collector.

## Requirements

- Node.js 24.18.0 (pinned in `.nvmrc` and the container)
- npm and the committed `package-lock.json`
- PostgreSQL 17 for a running service
- Docker for the explicit Testcontainers integration suite

## Setup And Run

```bash
npm ci
cp .env.example .env
npm run migrate
npm run build
node --env-file=.env dist/main.js
```

Local development defaults to a synthetic `local-dev` principal with reader and writer roles. Set `AUTH_ENABLED=true` and inject `AUTH_TOKEN` to require `Authorization: Bearer <token>` through the exemplar's deliberately limited static-token authenticator. Production identity-provider selection and protocol validation remain organization-owned decisions; this local authenticator is not presented as an internet-facing identity system.

Example requests:

```bash
curl -s localhost:3000/livez
curl -s localhost:3000/readyz

curl -s -X POST localhost:3000/widgets \
  -H 'content-type: application/json' \
  -H 'idempotency-key: create-meter-1' \
  -d '{"id":"550e8400-e29b-41d4-a716-446655440000","name":"Meter","description":null}'

curl -s 'localhost:3000/widgets?page_size=20'
curl -s 'localhost:3000/widgets?cursor=<next_cursor>&page_size=20'
```

## Configuration

| Key                   |      Required | Sensitive | Default     | Meaning                                           |
| --------------------- | ------------: | --------: | ----------- | ------------------------------------------------- |
| `NODE_ENV`            |           yes |        no | none        | `development`, `test`, or `production`            |
| `HOST`                |            no |        no | `127.0.0.1` | listen address                                    |
| `PORT`                |            no |        no | `3000`      | HTTP port, 1 through 65535                        |
| `DATABASE_URL`        |           yes |       yes | none        | PostgreSQL connection URL                         |
| `LOG_LEVEL`           |            no |        no | `info`      | Pino level                                        |
| `AUTH_ENABLED`        |            no |        no | `false`     | require the static local-development bearer token |
| `AUTH_TOKEN`          | conditionally |       yes | none        | required when authentication is enabled           |
| `SHUTDOWN_TIMEOUT_MS` |            no |        no | `10000`     | whole-process drain deadline                      |
| `DATABASE_TIMEOUT_MS` |            no |        no | `5000`      | connection, statement, and transaction bound      |
| `DATABASE_POOL_SIZE`  |            no |        no | `10`        | maximum PostgreSQL connections                    |
| `IDEMPOTENCY_TTL_MS`  |            no |        no | `86400000`  | durable create-replay retention                   |

Only `src/config/index.ts` reads runtime environment state. Migrations run as an explicit deployment step with `npm run migrate`; application replicas never migrate on startup.

## Package Map

| Path               | Responsibility                                                                                   |
| ------------------ | ------------------------------------------------------------------------------------------------ |
| `src/main.ts`      | fail-fast startup, fatal handlers, signals, and bounded shutdown                                 |
| `src/index.ts`     | composition, listener lifetime, readiness, pool and telemetry cleanup                            |
| `src/config/`      | strict environment selection, parsing, defaults, and conditional secret validation               |
| `src/core/`        | widget domain, authorization, errors, cursor, clock, and repository port                         |
| `src/api/`         | Fastify construction, auth hook, schemas, routes, cancellation, problem details, and DTO mapping |
| `src/db/`          | bounded pool, transaction helper, parsed rows, in-memory/PostgreSQL repositories, and migrations |
| `src/telemetry/`   | redacted operational logger, separate audit sink, readiness, and optional lifecycle seam         |
| `src/testutil/`    | fixed clock, builders, fakes, and injectable HTTP test composition                               |
| `src/integration/` | environment-gated PostgreSQL/Testcontainers proof                                                |

## Verification

The canonical offline gate runs formatting, ESLint with zero warnings, strict type checking, Jest with at most two workers, the high-severity npm audit policy, and the production TypeScript build:

```bash
npm run verify
# equivalent shim
make verify
```

The default Jest project excludes real infrastructure. Run the PostgreSQL suite explicitly; the script sets the required gate and caps Jest at one worker:

```bash
npm run test:integration
```

The integration suite starts pinned `postgres:17.6`, applies the same `node-pg-migrate` SQL files used by deployment, and proves timestamps, CRUD, optimistic versions, and atomic idempotent replay against PostgreSQL.

## Release And Recovery

Build the multi-stage image from the committed lockfile and promote one digest. Run `npm run migrate` as a single observable job before compatible application rollout. The image runs emitted ESM as numeric user `10001`; rollout configuration supplies secrets and a termination grace period longer than `SHUTDOWN_TIMEOUT_MS`. Prefer forward repair over destructive down migration when deployed data has changed.

## Related Exemplars And Handbook Docs

- [examplefrontend](../examplefrontend/) consumes this service's widget list, create, get, and delete wire contracts through a Zod-parsed browser client.
- [exampleworker](../exampleworker/) demonstrates the event-processing side of the same widget domain.
- [HTTP services](../../services/http-services.md), [database](../../services/database.md), and [testing](../../quality/testing.md) govern this package's boundaries and proof.
