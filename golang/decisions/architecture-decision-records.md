# Architecture Decision Records

How handbook-built repos capture the non-obvious decisions a future owner must understand to safely change the system.

## Default Approach

An Architecture Decision Record (ADR) is a short, immutable document that states one decision, the context that forced it, and the consequences accepted. It answers "why is it this way?" so a new owner does not have to reverse-engineer intent from code, commit history, or absent people. ADRs are not design proposals, status reports, or living docs: once accepted, an ADR is frozen and only superseded, never edited.

### When An ADR Is Required

Write an ADR before merging any decision that is non-obvious or hard to reverse. At minimum:

- **Datastore** choice (which engine, and why not the others).
- **Queue or broker** choice, including the delivery, ordering, and retry semantics relied on.
- **Transport** choice (HTTP vs gRPC vs messaging for a given boundary).
- A **major dependency or framework exception** — any escalation past a [framework-selection.md](framework-selection.md) default.
- A **service boundary** — splitting or merging a service, or where a domain edge falls.
- An **auth or tenancy model** — how identity, authorization, and tenant isolation work.
- Any **deviation from a handbook default** (layout, error model, logging, config posture, dependency posture).

If the decision is obvious, trivially reversible, and local to one package, skip the ADR — the code and its tests are the record. When in doubt, the test is: would a competent new owner be surprised by this, and is undoing it expensive? If yes, write the ADR.

### Where They Live And How They Are Numbered

ADRs live **in the project repo**, not in this handbook, under `decisions/`:

```
decisions/0001-use-postgres-for-primary-store.md
decisions/0002-grpc-for-internal-service-calls.md
decisions/0003-chi-router-exception.md
```

- One file per decision: `decisions/NNNN-kebab-title.md`.
- `NNNN` is a zero-padded, monotonically increasing integer. Never reuse or renumber.
- Use [../templates/adr-template.md](../templates/adr-template.md) as the skeleton for every record.
- The title line and the filename agree: `# 0003. chi router exception`.

The first record is effectively the project's stack: the [framework-selection.md](framework-selection.md) defaults this repo adopted and any exception it took are ADR-worthy and belong in `decisions/0001-...`. Recording the baseline makes every later deviation legible.

### Status Lifecycle

Each ADR carries exactly one status:

- **Proposed** — under review, not yet binding. Used only while the PR is open.
- **Accepted** — merged and binding. This is the steady state.
- **Superseded** — replaced by a later ADR. Set the status to `Superseded by 0007` and leave the body untouched.
- **Deprecated** — no longer applies and nothing replaced it (the capability was removed).

A decision changes by **adding a new ADR**, never by editing the old one. The superseding ADR links back in its `Supersedes` field (`Supersedes: 0002`), and the old ADR's status is flipped to `Superseded by NNNN` with a link forward. The chain of links is the audit trail: a reader can walk from current state back through every prior decision and see exactly what changed and why.

### How ADRs Deliver Handoff

The handbook's bar is that a new owner inherits a project with zero open questions. Code shows *what* the system does; ADRs are the only durable record of *why*. A handoff is complete when `decisions/` answers, without a meeting:

- why each irreversible technology was chosen, and what was rejected;
- which handbook defaults were deviated from, and the tradeoff accepted;
- what would trigger re-evaluating or removing each choice.

Link the live `decisions/` set from the project README and reference it from the handoff checklist so review confirms it exists and is current. See [../onboarding-and-handoff.md](../onboarding-and-handoff.md) and [../checklists/handoff.md](../checklists/handoff.md).

## Common Mistakes And Forbidden Patterns

- Editing an accepted ADR to reflect a new decision instead of writing a superseding record — this destroys the audit trail.
- Renumbering, reusing, or deleting ADR files; gaps and historical entries are expected and correct.
- Writing ADRs as design proposals or aspirational plans rather than decisions actually made.
- Recording only the chosen option with no `Alternatives Considered` — the rejected options are half the value.
- Deferring the record until "later"; an ADR written after the context is forgotten is fiction.
- Skipping the ADR for the baseline stack, leaving later deviations with nothing to deviate from.
- ADRs that restate the code instead of the reasoning, ordering, lock-in, or operational tradeoff behind it.
- Burying ADRs outside `decisions/` (wiki, tickets, chat) where they rot and are invisible at handoff.

## Verification And Proof

```bash
ls decisions/
```

A repo's ADR practice is in good shape when:

- every required decision above has a corresponding `decisions/NNNN-*.md`;
- numbering is contiguous and unique, and no accepted ADR has been edited post-merge;
- every `Superseded` ADR links forward and its replacement links back;
- the project README links the `decisions/` directory and the handoff checklist confirms it.

ADRs are done when a new owner can read `decisions/` and answer every "why is it this way?" without asking a person.

## Where To Go Next

- [framework-selection.md](framework-selection.md) — the rules that decide when a dependency exception (and thus an ADR) is warranted.
- [../templates/adr-template.md](../templates/adr-template.md) — the fill-in skeleton for new records.
- [../onboarding-and-handoff.md](../onboarding-and-handoff.md) and [../checklists/handoff.md](../checklists/handoff.md) — where ADRs are consumed.
