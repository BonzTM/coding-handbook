# Narrative Integration

Defaults for integrating story with systems: budgeting arc content against loops, keeping mechanics and narrative saying the same thing, and telling story through space before words.

## Default Approach

Narrative is a discipline in service of the pillars in [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md), not a parallel product. The core loop defined per [../foundations/core-loops.md](../foundations/core-loops.md) is the fixed point; story attaches to it. When story and loop conflict, the loop wins by default, and overriding that requires a Design Decision Record per [../decisions/design-decision-records.md](../decisions/design-decision-records.md).

### Loops Versus Arcs Budgeting

- Classify every narrative deliverable as **loop content** or **arc content** before it enters the schedule. The distinction is Daniel Cook's: loops are repeatable, mastery-driven systems; arcs are consumed-once content like story (`lostgarden.com/2012/04/30/loops-and-arcs/`). The classification is a cost statement, not a style label — arc content is spent once per player, loop content amortizes across every session.
- Cap arc spend explicitly. Arc-heavy asks (bespoke scenes, one-shot setpieces, branching that multiplies content) route through [../operations/scoping-and-production.md](../operations/scoping-and-production.md) as scope, not through narrative review as tone.
- Prefer narrative devices that attach to loops: reactive barks, systemic consequences, environmental dressing that recontextualizes on replay. A story beat the player revisits is loop content; a story beat the player consumes once and skips forever is an arc and is budgeted like one.
- Branching narrative multiplies arc cost by branch count while each player consumes one path. Default to branches that reconverge or that recolor loop content; full exclusive branches require the same DDR as any other scope multiplier.

### Mechanics-Narrative Harmony

- The failure mode is **ludonarrative dissonance** — the gameplay and the non-interactive narrative asserting opposing themes. Clint Hocking coined the term for BioShock in his October 2007 post "Ludonarrative Dissonance in Bioshock" (originally on `clicknothing.typepad.com`, now offline; summarized at `en.wikipedia.org/wiki/Ludonarrative_dissonance`): the game "promotes the theme of self-interest through its gameplay while promoting the opposing theme of selflessness through its narrative."
- Audit incentives, not scripts. For each mechanic, state what the reward structure teaches the player to value, and compare it against what the story claims to value. Players experience the game aesthetics-first, mechanics-last, per the MDA reading order owned by [../decisions/frameworks-and-models.md](../decisions/frameworks-and-models.md) — so a dissonant incentive is experienced as the game lying, regardless of how the script reads.
- When the audit finds a contradiction, change the reward structure or change the story; never paper over it with dialogue. The economy levers live in [economy-and-progression.md](economy-and-progression.md).
- Harmony is a standing item in [../checklists/design-review.md](../checklists/design-review.md); a mechanic change that inverts an incentive drags the narrative surface with it.

### Environmental Storytelling

- Default to the space before the script. Don Carson's theme-park principle: "the story element is infused into the physical space," with staged cause-and-effect areas "that lead the game player to come to their own conclusions about a previous event" (`gamedeveloper.com/design/environmental-storytelling-creating-immersive-3d-worlds-using-lessons-learned-from-the-theme-park-industry`, 2000).
- Build spaces the player interprets, not spaces that recite. Worch and Smith frame environmental storytelling as the player pulling information "in opposition to traditional fictional exposition" — spaces with "an inherent sense of history" that "invite the player's mind to piece together implied events," using props, scripted events, texturing, lighting, and scene composition, extended by game systems reacting to the player (`gdcvault.com/play/1012647/What-Happened-Here-Environmental`, GDC 2010).
- This is a shared surface with [level-design.md](level-design.md): a level that needs words to tell its story fails Dan Taylor's level-design principle list (`gamedeveloper.com/design/ten-principles-of-good-level-design-part-1-`). The narrative discipline supplies the implied event; the level discipline stages it.
- Delivery ladder, cheapest-and-strongest first: readable space, then systemic reaction, then incidental dialogue, then cutscene. Each step down the ladder requires the step above to be insufficient, not merely unwritten.
- Environmental scenes are testable content. A staged scene players cannot decode is a bug, verified in playtests per [../quality/playtesting.md](../quality/playtesting.md), not a subtlety to defend.

### Dialogue And Cutscene Restraint

- Use as few words as possible; teach and tell with visuals before text. This is George Fan's onboarding rule (`gdcvault.com/play/1015541/How-I-Got-My-Mom`, GDC 2012) and it governs narrative delivery the same way it governs tutorials in [ux-and-onboarding.md](ux-and-onboarding.md).
- A cutscene is arc content at its most expensive and least interactive. Each one carries a recorded justification: the specific beat that space, systems, and incidental dialogue cannot deliver. No justification, no scene.
- Every cutscene is skippable and every line is subtitled; subtitle presentation rules are owned by [accessibility.md](accessibility.md).
- Place beats deliberately: open with a hook, alternate peaks with rest, escalate to a climax, per Schell's interest-curve model (owned by [economy-and-progression.md](economy-and-progression.md) (### Progression Pacing And Interest Curves) alongside the rest of the pacing canon). Story beats compete with loop pacing for the same attention; schedule them against the loop's rhythm, not against a screenplay's.
- Narrative documentation follows the house rule in [../foundations/design-documentation.md](../foundations/design-documentation.md): beat sheets and one-pagers, not a story bible nobody reads.

## Common Mistakes And Forbidden Patterns

- Writing a story bible before a core loop exists — arcs specced against a loop that will change are content written twice.
- Shipping a known incentive-versus-theme contradiction and calling it tone; dissonance is a design defect with a named owner, not a mood.
- Exposition dumps: lore delivered as unprompted dialogue, scrolling text, or codex entries standing in for spaces that should carry the story themselves.
- An unskippable cutscene, or a skippable one that gates progress-critical information with no in-world backup.
- Budgeting arc content as if it amortizes — bespoke scenes costed like reusable systems.
- Branch explosion: exclusive story branches added without a DDR, multiplying arc cost for content most players never see.
- Narrative vetoing a loop or balance change late because "the story requires it" — pillar arbitration in [../foundations/design-pillars-and-vision.md](../foundations/design-pillars-and-vision.md) settles that conflict, not seniority.
- Staged environmental scenes that playtesters consistently misread or walk past, left in on the theory that attentive players will get it.

## Verification And Proof

- classification pass: every narrative deliverable in the schedule labeled loop or arc, with arc totals visible in the scope sheet per [../operations/scoping-and-production.md](../operations/scoping-and-production.md)
- dissonance audit: a table of mechanic → what its rewards teach → what the story claims, with zero unresolved contradictions at design review
- environmental readback test: fresh playtesters narrate "what happened here" for each staged scene without prompting; a scene most testers cannot reconstruct goes back to layout ([../quality/playtesting.md](../quality/playtesting.md))
- cutscene ledger: every scene lists its justification, its skip path, and its subtitle coverage
- word-budget check: dialogue and text counts tracked per beat, trending down across iterations, not up
- [../checklists/design-review.md](../checklists/design-review.md) harmony items answered with evidence, not assertion

## Related

- [../foundations/core-loops.md](../foundations/core-loops.md) — the loop the narrative attaches to
- [level-design.md](level-design.md) — staging the spaces that carry the story
- [economy-and-progression.md](economy-and-progression.md) — interest curves and the pacing canon story beats schedule against
- [../operations/scoping-and-production.md](../operations/scoping-and-production.md) — where arc budgets are enforced
