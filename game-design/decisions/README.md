# Decisions

The rules for making and recording hard-to-reverse design choices: which frameworks and models this handbook standardizes on, and what process a non-obvious design decision goes through before it changes the game.

This directory holds the handbook-level decision *process and defaults*. Project-level DDRs — the actual records of decisions a specific game made — live in **that project's own `design/decisions/` directory**, not here. This directory governs how those records are written and which frameworks they may safely assume.

## Decision Docs

- [design-decision-records.md](design-decision-records.md) - the DDR process and when a design decision requires one before it ships. Use it whenever a change is non-obvious, hard to reverse, or cuts against a design pillar in [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md).
- [frameworks-and-models.md](frameworks-and-models.md) - the default analytical frameworks (MDA, SDT-based motivation, flow, cost curves, Machinations) and the bar an alternative model must clear. Use it before analyzing or arguing a design in any other vocabulary — framework names live only here; topic docs link to this table instead of naming models inline.

## How Decisions Are Recorded

Project DDRs use [../templates/ddr-template.md](../templates/ddr-template.md). Copy it into the project's `design/decisions/` directory, fill in status, the pillar or framework the decision touches, alternatives considered, and expected player-facing consequences, and follow the process in [design-decision-records.md](design-decision-records.md). A DDR records the *decision and its rationale*; the resulting spec lives in the design docs governed by [../foundations/design-documentation.md](../foundations/design-documentation.md), not in the DDR.

## Relation To Change Routing

Route every decision-shaped change through the Change Routing table in [../AGENTS.md](../AGENTS.md) (## Change Routing) — it names the Start Here doc, the sync surface, and the proof for each change type. Two rules bind the routing table and this directory together:

- Any change a routing row marks as requiring a DDR does not ship without one.
- Weakening a handbook invariant or swapping a default framework is itself a decision: it needs a DDR here plus updates to every doc that cites the old default, landed in the same change.

## Where To Go Next

- Handbook overview: [../README.md](../README.md)
- Routing a decision-shaped change: [../AGENTS.md](../AGENTS.md) (## Change Routing)
- The DDR file to copy: [../templates/ddr-template.md](../templates/ddr-template.md)
- Pillars that decisions are tested against: [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md)
