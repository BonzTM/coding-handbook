# Recipe: Design A Level

Use this when you are building one level that teaches or stresses one mechanic, from paper layout to a first playable blockout.

## Files To Touch

- the level's one-page spec, from [../templates/one-page-gdd.md](../templates/one-page-gdd.md): mechanic taught, difficulty-curve position, the four beats
- a paper layout sketch (photo or scan attached to the spec) — paper before engine, per [../quality/prototyping.md](../quality/prototyping.md)
- the blockout scene in the engine: untextured geometry only, no art assets
- the playtest script and report for the blockout pass, from [../templates/playtest-script.md](../templates/playtest-script.md) and [../templates/playtest-report.md](../templates/playtest-report.md)
- the difficulty-curve document owned by [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md), if this level shifts the curve

## Steps

1. Name the one mechanic this level introduces or develops, sourced from [../foundations/mechanics-and-systems.md](../foundations/mechanics-and-systems.md), and its slot on the difficulty curve. Write both into the spec before sketching anything. A layout that goes looking for a purpose afterward violates the mechanics-driven principle in [../disciplines/level-design.md](../disciplines/level-design.md).
2. Sketch the level on paper as four labeled beats in order — **introduce, develop, twist, prove** — the Nintendo four-step structure defined in [../disciplines/level-design.md](../disciplines/level-design.md) (### Introduce-Develop-Twist-Prove Structure). Each beat is a distinct region of the sketch; annotate what the player must do in each.
3. Design the introduce beat as a safe space: failure is cheap, the new mechanic is the only new thing on screen, and its own feedback does the teaching — no tutorial text.
4. Design the develop beat as two or three escalating configurations of the mechanic: added hazard, tighter timing, or combination with exactly one already-mastered mechanic.
5. Design the twist beat to subvert the expectation the develop beat established — invert the mechanic, recontextualize it, or collide it with a system the player has not seen it touch. It must come after develop, never before.
6. Design the prove beat as an unscaffolded test of the mechanic the player must pass to finish. It certifies this level's teaching only; curve-level spikes belong to [tune-a-difficulty-curve.md](tune-a-difficulty-curve.md).
7. Plan guidance without words: light, geometry, and landmarks mark the intended path; state the goal (*what*), never the solution (*how*); pair every light or audio cue with a redundant channel per [../disciplines/accessibility.md](../disciplines/accessibility.md).
8. Build the blockout from the sketch and run a fresh-tester (Kleenex) playtest via [run-a-playtest.md](run-a-playtest.md), gated by [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md). No art pass until the blockout survives this test.

## Invariants To Preserve

- one new mechanic per level; a second introduction requires a documented reason in the spec
- the four beats stay in introduce-develop-twist-prove order
- the prove beat is a check on this level's teaching, not a difficulty spike
- zero instructional text inside the level; the space carries the guidance
- readability beats decoration: no prop, effect, or lighting choice may make game state ambiguous
- every guidance cue exists in at least two sensory channels
- no art pass before the blockout passes a fresh-tester playtest

## Proof

- beat audit: the spec names the mechanic and points at the geometry implementing each of the four beats; a beat with no geometry is a finding
- blockout playtest: a fresh tester completes the level with no verbal or written instruction from the team, reported on [../templates/playtest-report.md](../templates/playtest-report.md)
- navigation proof: testers state where they think they should go at each decision point; wrong answers are guidance failures, not tester failures
- encounter audit, if the level contains combat: cover classified soft/hard/half/full, sightlines walked from the player camera, no dominant position found
- [../checklists/design-review.md](../checklists/design-review.md) passed before the blockout is built

Governing doc: [../disciplines/level-design.md](../disciplines/level-design.md). The spec format is owned by [write-a-one-page-gdd.md](write-a-one-page-gdd.md); the playtest protocol by [../quality/playtesting.md](../quality/playtesting.md). If the level exposes a loop problem rather than a layout problem, go up a level to [design-a-core-loop.md](design-a-core-loop.md); if it exposes a curve problem, to [tune-a-difficulty-curve.md](tune-a-difficulty-curve.md).
