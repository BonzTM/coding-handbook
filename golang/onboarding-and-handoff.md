# Onboarding And Handoff

This guide is for taking over a Go repository that was built from this handbook. It defines what a new owner reads on day one, the questions they must be able to answer before they truly own the repo, and what the outgoing owner is responsible for.

This is not the handbook's own "Start Here" in [README.md](README.md). That section is about working *inside the handbook*. This guide is about owning a *project that was built with it*. If you are an agent or human contributing a single change, you want [AGENTS.md](AGENTS.md) and [maintainer-map.md](maintainer-map.md), not this file.

## Who This Is For

- A new primary owner or on-call owner inheriting a service, worker, CLI, or library.
- The outgoing owner running the transfer.
- A reviewer confirming a handoff is actually complete before sign-off.

If the repo follows the handbook, every artifact this guide references already exists in the project. If one is missing, that is a handoff defect, not an optional extra; surface it before accepting ownership.

## Day-One Reading Path, In Order

Read these in the project repo, not in the handbook. Do not skip ahead; each step assumes the previous one.

| Step | Read | What you must come away knowing |
|---|---|---|
| 1 | Project `README.md` | What the repo is, its shape (service/worker/CLI/library), how to build and run it locally, where the entrypoints are. |
| 2 | Project `AGENTS.md` | The repo's invariants, the task loop, and the exact baseline proof commands contributors must pass. |
| 3 | `decisions/` ADRs | Why the load-bearing choices were made, what was rejected, and which decisions are still open. Read newest and any `Proposed`/`Accepted`-but-unimplemented records first. |
| 4 | Project maintainer map / `maintainer-map.md` equivalent | How a given change routes to files, sync surfaces, and proof steps. |
| 5 | Run `make verify` locally | Confirm the safety gate passes on your machine from a clean checkout before you change anything. |
| 6 | The runbook | How the thing is deployed, what its SLOs and alerts are, what to do at 3 a.m., and who to escalate to. |

Step 5 is the gate between reading and owning. If `make verify` does not pass from a clean clone on your machine, you do not yet have a working environment and the handoff is not done. See [templates/Makefile](templates/Makefile) for the canonical target set and [quality/linting.md](quality/linting.md) for what `lint` enforces.

## Questions A New Owner Must Be Able To Answer

You do not own the repo until you can answer every one of these unaided. Treat any "I'd have to ask the previous owner" as an open handoff item.

### Build, Test, Deploy

- How do I build and run this locally? (Answer comes from the README plus `make build`.)
- What is the full proof gate, and can I run it? (`make verify`; see [operations/ci-and-release.md](operations/ci-and-release.md).)
- How does a change reach production: what triggers a deploy, how is a release tagged (`v1.2.3`), and how do I roll back?
- Can I perform a deploy dry-run without the previous owner present?

### Configuration And Secrets

- Where does configuration come from, and what are the required keys? (`internal/config`, env plus flags, validated at startup; see [foundations/configuration.md](foundations/configuration.md).)
- Where do secrets live, who grants access, and what is the rotation procedure for each one? (See [operations/security.md](operations/security.md).)
- What is in `.env.example`, and does it match the current config loader?

### Reliability And On-Call

- What are the SLOs, and which dashboards show them? (See [operations/observability.md](operations/observability.md).)
- What alerts fire, what do they mean, and what is the first response for each? (Runbook.)
- Who is on the escalation path, and who do I page when I am stuck?
- What are the known failure modes and their first fixes? (Project runbook plus [maintainer-reference.md](maintainer-reference.md) failure-mode table for patterns.)

### Decisions And Direction

- Why are the load-bearing architectural choices the way they are? (`decisions/` ADRs; see [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md).)
- Which decisions are still open or proposed, and what is blocking them?
- What dependency-update obligations exist (security patches, `make vuln`, version bumps), and who owns them now?

If a question has no documented answer, the fix is to write it down in the project, not to keep it in your head. Undocumented knowledge is the failure this guide exists to prevent.

## Outgoing Owner Responsibilities

The outgoing owner runs the handoff and is responsible for it being complete. Walk [checklists/handoff.md](checklists/handoff.md) item by item with the incoming owner; do not delegate it to the newcomer to discover.

- Update `CODEOWNERS` so reviews and notifications route to the new owner; see [templates/codeowners.md](templates/codeowners.md).
- Document where every secret lives and how it rotates, then grant the new owner access and revoke your own where it is no longer warranted.
- Transfer the on-call rotation and the escalation path, and confirm the new owner is reachable through it.
- Grant deploy and production access to the new owner and revoke access that should not outlive your ownership.
- Surface every open or proposed ADR and every undocumented decision; convert tribal knowledge into an ADR or runbook entry before you leave.
- Confirm the runbook is current: deploy steps, rollback, SLOs, alerts, and dashboards all match reality today, not how things worked a year ago.
- Transfer ownership of dashboards and alert routes so pages reach the right person.
- Hand off the dependency-update cadence, including who runs `make vuln` and acts on findings.

The transfer is complete only when the new owner can run `make verify` and a deploy dry-run independently and can answer the day-one questions without you. Verify that against [checklists/handoff.md](checklists/handoff.md) before signing off.

## Where To Go Next

- The transfer checklist: [checklists/handoff.md](checklists/handoff.md)
- Why decisions are recorded and how: [decisions/architecture-decision-records.md](decisions/architecture-decision-records.md)
- Release and deploy mechanics: [operations/ci-and-release.md](operations/ci-and-release.md)
- Secrets and access posture: [operations/security.md](operations/security.md)
- SLOs, dashboards, and alerts: [operations/observability.md](operations/observability.md)
- The canonical proof gate: [templates/Makefile](templates/Makefile)
