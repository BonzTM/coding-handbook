# Checklist: New Project

Use this after spec intake to create a TypeScript repository.

## Scaffold

- [ ] Copy applicable artifacts from [../templates/README.md](../templates/README.md) and replace every `<PLACEHOLDER>`.
- [ ] Choose backend, frontend, worker, CLI, or library shape; compare the matching [service](../reference/exampleservice/), [worker](../reference/exampleworker/), or [frontend](../reference/examplefrontend/) exemplar; and remove unused template blocks.
- [ ] `package.json`, `.nvmrc`, CI, and Docker agree on Node 24.
- [ ] One npm package, `type: module`, and one committed `package-lock.json` exist.
- [ ] Backend NodeNext or frontend Bundler tsconfig matches [project setup](../foundations/project-setup.md).
- [ ] Flat ESLint and Prettier match [linting](../quality/linting.md).

## Boundaries

- [ ] Backend layout preserves `src/api -> src/core <- src/db` with composition in `src/index.ts`.
- [ ] React layout uses `src/app`, `src/components`, `src/features`, `src/routes`, and `src/lib` as needed.
- [ ] Zod parses env and each initial trust boundary.
- [ ] Pino redaction, safe fatal handling, and signal-driven shutdown are wired.
- [ ] `.env.example` contains safe placeholders only; `.env` is ignored.

## Proof And Operations

- [ ] Jest 30 Babel transform config matches [testing](../quality/testing.md); production remains ESM.
- [ ] Unit/component tests are offline; real PostgreSQL is an explicit Testcontainers integration job.
- [ ] `/livez`, `/readyz`, ownership, SLO/runbook links, and deploy rollback exist for a service.
- [ ] CI performs `npm ci` then `npm run verify` with least permissions.
- [ ] Docker image runs emitted JavaScript as numeric non-root.
- [ ] CODEOWNERS, CONTRIBUTING, SECURITY, changelog, and PR template are customized.

## Proof

- [ ] `rg '<[A-Z][A-Z0-9_-]*>'` finds no unresolved required placeholder.
- [ ] Clean Node 24 checkout completes `npm ci` without lockfile change.
- [ ] `npm run verify` is green.
- [ ] Built artifact starts, becomes ready, receives `SIGTERM`, drains, and exits within budget.
