# Game Feel

Owns real-time control quality and the polish (juice) budget: feel amplifies a working core and never substitutes for one, and every effect must preserve game-state readability.

## Default Approach

Treat feel as a scoped, budgeted layer on top of a proven core loop. Steve Swink's definition is the working contract: game feel is "real-time control of virtual objects in a simulated space, with interactions emphasized by polish" (*Game Feel*, 2009 — `archive.org/details/gamefeelgamedesi0000swin`). All three parts matter: get the real-time control right first, then emphasize interactions with polish — in that order. The core loop being amplified is owned by [core-loops.md](core-loops.md); do not start feel work on a loop that has not passed a gray-box playtest.

### Real-Time Control Baseline

Control quality is the substrate; no amount of particles fixes bad control. Before any juice work, the controlled object must meet this baseline:

- **Immediate response**: visible reaction to input on the next frame the simulation can produce. Input-to-response latency is a feel parameter — measure it, do not guess it.
- **Tuned response curves**: acceleration, deceleration, and turning are deliberate values under designer control, not physics-engine defaults. Record chosen values and their rationale in the design record ([design-documentation.md](design-documentation.md)).
- **Forgiveness affordances where the design pillars call for them**: input buffering and grace windows are control-quality decisions, decided against the pillars in [design-pillars-and-vision.md](design-pillars-and-vision.md), not sprinkled in ad hoc.
- **Perceived control feeds motivation**: Ryan, Rigby & Przybylski found competence and autonomy perceptions were tied to intuitive controls (`link.springer.com/article/10.1007/s11031-006-9051-8`) — control quality is a psychology concern, not only an aesthetic one. The motivation model is owned by [player-psychology.md](player-psychology.md).

Swink's own criteria for great-feeling games — aesthetic sensation of control, the pleasure of learning and mastering a skill, extension of the senses and identity, a unique physical reality — are the review questions for this baseline. If the unpolished gray-box version is not satisfying to steer, stop and fix control before spending polish budget.

### Polish And Juice Budget

Juice means maximal feedback for minimal input — tweens, particles, screen shake, sound — demonstrated by Jonasson & Purho by iteratively juicing a plain Breakout clone ("Juice It or Lose It", GDC Europe 2012 — `gdcvault.com/play/1016487/Juice-It-or-Lose`). Nijman's "The Art of Screenshake" (INDIGO 2013 — `youtube.com/watch?v=AJdEqssNZ-U`) shows the same result from ~30 individually tiny tricks: bigger bullets, muzzle flash, screen shake, camera lerp and kick, hitstop, permanence (ejected shells, corpses), sound. The practical consequences:

- **Feel is built from many cheap effects, not one expensive one.** Budget polish as a list of small, independently reversible tricks. Each trick is a line item with an owner and an on/off switch, so a playtest can isolate its contribution.
- **Juice is a budget, not a virtue.** Allocate polish time per interaction in proportion to how often the player performs it — the core-loop verbs ([core-loops.md](core-loops.md)) get juiced first; rare interactions get the leftovers.
- **Every effect earns its place.** Practitioner pushback is on record that juice can mask weak mechanics ("Indies, resist the urge to 'juice it or lose it'" — `gamedeveloper.com/design/video-indies-resist-the-urge-to-juice-it-or-lose-it-`). If a playtest finding says the loop is boring, the fix is routed through [core-loops.md](core-loops.md) or [mechanics-and-systems.md](mechanics-and-systems.md), never through more juice.

### Readability Over Spectacle

Every effect must preserve the player's ability to read game state. This is the hard gate on the juice budget, grounded in game UX practice: usability — signs and feedback, clarity, minimizing cognitive load — is half of Hodent's UX model (*The Gamer's Brain*, 2017 — `thegamersbrain.com`), and an effect that obscures state fails it.

- Screen shake, hitstop, flashes, and particle bursts may never hide information the player needs for the next input decision. If a tester misreads state during an effect, the effect shrinks or moves.
- Effects that carry meaning (damage direction, hit confirmation, resource change) rank above pure spectacle. Cut spectacle first when readability and juice conflict.
- Screen shake, flashes, and hitstop intensity are also accessibility surfaces — expose intensity and toggle options per [../disciplines/accessibility.md](../disciplines/accessibility.md).
- Readability failures are a standing observation item in every playtest script ([../templates/playtest-script.md](../templates/playtest-script.md)).

### Feel Iteration Loop

Feel cannot be specified up front; it is tuned by iteration, following the playcentric prototype-playtest-revise cycle owned by [../quality/prototyping.md](../quality/prototyping.md) and [../quality/playtesting.md](../quality/playtesting.md) (Fullerton, *Game Design Workshop* — `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`).

1. Prove the core interaction gray-box: no polish, control baseline only.
2. Add one feel trick at a time from the budget list, each behind a toggle.
3. Playtest with the trick on and off; keep it only if observers see improved feel or readability, not just "more happening."
4. Record accepted tricks, their tuning values, and rejected tricks with reasons in the feel section of the design record ([design-documentation.md](design-documentation.md)).
5. Re-run the readability check whenever effects stack: tricks that pass alone can fail in combination.

Decisions that change the control contract itself — input latency targets, forgiveness windows, camera behavior — are design decisions and get a DDR ([../decisions/design-decision-records.md](../decisions/design-decision-records.md)).

## Common Mistakes And Forbidden Patterns

- Juicing a core loop that has not survived a gray-box playtest — polish spent proving the wrong thing.
- Treating juice as the fix for "the game is boring" playtest findings; that finding routes to the loop, not the polish layer.
- Physics-engine default movement shipped as the control scheme, with no tuned response curves and no recorded values.
- Effects with no off switch, making it impossible to isolate what any single trick contributes.
- Screen shake or flash intensity that obscures the state needed for the next input, or stacked effects never tested in combination.
- Spectacle prioritized over meaning-carrying feedback when the two conflict on screen.
- No intensity or toggle options for shake, flash, and hitstop, ignoring the accessibility surface.
- Feel tuning values that live only in a build, with no record of what was chosen or why.
- Polish budget spread evenly across all interactions instead of weighted toward the core-loop verbs.

## Verification And Proof

- The gray-box version of the core interaction has a recorded playtest verdict before any polish line item starts ([../checklists/playtest-readiness.md](../checklists/playtest-readiness.md)).
- Every feel trick appears in the budget list with an owner and a toggle; a build with all toggles off still plays and reads correctly.
- A playtest observation log ([../templates/playtest-report.md](../templates/playtest-report.md)) shows each accepted trick tested on and off, and includes at least one readability observation with effects stacked.
- Measured input-to-response latency and tuned response-curve values are recorded in the design record, not only embedded in the build.
- Accessibility options for shake, flash, and hitstop intensity exist and are checked against [../disciplines/accessibility.md](../disciplines/accessibility.md).
- Control-contract changes since the last review each have a DDR ([../decisions/design-decision-records.md](../decisions/design-decision-records.md)).

Related: [core-loops.md](core-loops.md), [player-psychology.md](player-psychology.md), [../quality/prototyping.md](../quality/prototyping.md), [../quality/playtesting.md](../quality/playtesting.md), [../disciplines/accessibility.md](../disciplines/accessibility.md), [../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md)
