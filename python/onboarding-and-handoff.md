# Onboarding And Handoff

> **Team-process document.** This governs project ownership transfer between people. It is not part of the app-generation contract; agents building or changing code do not read it.

This guide is for taking over a Python repository built from this handbook. It defines the day-one reading path, the questions a new owner must answer, and the outgoing owner's responsibilities.

This is not the handbook's own Start Here in [README.md](README.md). That section is about using the handbook. This guide is about owning a project built with it. A contributor making one change starts from [AGENTS.md](AGENTS.md), not here.

## Who This Is For

- A new primary or on-call owner inheriting a service, worker, CLI, or library.
- The outgoing owner running the transfer.
- A reviewer confirming the handoff is complete before sign-off.

Every referenced artifact should already exist in the project. A missing artifact is a handoff defect, not an optional extra.

## Day-One Reading Path, In Order

Read these in the project repo, not the handbook.

| Step | Read or run | What you must come away knowing |
|---|---|---|
| 1 | Project `README.md` | Repository purpose and shape, local setup, run commands, entry points, and primary owners. |
| 2 | Project `AGENTS.md` | Repo invariants, change routing, Python floor/pin, and exact proof gate. |
| 3 | `pyproject.toml`, `.python-version`, `uv.lock` policy | Import/distribution package names, dependency groups, tools, supported Python floor, and pinned development runtime. |
| 4 | Project `decisions/` ADRs | Load-bearing choices, rejected alternatives, accepted consequences, and open or superseded decisions. |
| 5 | Source package map and Import Linter contracts | Where core, adapters, config, telemetry, and workers live; which dependency directions are forbidden. |
| 6 | Run `make verify` | Confirm lock-check, frozen sync, format-check, lint, imports, types, test, and audit pass from a clean checkout. |
| 7 | Runbook | Deploy and rollback, SLOs, alerts, dashboards, failure modes, and escalation path. |

Step 6 is the gate between reading and owning. If `make verify` does not pass from a clean clone, the environment or repository is not reproducible and the handoff is not done. See [templates/Makefile](templates/Makefile) and [quality/linting.md](quality/linting.md).

## Questions A New Owner Must Be Able To Answer

Treat every “I would have to ask the previous owner” as an open handoff item.

### Build, Test, Package, Deploy

- Which Python versions are supported, and which exact interpreter does local development and CI use?
- How does uv create the environment, how is `uv.lock` updated, and which dependency groups ship at runtime?
- What does each `make verify` stage prove, which tests require Docker, and where are their artifacts?
- How do I run the service, worker, CLI, or library consumer locally through its installed entry point?
- For a library, how are wheel and sdist built, inspected, versioned, and published?
- How does a change reach production, what triggers release, and how do I perform a deploy dry-run and rollback unaided?

### Architecture And Runtime

- Where is the composition root, and which resources does FastAPI lifespan or the worker root task own?
- Which package is core, which `Protocol` ports does it own, and how does `lint-imports` enforce the direction?
- Where are Pydantic boundary models mapped into plain domain values?
- Which tasks run in the background, who owns them, how are they cancelled, and what is the drain deadline?
- Which blocking calls move to threads or executors, and how is their concurrency bounded?

### Configuration And Secrets

- Which pydantic-settings model defines configuration, what is required, and what precedence applies?
- Does `.env.example` match every documented non-secret key and startup validation rule?
- Where does each secret live, who grants access, and what is its rotation and revocation procedure?
- Which fields are redacted from logs, exceptions, traces, metrics, and object representations?

### Data And External Systems

- Which Alembic revision is current, who applies migrations, and how are expand/contract changes rolled out?
- How do real-PostgreSQL integration tests run, and what data cleanup or retention rules apply?
- Which shared HTTPX/gRPC clients exist, what deadlines and retry limits apply, and what SSRF restrictions exist?
- For messages, what are the ordering, idempotency, settlement, retry, and DLQ contracts?

### Reliability And On-Call

- What are the SLOs, and which Prometheus metrics, logs, and OpenTelemetry traces calculate or explain them?
- What do `/livez`, `/readyz`, and `/metrics` prove, and which dependencies affect readiness?
- What alerts fire, what is the first response, and who owns the escalation path?
- What are the known failure modes and first fixes? Use the project runbook plus [maintainer-reference.md](maintainer-reference.md).

### Decisions And Direction

- Why were the datastore, transport, framework exceptions, broker, and deployment model chosen?
- Which ADRs are proposed, accepted but not fully implemented, deprecated, or superseded?
- Who owns dependency updates, `uv lock --upgrade` review, `make audit` findings, and Python-floor changes?

If a question has no documented answer, write it in the project. Undocumented knowledge is the defect this guide prevents.

## Outgoing Owner Responsibilities

Walk [checklists/handoff.md](checklists/handoff.md) item by item with the incoming owner.

- Update `CODEOWNERS` using [templates/codeowners.md](templates/codeowners.md).
- Confirm README, AGENTS, architecture contracts, `.env.example`, ADRs, and runbook match the current system.
- Grant repository, package registry, CI, secret-manager, observability, on-call, deploy, database, and broker access; revoke access that should not survive the transfer.
- Document every secret location and rotation path without copying secret values into handoff artifacts.
- Transfer dashboards, alert routes, SLO/error-budget ownership, package publishing, and dependency-update automation.
- Surface every open decision, known vulnerability, migration risk, deprecation window, flaky test, and operational workaround.
- Pair on `make verify`, local startup, a migration dry-run where applicable, a deploy dry-run, rollback, and one alert/runbook walkthrough.
- For libraries, transfer package-index ownership and prove a built wheel installs with typing metadata intact.

The transfer is complete only when the new owner runs `make verify` and a deploy or release dry-run independently, explains the package boundaries and runtime lifecycle, and answers the day-one questions without the outgoing owner.

## Where To Go Next

- Transfer checklist: [checklists/handoff.md](checklists/handoff.md)
- Decision process: [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md)
- Release and deploy: [operations/ci-and-release.md](operations/ci-and-release.md)
- Secrets and access: [operations/security.md](operations/security.md)
- SLOs and alerts: [operations/observability.md](operations/observability.md), [operations/operability.md](operations/operability.md)
- Canonical proof gate: [templates/Makefile](templates/Makefile)
