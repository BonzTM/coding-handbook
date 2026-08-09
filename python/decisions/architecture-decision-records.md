# Architecture Decision Records

How handbook-built repos capture non-obvious decisions a future owner must understand to safely change the system.

## Default Approach

An Architecture Decision Record (ADR) is a short immutable document stating one decision, the context that forced it, alternatives considered, and consequences accepted. It answers “why is it this way?” without relying on code archaeology or absent people. Accepted ADRs are not living docs: supersede them; do not rewrite them.

### When An ADR Is Required

Write an ADR before merging a decision that is non-obvious or hard to reverse. At minimum:

- **Datastore** engine or persistence-model choice.
- **Queue or broker** choice and relied-on delivery, ordering, retry, and DLQ semantics.
- **Transport** choice: HTTP, gRPC, or messaging for a boundary.
- **Major dependency or framework exception** beyond [framework-selection.md](framework-selection.md): Litestar/Starlette in place of FastAPI, Tenacity, structlog, Typer/Click, a distributed cache, or a framework-heavy messaging stack.
- **Build-backend exception** from `uv_build`, native extension introduction, or unusual packaging/publishing model.
- **Python-floor change** or compatibility-policy change that enables newer language/stdlib behavior or drops supported consumers.
- **Boundary exception** to the `core` import rule or a split/merge of service/domain ownership.
- **Auth, tenancy, secrets, or data-classification model**.
- **Runtime model** that changes asyncio ownership, scheduling, process topology, or migration execution.
- Any **deviation from a handbook invariant**: layout, validation boundary, error model, logging, typing, config, testing, or verification posture.

Skip an ADR only when the choice is obvious, local, cheap to reverse, and fully expressed by code plus tests. Ask: would a competent new owner be surprised, and would reversal cross packages, data, contracts, or operations? If yes, write the ADR.

### Where They Live And How They Are Numbered

ADRs live in the project repo under `decisions/`:

```text
decisions/0001-adopt-handbook-python-stack.md
decisions/0002-use-postgres-for-primary-store.md
decisions/0003-use-litestar-for-mandated-platform-integration.md
```

- One decision per `decisions/NNNN-kebab-title.md`.
- `NNNN` is zero-padded and monotonically increasing. Never reuse or renumber it.
- Copy [../templates/adr-template.md](../templates/adr-template.md).
- Filename and title agree: `# 0003. use Litestar for mandated platform integration`.
- The first ADR records the baseline stack from [framework-selection.md](framework-selection.md), including accepted exceptions.

### Required Content

Every ADR records:

- status, date, owners, and decision scope;
- context and constraints, including requester mandates;
- decision in concrete terms, including which package layers may depend on it;
- alternatives considered and why each lost;
- positive and negative consequences across security, reliability, operations, compatibility, testing, packaging, and maintenance where relevant;
- migration, rollout, rollback, and removal plan when state or public contracts change;
- proof required before the decision is considered implemented;
- re-evaluation triggers and any superseded ADR.

Do not use an ADR to compensate for an absent acceptance criterion. Resolve WHAT through [../checklists/spec-intake.md](../checklists/spec-intake.md); use the ADR to record HOW and WHY.

### Status Lifecycle

Each ADR carries exactly one status:

- **Proposed** — under review, not binding.
- **Accepted** — merged and binding.
- **Superseded** — replaced by a later ADR; state `Superseded by NNNN` and link forward.
- **Deprecated** — no longer applies because the capability was removed.

A decision changes through a new ADR. The replacement states `Supersedes: NNNN`; the old record changes only its status and forward link. The linked chain is the audit trail.

### How ADRs Deliver Handoff

Code shows what the system does; ADRs preserve why. A handoff is complete when `decisions/` answers without a meeting:

- why each irreversible technology, runtime, and data choice won;
- which handbook defaults changed and what cost was accepted;
- which package layers may use an exception;
- what proof implemented it and what would trigger reversal or removal.

Link the live decision set from the project README and consume it from [../onboarding-and-handoff.md](../onboarding-and-handoff.md) and [../checklists/handoff.md](../checklists/handoff.md).

## Common Mistakes And Forbidden Patterns

- Editing an accepted ADR to hide a new decision instead of superseding it.
- Renumbering, deleting, or reusing records; gaps and history are correct.
- Writing an aspirational design proposal rather than a decision with owners and implementation proof.
- Omitting rejected alternatives, negative consequences, or re-evaluation triggers.
- Recording only the package name without permitted layer, runtime lifecycle, lockfile, or operational impact.
- Deferring the record until context is forgotten.
- Skipping the baseline-stack ADR, leaving exceptions with no baseline.
- Burying decisions in tickets, wikis, chat, docstrings, or `pyproject.toml` comments.
- Using an ADR to approve unbounded concurrency, swallowed cancellation, auto-migrations on startup, secret exposure, or a weakened verify gate.

## Verification And Proof

```bash
ls decisions/
```

ADR practice is healthy when:

- every required decision has one unique `decisions/NNNN-*.md`;
- numbering is monotonic and accepted bodies are unchanged after merge;
- every superseded record links forward and replacement links back;
- accepted decisions name implementation proof and that proof is green;
- the project README links the directory and handoff review confirms it is current.

ADRs are done when a new owner can answer every “why is it this way?” and identify the reversal cost without asking a person.

## Where To Go Next

- [framework-selection.md](framework-selection.md) — defaults and dependency exceptions requiring ADRs.
- [../templates/adr-template.md](../templates/adr-template.md) — fill-in skeleton.
- [../onboarding-and-handoff.md](../onboarding-and-handoff.md), [../checklists/handoff.md](../checklists/handoff.md) — where records are consumed.
