# TypeScript Maintainer Reference

> **Slow-path document.** Read [AGENTS.md](AGENTS.md) first. Use this when architecture, lifecycle, taxonomy, or failure diagnosis needs more context.

## Architecture Rationale

The handbook standardizes one package and one deployable first because repository, package, runtime, and ownership boundaries should not diverge before the product requires them. npm workspaces, shared packages, and extra deployables add independent release and compatibility surfaces; they require an ADR.

Node 24 LTS is the production runtime. Production backend code is emitted by TypeScript and runs as ESM. Node's native type stripping is not the compiler, ignores `tsconfig.json`, and is not the deployed build. Frontends use Vite-owned emit and retain `tsc --noEmit` as the type gate.

Strict compiler and typed-lint settings make accidental uncertainty visible. They do not validate runtime input. Zod establishes trust at HTTP, environment, database, message, file, browser-storage, and external-response boundaries; internal code then works with typed domain values.

The backend dependency shape is:

```text
src/api/  ->  src/core/  <-  src/db/
                     ^
src/lib/ adapters ----|
src/index.ts composes concrete dependencies
src/main.ts owns process lifetime
```

Core owns business vocabulary, decisions, and ports. API, database, HTTP-client, broker, and telemetry modules translate external behavior. Composition is allowed to know every concrete dependency; no other module is.

React follows the same ownership principle. Routes translate navigation, components translate user intent, features own behavior, `src/lib/` owns browser boundary policy, and TanStack Query owns remote state. Client guards improve UX but never replace server authorization.

## Module Map

| Path | Owns | Must not own |
|---|---|---|
| `src/main.ts` | fatal startup, signals, exit | business rules, SQL, request mapping |
| `src/index.ts` | composition, readiness, lifecycle | domain decisions |
| `src/config/` | environment selection and Zod parsing | ambient mutable config |
| `src/core/` | domain values, use cases, consumer-owned ports | Fastify, React, Pino, `pg`, env reads |
| `src/api/` | transport schemas, auth enforcement, problem mapping | SQL and business decisions |
| `src/db/` | SQL, transactions, row schemas, persistence mapping | transport DTOs |
| `src/lib/http/` | URL policy, fetch, timeout, response parsing | domain decisions |
| `src/telemetry/` | Pino and OpenTelemetry construction | sensitive payloads, domain policy |
| `src/testutil/` | controlled fakes and valid builders | production imports |
| `src/app/` | frontend composition | feature internals |
| `src/features/` | feature-owned UI behavior | unrelated shared utilities |
| `src/routes/` | navigation boundaries | server authorization |

## Process Lifecycle

Startup follows one direction:

1. read and parse configuration before side effects;
2. construct redacted Pino and telemetry providers;
3. create bounded pools, clients, brokers, and adapters;
4. compose core use cases with those adapters;
5. register Fastify plugins/routes or worker handlers;
6. initialize required dependencies and migrations only through the explicit deploy job;
7. mark readiness true and accept work.

Shutdown reverses ownership:

1. mark readiness false;
2. stop HTTP, broker, scheduler, and queue intake;
3. abort the process-lifetime signal;
4. drain owned work within a fixed deadline;
5. close servers, consumers, pools, workers, and exporters;
6. exit zero for a clean signal or nonzero for fatal failure.

Every promise is awaited, returned, or supervised by a lifecycle object that observes rejection. Every queue, batch, page loop, retry, timer, stream, and subprocess has a hard bound. No work starts merely because a module was imported.

## Contract And Data Lifecycle

Transport, event, database, browser, and package contracts evolve additively by default. Old and new versions must coexist during rollout and rollback. Required fields, removed fields, narrowed values, changed nullability, changed error semantics, and declaration incompatibilities are breaking changes.

PostgreSQL uses explicit parameterized SQL and Zod row parsing. TypeScript query generics describe an expectation; they do not prove runtime rows. Migrations run as one observable deployment job and use expand/migrate/contract for destructive evolution.

At-least-once delivery is the messaging assumption. Producers use a transactional outbox when state and publish intent must agree. Consumers use an inbox or equivalent durable dedupe when PostgreSQL effects must survive duplicate and concurrent duplicate delivery.

## Test Taxonomy

| Test | Default boundary | Proves |
|---|---|---|
| unit | pure module beside source | domain decision, parser, mapper, retry policy |
| transport | Fastify `app.inject()` | parsing, auth, status/problem mapping, hook scope |
| component | React Testing Library | accessible user behavior and state rendering |
| integration | Testcontainers PostgreSQL | SQL, constraints, transactions, migrations, row mapping |
| external boundary | local HTTP server or MSW | request/response policy, timeout, abort, malformed data |
| contract | schemas, golden fixtures, packed consumer | independent compatibility and package exports |
| lifecycle smoke | built process or container | emitted ESM, readiness, signal, drain, resource close |

