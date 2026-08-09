# Shared Constructs

Rules for shared helpers, ports, test utilities, and cross-cutting abstractions.

## Default Approach

Share only stable concepts with a clear owner and at least two real consumers.

### Ownership Before Reuse

Keep code with the feature that owns its meaning. Duplicate a small, volatile helper rather than create a dependency magnet. Promote it only when the behavior, vocabulary, and lifecycle are genuinely common.

Every shared module states its owner, consumers, supported behavior, and dependency constraints. Shared code is not ownerless code.

### Narrow Constructs

Good shared seams are small and semantic: `Clock`, ID construction, result types, safe pagination bounds, logger interface, or an outbound-client policy. They depend inward and avoid framework objects.

Prefer functions and readonly values. A class is justified by identity, lifecycle, or encapsulated mutable state, not as a namespace. Keep generic helpers constrained to one understandable relationship.

Shared error codes and result shapes remain small, stable, and independent of transport status. Translation to HTTP, UI, database, or broker concepts stays in the owning adapter.

### Ports And Adapters

Define a port where the consumer needs it. The adapter implements that contract and is wired by composition. Do not create an enterprise-wide interface package detached from its callers.

Cross-cutting adapters such as logging, HTTP, telemetry, and configuration remain under their owned infrastructure modules. Core receives only the narrow behavior it needs.

### Test Utilities

Place reusable test builders, fixed clocks, controlled promises, and fixture factories under `src/testutil` or a test-only path. Production code never imports it.

Builders return valid objects by default and accept explicit overrides. Fakes implement observable behavior; avoid broad mock frameworks and call-order scripts unless an ADR justifies them.

### Package Extraction

Do not extract a shared npm package merely to avoid copying. Extraction requires independent ownership, versioning, compatibility, release, security, and consumer proof. npm workspaces or cross-repository packages require the routing and ADR process.

## Common Mistakes And Forbidden Patterns

- `utils`, `common`, or `helpers` modules containing unrelated behavior.
- Shared constants whose meanings differ by caller.
- A base class or generic repository imposed across unrelated domains.
- Shared transport DTOs presented as domain models across independent boundaries.
- Framework and domain types mixed in one shared package.
- Production imports from test utility paths.
- Global mutable singleton used as a convenience seam.
- Package extraction without an owner and compatibility contract.

## Verification And Proof

- The PR names the owner and at least two genuine consumers.
- Static analysis proves shared modules do not introduce import cycles.
- Public signatures expose semantic values rather than framework internals.
- Error and result helpers remain exhaustively handled without transport coupling.
- Unit tests exercise shared behavior without process-global state.
- Test-only imports are excluded from production build and exports.
- Package extraction includes pack/install, versioning, rollback, and consumer tests.

Related: [module-design.md](module-design.md), [type-system.md](type-system.md), and [../decisions/architecture-decision-records.md](../decisions/architecture-decision-records.md).
