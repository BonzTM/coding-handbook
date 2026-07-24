# Recipes

Step-by-step procedures for the common design changes this handbook governs. Each recipe is a fixed-shape contract — **Files To Touch / Steps / Invariants To Preserve / Proof** — and links the topical doc that owns the rules it applies. Use a recipe when you know what kind of design work you are doing and want the exact artifact set and proof steps without rediscovering them.

For routing a change to its recipe and related obligations, start at the Change Routing table in [../AGENTS.md](../AGENTS.md). For the handbook overview, see [../README.md](../README.md).

## Loop And System Recipes

- [design-a-core-loop.md](design-a-core-loop.md) - design or rework the repeatable action-reward-progression loop, diagram it, and prove it against the pillars. Governed by [../foundations/core-loops.md](../foundations/core-loops.md).
- [design-a-level.md](design-a-level.md) - design one level as a mechanic-delivery structure: introduce, develop, twist, prove mastery. Governed by [../disciplines/level-design.md](../disciplines/level-design.md).

## Tuning Recipes

- [tune-a-difficulty-curve.md](tune-a-difficulty-curve.md) - adjust challenge pacing against the flow channel using playtest evidence and telemetry, not intuition alone. Governed by [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md).
- [balance-an-economy.md](balance-an-economy.md) - balance resource faucets and sinks with a cost-curve spreadsheet and a simulated model before shipping values. Governed by [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md) and [../templates/balance-spreadsheet-spec.md](../templates/balance-spreadsheet-spec.md).

## Validation Recipes

- [run-a-playtest.md](run-a-playtest.md) - script, run, and report a structured playtest with fresh testers and observation over interrogation. Governed by [../quality/playtesting.md](../quality/playtesting.md); uses [../templates/playtest-script.md](../templates/playtest-script.md) and [../templates/playtest-report.md](../templates/playtest-report.md).

## Documentation Recipes

- [write-a-one-page-gdd.md](write-a-one-page-gdd.md) - compress one system to a single dense, visual, actionable page and keep it synced with the design it describes. Governed by [../foundations/design-documentation.md](../foundations/design-documentation.md); uses [../templates/one-page-gdd.md](../templates/one-page-gdd.md).

## Where To Go Next

- Routing a change to the right files: [../AGENTS.md](../AGENTS.md) (## Change Routing)
- Handbook overview: [../README.md](../README.md)
- Checklists for lifecycle gates: [../checklists/README.md](../checklists/README.md)
- Fill-in skeletons the recipes consume: [../templates/README.md](../templates/README.md)
