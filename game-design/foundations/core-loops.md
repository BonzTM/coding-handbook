# Core Loops

The core loop is the structural unit of a game: the repeatable action-feedback cycle every mechanic, level, and economy hangs off. A concept without a stated core loop is not a design yet.

## Default Approach

Name the loop first, in one sentence of player verbs ("scout -> fight -> loot -> upgrade -> scout deeper"), before writing mechanics, content, or story. Practitioner consensus treats the core loop as "the primary game system or mechanic that defines your title" (`gamedeveloper.com/business/why-the-core-gameplay-loop-is-critical-for-game-design`). Every other foundations doc assumes this loop exists: pillars judge it ([design-pillars-and-vision.md](design-pillars-and-vision.md)), mechanics implement it ([mechanics-and-systems.md](mechanics-and-systems.md)), feel polishes it ([game-feel.md](game-feel.md)).

### Loop Altitudes

Loops are fractal: the same action-feedback shape recurs at multiple frequencies, and each altitude must close — the player must return to the start of the cycle with something changed (`gameanalytics.com/blog/how-to-perfect-your-games-core-loop`).

| Altitude | Cycle time | What repeats | Owned by |
|---|---|---|---|
| Moment | seconds | input -> response -> readable feedback | this doc + [game-feel.md](game-feel.md) |
| Session | minutes to an hour | goal -> challenge -> reward -> next goal | this doc + [../disciplines/level-design.md](../disciplines/level-design.md) |
| Meta | days to weeks | accumulate -> unlock -> re-enter sessions stronger | [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md) |

- Design and validate the moment loop first. A meta loop cannot rescue a moment loop that is not worth repeating; retention systems bolted onto a weak core are a forbidden pattern below.
- State explicitly what each altitude feeds into the one above it (loot from the moment loop funds the meta loop, meta upgrades change the moment loop). A loop that feeds nothing is decoration.

### Skill Atoms And Feedback

Decompose the loop into Daniel Cook's atomic unit, the skill atom — a feedback cycle in which the player acts, the simulation responds, feedback is delivered, and the player updates their mental model, gaining a skill (Daniel Cook, "The Chemistry of Game Design", `lostgarden.com/2021/03/13/the-chemistry-of-game-design-2/`).

1. **Action**: the player exercises an input or decision.
2. **Simulation**: the rules compute a result.
3. **Feedback**: the result is made perceivable — see [game-feel.md](game-feel.md) for the polish layer and [../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md) for readability.
4. **Modeling**: the player learns something and acts again, better.

- Chain atoms into a skill graph: which skills are prerequisites for which. Cook's claim is that mapping atoms shows where players get lost and fail to master skills — use the map to place teaching moments ([../disciplines/ux-and-onboarding.md](../disciplines/ux-and-onboarding.md)) and to script playtest probes ([../quality/playtesting.md](../quality/playtesting.md)).
- An atom with weak or absent feedback is a broken atom: the player acted, the simulation resolved, and nothing taught them anything. Fix the feedback before tuning the numbers.

### Fun As Learning

The working theory of why loops retain players is Raph Koster's: fun is the act of learning new patterns; the brain chunks mastered patterns, and when a game has no more patterns to teach, it becomes boring (Raph Koster, *A Theory of Fun for Game Design*, `theoryoffun.com/press.shtml`). Consequences this handbook adopts as defaults:

- Every loop iteration must offer either a pattern still being learned or a meaningful variation on a mastered one. "More of the same, bigger numbers" is stalling, not designing.
- Exhaustion of patterns, not exhaustion of content, is the real end of a game. Budget accordingly — see the next section.
- Player motivation beyond pattern-learning (autonomy, competence, relatedness, motivation profiles) is owned by [player-psychology.md](player-psychology.md); do not restate it here.

### Loops Versus Arcs

