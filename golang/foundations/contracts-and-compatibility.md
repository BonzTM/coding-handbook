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
| public library API | exported Go types and behavior | the documented exception to internal-first: an intentional `pkg/` surface or a separate module, never the repo root by default |

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
- Any contract consumed outside one package follows the deprecation policy below before removal.
- Internal contracts can change faster than public ones, but they still need coordinated callers and proof.

## Versioning And Deprecation Policy

A published surface is anything a caller depends on that is not a single internal package: a library's exported API, an HTTP or gRPC API, an event schema. Every published surface is versioned, and nothing published is removed without going through a deprecation window. Internal (non-published) packages are exempt and may move as fast as their callers can be updated in the same change.

### Semver Discipline

- Published surfaces follow semantic versioning. The release tag is canonical `v1.2.3` — the `v` prefix is required and the tag is the source of truth, not a `VERSION` file or a constant.
- A MAJOR bump is triggered by any breaking change to an exported or wire contract: removing or renaming an exported identifier, changing a function signature or struct field in an incompatible way, removing or renaming a proto field or RPC, changing the meaning of an existing field, or tightening validation that previously accepted valid input. When in doubt, it is breaking.
- MINOR covers backward-compatible additions: a new exported function, a new optional field, a new endpoint, a new proto field with a fresh tag number.
- PATCH covers backward-compatible fixes with no contract change.
- `v0.x.y` signals an unstable surface where MINOR may break. Reaching `v1.0.0` is a commitment to the policy above; do not ship a `v1` you are not ready to keep stable.

### Module Major Versions

- Go encodes the major version in the module path for v2 and later. A v2+ module declares `module example.com/foo/v2` in `go.mod`, and importers write `import "example.com/foo/v2"`. The matching tags are `v2.0.0`, `v2.1.0`, and so on.
- v0 and v1 use the unsuffixed path (`example.com/foo`); there is no `/v0` or `/v1`.
- A v2+ major is a new import path that coexists with the old one, so consumers migrate on their own schedule. This is the mechanism that makes a MAJOR bump non-breaking to ship: old and new live side by side until the old path is deprecated and removed.

### Deprecation Markers

- Mark a deprecated Go identifier with the tooling-recognized godoc convention: a paragraph in the doc comment beginning `// Deprecated:` stating the reason and the replacement, for example `// Deprecated: use NewClient; Connect will be removed in v3.`. `staticcheck` (the `SA1019` check, which runs under golangci-lint per [../quality/linting.md](../quality/linting.md)) and editors flag callers of the symbol; nothing else is needed to mark it. Note `go vet` itself does not report use of deprecated identifiers — `SA1019` does.
- For HTTP APIs, signal deprecation on the wire as routed in [../decisions/framework-selection.md](../decisions/framework-selection.md) (the API deprecation signaling row): a `Deprecation` header plus a `Sunset` header (RFC 8594) carrying the removal date.
- For proto surfaces, mark the field, RPC, or message with `option deprecated = true;` (`[deprecated = true]` on a field). This is the canonical marker that `buf` lint and protobuf-aware tooling key off; note the effect on generated code is language-dependent (some generators emit a deprecation annotation, others do not), so do not rely on a compiler warning alone — pair it with the announcement step below.
- For events, record the deprecation against the schema's authoritative source and announce it on the same channel consumers watch for the schema.

### Deprecation Window

Removal is a sequence, not a single commit. Every removal of a published contract runs this window:

1. Mark deprecated in code with the marker above and a stated target release for removal.
2. Announce it in the CHANGELOG and release notes for the release that ships the marker. See [../operations/ci-and-release.md](../operations/ci-and-release.md).
3. Measure usage. Removal does not happen until usage is observably zero (telemetry on the deprecated path, request logs, or consumer confirmation), and not before the stated target release.
4. Remove in the stated later release, which is a MAJOR bump if the surface is at v1+.

The full mechanics — choosing the target release, instrumenting the deprecated path, and the removal change itself — live in [../recipes/deprecate-and-remove-contract.md](../recipes/deprecate-and-remove-contract.md). This policy defines when the window applies and what each step must produce; the recipe is the executable how.

## Common Mistakes And Forbidden Patterns

- Proto files or JSON responses that mirror database tables instead of transport needs.
- Changing a public payload shape without compatibility review.
- Treating generated code as the contract while letting the source schema drift.
- Shipping migrations that assume every process upgrades at once.
- Relying on tribal knowledge instead of one obvious source of truth for the boundary.
- Shipping a breaking change to a v1+ surface as a MINOR or PATCH, or removing a contract with no marker, no announcement, and no measured-zero-usage step.
- Releasing a v2+ module without the `/v2` major-version suffix in the module path and import path, so the breaking version overwrites v1 for every consumer.
- A bare `// Deprecated:` comment with no replacement and no removal target, which flags the symbol but gives callers nowhere to go.
- Removing a deprecated surface on the next release because it "felt long enough" instead of waiting for the stated release and observed-zero usage.

## Verification And Proof

- transport tests that prove request decoding, validation, and response shape
- proto lint or generation checks for gRPC surfaces
- migration apply tests and compatibility review for schema changes
- consumer-focused tests for event or external payload changes
- release notes when contracts, env vars, migrations, or compatibility expectations change
- a canonical `v1.2.3` tag whose major component matches the nature of the change, and a `/vN` module path for any v2+ release
- `go vet` and `staticcheck` clean; `staticcheck`'s `SA1019` is what flags callers of a `// Deprecated:` symbol that is still supported but on its way out (run under `make verify` via golangci-lint, not by `go vet`)
- for any removal of a published contract: the deprecation marker in a prior release, the CHANGELOG and release-note entry that announced it, and the usage measurement showing zero before removal
