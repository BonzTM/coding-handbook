# Mechanics And Systems

This doc owns the handbook's shared analysis vocabulary — mechanics, dynamics, aesthetics — and the rules for applying it. The decision to standardize on MDA, and the criteria for adopting any other framework, are recorded in [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md); this doc defines the terms and the working method.

## Default Approach

Analyze and specify every design at three causally linked layers, using the MDA framework (Hunicke, LeBlanc & Zubek, 2004 — `users.cs.northwestern.edu/~hunicke/MDA.pdf`).

- **Specify experience-first.** State the target aesthetics before naming mechanics. The paper is explicit that "thinking about the player encourages experience-driven (as opposed to feature-driven) design." A spec that opens with a feature list and never names the intended player experience is incomplete; the experience targets themselves derive from the pillars in [design-pillars-and-vision.md](design-pillars-and-vision.md).
- **Edit only at the mechanics layer.** Dynamics and aesthetics are outcomes, not inputs. Every proposed change names the specific rule, number, or affordance being changed, the dynamic it is expected to shift, and the aesthetic that shift serves.
- **Model dynamics before playtesting them.** Per the paper, "by developing models that predict and describe gameplay dynamics, we can avoid some common design pitfalls." Cheap models (probability tables, spreadsheets, resource-flow diagrams) come before builds; playtests then check the model, they do not replace it.
- **Use the vocabulary consistently.** Each term below has exactly one meaning, defined here and indexed in [../glossary.md](../glossary.md). Do not use "mechanic," "system," and "feature" interchangeably in specs or reviews.

## MDA Vocabulary

The framework breaks a game into rules -> system -> "fun" and establishes their design counterparts, defined verbatim in the paper:

| Layer | Definition (Hunicke, LeBlanc & Zubek) | Designer control |
|---|---|---|
| Mechanics | "the particular components of the game, at the level of data representation and algorithms" — the actions, behaviors, and control mechanisms afforded to the player, plus content | Direct: this is the only layer you author |
| Dynamics | "the run-time behavior of the mechanics acting on player inputs and each others' outputs over time" | Indirect: emerges from mechanics in play |
| Aesthetics | "the desirable emotional responses evoked in the player, when she interacts with the game system" | Indirect only: the target, never the edit point |

The load-bearing insight is the direction of reading: "From the designer's perspective, the mechanics give rise to dynamic system behavior, which in turn leads to particular aesthetic experiences. From the player's perspective, aesthetics set the tone, which is born out in observable dynamics and eventually, operable mechanics." Design work lives in the gap between those two readings — you build bottom-up, the player experiences top-down.

For aesthetics, replace the word "fun" with the paper's taxonomy, which "includes but is not limited to": sensation, fantasy, narrative, challenge, fellowship, discovery, expression, submission. A game pursues several in stated priority order — the paper's own example ranks Charades as fellowship over challenge. Name the two or three that a feature serves; "make it fun" is not a spec. Why these aesthetics motivate players — SDT, flow, motivation models — is owned by [player-psychology.md](player-psychology.md).

## Mechanics To Dynamics

- **Predict, then verify.** The paper's Monopoly worked example is the house method: a two-dice probability model predicts pacing around the board; identifying the feedback system ("as the leader or leaders become increasingly wealthy, they can penalize players with increasing effectiveness") predicts the runaway-leader dynamic and the resulting loss of "dramatic tension and agency."
- **Trace complaints downward.** A playtest complaint arrives at the aesthetic layer ("boring," "unfair," "grindy"). Diagnose the dynamic producing it, then locate the mechanics feeding that dynamic. Fixes are mechanic edits — the paper's Monopoly fixes are subsidies for lagging players or progressive taxes on leaders, not "make it more tense."
- **Know the canonical chains.** The paper gives concrete mechanic-to-aesthetic examples worth memorizing as patterns: "challenge is created by things like time pressure and opponent play"; fellowship "can be encouraged by sharing information across certain members of a session (a team) or supplying winning conditions that are more difficult to achieve alone"; expression "comes from dynamics that encourage individual users to leave their mark."
- **Identify feedback loops explicitly.** For every core system, state whether its dominant loop amplifies leads (rich-get-richer) or corrects them, and whether that matches the target aesthetic. Runaway loops are acceptable only when shortening the game toward a climax is the intent.

