# Scoping And Production

Owns the milestone ladder, the vertical-slice policy and its known costs, and the discipline for cutting scope without cutting the game's identity.

## Default Approach

- Scope is a budget, not a wish list. Every milestone has written entry and exit criteria before the milestone starts, and a milestone exits on criteria, never on the calendar.
- Milestones gate on playable builds, not documents. The playcentric process (Fullerton, *Game Design Workshop*, 5th ed., `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`) is prototype -> playtest -> revise against stated experience goals; a milestone that cannot be played cannot be evaluated. Playtest protocol lives in [../quality/playtesting.md](../quality/playtesting.md).
- Milestone documentation is one-pagers, not novels. Each milestone's scope statement fits the one-page shape defined in [../foundations/design-documentation.md](../foundations/design-documentation.md) (Librande, "One-Page Designs", GDC 2010, `gdcvault.com/play/1012356/One-Page`), started from [../templates/one-page-gdd.md](../templates/one-page-gdd.md).
- The design pillars in [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md) are the cut criterion. Work that serves no pillar is cut-eligible from day one; the cut list below is maintained continuously, not invented in a crisis.
- Scope changes that alter an invariant — a pillar, a core-loop verb, the systems roster — are recorded as DDRs per [../decisions/design-decision-records.md](../decisions/design-decision-records.md).

## Milestone Ladder

The default ladder is prototype -> first playable -> vertical slice -> production (Ask a Game Dev, "Game Dev Glossary: Prototype, Vertical Slice, First Playable, MVP, Demo", `tumblr.com/askagamedev/746300998961741824/game-dev-glossary-prototype-vertical-slice`). Rami Ismail's framing is the decision rule for the first two questions: prototypes answer whether you *should* make the game; the vertical slice answers whether you *can* (`ltpf.ramiismail.com/prototypes-and-vertical-slice/`).

| Milestone | Question It Answers | Exit Criterion | Quality Bar |
|---|---|---|---|
| Prototype | Should we make this? | Core loop is fun in isolation per [../quality/prototyping.md](../quality/prototyping.md); a stated experience goal survived playtests | Throwaway; ugly on purpose |
| First playable | Does the loop hold up in context? | Core loop from [../foundations/core-loops.md](../foundations/core-loops.md) playable start to finish with placeholder assets | Placeholder everything |
| Vertical slice | Can we make this? | One small portion with all major systems working together at intended final quality | Final quality, narrow scope |
| Production | Can we make this repeatably? | Content pipeline produces shippable units at a measured, sustainable rate against a frozen systems roster | Shippable |

- Do not skip rungs. Entering slice work before a prototype has answered the "should" question means polishing an unproven loop — the most expensive way to discover it is not fun.
- Each milestone exit is a design review; run [../checklists/design-review.md](../checklists/design-review.md) against the build, and gate playtest-driven exits on [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md).

## Vertical Slice Policy

- **What the slice is for.** It de-risks the full pipeline — design, code, art, audio, and QA integrating on one goal — and it is the artifact publishers use to judge funding; typical build time for the slice itself is 1–3 months (Ask a Game Dev, `askagamedev.tumblr.com/post/77406994278/game-development-glossary-the-vertical-slice`). Treat it as a de-risking and funding tool, not an automatic best practice every project owes itself.
- **Known cost: slice content is often rebuilt.** Building final-quality content early is criticized in practice as wasteful, because slice content frequently gets remade once systems settle, and the slice pushes teams to polish before systems are proven. Budget the slice as spend-to-learn, not as shipped content; any plan that counts slice assets as production output needs a DDR saying why.
- **Known cost: premature polish pressure.** No polish or game-feel pass ([../foundations/game-feel.md](../foundations/game-feel.md)) starts on slice content before the underlying loop has survived playtests. Juice amplifies a good core; it does not substitute for one.
- **Alternatives are legitimate, but routed.** Horizontal-first prototyping (broad and shallow across systems) and pitch-oriented "fake" slices are the recognized counter-practices. Choosing one over the default ladder is a DDR, with the funding or de-risking need it serves stated explicitly.
- **The slice defines the quality bar.** Whatever quality the slice demonstrates is the bar production content is measured against; a slice polished beyond what production can sustain is a schedule lie told to yourself and the publisher.

