# Framework Selection

Standardize the TypeScript stack and require evidence before adding overlapping frameworks.

## Default Approach

Use the defaults below for new projects. An inherited repository keeps its established stack unless the task explicitly includes migration.

### Approval Questions

Before adding or replacing a dependency, answer:

1. What repo-owned problem cannot the platform or current dependency solve safely?
2. Is the package maintained, licensed appropriately, typed, ESM-compatible, and compatible with Node 24 or supported browsers?
3. What transitive code, install scripts, native binaries, runtime permissions, and vulnerability exposure enter the graph?
4. How will the boundary be tested, observed, upgraded, and removed?
5. Is the choice reversible? If not, where is the accepted ADR?

The lockfile diff is part of the review. Package popularity is not a substitute for these answers.

## Default Choices By Concern

| Concern | Default and rationale | Reach for something else when |
|---|---|---|
| Runtime | Node.js 24 LTS: stable production line with built-in fetch, Web Streams, AbortController, test runner, env-file support, and a stable permission model | the organization mandates another runtime and its compatibility is proven |
| Compiler | newest TypeScript supported by the pinned type-aware lint toolchain, locally pinned (currently TypeScript 6.0.3 because typescript-eslint 8 does not support TypeScript 7, which is current stable at verification time); `tsc` owns type checking and backend emission | a frontend framework owns emit; `tsc --noEmit` still gates types |
| Modules | ESM-only, `type: module`, NodeNext on backend and bundler resolution on frontend | publishing consumers require multiple formats, via ADR and export tests |
| Package manager | npm with committed `package-lock.json`; `npm ci` in CI | workspace scale and measured constraints justify pnpm; never mix lockfiles |
| HTTP framework | Fastify v5: explicit plugin encapsulation, hooks, schemas, injection testing, and Pino integration | Express 5 compatibility is required by inherited middleware or platform conventions |
| Validation | Zod 4: one runtime schema yields parsed, typed boundary values | a controlling JSON Schema/OpenAPI generation pipeline is already authoritative |
| Logging | Pino: structured JSON, child bindings, redaction, Fastify integration | a host mandates another sink adapter; retain an injected repo-owned logger interface |
| Data access | `pg`, parameterized handwritten SQL, Zod row parsing | Kysely is approved when query volume makes manual types costly; ORM adoption needs an ADR |
| ORM | none | Prisma or TypeORM provides measured value worth schema ownership, generated surface, migration coupling, and query opacity; document it in an ADR |
| Migrations | `node-pg-migrate`, explicit deployment job | an established platform migrator already owns PostgreSQL changes |
| Test runner | Jest 30 across backend and React for one assertion/mocking ecosystem and mature RTL integration | a Vite-heavy frontend proves Vitest materially simplifies transforms and parity gaps are covered, via ADR |
| Component tests | React Testing Library plus user-event | browser-only behavior requires a real-browser suite |
| External HTTP fakes | MSW on frontend; local HTTP fakes on backend | contract tests require a real sandbox |
| Server state | TanStack Query for caching, cancellation, mutation state, and invalidation | no remote state exists; use local React state |
| Global client state | none by default; local state, lifting, then narrow context | complex client-only state transitions justify a state machine or Redux through ADR |
| Routing | React Router | a selected application framework controls routing |
| Frontend build | Vite | a selected framework controls bundling or library output needs a specialized build |
| Telemetry | OpenTelemetry JS for traces and metrics; Pino for logs; OTLP export | the platform requires a compatible exporter; OTel JS logs remain non-default while their SDK status is development |
| Containers | multi-stage OCI image, Debian slim or distroless runtime, non-root | native dependencies have proven Alpine/musl support and size gains justify the tradeoff |

Fastify’s current reference identifies v5.11.x, while its TypeScript type-provider documentation lists Zod support; Zod 4 is stable. Jest 30 still marks ESM support experimental and requires Node’s VM modules path, so the repository must pin and test its transformer/configuration rather than pretend ESM is transparent. See [Fastify reference](https://fastify.dev/docs/latest/Reference/), [Fastify type providers](https://fastify.dev/docs/latest/Reference/Type-Providers/), [Zod versioning](https://zod.dev/v4/versioning), and [Jest ESM](https://jestjs.io/docs/30.0/ecmascript-modules).

### Jest Versus Vitest

Jest is the handbook default because one runner across service and browser-like tests reduces configuration and review variance, its fake timers and mocking model are mature, and React Testing Library examples commonly target it. This is a consistency decision, not a claim that Jest has the simplest ESM story.

Reach for Vitest when the application is inseparable from Vite, Jest’s ESM transform path is a persistent maintenance cost, and the team demonstrates equivalent coverage, mocking, timer, CI, and editor behavior. Record the change because test semantics are part of the repository contract.

### pg Versus Query Builders And ORMs

Handwritten SQL keeps PostgreSQL behavior, indexes, locks, and query plans visible. Parameter binding prevents values from becoming SQL syntax; Zod row parsing prevents database rows from becoming trusted domain objects by assertion.

Kysely is the first escalation when numerous compositional queries make manual types repetitive. Prisma and TypeORM are not defaults: each introduces a larger abstraction, schema/migration ownership questions, and query behavior that must be evaluated against operational needs. “Fewer lines” alone does not justify that coupling.

### npm Versus pnpm

npm is present with Node, `package-lock.json` is widely understood, and `npm ci` enforces lockfile consistency. pnpm can earn adoption for a real multi-package repository with measured disk/install or dependency-isolation needs. Workspaces and package-manager replacement are architectural changes, not bootstrap preferences.

## Mandated Frameworks

The selected defaults are mandatory for greenfield repositories until an accepted ADR says otherwise. Do not add a second validator, logger, HTTP client abstraction, test runner, state cache, or persistence abstraction beside the default.

Framework-specific objects stop at adapters. Domain code accepts typed values and narrow ports, not Fastify requests, React query results, Pino instances, or `pg` result objects.

## Common Mistakes And Forbidden Patterns

- Adding both Jest and Vitest, Zod and another validator, or Pino and another logger without a migration plan.
- Choosing Express because it is familiar when no compatibility requirement exists.
- Introducing Redux for server state already owned by TanStack Query.
- Treating an ORM’s generated types as validation of untrusted database or wire data.
- Running native TypeScript type stripping as the production build.
- Adding npm workspaces preemptively to a single deployable.
- Depending on an unmaintained package, broad install script, or floating version without review.
- Claiming a performance benefit without a representative benchmark.

## Verification And Proof

- `npm run verify` passes with one package manager and one lockfile.
- `npm ls --all` and the lockfile diff are reviewed for new dependencies.
- The dependency supports the repository’s Node, TypeScript, ESM, and browser targets.
- Security, license, maintenance, transitive, and install-script risks are recorded.
- The integration is behind a narrow adapter and has negative-path tests.
- Any exception has an accepted ADR with migration and rollback.

Current platform anchors: [Node.js releases](https://nodejs.org/en/about/previous-releases), [Node TypeScript support](https://nodejs.org/download/release/latest-v24.x/docs/api/typescript.html), [TypeScript download](https://www.typescriptlang.org/download/), [OpenTelemetry JS status](https://opentelemetry.io/docs/languages/js/), and [node-pg-migrate](https://www.npmjs.com/package/node-pg-migrate).
