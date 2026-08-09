# Module Design

Dependency and export rules that keep TypeScript modules comprehensible and replaceable.

## Default Approach

Organize by capability, point dependencies inward, and expose the smallest stable surface.

### Dependency Direction

For services, preserve `api -> core <- db`. `core` owns domain types, use cases, and ports. Transport and persistence modules translate framework values into core values. `src/index.ts` is the composition root and may know concrete adapters.

React features may depend on shared application primitives, never on another feature's internals. Promote code to a shared module only after ownership and reuse are demonstrated.

Allowed imports point inward through an owned public module:

```ts
// src/api/widgets-route.ts
import { createWidget } from "../core/widgets.js";

// src/db/postgres-widget-repository.ts
import type { WidgetRepository } from "../core/widgets.js";
```

The inverse directions are forbidden: `src/core` does not import `../api`, `../db`, Fastify, Pino, or `pg`; one feature does not deep-import another feature's private files. Composition is the only place that imports every concrete side.

### Imports And Exports

- Prefer named exports and direct module imports.
- Use `import type` when an import is type-only under `verbatimModuleSyntax`.
- Include runtime-correct file extensions for NodeNext output.
- Keep barrel files shallow and deliberate; do not re-export an entire subtree.
- Define package `exports` explicitly for libraries and test every supported entry point.
- Do not expose adapter-specific values through domain-facing signatures.

### Module Boundaries

A module owns its invariants, construction, and error vocabulary. Consumers receive typed values or narrow ports rather than mutable internals. Keep parsing, authorization, persistence mapping, and transport mapping at their boundaries.

Use dependency injection through constructors or functions at the narrowest useful seam. Avoid service locators, ambient configuration, and mutable module singletons.

### Cycles And Initialization

Import cycles are errors, even when ESM happens to initialize the current graph successfully. Enforce cycle detection in lint or a dedicated analysis step. Break cycles by moving a shared contract inward, inverting a dependency behind a port, or merging modules that are not actually independent.

Top-level module evaluation must be deterministic and side-effect-light. Do not connect to databases, read secrets, start timers, or register process handlers merely by importing a module.

### Boundary Enforcement

Apply a core-specific flat-config block in addition to repository cycle analysis:

```js
export default [
  {
    files: ["src/core/**/*.ts"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          paths: [
            { name: "fastify", message: "core must not import HTTP" },
            { name: "pino", message: "core must use an owned logging port" },
            { name: "pg", message: "core must not import persistence" },
          ],
          patterns: [
            {
              group: ["../api/**", "../db/**", "../telemetry/**"],
              message: "dependency direction points toward core",
            },
          ],
        },
      ],
    },
  },
];
```

Relative depth varies with layout; keep the groups proven by lint fixtures. Use a dedicated boundaries plugin when pattern complexity starts hiding the architecture.

### Public Libraries

Treat exported types, runtime values, subpaths, and error behavior as compatibility contracts. Keep internal files unavailable through `exports`. Produce declarations from the same build that produces JavaScript, and smoke-test a packed artifact from a consumer fixture.

Keep test helpers behind test-only paths and package conditions. Production consumers must not acquire a hidden dependency on fixture builders, Jest globals, or source-only aliases.

## Common Mistakes And Forbidden Patterns

- Core importing Fastify, React, Pino, `pg`, environment state, or generated transport models.
- Deep imports into another module's private files.
- A catch-all `utils.ts`, `common.ts`, or barrel that obscures ownership.
- Structural typing used to bypass an intended mapping boundary.
- Import cycles accepted because tests currently pass.
- Side effects during import or test-only hooks in production exports.
- Module aliases that resolve in editors but not in emitted Node.js or packed consumers.
- Publishing paths that exist in the repository but not in package `exports`.

## Verification And Proof

- Static cycle analysis reports no cycles.
- Lint and typecheck enforce allowed import directions and type-only imports.
- Unit tests exercise core modules without network, database, process environment, or framework setup.
- Adapter tests prove explicit mapping into and out of core values.
- A library pack/install smoke test resolves every documented export and its declarations.
- Review can identify one owner and one primary responsibility for each new module.

Related: [project-setup.md](project-setup.md), [shared-constructs.md](shared-constructs.md), and [contracts-and-compatibility.md](contracts-and-compatibility.md).
