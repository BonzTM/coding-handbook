# Checklists

Executable gates for the moments where missing a step is expensive: accepting a concept into work, reviewing a design before it enters production, and confirming a build is ready to put in front of playtesters. Each checklist is a grouped set of `- [ ]` items plus a closing Proof section; run it top to bottom and do not skip items.

For routing a change to its related obligations, see the Change Routing table in [../AGENTS.md](../AGENTS.md). For the handbook overview, see [../README.md](../README.md).

## Intake And Review Checklists

- [concept-intake.md](concept-intake.md) - resolving the WHAT decisions (pillars, core loop, target player, scope class, kill criteria) before a concept consumes prototype time. Governed by [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md) and [../operations/scoping-and-production.md](../operations/scoping-and-production.md).
- [design-review.md](design-review.md) - reviewing a design (system, level, economy, or one-page GDD) before it enters production, including the pillar-alignment and MDA gap check. Governed by [../foundations/design-documentation.md](../foundations/design-documentation.md) and [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md).

## Validation Checklists

- [playtest-readiness.md](playtest-readiness.md) - confirming a build, script, observers, and fresh testers are in place before a session burns Kleenex testers who can never be first-time players again. Governed by [../quality/playtesting.md](../quality/playtesting.md).

## How Checklists Are Written

- Every box is a question or verifiable assertion a reviewer can answer with evidence they can point at — a document section, a diagram, a build, or a playtest observation. "Feels good" is not a checkable item; "did five of six testers complete the first level unaided" is.
- Every checklist ends with a `## Proof` group naming the artifacts that show the checklist was actually run: a filled template from [../templates/README.md](../templates/README.md), a playtest report, or a Design Decision Record per [../decisions/design-decision-records.md](../decisions/design-decision-records.md).
- Checklists gate moments, not topics. The rules they check are owned by the linked topic docs; a checklist never restates a rule, it asks whether the rule was followed. Adding a rule means updating the owning doc first, then the checklist that gates it.

## Where To Go Next

- Handbook overview: [../README.md](../README.md)
- Routing a change to the right files: [../AGENTS.md](../AGENTS.md) (## Change Routing)
- Recipes for step-by-step execution: [../recipes/README.md](../recipes/README.md)
- Blank artifacts the Proof sections expect: [../templates/README.md](../templates/README.md)
