# Templates

Copy these files into a new repository, rename encoded destinations, replace every `<PLACEHOLDER>`, remove shape-specific blocks that do not apply, run `npm install` to create the lockfile, then make `npm run verify` green.

## Filename Convention

A filename may encode its destination when the literal name would be awkward in this documentation repository. Remove `.txt` and expand hyphens as shown below. `github-workflows-ci.yml`, for example, becomes `.github/workflows/ci.yml`; `src-main.ts.txt` becomes `src/main.ts`.

## Version Pins

The package template pins direct dependencies verified on 2026-08-09. Treat the manifest and generated lockfile as one reviewed unit. Runtime surfaces use Node `24.18.0`, TypeScript `6.0.3` (the newest compiler supported by typescript-eslint `8.65.0`), Fastify `5.11.3`, fastify-type-provider-zod `6.1.0`, Zod `4.4.3`, Pino `10.3.1`, `pg` `8.22.0`, `node-pg-migrate` `9.0.0`, Jest `30.4.2` with Babel `7.29.7`, Testcontainers PostgreSQL `12.0.4`, ESLint `9.39.5` (the newest line supported by jsx-a11y `6.10.2`), and Prettier `3.9.6`. The frontend reference additionally pins React `19.2.8`, Vite `8.2.1`, `@vitejs/plugin-react` `6.0.5`, React Router `7.18.2`, TanStack Query `5.101.4`, and MSW `2.15.0`. Verify current compatibility through [Node releases](https://nodejs.org/en/about/previous-releases), [npm package metadata](https://www.npmjs.com/), and the governing docs before refreshing a pin.

Use [exampleservice](../reference/exampleservice/) for the complete backend assembly, [exampleworker](../reference/exampleworker/) for the event-worker assembly, and [examplefrontend](../reference/examplefrontend/) for the React/Vite variant. Templates are intentionally shape-neutral where the reference packages need different composition.

## Project Baseline

| Template | Destination | Purpose | Governing doc |
|---|---|---|---|
| [package.json.txt](package.json.txt) | `package.json` | Backend package and canonical scripts. | [Project setup](../foundations/project-setup.md) |
| [tsconfig.json](tsconfig.json) | `tsconfig.json` | Backend NodeNext strict compiler baseline. | [Project setup](../foundations/project-setup.md) |
| [tsconfig.frontend.json](tsconfig.frontend.json) | frontend `tsconfig.json` | React/Vite strict compiler baseline. | [Project setup](../foundations/project-setup.md) |
| [eslint.config.mjs](eslint.config.mjs) | `eslint.config.mjs` | Typed backend and frontend lint policy. | [Linting](../quality/linting.md) |
| [.prettierrc.json](.prettierrc.json) | `.prettierrc.json` | Formatting policy. | [Linting](../quality/linting.md) |
| [.editorconfig](.editorconfig) | `.editorconfig` | Editor whitespace baseline. | [Style](../foundations/style-and-review.md) |
| [gitignore](gitignore) | `.gitignore` | Generated and local-state exclusions. | [Project setup](../foundations/project-setup.md) |
| [.nvmrc](.nvmrc) | `.nvmrc` | Exact Node runtime. | [Project setup](../foundations/project-setup.md) |
| [.dockerignore](.dockerignore) | `.dockerignore` | Image build context exclusions. | [Deployment](../operations/deployment.md) |
| [.env.example](.env.example) | `.env.example` | Safe configuration names. | [Configuration](../foundations/configuration.md) |
| [babel.config.cjs](babel.config.cjs) | `babel.config.cjs` | Jest-only TypeScript-to-CJS transform. | [Testing](../quality/testing.md) |
| [jest.config.cjs](jest.config.cjs) | `jest.config.cjs` | Jest 30 backend configuration. | [Testing](../quality/testing.md) |
| [src-main.ts.txt](src-main.ts.txt) | `src/main.ts` | Thin lifecycle entrypoint. | [Async and cancellation](../foundations/async-and-cancellation.md) |
| [Makefile](Makefile) | `Makefile` | npm script shims. | [CI and release](../operations/ci-and-release.md) |

## Delivery And Team Files

| Template | Destination | Governing doc |
|---|---|---|
| [Dockerfile](Dockerfile) | `Dockerfile` | [Deployment](../operations/deployment.md) |
| [docker-compose.yml](docker-compose.yml) | `docker-compose.yml` | [Database](../services/database.md) |
| [github-workflows-ci.yml](github-workflows-ci.yml) | `.github/workflows/ci.yml` | [CI and release](../operations/ci-and-release.md) |
| [github-workflows-release.yml](github-workflows-release.yml) | `.github/workflows/release.yml` | [CI and release](../operations/ci-and-release.md) |
| [dependabot.yml](dependabot.yml) | `.github/dependabot.yml` | [CI and release](../operations/ci-and-release.md) |
| [k8s-deployment.yaml](k8s-deployment.yaml) | `deploy/k8s/deployment.yaml` | [Deployment](../operations/deployment.md) |
| [adr-template.md](adr-template.md) | `decisions/NNNN-<SHORT_TITLE>.md` | [ADRs](../decisions/architecture-decision-records.md) |
| [changelog.md](changelog.md) | `CHANGELOG.md` | [CI and release](../operations/ci-and-release.md) |
| [codeowners.md](codeowners.md) | `.github/CODEOWNERS` | [Handoff](../onboarding-and-handoff.md) |
| [project-readme.md](project-readme.md) | `README.md` | [Project setup](../foundations/project-setup.md) |
| [project-agents.md](project-agents.md) | `AGENTS.md` | [Handbook contract](../AGENTS.md) |
| [project-contributing.md](project-contributing.md) | `CONTRIBUTING.md` | [Git workflow](../foundations/git-workflow.md) |
| [pull_request_template.md](pull_request_template.md) | `.github/pull_request_template.md` | [PR review](../checklists/pr-review.md) |
| [runbook.md](runbook.md) | `docs/runbook.md` | [Operability](../operations/operability.md) |
| [security-policy.md](security-policy.md) | `SECURITY.md` | [Security](../operations/security.md) |

These are fill-in skeletons, not generated policy. Project decisions and accepted ADRs override template placeholders.