Cook's second distinction: loops are repeatable mastery-driven systems; arcs are consumed-once content — story beats, set pieces, hand-authored levels (Daniel Cook, "Loops and Arcs", `lostgarden.com/2012/04/30/loops-and-arcs/`). Use the split as a production-budget lens:

- Loop spend buys replayable hours; arc spend buys each hour once. Know which one every line item on the content plan is.
- Arcs are legitimate — they carry narrative and pacing ([../disciplines/narrative-integration.md](../disciplines/narrative-integration.md)) — but a game whose engagement is all arcs has a content treadmill, not a core loop, and its scope risk belongs in [../operations/scoping-and-production.md](../operations/scoping-and-production.md).
- When a pitch claims "40 hours of gameplay", ask how many are loop hours and how many are arc hours. The answers have different price tags.

### Diagramming The Loop

Every concept ships a core-loop diagram: a deliberately zoomed-out, one-page view showing how player actions feed into each other — not a detailed spec (`gamedeveloper.com/business/why-the-core-gameplay-loop-is-critical-for-game-design`). This is a required artifact of [../checklists/concept-intake.md](../checklists/concept-intake.md) and of the one-page GDD ([../templates/one-page-gdd.md](../templates/one-page-gdd.md)).

```text
        +--> fight enemies --> collect loot --+
        |                                     |
   choose route                          spend on upgrades
        ^                                     |
        +------- return stronger <------------+
```

- Nodes are player verbs; edges are what flows between them (resources, unlocks, information). If an edge has no label, either label it or delete it.
- One diagram per altitude when the game has a real meta loop; draw the arrows that connect altitudes.
- Resource flows on the diagram must reconcile with the economy model — faucets and sinks are owned by [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md).
- The step-by-step construction procedure, including the paper-prototype check, lives in [../recipes/design-a-core-loop.md](../recipes/design-a-core-loop.md).

## Common Mistakes And Forbidden Patterns

- Pitching theme, story, or feature lists with no stated loop of player verbs.
- Building the meta loop (progression, retention, monetization hooks) before the moment loop is proven fun in a prototype ([../quality/prototyping.md](../quality/prototyping.md)).
- Skill atoms with missing feedback: the simulation resolves but the player cannot perceive the result, so no learning occurs.
- Loops that do not close — a reward that feeds nothing, an upgrade that changes nothing about the next iteration.
- Confusing arc content for loop depth: shipping a content treadmill and calling it replayability.
- Difficulty or reward escalation with no new pattern to learn — bigger numbers standing in for design.
- Loop diagrams that are secretly full specs: ten-page flowcharts with UI states and edge cases. The diagram is a zoomed-out communication tool; the spec lives in [design-documentation.md](design-documentation.md).
- Restating motivation theory, economy math, or feel polish here instead of linking their owning docs.

## Verification And Proof

- The loop is writable as one sentence of player verbs, and the team repeats the same sentence back unprompted.
- The one-page loop diagram exists, every node is a player verb, every edge is labeled, and it passes [../checklists/concept-intake.md](../checklists/concept-intake.md).
- For each altitude (moment, session, meta), name what it feeds into the altitude above; any loop feeding nothing is cut or connected.
- Walk the skill-atom chain: for each atom, name the action, the rule, the perceivable feedback, and the skill learned. An atom missing any of the four is defective.
- A gray-box prototype of the moment loop alone — no meta, no arcs, no polish — is still worth ten consecutive iterations to a fresh playtester ([../quality/playtesting.md](../quality/playtesting.md)). If it is not, stop and redesign before adding anything.
- Classify the content plan line-by-line as loop or arc spend and confirm the split matches the pitch's replayability claim.

## Related

- [../recipes/design-a-core-loop.md](../recipes/design-a-core-loop.md) - the construction recipe.
- [mechanics-and-systems.md](mechanics-and-systems.md) - the mechanics that implement loop steps.
- [player-psychology.md](player-psychology.md) - why players stay in the loop.
- [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md) - where loop/atom vocabulary sits among the formal frameworks.