Jest uses Babel to transform TypeScript to CommonJS only inside the test runtime. Babel performs no type checking, so `tsc --noEmit` remains mandatory. The production-artifact smoke test catches NodeNext problems hidden by the test transform.

## Verification Model

`npm run verify` is canonical and runs format check, typed lint, typecheck, deterministic Jest tests, audit policy, and build. CI performs `npm ci` before the script. Integration runs separately with pinned real dependencies. The Makefile is a shim and does not duplicate policy.

The executable proofs are [exampleservice](reference/exampleservice/) for Fastify and PostgreSQL, [exampleworker](reference/exampleworker/) for bounded event processing, and [examplefrontend](reference/examplefrontend/) for React, routing, TanStack Query, accessible forms, MSW, and static assets. Each package owns a committed lockfile and runs the canonical gate independently.

The narrowest meaningful test runs first. Broader proof follows the blast radius: contract changes need mixed-version fixtures, persistence changes need PostgreSQL, UI changes need accessibility behavior, libraries need pack/install, and lifecycle changes need a built-process smoke test.

## Troubleshooting

| Symptom | Likely cause | First fix |
|---|---|---|
| `ERR_MODULE_NOT_FOUND` only from `dist/` | NodeNext relative import omitted runtime `.js` extension or points at a source-only alias | inspect emitted import; author the TypeScript relative specifier with `.js`; run built-artifact smoke |
| Jest passes but emitted app fails | Babel mapper concealed production resolution or export error | build and run `node dist/main.js`; fix NodeNext path/package exports |
| Jest reports `Cannot use import statement outside a module` | transform/config mode drifted between native ESM and CommonJS | restore `jest.config.cjs`, Babel CommonJS modules, and plain `jest`; remove VM-modules/ESM options |
| `ts-jest` transform diagnostics or preset conflicts | repository mixed an unapproved transformer with the Babel baseline | remove overlapping transform or record and prove an ADR-governed alternative |
| a value import disappears at runtime | it was type-only under `verbatimModuleSyntax`, or a type import was used for a runtime side effect | separate `import type`; import runtime initialization explicitly and test it |
| CommonJS dependency has no expected named export | Node ESM/CJS interop shape differs from its declarations or bundler behavior | inspect the package's documented exports; isolate a default/namespace import in one adapter |
| ESLint says a file is outside the project service | the file is not included by the relevant tsconfig | include it deliberately or add a non-type-checked config-file override |
| tests hang after completion | server, pool, container, timer, MSW server, or worker was not closed | find the owner; close in `finally`/teardown; diagnose with open-handle reporting |
| config works locally but fails in deploy | `.env.example`, platform key, and Zod schema drifted | compare exact names; fail startup before readiness; update all sync surfaces |
| duplicate messages create duplicate state | acknowledgement precedes durable effect or dedupe is not atomic | move inbox receipt and effect into one transaction; acknowledge afterward |
| React effect runs twice in development | Strict Mode exposed non-idempotent setup or missing cleanup | make setup reversible, return cleanup, and move irreversible work to an event/server boundary |
| metrics backend cardinality grows | raw URL, ID, tenant, SQL, key, or error text became an attribute | replace with route template and closed result vocabulary |

## Primary Sources

- [Node.js releases](https://nodejs.org/en/about/previous-releases)
- [Node.js TypeScript support](https://nodejs.org/download/release/latest-v24.x/docs/api/typescript.html)
- [TypeScript downloads](https://www.typescriptlang.org/download/)
- [typescript-eslint typed linting](https://typescript-eslint.io/getting-started/typed-linting/)
- [Jest TypeScript setup](https://jestjs.io/docs/30.0/getting-started#using-typescript)
- [Jest ESM status](https://jestjs.io/docs/30.0/ecmascript-modules)
- [Fastify testing](https://fastify.dev/docs/latest/Guides/Testing/)
- [React input accessibility](https://react.dev/reference/react-dom/components/input#providing-a-label-for-an-input)

## Related Docs

- Fast path and routing: [AGENTS.md](AGENTS.md)
- Human entrypoint: [README.md](README.md)
- Procedures: [recipes/README.md](recipes/README.md)
- Gates: [checklists/README.md](checklists/README.md)
- Ownership transfer: [onboarding-and-handoff.md](onboarding-and-handoff.md)
