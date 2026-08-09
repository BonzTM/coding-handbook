# Project Setup

Standard repository shape for TypeScript services, workers, libraries, and React applications.

## Default Approach

Start with one npm package, one deployable, ESM, Node.js 24 LTS, and TypeScript 7, the current stable compiler. Node.js 26 is Current, but Current is not the production pin.

### Runtime And Package Contract

- Set `"type": "module"` and `"engines": { "node": "24.x" }`.
- Pin the same Node 24 release line in `.nvmrc` and CI.
- Commit `package-lock.json`; use `npm ci` in CI and release builds.
- Pin direct tool dependencies in `devDependencies`; invoke them through npm scripts.
- Do not introduce workspaces until multiple independently owned packages justify an ADR.

Node's native TypeScript support strips erasable syntax only. It ignores `tsconfig.json` and performs no type checking. Use it only for bounded development commands; production backend artifacts come from `tsc`.

### Source Layout

Use the smallest layout matching the process:

```text
src/
  api/        # inbound transport adapters
  config/     # environment parsing
  core/       # domain rules and ports
  db/         # PostgreSQL adapter
  lib/        # owned infrastructure adapters
  telemetry/  # logger and OpenTelemetry construction
  index.ts    # composition and process lifetime
```

React applications use `src/app`, `src/components`, `src/features`, `src/routes`, and `src/lib`; keep feature code together until a proven shared seam exists. Tests live beside their subject or under a clearly scoped integration directory.

### Compiler Baseline

Enable `strict`, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`, `noImplicitOverride`, `verbatimModuleSyntax`, `isolatedModules`, and `noFallthroughCasesInSwitch`. Keep `skipLibCheck: true` and source maps enabled.

Backends use `module` and `moduleResolution` set to `NodeNext`. Frontends use bundler resolution and Vite-owned emit. `tsc --noEmit` remains the type gate for every shape.

The backend `tsconfig.json` starts here:

```jsonc
{
  "compilerOptions": {
    "target": "ES2024",
    "lib": ["ES2024"],
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "rootDir": "src",
    "outDir": "dist",
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noImplicitOverride": true,
    "verbatimModuleSyntax": true,
    "isolatedModules": true,
    "noFallthroughCasesInSwitch": true,
    "skipLibCheck": true,
    "sourceMap": true,
    "declaration": true
  },
  "include": ["src/**/*.ts"],
  "exclude": ["dist", "node_modules"]
}
```

Frontend configuration extends the same strict base and changes only the build boundary:

```jsonc
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2022", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "rootDir": ".",
    "jsx": "react-jsx",
    "noEmit": true
  },
  "include": ["src/**/*.ts", "src/**/*.tsx", "vite.config.ts"]
}
```

### Scripts And Artifacts

Provide `format`, `format:check`, `lint`, `typecheck`, `test`, `audit`, `build`, and `verify` scripts. `npm run verify` is canonical; a Makefile may expose only `verify: ; npm run verify`.

Keep the package anatomy explicit and the gate readable:

```jsonc
{
  "name": "example-service",
  "private": true,
  "type": "module",
  "engines": { "node": "24.x" },
  "scripts": {
    "format": "prettier --write .",
    "format:check": "prettier --check .",
    "lint": "eslint . --max-warnings 0",
    "typecheck": "tsc --noEmit",
    "test": "jest",
    "test:integration": "RUN_INTEGRATION=1 jest --selectProjects integration",
    "audit": "npm audit --audit-level=high",
    "build": "tsc",
    "verify": "npm run format:check && npm run lint && npm run typecheck && npm test && npm run audit && npm run build"
  }
}
```

CI runs `npm ci` before this chain. Do not place `npm ci` inside a script executed from the installed dependency tree.

Emit backend JavaScript and source maps into `dist/`. Never execute source TypeScript as the production artifact. Exclude `dist`, coverage, local environment files, and editor state from version control.

### Environment Separation

Commit `.env.example` with names and safe placeholders, never real values. Parse configuration once at startup. Do not make source layout, compiler settings, or dependency resolution depend on a developer's global tools.

## Common Mistakes And Forbidden Patterns

- CommonJS in a new package, mixed module modes, or extensionless NodeNext imports that fail after emit.
- Floating Node, npm, or TypeScript versions across local, CI, and container builds.
- Treating native type stripping as type checking or as the production backend build.
- Multiple lockfiles, install with `npm install` in CI, or unreviewed lockfile rewrites.
- Path aliases that compile but cannot resolve at runtime.
- A premature monorepo, shared package, or build orchestrator without an ADR.
- Business rules in `index.ts`, route modules, React components, or persistence adapters.

## Verification And Proof

- A clean checkout succeeds with `npm ci` on the pinned Node 24 runtime.
- `npm run typecheck` reports zero diagnostics under the strict baseline.
- The backend artifact runs from `dist/`; the frontend production bundle loads directly.
- `npm run verify` performs format check, lint, typecheck, Jest, audit policy, and build.
- The lockfile, `.nvmrc`, `engines`, CI image, and container builder agree.
- No ignored local state or secret appears in the committed tree or artifact.

Related: [module-design.md](module-design.md), [../quality/linting.md](../quality/linting.md), and [../operations/ci-and-release.md](../operations/ci-and-release.md). Platform anchors: [TypeScript download](https://www.typescriptlang.org/download/), [Node.js releases](https://nodejs.org/en/about/previous-releases), and [Node.js TypeScript support](https://nodejs.org/download/release/latest-v24.x/docs/api/typescript.html).
