# Contracts and Compatibility

Schemas, APIs, and data contracts are first-class engineering surfaces. Treat them like code, not like incidental documentation.

## Default Approach

Every boundary should have a clear owner and a clear contract.

| Boundary | Contract form | Primary owner |
|---|---|---|
| HTTP API | request and response structs, documented status and error model, optional OpenAPI if the repo publishes one | `internal/api/http` plus the owning core service |
| gRPC API | `.proto` files and generated stubs | `api/<service>/v1` plus `internal/api/grpc` |
| database | schema migrations, query shape, transaction rules | `internal/db` |
| event or queue payload | explicit envelope struct or schema, versioning and idempotency rules | owning producer and consumer packages |
| public library API | exported Go types and behavior | root or `pkg/` public surface |

## Wire Contracts

- HTTP handlers should define stable request and response shapes at the transport boundary.
- Error responses need a consistent shape, not ad hoc JSON per handler.
- gRPC services should use versioned proto packages from day one.
- Generated code policy should be explicit and enforced consistently.

## Data Contracts

- Every schema change ships as a versioned migration.
- Application code and schema changes need a mixed-version rollout story when the system deploys gradually.
- Queries, indexes, and migration order are part of the contract, not just implementation detail.
- Destructive migrations require an explicit rollback or compatibility plan.

## Event And Message Contracts

- Event payloads are contracts, not internal implementation residue.
- Give each published event one authoritative schema source and one stable event name.
- Use explicit metadata for event ID, type, source, time, and correlation context.
- Treat additive evolution as the default. If meaning changes incompatibly, publish a new contract rather than mutating the old one in place.
- Delivery semantics are part of the contract too: ordering guarantees, retry behavior, idempotency expectations, and DLQ policy should be documented where operators and consumers can find them.

Runtime guidance for producers, consumers, retries, outbox, inbox, and DLQ behavior lives in [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).

## Compatibility Rules

- Additive changes are usually safest: new optional fields, new endpoints, new enum values handled carefully.
- Renames, required-field additions, and destructive schema changes need explicit transition plans.
- Deprecation should be documented before removal when a boundary is consumed outside one package.
- Internal contracts can change faster than public ones, but they still need coordinated callers and proof.

## Common Mistakes And Forbidden Patterns

- Proto files or JSON responses that mirror database tables instead of transport needs.
- Changing a public payload shape without compatibility review.
- Treating generated code as the contract while letting the source schema drift.
- Shipping migrations that assume every process upgrades at once.
- Relying on tribal knowledge instead of one obvious source of truth for the boundary.

## Verification And Proof

- transport tests that prove request decoding, validation, and response shape
- proto lint or generation checks for gRPC surfaces
- migration apply tests and compatibility review for schema changes
- consumer-focused tests for event or external payload changes
- release notes when contracts, env vars, migrations, or compatibility expectations change
