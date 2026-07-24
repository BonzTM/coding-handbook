# Prototyping

Prototype discipline for design teams: the cheapest artifact that answers one design question, built to be thrown away.

## Default Approach

Follow the playcentric process: set experience goals first, then prototype, playtest, and revise against those goals continuously (Fullerton, *Game Design Workshop*, 5th ed. — `routledge.com/Game-Design-Workshop-A-Playcentric-Approach-to-Creating-Innovative-Games/Fullerton/p/book/9781032607009`). A prototype exists to answer a question about the design, not to start the codebase. Rami Ismail's framing is the scope test: prototypes tell you whether you *should* make the game; the vertical slice tells you whether you *can* (`ltpf.ramiismail.com/prototypes-and-vertical-slice/`). The experience goals a prototype is tested against are owned by [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md); the loop it usually probes is owned by [../foundations/core-loops.md](../foundations/core-loops.md).

Every prototype starts from three written lines before anything is built:

1. The question, phrased so a playtest can answer it ("Is chaining dashes into attacks legible to a first-time player?").
2. The pass and kill criteria — what observed player behavior counts as yes, no, or inconclusive.
3. The timebox — a calendar deadline after which the answer is recorded and the prototype stops, whatever state it is in.

### Paper First

Default to a physical prototype before any digital build. Fullerton calls physical prototyping the heart of the playcentric process, for reasons that are operational, not nostalgic (`taylorfrancis.com/chapters/mono/10.1201/9781003460268-9/prototyping-tracy-fullerton`):

- It is the fastest and cheapest method available, so a bad idea dies in hours instead of weeks.
- It keeps attention on gameplay, not technology — there is no engine to fiddle with.
- Rules can change mid-test: when a mechanic fails in front of you, you amend it and keep playing.
- Every discipline can participate; no one is blocked on a programmer.

Paper answers questions about rules, turn structure, resource flow, spatial layout, and decision pressure. Sketch economy questions in a resource-flow diagram before committing them to paper rules; the diagram grammar is owned by [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md). Skipping paper is allowed only when the question is one paper cannot represent — see the ladder below — and the skip is noted in the prototype's written question.

### Prototype Fidelity Ladder

Enter the ladder at the lowest rung that can answer the question, and climb only when the current rung has produced an answer that justifies the next investment. The rung names follow the standard milestone vocabulary (`askagamedev.tumblr.com/post/77406994278/game-development-glossary-the-vertical-slice`).

| Rung | Medium | Answers questions about | Typical cost |
|---|---|---|---|
| paper | cards, tokens, sketches, humans running the rules | rules, turn structure, economy flows, decision pressure, spatial layout | hours |
| digital greybox | placeholder art, one mechanic, hardcoded values | real-time interaction of one system, simulation-scale behavior paper cannot run | days |
| feel prototype | playable build with tuned input, camera, and feedback | control, responsiveness, readability of moment-to-moment play | days to weeks |
| first playable | core loop assembled end to end, placeholder content | whether the loop holds attention across a full session | weeks |
| vertical slice | one slice of the game at intended final quality | whether the team and pipeline can ship it | months |

Two rules govern the ladder:

- **Feel questions cannot be answered on paper.** Game feel is real-time control of virtual objects with interactions emphasized by polish (Swink — definition owned by [../foundations/game-feel.md](../foundations/game-feel.md)), so a control or juice question enters at the feel rung, never below it, and never above it.
- **The vertical slice is not a prototype.** It is a production and funding artifact with its own contested trade-offs, owned by [../operations/scoping-and-production.md](../operations/scoping-and-production.md). Do not climb into slice territory to answer a design question — that is the exact waste the ladder exists to prevent.

### One Question Per Prototype

A prototype that answers two questions answers neither: when the playtest goes badly you cannot tell which answer was no. One built artifact, one written question.

