<!--
DDR template (MADR-style, adapted for design decisions). Copy to the PROJECT repo as design/decisions/NNNN-kebab-title.md.
NNNN is the next zero-padded, monotonically increasing number. Never reuse or renumber.
Once Accepted, do not edit this record — supersede it with a new DDR.
A DDR without playtest evidence is a hypothesis, not a decision; record it as Proposed until evidence lands.
Process: ../decisions/design-decision-records.md
-->

# NNNN. Short Imperative Title

- **Status:** Proposed | Accepted | Superseded by NNNN | Deprecated
- **Date:** YYYY-MM-DD
- **Pillar(s) affected:** <PILLAR NAME(S)> — see the project's pillar doc, per [design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md)

## Context

The design forces in play: the player problem, the pillar tension, the constraint (scope, platform, rating, team), and the assumptions that make this decision necessary and non-obvious. State what is hard to reverse — economy structures, control schemes, and progression shapes calcify once content is built on them. Keep it factual — no solution yet.

## Decision

The decision, stated in one clear sentence, then any detail needed to act on it: the tuned values chosen, the systems it touches, the content it obligates. Write it as a settled choice ("Stamina regenerates only out of combat"), not a proposal.

## Alternatives Considered

- **Option A** — what it was, whether it was prototyped or only argued, and why it was rejected.
- **Option B** — same. "Rejected on discussion alone" is a legal but weaker entry; say so explicitly.

## Consequences

What becomes true once this is in effect.

### Good

- Positive player-experience outcome enabled by this decision, tied to the pillar it serves.

### Bad

- Cost accepted: content invalidated, tuning debt, player expectation set. Name the observable signal that would trigger re-evaluation (a playtest metric, a telemetry threshold, a review theme).

### Neutral

- Follow-on facts a future owner must know: knobs now exposed, docs and balance sheets that must track this value.

## Playtest Evidence

The evidence this decision rests on. Every entry names its source and its limits — telemetry says what players did, playtests say why (see [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)).

- **Session(s):** link the playtest report(s) — [playtest-report.md](playtest-report.md) instances in the project repo, run per [../quality/playtesting.md](../quality/playtesting.md).
- **What was tested:** the prototype or build, the variant(s) compared, tester count and freshness (first-time testers or returning).
- **What was observed:** the finding that discriminates between the alternatives above, quoted or measured — not a general "players liked it".
- **Confidence and gaps:** sample size, tester bias, untested populations. If evidence is absent, write "None — accepted on design judgment" and keep Status at Proposed until a scheduled playtest confirms or kills it.

## Links

- **Supersedes:** NNNN (and flip that DDR's status to `Superseded by` this one), or "None".
- **Superseded by:** NNNN, or leave absent until replaced.
- Related DDRs, playtest reports, balance sheets, issues, or PRs.
- Handbook references: record process in
  [design-decision-records.md](../decisions/design-decision-records.md);
  framework vocabulary this record uses in
  [frameworks-and-models.md](../decisions/frameworks-and-models.md);
  handoff context in [onboarding-and-handoff.md](../onboarding-and-handoff.md).
