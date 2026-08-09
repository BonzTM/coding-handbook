# Onboarding And Handoff

> **Team-process document.** This governs ownership transfer for projects built from this handbook, not ordinary application changes.

## Who This Is For

- A new repository, deploy, or on-call owner.
- The outgoing owner responsible for a complete transfer.
- A reviewer deciding whether ownership can move safely.

Missing project documentation, access, proof, or operational knowledge is a handoff defect. Use [checklists/handoff.md](checklists/handoff.md) for sign-off.

## Day-One Reading Path

| Step | Project artifact | Required understanding |
|---|---|---|
| 1 | `README.md` | purpose, shape, setup, entrypoints, configuration |
| 2 | `AGENTS.md` | invariants, routing, verification contract |
| 3 | `decisions/` | accepted, superseded, and open consequential choices |
| 4 | source boundaries | `api -> core <- db`, frontend features, composition, lifecycle |
| 5 | `npm ci && npm run verify` | clean Node 24 environment and executable proof gate |
| 6 | built artifact and runbook | production runtime, deploy, SLOs, alerts, mitigation, rollback |

The new owner does not own the project until the clean gate and production-shaped smoke test work on their machine without the outgoing owner acting.

## Questions The New Owner Must Answer

### Build And Delivery

- What does the repository build, and which emitted ESM, Vite, image, or packed artifact is released?
- Which commands select one test, run integration, and execute the canonical gate?
- What commit, tag, digest, lockfile, and provenance identify the running release?
- How are migrations ordered, how does rollout stop, and how is the prior artifact restored?

### Architecture And Contracts

- Which module owns each domain decision, adapter, schema, port, and lifecycle resource?
- Which HTTP, event, persistence, configuration, UI, and package contracts are independently consumed?
- Which compatibility windows, deprecations, feature flags, and proposed ADRs remain open?
- Where are cancellation, timeout, idempotency, retry, concurrency, and overload policies owned?

### Configuration, Security, And Data

- Which configuration keys exist, where are they injected, and which are public frontend values?
- Where do secrets live, who grants access, and how is each secret rotated?
- What data classifications, tenant boundaries, retention, deletion, audit, backup, and restore obligations apply?
- How are vulnerabilities reported privately and dependency advisories triaged?

### Reliability And On-Call

- Which user-visible SLIs and SLOs govern the service, and who owns the error budget?
- What pages, what does each mean, and what safe first action does its runbook prescribe?
- What are the capacity limits and known failure modes for database, queue, external dependency, event loop, and memory?
- Can the owner deploy, diagnose, mitigate, replay, rollback, and verify recovery unaided?

## Outgoing Owner Responsibilities

- Bring README, AGENTS, ADRs, changelog, `.env.example`, schemas, runbook, SLOs, dashboards, and alerts current.
- Transfer CODEOWNERS, repository, CI, registry, deploy, database, secret, telemetry, and on-call access.
- Exercise the clean verification gate, built-artifact smoke, deploy dry-run, rollback walkthrough, and secret rotation with the incoming owner.
- Surface open incidents, risks, flags, deprecations, migrations, dependency exceptions, and tribal knowledge with owners and dates.
- Revoke access that should not survive transfer and record the effective ownership time.

## Completion

Transfer completes only when the incoming owner can run `npm run verify`, identify and operate the production artifact, answer the questions above, respond to an alert through the runbook, and perform a deploy/rollback dry-run without assistance.

Related: [checklists/handoff.md](checklists/handoff.md), [operations/operability.md](operations/operability.md), [operations/ci-and-release.md](operations/ci-and-release.md), and [operations/security.md](operations/security.md).