## Scope Cuts And The Cut List

- Maintain a ranked cut list from the first milestone — a convention this handbook adopts. Every feature and content batch appears on it, ordered by distance from the pillars; when time pressure arrives, cuts come off the top of the list instead of landing on whatever happens to be unfinished.
- Cut arcs before loops. Daniel Cook's distinction — loops are repeatable mastery-driven systems, arcs are consumed-once content (`lostgarden.com/2012/04/30/loops-and-arcs/`) — is the triage rule: cutting an arc shrinks the game; cutting a loop changes what game it is.
- A cut that touches a pillar is not a cut, it is a re-vision. It goes through [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md) and a DDR, not the cut list.
- Every executed cut is recorded as a DDR with the reason and the expected saving, so the next owner knows what was removed on purpose versus never built.
- Cuts cascade. After a cut, re-check the sync surface it drags along: the core-loop diagram, the economy flows in [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md), and the difficulty curve in [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md). A cut that leaves a dangling faucet, sink, or difficulty spike is only half done.

## Content Versus Systems Budget

- Classify every line item as loop (system) or arc (content) using Cook's "Loops and Arcs" (`lostgarden.com/2012/04/30/loops-and-arcs/`). The classification is the budget: arcs cost roughly per minute of play and are consumed once; loops cost up front and pay back across every session the player masters them in.
- Systems generate play only while they still have patterns to teach — a game with no more patterns to learn becomes boring (Koster, *A Theory of Fun for Game Design*, `theoryoffun.com/press.shtml`). Depth budgets belong on the loops that carry the game, not spread thin across many shallow ones; the owning doc is [../foundations/mechanics-and-systems.md](../foundations/mechanics-and-systems.md).
- Freeze the systems roster at slice exit — a convention this handbook adopts. Production adds content against frozen systems; a new system added mid-production reopens integration, balance, and tutorialization work across everything already built, and therefore requires a DDR.
- Content quantity is decided late, systems count early. Level and content targets flex with the cut list; the systems roster does not flex without a DDR.

## Common Mistakes And Forbidden Patterns

- entering vertical-slice work before a prototype has answered the "should we make this" question
- counting slice content as shipped production output instead of spend-to-learn
- polishing slice assets while the core loop is still failing playtests
- exiting a milestone on the calendar ("it is May, so we are in production") instead of on written criteria
- crisis-cutting with no maintained cut list, so cuts land on whatever is unfinished rather than what matters least
- cutting a loop to save an arc because the arc was further along
- executing cuts without DDRs, so the next owner cannot tell deliberate removals from gaps
- adding a new system mid-production without a DDR, reopening balance and onboarding work across finished content
- a slice polished beyond the quality bar production can actually sustain
- a scope statement that is a feature list with no exit criteria attached

## Verification And Proof

- every milestone has entry and exit criteria written before it started, and the current build is being checked against them, not against the date
- the prototype's "should" answer exists as a playtest report per [../templates/playtest-report.md](../templates/playtest-report.md) before any slice work is scheduled
- the slice demonstrates all major systems together in one playable path, and a first-time player can traverse it unaided per [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md)
- the cut list exists, is ranked against the pillars, and every executed cut has a DDR
- the systems roster in production matches the slice-exit roster, or the diff is covered by DDRs
- each milestone's one-page scope statement is current and matches the build a reviewer can play

## Related

- [../quality/prototyping.md](../quality/prototyping.md) — what a prototype must prove before the ladder advances
- [../quality/playtesting.md](../quality/playtesting.md) — the evidence source milestone exits gate on
- [live-tuning-and-telemetry.md](live-tuning-and-telemetry.md) — post-ship scope: tuning shipped systems instead of adding new ones
- [../decisions/design-decision-records.md](../decisions/design-decision-records.md) — the record every cut, freeze exception, and ladder deviation flows through
