# Recipe: Tune A Difficulty Curve

Use this when playtest or telemetry evidence shows challenge pacing off the flow channel — a spike players quit at, a stretch they coast through, or an onboarding segment that loses fresh players.

## Files To Touch

- the intended-curve section of the balance model — per level/chapter: skill tested, target fail rate, target completion time, planned rest valleys (see [../templates/balance-spreadsheet-spec.md](../templates/balance-spreadsheet-spec.md))
- the tuning values under test (encounter stats, timings, resource grants) — in the balance model first, the build second
- the playtest script and report for the affected segment ([../templates/playtest-script.md](../templates/playtest-script.md), [../templates/playtest-report.md](../templates/playtest-report.md))
- the telemetry event list, if the game is live — the events that answer this curve's questions ([../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md))
- a DDR only if the intended curve itself changes, or if hidden DDA is being introduced ([../decisions/design-decision-records.md](../decisions/design-decision-records.md))

## Steps

1. Write down the intended curve before touching a number. For each level or chapter: which skill it tests, where it sits relative to the segments around it, and the target evidence (expected attempts-to-clear, completion time, acceptable quit rate). The rationale is the flow channel — challenge far above skill produces anxiety, far below produces boredom (Csikszentmihalyi, *Flow*, 1990; policy owned by [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) (### Difficulty Curve Design)). Without a stated intent, "too hard" is unfalsifiable.
2. Instrument the failure points against that intent: per-encounter attempt counts, death/fail locations, time-to-complete, retry counts, and quit points. Pre-launch this is observer tally data from [../quality/playtesting.md](../quality/playtesting.md); live it is telemetry events. Instrument the design question, not everything — telemetry shows what players do, never why (Seif El-Nasr, Drachen & Canossa, *Game Analytics*, Springer 2013).
3. Collect the evidence, then locate hot spots by clustering: a spike is a fail/quit cluster shared across testers, not one loud complaint. Completion-time outliers in the easy direction mark boredom valleys the same way.
4. Diagnose each hot spot before tuning. Three distinct root causes look identical in the data: a **teaching failure** (the spike tests a skill never introduced — a sequencing bug owned by [../disciplines/level-design.md](../disciplines/level-design.md), fixed by reordering, not by numbers), a **readability failure** (the player cannot see the threat or the feedback — route to [../foundations/game-feel.md](../foundations/game-feel.md) and [../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md)), and a genuine **tuning failure**. Only the last one is fixed in this recipe.
5. For a tuning failure, change one variable at a time. Record the change in the balance model with the predicted effect on the target evidence before testing it. A multi-variable change that works teaches nothing about which variable worked; the next tuning pass starts blind.
6. Prefer adding a self-selection point over flattening the curve: an optional hard route, a skippable challenge, a loadout risk choice lets players adjust challenge in-fiction (Jenova Chen, "Flow in Games", `jenovachen.com/flowingames/Flow_in_games_final.pdf`). Hidden DDA is not a default tuning tool — it requires a DDR per the policy in [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) (### Dynamic Difficulty Policy).
7. Re-test with fresh players. Returning testers have practiced skill the target audience will not have, so they systematically overstate ease — a first-time tester can only be first-time once (Kleenex testing, Will Wright, `masterclass.com/classes/will-wright-teaches-game-design-and-theory/chapters/playtesting`). Run the session per [run-a-playtest.md](run-a-playtest.md).
8. Compare observed evidence to the intended curve. If the cluster is gone and neighboring segments did not regress, commit the value; otherwise revert or iterate from step 5. For a live game, ship the committed value through the segment-first remote-config rollout in [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md).

## Invariants To Preserve

- the intended curve rises with player skill and keeps its rest valleys; tuning removes accidental spikes, not the designed ones
- every difficulty spike tests a skill the game has already taught and let the player practice (kishōtenketsu sequencing, owned by [../disciplines/level-design.md](../disciplines/level-design.md))
- one variable changes per test iteration, recorded in the balance model with prediction and outcome
- fresh (Kleenex) testers for any verdict on onboarding or first-encounter difficulty; veteran data never signs off a first-time experience
- tuning values live in the committed balance model, never only in the build
- hidden DDA only through a DDR with stated bounds, and never in competitive or leaderboard modes ([../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md))
- difficulty settings are not the accessibility plan; accommodations are owned by [../disciplines/accessibility.md](../disciplines/accessibility.md)
- live changes roll out segment-first with the justifying metric named in the change note

## Proof

- an intended-curve entry exists for every affected segment with target fail rate, completion time, and quit rate
- the playtest report attaches fail-point, retry, and quit-point data per encounter, and the diagnosed root cause (teaching / readability / tuning) for each hot spot
- each tuning change is a single-variable balance-model entry with predicted effect and the observed re-test result beside it
- a fresh-player re-test shows the target cluster resolved and no regression in adjacent segments
- for live changes: the segment rollout record and the before/after metric per [../operations/live-tuning-and-telemetry.md](../operations/live-tuning-and-telemetry.md)

See [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md) for the flow-channel rationale, the DDA policy, and the balance method this recipe assumes, and [../quality/playtesting.md](../quality/playtesting.md) for the observation protocol the evidence comes from.
