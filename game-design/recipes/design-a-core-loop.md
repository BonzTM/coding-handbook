# Recipe: Design A Core Loop

Use this when a new concept, mode, or major system needs its repeatable action loop defined and proven before production work starts.

## Files To Touch

- experience goals section of the concept one-pager — [../templates/one-page-gdd.md](../templates/one-page-gdd.md)
- loop diagram and skill-atom breakdown, governed by [../foundations/core-loops.md](../foundations/core-loops.md)
- pillar alignment against [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md)
- paper prototype materials per [../quality/prototyping.md](../quality/prototyping.md)
- a DDR if the loop sets or changes a design invariant — [../decisions/design-decision-records.md](../decisions/design-decision-records.md)

## Steps

1. Write the experience goals first — what the player should feel and do per session — before naming any mechanic. The playcentric process orders goals before prototypes (Fullerton, *Game Design Workshop*, `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`).
2. Name the loop as an action sequence the player repeats, and keep the diagram zoomed out: it shows how player actions feed each other, not a detailed spec (`gamedeveloper.com/business/why-the-core-gameplay-loop-is-critical-for-game-design`).
3. Diagram the loop at three frequencies — moment-to-moment, session, and meta — since loops recur fractally at each level ([../foundations/core-loops.md](../foundations/core-loops.md)).
4. Decompose the moment-to-moment loop into skill atoms per Daniel Cook's "The Chemistry of Game Design" (`lostgarden.com/2021/03/13/the-chemistry-of-game-design-2/`): for each atom, name the player action, the simulation response, the feedback, and the skill the player gains.
5. Classify each element as loop (repeatable, mastery-driven) or arc (consumed once) using Cook's "Loops and Arcs" (`lostgarden.com/2012/04/30/loops-and-arcs/`); arcs mislabeled as loops are where iteration budget silently drains.
6. Build a paper prototype of one full loop cycle — paper is the fastest, cheapest test and rules can change mid-session (Fullerton, prototyping chapter, `taylorfrancis.com/chapters/mono/10.1201/9781003460268-9/prototyping-tracy-fullerton`).
7. Run the paper test with at least one fresh player following [run-a-playtest.md](run-a-playtest.md); watch where they stall — a stalled skill atom marks where players get lost and fail to master the skill (Cook).
8. Decide keep, revise, or kill against the experience goals; record a keep decision that sets an invariant as a DDR.

## Invariants To Preserve

- experience goals precede and outrank mechanics; the loop exists to serve them
- every skill atom names a skill the player is learning; an atom that teaches nothing is cut or reclassified as an arc
- the loop diagram stays zoomed out; mechanic detail lives in [../foundations/mechanics-and-systems.md](../foundations/mechanics-and-systems.md)
- the loop supports every design pillar, or the conflict is escalated via DDR before production
- the loop keeps teaching new patterns as skill grows — fun is pattern-learning, and a loop with nothing left to teach goes boring (Koster, *A Theory of Fun*, `theoryoffun.com/press.shtml`)

## Proof

- a one-page loop diagram at all three frequencies, reviewable without the author present
- a skill-atom map with no dead atoms and a named skill per atom
- a completed paper test where a fresh player repeated the loop three or more times unaided
- observed behavior matching the stated experience goals, or the divergence logged in the playtest report
