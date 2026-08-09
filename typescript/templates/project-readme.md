# <PROJECT_NAME>

<ONE_SENTENCE_PURPOSE>

## Shape And Ownership

- Shape: <HTTP_SERVICE_WORKER_CLI_LIBRARY_OR_REACT_APP>
- Owners: <TEAM_OR_ROTATION>
- Runbook: <RUNBOOK_PATH_OR_NOT_APPLICABLE>
- SLO/dashboard: <LINKS_OR_NOT_APPLICABLE>

## Requirements

- Node.js 24.18.0
- npm and the committed `package-lock.json`
- Docker for PostgreSQL integration tests when applicable

## Setup

```bash
npm ci
cp .env.example .env
npm run verify
```

## Run

```bash
npm run build
npm start
```

## Configuration

| Key | Required | Sensitive | Default | Meaning |
|---|---|---|---|---|
| `<CONFIG_KEY>` | <YES_NO> | <YES_NO> | <SAFE_DEFAULT_OR_NONE> | <MEANING_AND_UNIT> |

## Architecture

`src/api -> src/core <- src/db`; `src/main.ts` and `src/index.ts` own composition and process lifetime. See `AGENTS.md` and `decisions/`.

## Verification

`npm run verify` is the canonical format, lint, type, test, audit, and build gate. Run `npm run test:integration` for real dependency proof.

## Release And Recovery

<RELEASE_TRIGGER_ARTIFACT_IDENTITY_MIGRATION_ORDER_AND_ROLLBACK>