## Systemic Interactions

- Every new or changed mechanic declares its interaction surface: which existing mechanics it feeds, competes with, or gates, and the dynamics expected from each pairing. A mechanic reviewed in isolation is not reviewed.
- Distinguish loops from arcs before spending: repeatable mastery-driven systems versus consumed-once content (Daniel Cook, "Loops and Arcs" — `lostgarden.com/2012/04/30/loops-and-arcs/`). Loop structure and skill-atom mapping are owned by [core-loops.md](core-loops.md).
- Model resource-producing and resource-consuming mechanics as an internal economy — pools, sources, drains, converters — before implementation (Adams & Dormans, *Game Mechanics: Advanced Game Design* — `oreilly.com/library/view/game-mechanics-advanced/9780132946728/ch05.html`). Economy structure, faucets, and sinks are owned by [../disciplines/economy-and-progression.md](../disciplines/economy-and-progression.md).
- When multiple options must stay viable against each other, prefer intransitive (counter) relationships over transitive power ladders; tuning method is owned by [../disciplines/difficulty-and-balance.md](../disciplines/difficulty-and-balance.md).
- A mechanic change that touches a pillar, an economy invariant, or another team's system gets a decision record per [../decisions/design-decision-records.md](../decisions/design-decision-records.md).

## Framework Limits And Alternatives

MDA is this handbook's shared vocabulary, not a settled theory of everything. Apply it with its documented limits in view:

- **Known critiques.** MDA has been criticized for neglecting narrative and audiovisual design and for fitting mechanics-first work better than experience-first work. DDE (Walk, Görlich & Barrett, 2017 — `link.springer.com/chapter/10.1007/978-3-319-53088-8_3`) positions itself as "an advancement of the MDA framework," yet the same paper concedes MDA remains "the most widely accepted and practically employed approach to game design." No successor has earned replacement status; treat DDE and similar frameworks as critiques to learn from, not new canon.
- **Coarse aesthetics granularity.** The eight-aesthetic taxonomy names experience categories; it does not explain moment-to-moment control feel or player motivation. Supplement with the game-feel model in [game-feel.md](game-feel.md) and the motivation models in [player-psychology.md](player-psychology.md) rather than stretching MDA terms to cover them.
- **Narrative work needs more than "narrative: game as drama".** Structure for story-mechanics integration is owned by [../disciplines/narrative-integration.md](../disciplines/narrative-integration.md).
- **Switching or supplementing frameworks is a routed decision.** Adopting DDE, lenses-as-primary, or any other analysis frame for a project requires a row change in [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md), not an ad hoc vocabulary fork in one spec.

## Common Mistakes And Forbidden Patterns

- Specifying at the wrong layer: "make it feel epic" or "increase tension" with no named mechanic edit and no expected dynamic.
- Feature-driven specs that never state a target aesthetic, or state "fun" instead of terms from the taxonomy.
- Treating aesthetics as art direction or graphics polish rather than the emotional response the systems produce.
- Using "mechanic," "system," "dynamic," and "feature" interchangeably, or redefining them per document instead of deferring to [../glossary.md](../glossary.md).
- Shipping a mechanic with no declared interaction surface, then discovering the emergent dynamic in a live build.
- Tuning several mechanics in one change so a dynamic shift cannot be attributed to any of them.
- Ignoring an identified runaway feedback loop because the mechanic "tests fine" in short sessions.
- Presenting DDE or any other successor framework as settled canon, or forking the analysis vocabulary without touching [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md).

## Verification And Proof

- every feature spec names its target aesthetics (taxonomy terms, priority-ordered), the expected dynamics, and the exact mechanics changed — the one-page format in [../templates/one-page-gdd.md](../templates/one-page-gdd.md) has slots for all three
- a pre-build model exists for any system with resource flows or probability: spreadsheet or diagram per [../templates/balance-spreadsheet-spec.md](../templates/balance-spreadsheet-spec.md)
- playtest findings are recorded as aesthetic observation -> inferred dynamic -> candidate mechanic fix, per [../quality/playtesting.md](../quality/playtesting.md) and [../templates/playtest-report.md](../templates/playtest-report.md)
- design reviews walk the [../checklists/design-review.md](../checklists/design-review.md) items covering layer discipline and interaction surfaces
- any framework deviation appears in [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md) before the deviating spec merges
