# Linting

Flat ESLint, type-aware correctness, import-boundary, React, and formatting policy.

## Default Approach

Use ESLint flat config for correctness and Prettier for formatting; both fail CI with zero warnings.

### Blessed Flat Configuration

Build `eslint.config.mjs` with typescript-eslint v8 flat config. Apply `tseslint.configs.strictTypeChecked` and `tseslint.configs.stylisticTypeChecked` to TypeScript source:

```js
import eslint from "@eslint/js";
import eslintConfigPrettier from "eslint-config-prettier";
import jsxA11y from "eslint-plugin-jsx-a11y";
import reactHooks from "eslint-plugin-react-hooks";
import globals from "globals";
import tseslint from "typescript-eslint";

export default tseslint.config(
  { ignores: ["dist/**", "coverage/**", "node_modules/**"] },
  eslint.configs.recommended,
  {
    files: ["**/*.{ts,tsx}"],
    extends: [
      ...tseslint.configs.strictTypeChecked,
      ...tseslint.configs.stylisticTypeChecked,
    ],
    languageOptions: {
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      "@typescript-eslint/no-floating-promises": "error",
      "@typescript-eslint/no-misused-promises": "error",
      "@typescript-eslint/switch-exhaustiveness-check": "error",
      "@typescript-eslint/consistent-type-imports": "error",
      "@typescript-eslint/no-explicit-any": "error",
      "@typescript-eslint/no-non-null-assertion": "error",
    },
  },
  {
    files: ["src/**/*.{ts,tsx}"],
    languageOptions: { globals: globals.browser },
    plugins: {
      "react-hooks": reactHooks,
      "jsx-a11y": jsxA11y,
    },
    rules: {
      ...reactHooks.configs.flat.recommended.rules,
      ...jsxA11y.flatConfigs.recommended.rules,
    },
  },
  eslintConfigPrettier,
);
```

Backend-only repositories omit the frontend block and browser globals. Mixed repositories narrow its `files` glob to the actual frontend root.

Use type-aware config only for files included by the relevant tsconfig. Give config files, generated files, and other JavaScript an explicit non-type-checked override rather than broad parser exceptions.

Prettier owns layout. Disable conflicting stylistic rules through the established Prettier compatibility config and run `prettier --check .` separately. ESLint fix does not replace formatting proof.

### Required Correctness Rules

Keep the strict presets and explicitly protect these high-value rules:

- `@typescript-eslint/no-floating-promises`
- `@typescript-eslint/no-misused-promises`
- `@typescript-eslint/switch-exhaustiveness-check`
- unsafe assignment, argument, call, member access, and return rules
- unnecessary condition and promise checks
- consistent type imports under `verbatimModuleSyntax`
- no explicit `any` and no non-null assertions

Configure void-return callback checks so event handlers cannot silently lose rejected promises. A boundary wrapper must observe failure and transfer ownership.

### Imports And Architecture

Enforce import cycles with an established import-cycle plugin or dedicated analyzer. Also enforce the repository dependency direction: API and DB may import core; core may not import adapters, Fastify, React, Pino, `pg`, environment state, or test utilities.

For NodeNext, lint import extensions and unresolved paths using a resolver proven against the repository tsconfig. For libraries, prevent deep imports outside declared package exports.

### React Configuration

Frontend config includes `eslint-plugin-react-hooks` recommended rules and `eslint-plugin-jsx-a11y` recommended flat rules. Hooks rules are mandatory; do not suppress dependency analysis to force an effect shape.

Treat accessibility findings as merge blockers. A local suppression requires a documented semantic alternative and a test proving accessible behavior.

### Scope And Generated Code

Lint source, tests, scripts, and configuration with the correct environment. Ignore generated output, coverage, dependencies, and build artifacts by explicit patterns. Generated source is checked through its generator and build unless it is intended for hand maintenance.

Do not globally enable browser globals in backend files or Node globals in browser code. Split config by file pattern and project shape.

### Suppression Policy

Fix the design first. When a false positive remains, suppress the exact rule on the smallest line and explain the invariant immediately above it.

`eslint-disable` without a rule, file-wide disable blocks, `@ts-ignore`, warning-only policy, and blanket config downgrades are forbidden. Use `@ts-expect-error` only for intentional negative type proof and include the expected reason.

Every suppression is searchable and reviewed. Time-bound dependency or generated-code exceptions name an owner and removal issue.

### Performance And Editor Parity

Keep a dedicated typecheck command because ESLint is not the compiler. Use project service rather than hand-maintained project globs. Editor and CI load the same committed flat config and local tool versions.

Measure lint performance before splitting configs or dropping type information. Cache may accelerate local runs, but CI proof must not reuse untrusted or incompatible caches.

## Common Mistakes And Forbidden Patterns

- Legacy `.eslintrc` mixed with flat config.
- Syntax-only ESLint presented as type-aware correctness proof.
- Prettier rules recreated manually in ESLint.
- Floating or misused promise rules weakened for callbacks.
- Cycle detection omitted because ESM currently initializes successfully.
- React Hooks or accessibility rules disabled at repository scope.
- `eslint-disable` without an exact rule and local rationale.
- Generated output or `dist` included in routine lint.
- CI accepting warnings that local development accumulates indefinitely.

## Verification And Proof

- `prettier --check .`, `eslint . --max-warnings 0`, and `tsc --noEmit` all pass.
- A known floating promise, misused promise callback, and non-exhaustive switch fail a lint fixture or config test.
- A deliberate import cycle and forbidden core-to-adapter import are rejected.
- Frontend fixtures prove invalid hook usage and inaccessible JSX fail lint.
- Node and browser file patterns receive only their intended globals.
- Search finds no broad disables, `@ts-ignore`, or unexplained suppressions.
- Editor and clean-checkout CI resolve the same local flat configuration.

Related: [../foundations/type-system.md](../foundations/type-system.md), [testing.md](testing.md), and [../operations/ci-and-release.md](../operations/ci-and-release.md). Configuration anchors: [typescript-eslint shared configs](https://typescript-eslint.io/users/configs/), [typed linting with project service](https://typescript-eslint.io/getting-started/typed-linting/), and [jsx-a11y flat config](https://github.com/jsx-eslint/eslint-plugin-jsx-a11y#shareable-configs).
