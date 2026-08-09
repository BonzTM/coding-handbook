# Contracts and Compatibility

Schemas, APIs, messages, and public Python surfaces are first-class engineering contracts.

## Default Approach

Every boundary has one authoritative source, owner, compatibility policy, and proof.

| Boundary | Contract form | Primary owner |
|---|---|---|
| HTTP API | FastAPI-generated OpenAPI plus adapter DTOs, status codes, headers, and problem shape | `api/http` plus owning core use case |
| gRPC API | versioned `.proto` source plus generated Python/stubs | `api/` plus `api/grpc` adapter |
| database | Alembic migrations, query/transaction behavior | `db` |
| event/message | explicit Pydantic envelope or schema, stable name, delivery/idempotency rules | owning producer and consumer adapters |
| public library | documented exported Python API and wheel metadata | distribution owner |

### OpenAPI Is Published Evidence

FastAPI DTOs and router declarations generate the HTTP OpenAPI contract. Export and commit a normalized spec when consumers, generated clients, or compatibility tooling depend on it; CI regenerates and fails on drift. If the service publishes through an external spec registry, that registry path and generation direction are documented. Runtime code and published spec never evolve independently.

Generated clients consume the published spec and are generated in a reproducible pinned workflow. Do not hand-edit them. Generated code is evidence derived from the contract, not the authoritative source.

### Additive Evolution

- Add optional response/event fields; tolerant consumers ignore fields they do not use.
- Add optional request fields with server defaults only when old clients retain their prior behavior.
- Treat required-field additions, removals, renames, type changes, tighter validation, changed defaults, and changed meaning as breaking.
- Treat enum additions as compatibility-sensitive: consumers must have a documented unknown-value path before producers emit them.
- Keep old and new database/application versions interoperable during rolling deploys through expand/contract migrations.

Strict inbound DTOs reject fields not in the currently published command contract. This does not conflict with additive rollout: update the contract and server before a client sends the new field. External response/event consumers follow the tolerant-reader rule in [serialization](serialization.md).

### Versioning

Published Python distributions follow [PEP 440](https://peps.python.org/pep-0440/) for version syntax and Semantic Versioning for compatibility meaning. HTTP/gRPC breaking versions use an explicit path/package version and coexist during migration. Events publish a new type/schema version when meaning changes incompatibly; never mutate an old event's meaning in place.

Internal modules may change with all in-repo callers in one change. Anything consumed outside one module/repo is published and follows the deprecation window.

### Deprecation And Removal

Removal is a sequence:

1. Mark the surface deprecated in code/schema and name the replacement and target removal release/date.
2. Announce it in CHANGELOG, release notes, and the channel its consumers watch.
3. Instrument usage and wait for observed zero or explicit consumer confirmation.
4. Remove only at the stated breaking version after the window.

Python public APIs use a runtime `DeprecationWarning` with an appropriate `stacklevel` plus documentation; tests enable and assert it because that warning category is normally hidden from end users. HTTP responses send a standards-based `Deprecation` signal and a `Sunset` header whose date follows [RFC 8594](https://www.rfc-editor.org/rfc/rfc8594). Proto fields/RPCs use their schema deprecation option. Event deprecation lives in the authoritative schema/catalog.

The executable procedure is [deprecate and remove contract](../recipes/deprecate-and-remove-contract.md).

### Consumer Compatibility

Contract tests are written from observable consumer behavior, not internal implementation. Provider tests pin response/problem examples and schema. Consumers validate used fields, tolerate additive unknown fields, and test unknown enum/version behavior. Do not require every consumer to upgrade atomically.

### Message Payload Contracts

Every published message has a stable event type and explicit envelope fields for event ID, source, occurred-at time, schema version, and correlation context. Document ordering key, retry/settlement, idempotency, and DLQ behavior. Producer storage and publication follow the outbox policy when one state change must atomically imply an event; see [eventing and messaging](../services/eventing-and-messaging.md).

## Common Mistakes And Forbidden Patterns

- Database or domain objects reflected directly into a public payload.
- Runtime DTOs and a committed/registered OpenAPI spec drifting independently.
- Generated client code treated as the source or edited by hand.
- Required additions, tighter validation, enum emission, or changed defaults called non-breaking without consumer proof.
- A rolling migration that assumes all processes upgrade at once.
- A public Python symbol removed after only a prose note or without a warning test.
- HTTP deprecation with no replacement, Sunset date, usage measurement, or consumer announcement.
- Event meaning changed in place under the same name/version.

## Verification And Proof

```bash
uv run pytest -k "contract or compatibility or deprecat"
make verify
```

Regenerate and diff OpenAPI/stubs; run provider and consumer tests; apply migrations across mixed application versions where relevant; assert warnings/headers and replacement guidance; and require usage evidence before removal. Release notes identify every payload, config, migration, or public API compatibility effect.

Related: [serialization](serialization.md), [gRPC services](../services/grpc-services.md), and [CI and release](../operations/ci-and-release.md).
