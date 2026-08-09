# AGENTS.md - <PROJECT_NAME> Contract

Fast-path contract for this repository. The TypeScript handbook at `<HANDBOOK_PATH_OR_URL>` governs where this file is silent.

## Purpose

- Shape: <SERVICE_WORKER_CLI_LIBRARY_OR_FRONTEND>.
- Purpose: <ONE_SENTENCE_PURPOSE>.
- Owners: <OWNERS>.

## Invariants

- Node 24, npm, one package, committed lockfile, ESM only.
- Strict TypeScript; Zod parses every trust boundary.
- Backend dependencies preserve `src/api -> src/core <- src/db`.
- Async work has owner, timeout, cancellation, bounds, and observed failure.
- Pino JSON is redacted at construction and emitted once at the acting boundary.
- PostgreSQL uses `pg`, parameterized SQL, parsed rows, and explicit `node-pg-migrate` jobs.
- React uses function components, semantic HTML, and TanStack Query for server state.
- `npm run verify` is mandatory before merge.
- Project-specific: <PROJECT_INVARIANTS_OR_NONE>.

## Change Routing

| Change | Start in | Also update |
|---|---|---|
| domain behavior | `src/core/` | tests and owning adapters |
| HTTP | `src/api/` | schemas, core, contract, telemetry |
| database | `src/db/`, `migrations/` | core port, rollout, integration tests |
| config | `src/config/` | `.env.example`, deployment, runbook |
| frontend | `src/features/`, `src/routes/` | accessibility and RTL tests |
| delivery | `.github/`, `Dockerfile`, `deploy/` | changelog and rollback |
| <PROJECT_AREA> | `<PATH>` | <SYNC_SURFACES> |

## Verification

Run the narrow test first, then:

```bash
npm run verify
```

Report exact failures. Never claim a red gate is green.
