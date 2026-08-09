# Contracts And Compatibility

Versioning rules for HTTP, event, library, and persisted contracts that evolve across independent deploys.

## Default Approach

Make contracts explicit, validate both sides, and prefer additive evolution with a measured deprecation window.

### Contract Sources

Publish the controlling artifact for each boundary: OpenAPI or JSON Schema for HTTP, a versioned schema for events, package exports and declarations for libraries, and migrations plus row parsing for persistence.

Generated artifacts are committed or reproduced deterministically in CI. Generated TypeScript types never replace runtime validation at a trust boundary.

### Compatibility Rules

- Adding an optional response or event field is normally backward compatible when readers ignore allowed unknown fields.
- Adding a required request field, removing or renaming a field, narrowing accepted values, or changing meaning is breaking.
- Changing nullability, omission rules, units, precision, ordering, or error semantics is a contract change.
- Enum expansion is breaking for consumers that exhaustively reject unknown values; document an unknown-value strategy.
- HTTP status and problem `type` values are part of the public contract.

Assess both directions during rolling deployment: old producer/new consumer and new producer/old consumer. Internal APIs still cross version boundaries during rollout and rollback.

### Versioning And Deprecation

Avoid version numbers for additive changes. When an incompatible change is necessary, introduce a parallel field, route, event version, or package major; migrate consumers; observe remaining use; then remove the old contract.

Every deprecation names an owner, announcement date, removal criteria, earliest removal date, telemetry, and consumer migration path. Follow [../recipes/deprecate-and-remove-contract.md](../recipes/deprecate-and-remove-contract.md).

### Shared Schemas

Share schemas only through a versioned, owned package when producer and consumers truly share a release contract. Do not import service source across repositories or assume a shared TypeScript type proves wire compatibility.

Consumers remain responsible for parsing received values and handling forward-compatible alternatives.

### Idempotency And Retries

Document whether writes are safe to retry, how idempotency keys are scoped, the retention window, and which response is replayed. Message contracts state delivery, ordering, duplication, and replay expectations.

### Library Compatibility

Treat exports, types, runtime behavior, errors, side effects, and supported Node/browser ranges as API. Use semantic versioning. Smoke-test the packed package from a consumer instead of importing repository source.

## Common Mistakes And Forbidden Patterns

- Sharing compile-time interfaces while wire payloads drift independently.
- A generated client accepted as the only compatibility test.
- Required fields added during a mixed-version rollout.
- Reusing an event name while changing its meaning.
- Silent field removal because telemetry was never added.
- Breaking library declaration changes released as a patch.
- Versioning every additive endpoint change and maintaining unnecessary parallel APIs.

## Verification And Proof

- Contract tests validate producer examples and consumer parsing against the controlling schema.
- Golden fixtures include each supported version and unknown/additive fields where allowed.
- Compatibility review covers requests, responses, errors, events, storage, and packed library exports.
- Mixed-version rollout and rollback cases are documented and tested.
- Deprecation telemetry identifies remaining use without high-cardinality metrics.
- Breaking changes have an accepted ADR, migration path, release note, and removal proof.

Related: [serialization.md](serialization.md), [../services/http-services.md](../services/http-services.md), and [../services/eventing-and-messaging.md](../services/eventing-and-messaging.md).