- Generate questions deliberately. Schell's lenses are the house hypothesis generator — a lens interrogates the design and produces a testable question; the prototype and playtest check it (Schell, *The Art of Game Design*, 3rd ed. — `routledge.com/The-Art-of-Game-Design-A-Book-of-Lenses-Third-Edition/Schell/p/book/9781138632059`; lens usage owned by [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md)).
- Related questions get separate prototypes, sequenced by risk: prototype the question whose "no" kills the concept first. That ordering is the concept-intake gate in [../checklists/concept-intake.md](../checklists/concept-intake.md).
- The prototype gets its answer from observed play, not team opinion. Run it through the protocol in [playtesting.md](playtesting.md); a prototype session still meets [../checklists/playtest-readiness.md](../checklists/playtest-readiness.md).
- When the answer is inconclusive at the timebox, the recorded answer is "inconclusive" plus a decision: rephrase the question, climb one rung, or kill the line of inquiry. Never silently extend the timebox.

### Throwaway Discipline

A prototype is throwaway quality by definition (`askagamedev.tumblr.com/post/77406994278/game-development-glossary-the-vertical-slice`). The failure mode this section exists to block: a prototype that tests well quietly becomes the foundation of the production codebase, and the team spends the next year paying interest on code written to be deleted.

- Prototype code lives outside the production mainline — a separate repository or a branch that is never merged. There is no path from prototype branch to main.
- Build to the question's fidelity and no further. Polish, error handling, save systems, and options menus on a prototype are scope theft from the question.
- When a prototype's answer is "yes, build this," the mechanic is **rebuilt** under production standards; the prototype is the spec, not the seed. The rebuild decision is recorded as a design decision record ([../decisions/design-decision-records.md](../decisions/design-decision-records.md)) citing the prototype's question, build, and playtest result.
- Dead prototypes are still results. Archive the written question, the pass/kill criteria, the playtest findings, and the decision — a killed question that goes unrecorded gets re-asked and re-built next year. Findings use the report shape in [../templates/playtest-report.md](../templates/playtest-report.md).
- Promoting prototype code directly into production is the ADR-level exception, not a shortcut: it requires a DDR stating why a rewrite is not warranted and what debt is being accepted.

## Common Mistakes And Forbidden Patterns

- Building a digital prototype for a rules or economy question that a deck of index cards would have answered by lunch.
- Prototyping with no written question — "exploring the mechanic" produces a demo, not an answer.
- Stacking multiple questions into one build, then arguing about which one the failed playtest actually answered.
- Answering the question by team consensus in a meeting instead of by watching players ([playtesting.md](playtesting.md)).
- Polishing a prototype — juice, menus, save support — before the question it exists for has an answer.
- Letting a successful prototype's codebase become the production codebase without a rebuild or a DDR accepting the debt.
- Silently extending a timebox instead of recording "inconclusive" and making an explicit rephrase/climb/kill decision.
- Skipping rungs upward — jumping from paper to first playable because the team is excited, spending weeks to learn what days would have taught.
- Using a vertical slice to answer a design question; slices prove the team and pipeline, not the design ([../operations/scoping-and-production.md](../operations/scoping-and-production.md)).
- Discarding the findings with the prototype — deleting the build is required, deleting the answer is a repeat purchase of the same lesson.

## Verification And Proof

- Every active prototype has its question, pass/kill criteria, and timebox written down before the build started — point at the document.
- The question maps to exactly one rung of the fidelity ladder, and the build sits on that rung; any rung skip is justified in the written question.
- The answer came from a playtest run per [playtesting.md](playtesting.md), with findings recorded in the [../templates/playtest-report.md](../templates/playtest-report.md) shape.
- Closed prototypes show a recorded outcome — yes, no, or inconclusive-plus-decision — and dead ones are archived with their findings.
- No prototype code is reachable from the production mainline; any exception has a DDR ([../decisions/design-decision-records.md](../decisions/design-decision-records.md)).
- A "yes" that entered production points at the rebuild commit or the DDR that waived it.

Prototyping is done when the concept's riskiest questions have recorded answers — not when the prototype is fun to show.
