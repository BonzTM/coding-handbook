# Architecture Decision Records

Use ADRs to make consequential TypeScript architecture choices explicit, reviewable, and durable.

## Default Approach

Write an ADR before implementation when a decision is expensive to reverse, changes a handbook invariant, or creates a new long-lived dependency boundary.

### When An ADR Is Required

- adopting or replacing a framework, ORM, broker abstraction, package manager, runtime, or build system
- introducing npm workspaces, multiple deployables, shared packages, code generation, or a new public API
- changing ESM, Node, TypeScript strictness, validation, persistence, testing, or verification defaults
- accepting a security exception, irreversible migration, compatibility break, or unusual operational dependency
- choosing between materially different data consistency, delivery, caching, or rollout models

Routine implementation within an accepted design does not need an ADR. A PR description is enough for a small, reversible library addition that follows [framework-selection.md](framework-selection.md).

### Where They Live And How They Are Numbered

Project ADRs live in `decisions/` at the repository root, not in this handbook directory. Name them `NNNN-short-kebab-title.md`, beginning with `0001`. Never reuse a number.

Each record contains:

- title, date, owners, and status
- context and forces, including security and operational constraints
- the decision in testable language
- alternatives considered and why they lost
- consequences, migration, rollback, and follow-up work
- links to affected contracts, evidence, and superseded ADRs

Use [templates/adr-template.md](../templates/adr-template.md). Keep the decision self-contained; links support the record but do not replace its reasoning.

### Status Lifecycle

Allowed statuses are `Proposed`, `Accepted`, `Rejected`, `Deprecated`, and `Superseded by ADR-NNNN`.

- Draft as `Proposed` before implementation.
- Change to `Accepted` only after the decision is approved.
- Mark an abandoned proposal `Rejected`; do not delete it.
- Mark a still-valid but discouraged decision `Deprecated` with a migration path.
- Supersede through a new ADR and cross-link both records.

### Evidence And Scope

State current constraints without pretending forecasts are facts. Pin exact package versions in lockfiles and templates, not in ADR prose unless the exact version is the decision.

Include the affected surfaces: runtime, source layout, API compatibility, data migration, privacy, observability, performance, deployment, and rollback. Mark irrelevant concerns explicitly rather than omitting a material one accidentally.

### How ADRs Deliver Handoff

An ADR lets a future maintainer distinguish intentional constraints from accidents. Record the trigger that would justify revisiting the decision, the metrics that would demonstrate it, and the owner of deferred work.

## Common Mistakes And Forbidden Patterns

- Writing the ADR after the implementation to rationalize an already-made choice.
- Listing only the winner and omitting credible alternatives.
- Using “industry standard” or popularity as evidence without project-specific forces.
- Recording package syntax that belongs in a recipe instead of a durable decision.
- Editing an accepted ADR until history says something different.
- Omitting migration and rollback for schema, runtime, build, or package-manager changes.
- Creating an ADR for every local refactor and burying consequential decisions in noise.

## Verification And Proof

- The ADR exists and is accepted before the dependent implementation merges.
- Every claimed constraint has repository evidence or an authoritative source.
- Alternatives and consequences are specific enough for a reviewer to challenge.
- The change routing table, README defaults, topical docs, templates, and reference examples agree.
- Superseded records link in both directions and no accepted records contradict one another.
- Follow-up work has owners and completion criteria.

## Where To Go Next

Use [framework-selection.md](framework-selection.md) for dependency choices and [../foundations/git-workflow.md](../foundations/git-workflow.md) for the PR that carries the decision.
