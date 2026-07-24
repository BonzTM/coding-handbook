# Design Review Checklist

Review checklist for any design change that alters mechanics, systems, content, tuning values, or player-facing flow.

## Intent And Pillars

- [ ] Does the change serve at least one named pillar in [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md), and contradict none?
- [ ] Is the intended player experience stated in one sentence, and can the reviewer restate it from the change alone?
- [ ] Was the change read both ways — mechanics-up and experience-down — across the MDA gap per [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md)?
- [ ] If the change weakens or reinterprets a pillar, is a superseding record filed per [../decisions/design-decision-records.md](../decisions/design-decision-records.md)?

## Loop Integrity

- [ ] Is the core loop diagram still accurate after the change per [../foundations/core-loops.md](../foundations/core-loops.md)?
- [ ] Does every new action close its feedback loop — action, feedback, reward, next action — with no dead-end mechanic?
- [ ] Does the change add a learnable pattern or deepen mastery of an existing one, rather than noise, per the fun-as-learning model in [../foundations/player-psychology.md](../foundations/player-psychology.md)?
- [ ] If the change is arc content (consumed once) rather than loop content, is that cost acknowledged against the loops-versus-arcs split in [../foundations/core-loops.md](../foundations/core-loops.md)?

## Balance And Economy Impact

- [ ] Is every new or changed object plotted against the cost curve, and does no strictly dominant option result, per [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md)?
- [ ] Are previously viable options still viable after the change, with counterplay intact?
- [ ] Are new resource faucets matched by sinks, and is the economy model re-run, per [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md)?
- [ ] If tuning values changed, is the balance sheet updated in the same change per [../templates/balance-spreadsheet-spec.md](../templates/balance-spreadsheet-spec.md), not just the build?
- [ ] Does the difficulty curve still track the flow channel — no spike or dead zone introduced — per [../recipes/tune-a-difficulty-curve.md](../recipes/tune-a-difficulty-curve.md)?

## UX And Accessibility

- [ ] Can a player perceive, understand, and act on the change through in-game signs and feedback per [../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md)?
- [ ] If the change adds something the player must learn, is it taught by doing and sequenced into the existing onboarding beats, not explained in text?
- [ ] Does any added polish or juice preserve game-state readability rather than obscure it, per [../foundations/game-feel.md](../foundations/game-feel.md)?
- [ ] Does the change hold the accessibility baseline — remappable inputs, text size, colorblind-safe signaling, subtitle coverage — per [../disciplines/accessibility.md](../disciplines/accessibility.md)?

## Documentation Sync

- [ ] Is the one-page design for the affected system updated in the same change per [../foundations/design-documentation.md](../foundations/design-documentation.md)?
- [ ] Is any new term added to [../glossary.md](../glossary.md) with exactly one meaning and an owning doc?
- [ ] If the change overturns a recorded decision, is the record superseded using [../templates/ddr-template.md](../templates/ddr-template.md), not silently contradicted?
- [ ] Do telemetry event definitions still answer the design questions this change raises, per [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)?

## Proof

- [ ] The stated experience goal was checked in a prototype or playtest, not argued from intuition, per [../quality/playtesting.md](../quality/playtesting.md).
- [ ] Balance claims are backed by the updated cost-curve sheet or an economy simulation run, not adjectives.
- [ ] For onboarding-affecting changes, at least one fresh first-time tester exercised the changed flow.
- [ ] Instrumentation for the open design questions ships with the change, or the gap is recorded as a follow-up.
- [ ] If the change is headed for a scheduled playtest, it meets every gate in [playtest-readiness.md](playtest-readiness.md).
