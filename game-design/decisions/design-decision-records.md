# Design Decision Records

How game projects capture the hard-to-reverse design decisions a future owner must understand to safely change the game — and the only mechanism through which a handbook invariant may be weakened.

## Default Approach

A Design Decision Record (DDR) is a short, immutable document that states one design decision, the context that forced it, and the consequences accepted. It is the design-side counterpart of an ADR: where an ADR records *how* the system is built, a DDR records *what* the game is and why. It answers "why is the game this way?" so a new design owner does not have to reverse-engineer intent from the build, the balance sheet, or absent people. DDRs are not pitches, proposals, or living docs: once accepted, a DDR is frozen and only superseded, never edited. Every record uses [../templates/ddr-template.md](../templates/ddr-template.md).

### When A DDR Is Required

Write the DDR before the change lands — before design review sign-off, not after the build ships. At minimum:

- A **pillar set change** — adding, removing, or reweighting a pillar, or shipping a feature that violates one. Pillars gate scope, so changing them rebinds every downstream decision ([../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md)).
- A **core loop change** at any tier — moment-to-moment, session, or meta ([../foundations/core-loops.md](../foundations/core-loops.md)).
- A **framework or model escalation** past the defaults in [frameworks-and-models.md](frameworks-and-models.md) — adopting DDE framing over MDA, a different motivation model, an alternative economy notation.
- An **economy structure change** — adding or removing a currency, rewiring faucet/sink topology, or changing the monetization model ([../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md)).
- Adopting **hidden dynamic difficulty adjustment** — contested practice, forbidden by default ([../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md)).
- An **accessibility gap below the basic tier** of `gameaccessibilityguidelines.com` — the gap is recorded, not silent ([../disciplines/accessibility.md](../disciplines/accessibility.md)).
- A **scope cut of a pillar-serving system** — every cut-list entry that removes something a pillar depends on ([../operations/scoping-and-production.md](../operations/scoping-and-production.md)).
- An **audience, platform, or genre pivot** — anything that invalidates the motivation profile playtest recruiting is built on ([../foundations/player-psychology.md](../foundations/player-psychology.md)).
- Any **deviation from a handbook default** — projects design with the repo defaults unless a DDR documents the exception ([../AGENTS.md](../AGENTS.md) (## Fast Path)).

If the decision is a tuning value — a drop rate, a damage number, a timer — skip the DDR: the balance spreadsheet and playtest reports are the record, and DDR churn over reversible numbers drowns the signal. The test is the same as for ADRs: would a competent new owner be surprised by this, and is undoing it expensive? If yes, write the DDR.

A DDR must record the evidence behind it, not taste alone: the playtest report, prototype outcome, or telemetry finding that forced the decision ([../quality/playtesting.md](../quality/playtesting.md), [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)). A lens answer is a hypothesis, never proof on its own ([../glossary.md](../glossary.md)). Where evidence was genuinely unobtainable — a pre-prototype bet — say so explicitly in the record, and name what future evidence would trigger re-evaluation.

### DDR Lifecycle

Each DDR carries exactly one status:

- **Proposed** — under review, not yet binding. Used only while the review is open.
- **Accepted** — reviewed and binding. This is the steady state.
- **Superseded** — replaced by a later DDR. Set the status to `Superseded by 0007` and leave the body untouched.
- **Deprecated** — no longer applies and nothing replaced it (the system was cut without a successor).

A decision changes by **adding a new DDR**, never by editing the old one. The superseding record links back in its `Supersedes` field (`Supersedes: 0002`), and the old record's status is flipped forward to `Superseded by NNNN`. The chain of links is the audit trail: a reader can walk from the current design back through every prior decision and see exactly what changed, on what evidence, and why.

### Storage And Indexing

DDRs live **in the project repo**, not in this handbook, under `design/decisions/`:

```
design/decisions/0001-pillars-and-framework-baseline.md
design/decisions/0002-single-soft-currency.md
design/decisions/0003-cut-coop-mode.md
```

- One file per decision: `design/decisions/NNNN-kebab-title.md`.
- `NNNN` is a zero-padded, monotonically increasing integer. Never reuse or renumber; gaps and historical entries are expected and correct.
- The title line and the filename agree: `# 0003. cut coop mode`.
- Projects that also follow an engineering handbook keep code ADRs separately in `decisions/` per that handbook (e.g. [../../golang/decisions/architecture-decision-records.md](../../golang/decisions/architecture-decision-records.md)). A decision spanning both — an engine choice that constrains the design, a design pivot that forces a rebuild — gets one record on each side, cross-linked.

The first record is the project's design baseline: the pillar set, the target experience, and the framework choices adopted from [frameworks-and-models.md](frameworks-and-models.md). Recording the baseline makes every later deviation legible. Link the live `design/decisions/` set from the one-page GDD and the project README; open DDRs surface in [../checklists/design-review.md](../checklists/design-review.md) and at handoff ([../onboarding-and-handoff.md](../onboarding-and-handoff.md)).

### Weakening A Handbook Invariant

The repo-wide invariants in [../AGENTS.md](../AGENTS.md) and the Non-Negotiables in [../README.md](../README.md) change through exactly one path: an accepted DDR recording the alternatives considered and the consequences accepted. There is no second path — not a topic-doc edit, not a review comment, not precedent from a project that quietly deviated. Once the DDR is accepted, the fast path, slow path, recipes, and templates move together in the same change, per [../CONTRIBUTING.md](../CONTRIBUTING.md) (### Changing An Invariant Requires A DDR). A project-level exception to a default does not edit the handbook at all: the project records its own DDR and the handbook default stands for everyone else.

## Common Mistakes And Forbidden Patterns

- Editing an accepted DDR to reflect a new decision instead of writing a superseding record — this destroys the audit trail.
- Renumbering, reusing, or deleting DDR files.
- Writing DDRs as pitches or aspirational plans rather than decisions actually made.
- Recording only the chosen option with no `Alternatives Considered` — the rejected options are half the value.
- Justifying a DDR by taste when playtest or telemetry evidence was obtainable, or omitting what future evidence would trigger re-evaluation.
- Deferring the record until "later"; a DDR written after the context is forgotten is fiction.
- Using a DDR to launder scope creep — a DDR records a decision; it does not exempt a feature from pillar gating.
- Filing tuning-value changes as DDRs, burying the hard-to-reverse decisions in noise.
- Weakening a handbook invariant in a passing doc edit, review comment, or project deviation instead of through an accepted DDR.
- Burying DDRs outside `design/decisions/` (wiki, tickets, chat) where they rot and are invisible at handoff.

## Verification And Proof

```bash
ls design/decisions/
```

A project's DDR practice is in good shape when:

- every required decision above has a corresponding `design/decisions/NNNN-*.md`, including the `0001` baseline;
- numbering is contiguous and unique, and no accepted DDR has been edited post-acceptance;
- every `Superseded` DDR links forward and its replacement links back;
- every DDR names its evidence or explicitly declares the bet and its re-evaluation trigger;
- the one-page GDD and project README link `design/decisions/`, and [../checklists/design-review.md](../checklists/design-review.md) confirms open DDRs are resolved before build commitment;
- any handbook-invariant change references its accepted DDR in the same change.

DDRs are done when a new design owner can read `design/decisions/` and answer every "why is the game this way?" without asking a person.

## Where To Go Next

- [../templates/ddr-template.md](../templates/ddr-template.md) — the fill-in skeleton for new records.
- [frameworks-and-models.md](frameworks-and-models.md) — the framework defaults whose escalation requires a DDR.
- [../CONTRIBUTING.md](../CONTRIBUTING.md) — the sync surfaces that move together when a DDR changes a handbook invariant.
- [../onboarding-and-handoff.md](../onboarding-and-handoff.md) and [../checklists/design-review.md](../checklists/design-review.md) — where DDRs are consumed.
